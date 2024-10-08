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
	Crypto           Cipher
	encryptionKey    []byte
	RootDir          string
	Transport        p2p.Transport
	PathTranformFunc PathTranformSignature
	BootstrapNodes   []string
	IsBootstrapNode  bool
}

type FileServer struct {
	Config FileServerOPT

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	Storage Storage
	quitCh  chan struct{}

	// PeersAddresses holds the addresses of peer nodes in the distributed network.
	PeersAddresses []string
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

type PeersInfoMessage struct {
	Addresses []string
}

type NodeIntroductionMessage struct {
	Address string
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

	if s.Config.IsBootstrapNode {
		msg := Message{
			Payload: PeersInfoMessage{
				Addresses: s.PeersAddresses,
			},
		}
		
		if err := p.Send([]byte{p2p.IncomingMessage}); err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(msg); err != nil {
			return err
		}
		if err := p.Send(buf.Bytes()); err != nil {
			log.Printf("Failed to send addresses to peer %s: %v\n", p.RemoteAddr(), err)
		}
	}

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

// Get retrieves a file by its key from local storage or peers.
//
// Checks if the file exists locally and returns it if found.
// If not found, broadcasts a request to peers for the file.
// Waits briefly before attempting to read the file from peers.
// Reads the file size from each peer and stores the file locally.
// Logs the successful retrieval and storage of the file from a peer.
// Sets a timeout for reading the file size. If the read operation times out, it proceeds to the next peer.
func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.Storage.HasKey(key) {
		fmt.Printf("[%s] file with key (%s) found locally\n", s.Config.Transport.RemoteAddr(), key)
		// r, _, err := s.Storage.ReadFile(key)
		r, _, err := s.Storage.ReadFileDecrypted(key, s.Config.Crypto.Decrypt, s.Config.encryptionKey)
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

		// Set a timeout duration for reading the file size
		readTimeout := 5 * time.Second
		done := make(chan error) // Channel to signal completion of read operation

		go func(peer p2p.Peer) {
			err := binary.Read(peer, binary.LittleEndian, &fileSize)
			done <- err // Send the result (error or nil) back to the main routine
		}(peer)

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("Error reading file size from peer %s: %v\n", peer.RemoteAddr().String(), err)
				continue
			}
		case <-time.After(readTimeout):
			fmt.Printf("Timeout while waiting for file size from peer %s\n", peer.RemoteAddr().String())
			continue
		}

		if _, err := s.Storage.StoreFile(key, io.LimitReader(peer, fileSize)); err != nil {
			return nil, err
		}

		fmt.Printf("[%s] successfully received and stored file with key (%s) of size %d bytes from peer %s\n", s.Config.Transport.RemoteAddr(), key, fileSize, peer.RemoteAddr().String())

		peer.CloseStream()
	}

	// r, _, err := s.Storage.ReadFile(key)
	r, _, err := s.Storage.ReadFileDecrypted(key, s.Config.Crypto.Decrypt, s.Config.encryptionKey)
	return r, err
}

// Store saves a file to storage and broadcasts the event.
//
// Reads from the provided io.Reader while storing data in a buffer.
// Stores the file with the specified key and retrieves its size.
// Constructs a message with the key and size for broadcasting.
// Sends the storage event to all peers and initiates data transfer.
// Logs the total bytes received and written to disk.
func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
	// size, err := s.Storage.StoreFile(key, tee)
	size, err := s.Storage.StoreFileEncrypted(key, tee, s.Config.Crypto.Encrypt, s.Config.encryptionKey)
	if err != nil {
		return err
	}

	message := Message{
		Payload: StoreFileMessage{
			Key: key,
			// Total size includes 16 bytes for the Initialization Vector (IV) at the beginning of the file.
			Size: 16 + size,
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
	// mw.Write(fileBuffer.Bytes())
	s.Config.Crypto.Encrypt(s.Config.encryptionKey, mw, fileBuffer)

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
	case PeersInfoMessage:
		return s.handlePeersInfoMessage(from, payloadType)
	case NodeIntroductionMessage:
		return s.handleNodeIntroductionMessage(payloadType)
	}

	return nil
}

// handleGetFileMessage processes a file retrieval request from a peer.
//
// Validates the existence of the requested file in storage.
// Logs the operation of serving the file over the network.
// Retrieves the file reader and its size from storage.
// Confirms the requesting peer is present in the peer map.
// Sends a stream initiation signal to the peer.
// Writes the file size to the peer using binary format.
// Copies the file data to the peer's stream and logs the transfer.
func (s *FileServer) handleGetFileMessage(from net.Addr, message GetFileMessage) error {
	if !s.Storage.HasKey(message.Key) {
		return fmt.Errorf("file with key %s not found on %s disk", message.Key, s.Config.Transport.RemoteAddr())
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

// Handles the reception of a file storage message from a peer.
//
// Verifies the existence of the sending peer in the peer map.
// Stores the file associated with the provided key from the peer's stream.
// Logs the successful storage of the file, including the key,
// size, and peer address.
// Closes the stream for the sending peer to manage resources.
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

// handlePeersInfoMessage handles incoming peer information messages, which contain a list of peer addresses.
// When a new list of peer addresses is received, the function attempts to establish connections with each address.
func (s *FileServer) handlePeersInfoMessage(from net.Addr, message PeersInfoMessage) error {
	fmt.Printf("%s received peer address list %v from %s\n", s.Config.Transport.RemoteAddr(), message.Addresses, from)
	for _, address := range message.Addresses {
		if len(address) == 0 {
			continue
		}

		go func() {
			fmt.Printf("%s Attempting to broadcasted node: %s\n", s.Config.Transport.RemoteAddr(), address)
			if err := s.Config.Transport.Dial(address); err != nil {
				log.Printf("Failed to connect to broadcasted node %s: %v\n", address, err)
			}
		}()
	}

	return nil
}

// handleNodeIntroductionMessage processes a NodeIntroductionMessage by adding the
// address from the message to the server's list of peer addresses.
func (s *FileServer) handleNodeIntroductionMessage(message NodeIntroductionMessage) error {
	fmt.Printf("Node %s received introduction message from Node %s\n", s.Config.Transport.RemoteAddr(), message.Address)
	s.PeersAddresses = append(s.PeersAddresses, message.Address)
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

	time.Sleep(5 * time.Millisecond)
	// After joining the network, send your address to all other nodes to share your real address
	message := Message{
		Payload: NodeIntroductionMessage{
			Address: s.Config.Transport.RemoteAddr(),
		},
	}
	if err := s.broadcast(&message); err != nil {
		return err
	}

	s.loop()

	return nil
}

func init() {
	gob.Register(StoreFileMessage{})
	gob.Register(GetFileMessage{})
	gob.Register(PeersInfoMessage{})
	gob.Register(NodeIntroductionMessage{})
}
