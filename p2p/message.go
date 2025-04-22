package p2p

// Message holds any arbitrary data that is being sent over the
// each transport betwean two nods in the nertwork.
type Message struct {
	Payload []byte
}