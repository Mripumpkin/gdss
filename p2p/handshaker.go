package p2p

import "errors"

// ErrInvalidPeer is returned if the handsjake beteen
// the local and remote
var ErrInvalidHandshake = errors.New("invalid handshake")

// HandshakeFunc is a function that performs a handshake with the peer.
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error { return nil }
