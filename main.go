package main

import (
	"go-distributed-storage/p2p"
	"log"

)


func main() {
	tcptransportOPT := p2p.TCPTransportOPT{
		ListenAddress : ":3030",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcptransportOPT);
	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}
	select{}
}