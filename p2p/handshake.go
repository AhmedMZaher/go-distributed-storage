// HandshakeFunc defines a function type that takes a Peer and returns an error.
// It is used to perform handshake operations between peers in the P2P network.
//
// NOPHandshakeFunc is a no-operation implementation of HandshakeFunc.
// It takes a Peer and always returns nil, effectively performing no handshake.
package p2p

type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error{
	return nil
}