package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"go-distributed-storage/p2p"
	"log"
	"net"
	"sync"
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

type Message struct{
	payload		any
}

type StoreFileMessage struct{
	Key		string
	Size 	int64
}

type GetFileMessage struct{
	Key		string
}

func (s *FileServer) broadcast(message Message) error {
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

			if err := gob.NewDecoder(bytes.NewBuffer(rpc.Payload)).Decode(&message); err != nil {
				fmt.Printf("Decoding error: ", err)
			}
			
			if err := s.handleMessage(rpc.From, message); err != nil {
				fmt.Printf("Handling message error: ", err)
			}
			
		case <-s.quitCh:
			return
		}
	}
}

// handleMessage processes incoming messages and delegates them to the appropriate handler
// based on the type of the message payload.
func (s *FileServer) handleMessage(from net.Addr, message Message) error {
	switch payloadType := message.payload.(type) {
	case GetFileMessage:
			return s.handleGetFileMessage(from, payloadType)
	case StoreFileMessage:
		return s.handleStoreFileMessage(from, payloadType)
	}

	return nil
}

func (s *FileServer) handleGetFileMessage (from net.Addr, message GetFileMessage) error {
	return nil	
}

func (s *FileServer) handleStoreFileMessage (from net.Addr, message StoreFileMessage) error {
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
			fmt.Printf("%s Attempting to connect to bootstrap node: %s\n", s.Config.Transport.RemoteAddr().Addr().String(), address)
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
