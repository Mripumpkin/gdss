package p2p

import "net"

// Message holds any arbitrary data that is being sent over the
// each transport betwean two nods in the nertwork.
type Message struct {
	From    net.Addr
	Payload []byte
}
