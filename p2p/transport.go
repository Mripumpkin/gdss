package p2p

// Peer is an interface that reperesent the remote node.
type Peer interface {
	
}

// Transport is anything that handles the communication
// between the nodes in the network. This is can be of the
// from (TCP, UDP, websockets, ...)
type Transport interface {
	ListenAndAccept() error
}
