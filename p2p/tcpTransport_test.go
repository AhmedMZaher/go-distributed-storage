package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAddress := "8080"
	tr := NewTCPTransport(listenAddress)
	assert.Equal(t, listenAddress, tr.listenAddress);
}