package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAdder := ":4344"
	tr := NewTCPTransport(listenAdder)
	assert.Equal(t, tr.listenAddress, listenAdder)

	tr.ListenAndAccept()
	assert.NotNil(t, tr.ListenAndAccept())

	select {}
}
