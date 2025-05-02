package log

// withStoreContext creates a logger with store-specific fields.
func WithStoreContext(key, path, localAddr string, id ...string) Logger {
	fields := Fields{
		"key":  key,
		"path": path,
		// "address": localAddr,
	}
	if len(id) > 0 && id[0] != "" {
		fields["id"] = id[0]
	}
	return WithFields(fields)
}
