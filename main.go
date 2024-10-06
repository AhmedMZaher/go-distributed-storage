package main

// import (
// 	"bytes"
// 	"fmt"
// 	"go-distributed-storage/p2p"
// 	"io"
// 	"log"
// 	"time"
// )

// func makeServer(listenAddress string, nodes ...string) *FileServer {
// 	tcptransportOpts := &p2p.TCPTransportOPT{
// 		ListenAddress: listenAddress,
// 		HandshakeFunc: p2p.NOPHandshakeFunc,
// 		Decoder:       p2p.DefaultDecoder{},
// 	}
// 	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

// 	fileServerOpts := FileServerOPT{
// 		encryptionKey:    []byte{0x0e, 0x02, 0x5d, 0x3d, 0xb7, 0xb1, 0xf1, 0xfa, 0xdb, 0xcd, 0x1b, 0x8e, 0xc9, 0xa4, 0x5f, 0x99, 0xa1, 0x0a, 0x3f, 0x1f, 0x27, 0x31, 0xab, 0xfa, 0x68, 0x9f, 0x91, 0x42, 0x75, 0x46, 0x28, 0xec},
// 		Crypto:           &BasicCrypto{},
// 		RootDir:          listenAddress + "_network",
// 		PathTranformFunc: HashPathBuilder,
// 		Transport:        tcpTransport,
// 		BootstrapNodes:   nodes,
// 	}

// 	server := NewFileServer(fileServerOpts)

// 	tcptransportOpts.OnPeer = server.OnPeer

// 	return server
// }
// func main() {
// 	s1 := makeServer("127.0.0.5:3000")
// 	s2 := makeServer("127.0.0.5:5000", "127.0.0.5:3000")
// 	s3 := makeServer("127.0.0.5:7000", "127.0.0.5:3000")

// 	go func() { log.Fatal(s1.Start()) }()
// 	time.Sleep(5 * time.Millisecond)
// 	go func() { log.Fatal(s2.Start()) }()
// 	time.Sleep(5 * time.Millisecond)
// 	go func() { log.Fatal(s3.Start()) }()
// 	time.Sleep(5 * time.Millisecond)

// 	data := bytes.NewReader([]byte("Hi my name is ahmed"))
// 	key := "myfile"

// 	// ////////////////////////////////////////////
// 	fmt.Println("---------------- TESTING STORE ---------------- ")
// 	if err := s1.Store(key, data); err != nil {
// 		fmt.Print(err)
// 	}
// 	time.Sleep(2 * time.Second)

// 	////////////////////////////////////////////
// 	fmt.Println(" ---------------- TESTING GET ---------------- ")
// 	r, err := s1.Get(key)
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	buf, err := io.ReadAll(r)
// 	if err != nil {
// 		fmt.Print(err)
// 	}
// 	fmt.Println(string(buf))
// }
