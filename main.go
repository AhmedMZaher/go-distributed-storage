package main

import (
	"fmt"
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