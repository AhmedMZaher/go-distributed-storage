package main

import (
	"fmt"
	"go-distributed-storage/p2p"
	"log"
	"sync"
)

type FileServerOPT struct {
	RootDir         string
	Transport       p2p.Transport
	PathTranformFunc PathTranformSignature
}

type FileServer struct {
	Config	FileServerOPT

	peerLock sync.Mutex
	peers map[string]p2p.Peer

	Storage Storage
	quitCh	chan struct{}
}

func NewFileServer(opt FileServerOPT) *FileServer {
	storageOPT := StoreOPT{
		RootDir: opt.RootDir,
		PathTranformFunc: opt.PathTranformFunc,
	}
	return &FileServer{
		Config: opt,
		Storage: *NewStorage(storageOPT),
		quitCh: make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Stop() {
	close(s.quitCh)
}

func (s *FileServer) loop() {
	defer func ()  {
		log.Println("FileServer has been closed")
		s.Config.Transport.Close()
	}()

	for {
		select{
		case rpc := <-s.Config.Transport.Consume() :
			fmt.Println(rpc)
		case <-s.quitCh:
			return
		}
	}
}

func (s *FileServer) Start() error {
	if err := s.Config.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.loop()

	return nil
}