package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAdder := ":4344"
	tr := NewTCPTransport(listenAdder)

	assert.Equal(t, tr.listenAdderss, listenAdder)

	tr.ListenAccept()
}
