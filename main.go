package main

import (
	"bytes"
	"fmt"
	"go-distributed-storage/p2p"
	"log"
	"time"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {
	tcptransportOpts := &p2p.TCPTransportOPT{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := FileServerOPT{
		RootDir:          listenAddress + "_network",
		PathTranformFunc: HashPathBuilder,
		Transport:        tcpTransport,
		BootstrapNodes:   nodes,
	}

	server := NewFileServer(fileServerOpts)
	
	tcptransportOpts.OnPeer = server.OnPeer

	return server
}
func main() {
	s1 := makeServer("127.0.0.5:3000", "")
	s2 := makeServer("127.0.0.5:5000", "127.0.0.5:3000")
	s3 := makeServer("127.0.0.5:7000", "127.0.0.5:3000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(5 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(5 * time.Millisecond)
	go func() { log.Fatal(s3.Start()) }()
	time.Sleep(5 * time.Millisecond)

	data := bytes.NewReader([]byte("Hi my name is ahmed"))
	key := "myfile"
	
	if err := s1.Store(key, data); err != nil {
		fmt.Print(err)
	}
	// if err := s2.Store(key, data); err != nil {
	// 	fmt.Print(err)
	// }
}
