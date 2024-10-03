package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"go-distributed-storage/p2p"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type FileServerOPT struct {
	RootDir          string
	Transport        p2p.Transport
	PathTranformFunc PathTranformSignature
	BootstrapNodes   []string
}

type FileServer struct {
	Config FileServerOPT

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	Storage Storage
	quitCh  chan struct{}
}

func NewFileServer(opt FileServerOPT) *FileServer {
	storageOPT := StoreOPT{
		RootDir:          opt.RootDir,
		PathTranformFunc: opt.PathTranformFunc,
	}
	return &FileServer{
		Config:  opt,
		Storage: *NewStorage(storageOPT),
		quitCh:  make(chan struct{}),
		peers:   make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type StoreFileMessage struct {
	Key  string
	Size int64
}

type GetFileMessage struct {
	Key string
}

func (s *FileServer) broadcast(message *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(message); err != nil {
		return err
	}
	
	for _, peer := range s.peers {
		if err := peer.Send([]byte{p2p.IncomingMessage}); err != nil {
			return err
		}

		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitCh)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("%s accepted peer connection from: %s\n", p.LocalAddr(), p.RemoteAddr())

	return nil
}
func (s *FileServer) loop() {
	defer func() {
		log.Println("FileServer has shut down and transport connection has been closed.")
		s.Config.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Config.Transport.Consume():
			var message Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&message); err != nil {
				fmt.Println("Decoding error: ", err)
			}

			if err := s.handleMessage(rpc.From, message); err != nil {
				fmt.Println("Handling message error: ", err)
			}

		case <-s.quitCh:
			return
		}
	}
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.Storage.HasKey(key) {
		fmt.Printf("[%s] file with key (%s) found locally\n", s.Config.Transport.RemoteAddr(), key)
		r, _, err := s.Storage.ReadFile(key)
		return r, err
	}

	fmt.Printf("[%s] file with key (%s) not found locally, broadcasting request to peers\n", s.Config.Transport.RemoteAddr(), key)

	message := Message{
		Payload: GetFileMessage{
			Key: key,
		},
	}

	if err := s.broadcast(&message); err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Millisecond)

	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		if _, err := s.Storage.StoreFile(key, io.LimitReader(peer, fileSize)); err != nil {
			return nil, err
		}
		fmt.Printf("[%s] successfully received and stored file with key (%s) of size %d bytes from peer %s\n", s.Config.Transport.RemoteAddr(), key, fileSize, peer.RemoteAddr().String())
		peer.CloseStream()
	}

	r, _, err := s.Storage.ReadFile(key)
	return r, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
 	size, err := s.Storage.StoreFile(key, tee)
	if err != nil {
		return err
	}

	message := Message{
		Payload: StoreFileMessage{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&message); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	mw.Write(fileBuffer.Bytes())
	// for _, peer := range s.peers {
	// 	peer.Send([]byte{p2p.IncomingStream})
	// 	_, err := io.Copy(peer, fileBuffer)
	// 	if err != nil {
	// 		return err
	// 	}

	// }

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Config.Transport.RemoteAddr(), size)
	return nil
}

// handleMessage processes incoming messages and delegates them to the appropriate handler
// based on the type of the message payload.
func (s *FileServer) handleMessage(from net.Addr, message Message) error {
	switch payloadType := message.Payload.(type) {
	case GetFileMessage:
		return s.handleGetFileMessage(from, payloadType)
	case StoreFileMessage:
		return s.handleStoreFileMessage(from, payloadType)
	}

	return nil
}

func (s *FileServer) handleGetFileMessage(from net.Addr, message GetFileMessage) error {
	if !s.Storage.HasKey(message.Key) {
		return fmt.Errorf("file with key %s not found on disk", message.Key)
	}

	fmt.Printf("[%s] serving file with key (%s) over the network\n", s.Config.Transport.RemoteAddr(), message.Key)

	r, fileSize, err := s.Storage.ReadFile(message.Key)
	if err != nil {
		return err
	}

	peer, isExist := s.peers[from.String()]
	if !isExist {
		return fmt.Errorf("peer %s not found in peer map", from.String())
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return fmt.Errorf("error copying file data to peer: %v", err)
	}

	fmt.Printf("[%s] successfully written %d bytes to peer %s\n", s.Config.Transport.RemoteAddr(), n, from.String())

	return nil
}

func (s *FileServer) handleStoreFileMessage(from net.Addr, message StoreFileMessage) error {
	peer, isExist := s.peers[from.String()]
	if !isExist {
		return fmt.Errorf("peer %s not found in peer map", from.String())
	}

	_, err := s.Storage.StoreFile(message.Key, io.LimitReader(peer, message.Size))
	if err != nil {
		return fmt.Errorf("error storing file with key %s from peer %s: %v", message.Key, from.String(), err)
	}

	fmt.Printf("[%s] successfully stored file with key (%s) of size %d bytes from peer %s\n", s.Config.Transport.RemoteAddr(), message.Key, message.Size, from.String())

	peer.CloseStream()

	return nil
}

// connectToBootstrapNodes attempts to connect to all bootstrap nodes specified in the server's configuration.
// It iterates over the list of bootstrap node addresses and spawns a goroutine for each non-empty address to
// establish a connection using the server's transport mechanism. If a connection attempt fails, an error is logged.
//
// Returns an error if any issues occur during the connection process.
func (s *FileServer) connectToBootstrapNodes() error {
	for _, address := range s.Config.BootstrapNodes {
		if len(address) == 0 {
			continue
		}

		go func() {
			fmt.Printf("%s Attempting to connect to bootstrap node: %s\n", s.Config.Transport.RemoteAddr(), address)
			if err := s.Config.Transport.Dial(address); err != nil {
				log.Printf("Failed to connect to bootstrap node %s: %v\n", address, err)
			}
		}()
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Config.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.connectToBootstrapNodes()

	s.loop()

	return nil
}

func init() {
	gob.Register(StoreFileMessage{})
	gob.Register(GetFileMessage{})
}