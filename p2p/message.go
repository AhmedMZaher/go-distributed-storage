package p2p


// The message struct represents a message in the peer-to-peer network.
// It contains the following fields:
// - From: The identifier of the sender.
// - Payload: The content of the message in bytes.
type Message struct{
	From string
	Payload []byte
}