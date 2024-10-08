package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface that defines a method for decoding messages from an io.Reader.
type Decoder interface {
	// Decode reads from the provided io.Reader and decodes the data into the provided message.
	// Returns an error if the decoding fails.
	Decode(io.Reader, *RPC) error
}

// GOBDecoder is a struct that implements the Decoder interface using the gob package.
type GOBDecoder struct{}

// Decode reads from the provided io.Reader and decodes the data into the provided message using gob.
// Returns an error if the decoding fails.
func (Decoder GOBDecoder) Decode(reader io.Reader, msg *RPC) error {
	return gob.NewDecoder(reader).Decode(msg)
}

// DefaultDecoder is a struct that implements the Decoder interface using a custom decoding method.
type DefaultDecoder struct{}

// Decode reads data from the provided io.Reader and decodes it into the given RPC message.
// It first peeks at the first byte to determine if the incoming data is a stream.
// If it is a stream, it sets the Stream field of the RPC message to true and returns without further reading.
// Otherwise, it reads up to 1024 bytes from the reader into the Payload field of the RPC message.
func (Decoder DefaultDecoder) Decode(reader io.Reader, msg *RPC) error {
	peekBuf := make([]byte, 1)
	if _, err := reader.Read(peekBuf); err != nil {
		return nil
	}

	// In case of a stream we are not decoding what is being sent over the network.
	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]
	return nil
}
