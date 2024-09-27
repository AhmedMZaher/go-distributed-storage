// Package p2p provides the implementation of peer-to-peer communication
// mechanisms for the distributed storage system.

package p2p

// Peer represents a node in the network.
type Peer interface{
	Close() error
}

// Transport represents the communication layer used by peers to exchange data.
// It can be implemented using various protocols such as TCP, UDP, etc.
type Transport interface{
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}