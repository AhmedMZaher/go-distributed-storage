package p2p

import (
	"fmt"
	"net"
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



// TCPTransport represents a transport layer for peer-to-peer communication over TCP.
// It holds the configuration options, a network listener, a channel for RPC messages,
// and a callback function to make the code more generic
// that is triggered when a new peer is connected.
type TCPTransport struct {
	tcpTransportOPT TCPTransportOPT
	listener      net.Listener
	rpcCh	chan RPC
	OnPeer func(Peer) error
}

func NewTCPTransport(tcpTransportOPT TCPTransportOPT) *TCPTransport {
	return &TCPTransport{
		tcpTransportOPT: tcpTransportOPT,
		rpcCh:	make(chan RPC, 1024),
	}
}

// Consume returns a read-only channel of RPCs that the TCPTransport has received.
// This channel can be used to process incoming RPCs in a non-blocking manner.
func (t *TCPTransport) Consume() <-chan RPC{
	return t.rpcCh
}

func (t *TCPTransport) ListenAndAccept() error {
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
// with the peer, invokes the OnPeer callback if set, and continuously decodes
// incoming RPC messages, sending them to the rpcCh channel.
//
// The function will log an error message and drop the connection if any error
// occurs during the handshake or while decoding messages.
func (t *TCPTransport) handleConn(conn net.Conn){
	var err error
	defer func(){
		fmt.Printf("ERROR: dropping the peer connection %s", err)
	}()

	peer := NewTCPPeer(conn, true)

		
	if err = t.tcpTransportOPT.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil{
		if err = t.OnPeer(peer); err != nil{
			return
		}
	}


	rpc := RPC{}
	for{
		if err := t.tcpTransportOPT.Decoder.Decode(conn, &rpc); err != nil{
			fmt.Println("TCP receiving message error")
			continue
		}

		rpc.From = conn.RemoteAddr();
		t.rpcCh <- rpc
	}	
}
