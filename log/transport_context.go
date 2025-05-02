package log

// withPeerContext creates a logger with peer-specific fields.
func WithPeerContext(peerAddr, localAddr string, id ...string) Logger {
	fields := Fields{
		"peer":       peerAddr,
		"listenaddr": localAddr,
	}
	if len(id) > 0 && id[0] != "" {
		fields["trace_id"] = id[0]
	}
	return WithFields(fields)
}
