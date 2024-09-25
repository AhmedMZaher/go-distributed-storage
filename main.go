package main

import (
	"fmt"
	"go-distributed-storage/p2p"
	"log"
)

func OnPeer(peer p2p.Peer) error{
	return peer.Close()
}

func main() {
	tcptransportOPT := p2p.TCPTransportOPT{
		ListenAddress : ":3030",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		OnPeer: OnPeer,
		
	}
	tr := p2p.NewTCPTransport(tcptransportOPT);

	go func(){
		for{
			msg := <- tr.Consume()
			fmt.Printf("New message arrived %+v\n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}
	select{}
}