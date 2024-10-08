package main

import (
	"bytes"
	"fmt"
	"go-distributed-storage/p2p"
	"io"
	"log"
	"math/rand"
	"testing"
	"time"
)

// makeServer initializes a FileServer with the specified options.
func makeServer(listenAddress string, isBootstrapNode bool, bootstrapNode ...string) *FileServer {
	tcptransportOpts := &p2p.TCPTransportOPT{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := FileServerOPT{
		encryptionKey:    []byte{0x0e, 0x02, 0x5d, 0x3d, 0xb7, 0xb1, 0xf1, 0xfa, 0xdb, 0xcd, 0x1b, 0x8e, 0xc9, 0xa4, 0x5f, 0x99, 0xa1, 0x0a, 0x3f, 0x1f, 0x27, 0x31, 0xab, 0xfa, 0x68, 0x9f, 0x91, 0x42, 0x75, 0x46, 0x28, 0xec},
		Crypto:           &BasicCrypto{},
		RootDir:          listenAddress + "_network",
		PathTranformFunc: HashPathBuilder,
		Transport:        tcpTransport,
		BootstrapNodes:   bootstrapNode,
		IsBootstrapNode:  isBootstrapNode,
	}

	server := NewFileServer(fileServerOpts)

	tcptransportOpts.OnPeer = server.OnPeer

	return server
}

// generateRandomData creates a random byte slice of specified size.
func generateRandomData(size int) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}

// TestFileServer tests the Store and Get functionalities of the FileServer.
func TestFileServer(t *testing.T) {
	// Start the initial peer on port 3000
	initialPeer := "127.0.0.5:3000"
	server := makeServer(initialPeer, true)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server on %s: %v", initialPeer, err)
		}
	}()
	time.Sleep(10 * time.Millisecond) // Give the initial server time to start

	// Create three more peers, each using the initial peer as the bootstrap node
	addresses := []string{"127.0.0.5:5000", "127.0.0.5:7000", "127.0.0.5:9000"}
	var servers []*FileServer
	servers = append(servers, server) // Include the initial server

	for _, addr := range addresses {
		s := makeServer(addr, false, initialPeer)
		servers = append(servers, s)

		go func(s *FileServer) {
			if err := s.Start(); err != nil {
				log.Fatalf("Failed to start server on %s: %v", addr, err)
			}
		}(s)

		time.Sleep(10 * time.Millisecond) // Give each server time to start
	}

	// Prepare multiple files for testing
	numFiles := 5
	files := make(map[string][]byte)
	for i := 0; i < numFiles; i++ {
		key := fmt.Sprintf("file_%d", i+1)
		content := generateRandomData(1024) // 1KB of random data per file
		files[key] = content
	}

	// Test Store functionality for multiple files
	t.Run("StoreMultipleFiles", func(t *testing.T) {
		for key, content := range files {
			data := bytes.NewReader(content)
			if err := server.Store(key, data); err != nil {
				t.Errorf("store error for key %s: %v", key, err)
			}
		}
	})

	// Test Get functionality for multiple files
	t.Run("GetMultipleFiles", func(t *testing.T) {
		for key, expectedContent := range files {
			r, err := server.Get(key)
			if err != nil {
				t.Errorf("get error for key %s: %v", key, err)
				continue
			}

			retrievedContent, err := io.ReadAll(r)
			if err != nil {
				t.Errorf("read error for key %s: %v", key, err)
				continue
			}
			if !bytes.Equal(retrievedContent, expectedContent) {
				t.Errorf("data mismatch for key %s: got %v, want %v", key, retrievedContent[:10], expectedContent[:10])
			}
		}
	})
}

func TestMultiHopFileRequest(t *testing.T) {
	// This test simulates a multi-hop file request in a peer-to-peer network.
	// NodeA acts as the bootstrap node, while NodeB and NodeC represent other nodes in the network.
	//
	// The network configuration is as follows:
	// - NodeA (Bootstrap Node): 127.0.0.5:3000
	// - NodeB (Requester Node): 127.0.0.5:5000
	// - NodeC (File Holder Node): 127.0.0.5:7000
	//
	// The test scenario involves the following steps:
	// 1. NodeC stores a file ("mybigfile") containing some data.
	// 2. NodeB starts after the file is stored on NodeC, and it only knows NodeA.
	// 3. NodeA provides NodeB with NodeC's address.
	// 4. NodeB sends a request to NodeC (via the address provided by NodeA) to retrieve the file.
	// 5. The test verifies that NodeB can successfully retrieve the file from NodeC.

	initialPeer := "127.0.0.5:3000"

	// Create and start NodeA (the bootstrap node)
	nodeA := makeServer(initialPeer, true)
	go func() {
		if err := nodeA.Start(); err != nil {
			log.Fatalf("Failed to start server on %s: %v", initialPeer, err)
		}
	}()
	time.Sleep(5 * time.Millisecond)

	// Create and start NodeC (the file holder node)
	nodeC := makeServer("127.0.0.5:7000", false, initialPeer)
	go func() {
		if err := nodeC.Start(); err != nil {
			log.Fatalf("Failed to start server on 127.0.0.5:7000: %v", err)
		}
	}()
	time.Sleep(5 * time.Millisecond)

	// Create NodeB (the requester node)
	nodeB := makeServer("127.0.0.5:5000", false, initialPeer)

	// Store a file in NodeC
	fileName := "mybigfile"
	data := bytes.NewReader([]byte("some bytes"))
	if err := nodeC.Store(fileName, data); err != nil {
		t.Error(err)
	}

	// Start NodeB after the file has been stored
	go func() {
		if err := nodeB.Start(); err != nil {
			log.Fatalf("Failed to start server on 127.0.0.5:5000: %v", err)
		}
	}()
	time.Sleep(5 * time.Millisecond)

	// Simulate NodeA deleting the file from its storage
	if err := nodeA.Storage.DeleteFile(fileName); err != nil {
		t.Error(err)
	}

	// NodeB requests the file from NodeC via the address provided by NodeA
	r, err := nodeB.Get(fileName)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(r)
}
