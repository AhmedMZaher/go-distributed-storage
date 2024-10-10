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
	net.Conn
	// outbound indicates whether the connection is outbound (true) or inbound (false).
	outbound bool

	// wg (WaitGroup) is used to block the handling loop while waiting for RPC streaming
	// messages to complete processing.
	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// This function implements the transport interface.
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func (p *TCPPeer) Send(bytes []byte) error {
	_, err := p.Conn.Write(bytes)
	return err
}

// TCPTransportOPT holds the configuration options for the TCP transport layer.
// It includes the address to listen on and a function for handling handshakes.
//
// Fields:
// - ListenAddress: The address on which the TCP transport will listen for incoming connections.
// - HandshakeFunc: A function that defines the handshake process for establishing connections.
type TCPTransportOPT struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport represents a transport layer for peer-to-peer communication over TCP.
// It holds the configuration options, a network listener, a channel for RPC messages,
// and a callback function to make the code more generic
// that is triggered when a new peer is connected.
type TCPTransport struct {
	tcpTransportOPT *TCPTransportOPT
	listener        net.Listener
	rpcCh           chan RPC
}

func NewTCPTransport(tcpTransportOPT *TCPTransportOPT) *TCPTransport {
	return &TCPTransport{
		tcpTransportOPT: tcpTransportOPT,
		rpcCh:           make(chan RPC, 1024),
	}
}

// Implements the transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Implements the transport interface
func (t *TCPTransport) RemoteAddr() string {
	return t.listener.Addr().String()
}

// Dial establishes a TCP connection to the specified address.
// Once connected, it spawns a goroutine to handle the connection asynchronously.
// Returns an error if the connection cannot be established.
func (t *TCPTransport) Dial(address string) error {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

// Consume returns a read-only channel of RPCs that the TCPTransport has received.
// This channel can be used to process incoming RPCs in a non-blocking manner.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.tcpTransportOPT.ListenAddress)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	fmt.Printf("TCPTransport listening on address: %s\n", t.listener.Addr())
	return nil
}

func (t *TCPTransport) startAcceptLoop() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			// Check if the error is due to the listener being closed
			// if it's stop the for loop.
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				fmt.Println("TCP listener closed, stopping accept loop")
				return
			}

			// For other errors, log and continue to try accepting connections
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}

		go t.handleConn(conn, false)
	}
}

// handleConn handles an incoming TCP connection. It performs a handshake
// with the peer, invokes the OnPeer callback if set, and continuously decodes
// incoming RPC messages, sending them to the rpcCh channel.
//
// If the message is part of a stream, the read is blocked by using the peer's
// WaitGroup (wg). This allows the server to directly use the net.Conn for reading
// streaming data without sending the message to the RPC channel.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection due to error: %v\n", err)
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.tcpTransportOPT.HandshakeFunc(peer); err != nil {
		return
	}

	if t.tcpTransportOPT.OnPeer != nil {
		if err = t.tcpTransportOPT.OnPeer(peer); err != nil {
			return
		}
	}

	// Continuously decode the incoming RPC messages.
	for {
		rpc := RPC{}
		rpc.From = conn.RemoteAddr()

		err = t.tcpTransportOPT.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("%s Received streaming RPC message from %v\n", t.RemoteAddr(), rpc.From)
			peer.wg.Wait()
			fmt.Printf("Finished processing stream from peer %v\n", rpc.From)
			continue
		}

		t.rpcCh <- rpc
	}
}
