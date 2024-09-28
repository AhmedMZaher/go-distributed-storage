package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcptransportOPT := &TCPTransportOPT{
		ListenAddress : ":3030",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder: DefaultDecoder{},
	}
	tr := NewTCPTransport(tcptransportOPT)
	assert.Equal(t, tr.tcpTransportOPT.ListenAddress, ":3030")
	assert.Nil(t, tr.ListenAndAccept())
}