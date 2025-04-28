package log

// withStoreContext creates a logger with store-specific fields.
func WithStoreContext(key, path, localAddr string, traceID ...string) Logger {
	fields := Fields{
		"key":  key,
		"path": path,
		// "address": localAddr,
	}
	if len(traceID) > 0 && traceID[0] != "" {
		fields["trace_id"] = traceID[0]
	}
	return WithFields(fields)
}
