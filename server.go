package main

import (
	"fmt"
	"go-distributed-storage/p2p"
	"log"
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

func (s *FileServer) Stop() {
	close(s.quitCh)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("%s accepted peer connection from: %s\n", s.Config.Transport.RemoteAddr().Addr().String(), p.RemoteAddr())

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
			fmt.Println(rpc)
		case <-s.quitCh:
			return
		}
	}
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
