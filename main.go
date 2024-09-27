package main

import (
	"go-distributed-storage/p2p"
	"time"
)



func main() {
	tcptransportOPT := p2p.TCPTransportOPT{
		ListenAddress : ":3030",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		// TODO onPeer		
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOPT);

	fileServerOPT := FileServerOPT {
		RootDir: ":8080",
		PathTranformFunc: HashPathBuilder,
		Transport: tcpTransport,
	}
	server := NewFileServer(fileServerOPT)
	
	go func ()  {
		time.Sleep(time.Second * 2)
		server.Stop()
		
	}()
	
	server.Start()
}