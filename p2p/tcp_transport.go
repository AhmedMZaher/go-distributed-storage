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

// TCPTransportOPT holds the configuration options for the TCP transport layer.
// It includes the address to listen on and a function for handling handshakes.
//
// Fields:
// - ListenAddress: The address on which the TCP transport will listen for incoming connections.
// - HandshakeFunc: A function that defines the handshake process for establishing connections.
type TCPTransportOPT struct{
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
}

// TCPTransport represents a transport mechanism using TCP for peer-to-peer communication.
// It holds the configuration options, a network listener, and a map of connected peers.
// The structure is thread-safe, utilizing a read-write mutex for concurrent access.
type TCPTransport struct {
	tcpTransportOPT TCPTransportOPT
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(tcpTransportOPT TCPTransportOPT) *TCPTransport {
	return &TCPTransport{
		tcpTransportOPT: tcpTransportOPT,
	}
}

func (t *TCPTransport)ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.tcpTransportOPT.ListenAddress)

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
		fmt.Printf("new incoming connection %+v\n", conn)

		go t.handleConn(conn)
	}
}

// handleConn handles an incoming TCP connection. It performs a handshake
// with the peer and then continuously decodes incoming messages from the connection.
// If an error occurs during the handshake, the connection is closed and an error message is printed.
// If an error occurs while decoding a message, an error message is printed and the loop continues.
func (t *TCPTransport) handleConn(conn net.Conn){
	peer := NewTCPPeer(conn, true)

		
	if err := t.tcpTransportOPT.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error %s\n", err)
		return
	}

	message := &RPC{}
	for{
		if err := t.tcpTransportOPT.Decoder.Decode(conn, message); err != nil{
			fmt.Println("TCP receiving message error")
			continue
		}

		message.From = conn.RemoteAddr();
		fmt.Printf("new message received %+v\n", message)
	}	
}
