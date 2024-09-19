package main

import (
	"go-distrbuted-storage/p2p"
	"log"

)


func main() {
	tr := p2p.NewTCPTransport(":8080");
	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}
	// select{}
}