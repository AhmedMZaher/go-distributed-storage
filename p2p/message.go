package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC represents a message in the peer-to-peer network.
// It contains the address of the sender, the payload of the message, and a flag
// to indicate wether the message is part of a stream.
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
