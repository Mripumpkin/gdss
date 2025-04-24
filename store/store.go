package store

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/jekki/gdss/log"
)

// CASPathTransformFunc generates a content-addressable storage path from a key.
func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockszie := 5
	sliceLen := len(hashStr) / blockszie
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockszie, (i*blockszie)+blockszie
		paths[i] = hashStr[from:to]
	}
	return strings.Join(paths, "/")
}

// PathTransformFunc defines a function to transform a key into a storage path.
type PathTransformFunc func(string) string

// StoreOpts holds configuration options for the Store.
type StoreOpts struct {
	PathTransformFunc PathTransformFunc
	NodeID            string // Unique identifier for the node
}

// Store manages file storage operations.
type Store struct {
	StoreOpts
}

// DefaultPathTransformFunc returns the key as the storage path.
var DefaultPathTransformFunc = func(key string) string {
	return key
}

// NewStore creates a new Store with the given options.
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	return &Store{
		StoreOpts: opts,
	}
}

// withStoreContext creates a logger with store-specific fields.
func withStoreContext(nodeID, key, path string) log.Logger {
	return log.WithFields(log.Fields{
		"node_id": nodeID,
		"key":     key,
		"path":    path,
	})
}

// writeStream writes data from a reader to a file, using the key to determine the path.
func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)
	logger := withStoreContext(s.NodeID, key, pathName)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		logger.Errorf("Failed to create directory: %v", err)
		return err
	}

	// Buffer data to compute MD5 hash
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		logger.Errorf("Failed to buffer data: %v", err)
		return err
	}

	// Compute filename using MD5 hash
	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])
	pathAndFilename := pathName + "/" + filename
	logger = withStoreContext(s.NodeID, key, pathAndFilename)

	// Create file
	f, err := os.Create(pathAndFilename)
	if err != nil {
		logger.Errorf("Failed to create file: %v", err)
		return err
	}
	defer f.Close()

	// Write buffered data to file
	reader := bytes.NewReader(buf.Bytes())
	n, err := io.Copy(f, reader)
	if err != nil {
		logger.Errorf("Failed to write to file: %v", err)
		return err
	}

	logger.Infof("Wrote %d bytes to disk", n)
	return nil
}
