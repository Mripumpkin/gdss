package log

// withStoreContext creates a logger with store-specific fields.
func WithServerContext(localAddr string, id ...string) Logger {
	fields := Fields{
		"listenaddr": localAddr,
	}
	if len(id) > 0 && id[0] != "" {
		fields["id"] = id[0]
	}
	return WithFields(fields)
}
