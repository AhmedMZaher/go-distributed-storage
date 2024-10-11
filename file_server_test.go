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

func TestComplexDFSScenario(t *testing.T) {
	// Create a network of 5 nodes
	initialPeer := "127.0.0.5:3000"
	nodes := make([]*FileServer, 5)
	addresses := []string{
		initialPeer,
		"127.0.0.5:9000",
		"127.0.0.5:5000",
		"127.0.0.5:6000",
		"127.0.0.5:7000",
	}

	// Start all nodes
	for i, addr := range addresses {
		if i == 0 {
			nodes[i] = makeServer(addr, true)
		} else {
			nodes[i] = makeServer(addr, false, initialPeer)
		}
		go func(node *FileServer) {
			if err := node.Start(); err != nil {
				log.Fatalf("Failed to start server on %s: %v", node.Config.Transport.RemoteAddr(), err)
			}
		}(nodes[i])
		time.Sleep(5 * time.Millisecond)
	}

	// Allow time for share addresses
	time.Sleep(1 * time.Second)

	// Define test files
	files := map[string]string{
		"file1.txt": "This is the content of file 1",
		"file2.txt": "File 2 has different content",
		"file1119.txt": "some jpg bytes",
		// "file3.txt": "Yet another file with unique content",
	}

	// Store files on different nodes
	for fileName, content := range files {
		nodeIndex := (len(files) - 1) % len(nodes)
		err := nodes[nodeIndex].Store(fileName, bytes.NewReader([]byte(content)))
		if err != nil {
			t.Errorf("Failed to store %s: %v", fileName, err)
		}
		fmt.Printf("Stored %s on Node %d\n", fileName, nodeIndex)
	}

	// Allow time for replication
	time.Sleep(1 * time.Second)

	// Verify files are accessible from all nodes
	for _, node := range nodes {
		for fileName, expectedContent := range files {
			r, err := node.Get(fileName)
			if err != nil {
				t.Errorf("Node %s failed to get %s: %v", node.Config.Transport.RemoteAddr(), fileName, err)
				continue
			}
			content := make([]byte, len(expectedContent))
			n, err := r.Read(content)
			if err != nil && err != io.EOF {
				t.Errorf("Failed to read content of %s: %v", fileName, err)
				continue
			}
			if n != len(expectedContent) {
				t.Errorf("Unexpected content length for %s. Got: %d, Want: %d", fileName, n, len(expectedContent))
				continue
			}
			if string(content) != expectedContent {
				t.Errorf("Unexpected content for %s. Got: %s, Want: %s", fileName, string(content), expectedContent)
			}
		}
	}

	// // Simulate node failure: stop node 2
	nodes[2].Stop()
	// Delete node 2 files
	nodes[2].Storage.Clear()
	fmt.Println("Node 2 has been stopped")

	// Delete a file from its original node
	deletedFile := "file2.txt"
	if err := nodes[1].Storage.DeleteFile(deletedFile); err != nil {
		t.Errorf("Failed to delete %s: %v", deletedFile, err)
	}
	fmt.Printf("%s has been deleted from its original node\n", deletedFile)

	// Verify the file is still accessible from other nodes
	for i, node := range nodes {
		if i == 1 || i == 2 { // Skip the node where we deleted the file and the stopped node
			continue
		}
		r, err := node.Get(deletedFile)
		if err != nil {
			t.Errorf("Node %s failed to get %s after deletion: %v", node.Config.Transport.RemoteAddr(), deletedFile, err)
			continue
		}
		expectedContent := files[deletedFile]
		content := make([]byte, len(expectedContent))
		n, err := r.Read(content)
		if err != nil && err != io.EOF {
			t.Errorf("Failed to read content of %s after deletion: %v", deletedFile, err)
			continue
		}
		if n != len(expectedContent) {
			t.Errorf("Unexpected content length for %s after deletion. Got: %d, Want: %d", deletedFile, n, len(expectedContent))
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("Unexpected content for %s after deletion. Got: %s, Want: %s", deletedFile, string(content), expectedContent)
		}
	}

	// // Restart the failed node
	nodes[2] = makeServer(addresses[2], false, initialPeer)
	go func() {
		if err := nodes[2].Start(); err != nil {
			log.Fatalf("Failed to restart server on %s: %v", addresses[2], err)
		}
	}()
	time.Sleep(1 * time.Second)
	fmt.Println("Node 2 has been restarted")

	// // Verify the restarted node can access all files
	for fileName, expectedContent := range files {
		r, err := nodes[2].Get(fileName)
		if err != nil {
			t.Errorf("Restarted node failed to get %s: %v", fileName, err)
			continue
		}
		content := make([]byte, len(expectedContent))
		n, err := r.Read(content)
		if err != nil && err != io.EOF {
			t.Errorf("Failed to read content of %s from restarted node: %v", fileName, err)
			continue
		}
		if n != len(expectedContent) {
			t.Errorf("Unexpected content length for %s from restarted node. Got: %d, Want: %d", fileName, n, len(expectedContent))
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("Unexpected content for %s from restarted node. Got: %s, Want: %s", fileName, string(content), expectedContent)
		}
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Complex DFS test completed successfully")
}
