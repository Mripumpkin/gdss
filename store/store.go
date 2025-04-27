package store

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/jekki/gdss/log"
)

const (
	defaultRootFolderName = "gdss"
)

// CASPathTransformFunc generates a content-addressable storage path from a key.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockszie := 5
	sliceLen := len(hashStr) / blockszie
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockszie, (i*blockszie)+blockszie
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

// PathTransformFunc defines a function to transform a key into a storage path.
type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

// StoreOpts holds configuration options for the Store.
type StoreOpts struct {
	// Root is the folder name of the root, containing all the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
	NodeID            string // Unique identifier for the node
}

// Store manages file storage operations.
type Store struct {
	StoreOpts
}

// DefaultPathTransformFunc returns the key as the storage path.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

// NewStore creates a new Store with the given options.
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)

	hasPathWithRoot := fmt.Sprintf("%s:%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(hasPathWithRoot)
	return err != fs.ErrNotExist
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		logger := withStoreContext(s.NodeID, key, pathKey.FullPath())
		logger.Infof("deleted [%s] form disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)
}

// withStoreContext creates a logger with store-specific fields.
func withStoreContext(nodeID, key, path string) log.Logger {
	return log.WithFields(log.Fields{
		"node_id": nodeID,
		"key":     key,
		"path":    path,
	})
}

func (s *Store) Read(key string) (io.ReadCloser, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	return f.(io.ReadCloser), nil
}

func (s *Store) readStream(key string) (io.Reader, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullPathWithRoot)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	logger := withStoreContext(s.NodeID, key, pathKey.FullPath())

	if err := os.MkdirAll(pathKeyWithRoot, 0755); err != nil {
		logger.Errorf("Failed to create directory: %v", err)
		return err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		logger.Errorf("Failed to create file: %v", err)
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		logger.Errorf("Failed to write to file: %v", err)
		_ = os.Remove(pathKey.FullPath())
		return err
	}

	logger.Infof("Wrote %d bytes to disk", n)
	return nil
}
