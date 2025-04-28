package log

// withPeerContext creates a logger with peer-specific fields.
func WithPeerContext(peerAddr, localAddr string, traceID ...string) Logger {
	fields := Fields{
		"peer":    peerAddr,
		"address": localAddr,
	}
	if len(traceID) > 0 && traceID[0] != "" {
		fields["trace_id"] = traceID[0]
	}
	return WithFields(fields)
}
