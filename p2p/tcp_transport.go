package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents a peer in the network using TCP for communication.
// It holds the connection information and whether the connection is outbound.
type TCPPeer struct {

	// conn represents a network connection that implements the net.Conn interface.
	// It is used for reading and writing data over a TCP connection.
	conn net.Conn
	// outbound indicates whether the connection is outbound (true) or inbound (false).
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// TCPTransport represents a transport mechanism using TCP for peer-to-peer communication.
// It includes the listening address, the network listener, and a map of peers with their respective addresses.
// The structure is protected by a read-write mutex to ensure thread-safe access.
type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport)ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("new incoming connection %+v\n", peer)
}
