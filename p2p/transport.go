package p2p

import "net"

// Peer is an interface that reperesent the remote node.
type Peer interface {
	RemoteAddr() net.Addr
	Close() error
}

// Transport is anything that handles the communication
// between the nodes in the network. This is can be of the
// from (TCP, UDP, websockets, ...)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	LocalAddr() string
}
