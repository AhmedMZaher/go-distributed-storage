package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface that defines a method for decoding messages from an io.Reader.
type Decoder interface {
	// Decode reads from the provided io.Reader and decodes the data into the provided message.
	// Returns an error if the decoding fails.
	Decode(io.Reader, *Message) error
}

// GOBDecoder is a struct that implements the Decoder interface using the gob package.
type GOBDecoder struct{}

// Decode reads from the provided io.Reader and decodes the data into the provided message using gob.
// Returns an error if the decoding fails.
func (Decoder GOBDecoder) Decode(reader io.Reader, msg *Message) error {
	return gob.NewDecoder(reader).Decode(msg)
}

// DefaultDecoder is a struct that implements the Decoder interface using a custom decoding method.
type DefaultDecoder struct{}

// Decode reads from the provided io.Reader and decodes the data into the provided message.
// It reads up to 1024 bytes into a buffer and assigns it to the message's Payload field.
// Returns an error if the reading fails.
func (Decoder DefaultDecoder) Decode(reader io.Reader, msg *Message) error {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]
	return nil
}
