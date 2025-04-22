package p2p

import "net"

// RPC holds any arbitrary data that is being sent over the
// each transport betwean two nods in the nertwork.
type RPC struct {
	From    net.Addr
	Payload []byte
}
