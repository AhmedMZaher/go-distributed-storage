package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPTransport represents a transport mechanism using TCP for peer-to-peer communication.
// It includes the listening address, the network listener, and a map of peers with their respective addresses.
// The structure is protected by a read-write mutex to ensure thread-safe access.
type TCPTransport struct {
	listenAddress string
	listener	net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func ListenAndAccept(t *TCPTransport) error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport)startAcceptLoop() {

	for {
		conn, err := t.listener.Accept();
		if err != nil{
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		go t.handleConn(conn);
	}
}

func (t *TCPTransport)handleConn(conn net.Conn){
	fmt.Printf("new incoming connection %+v\n", conn);
}