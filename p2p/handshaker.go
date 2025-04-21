package p2p

type Handshaker interface {
	Handshaker() error
}

type DefaultHandshaker struct {
}
