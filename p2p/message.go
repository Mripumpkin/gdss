package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC holds any arbitrary data that is being sent over the
// each transport betwean two nods in the nertwork.
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
