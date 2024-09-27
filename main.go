package main

import (
	"go-distributed-storage/p2p"
	"log"
	"time"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {
	tcptransportOpts := p2p.TCPTransportOPT{
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

	return server
}
func main() {
	s1 := makeServer("127.0.0.5:3000", "")
	s2 := makeServer("127.0.0.5:7000", "")
	s3 := makeServer("127.0.0.5:5000", "127.0.0.5:3000", "127.0.0.5:7000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)
}
