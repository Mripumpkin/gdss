package store

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	// Test the default path transform function
	key := "store_dir"
	pathKey := CASPathTransformFunc(key)
	expectOriginalKey := "c67780c83e93dd5f9964398bd97de9c296f923cb"
	expectPathName := "c6778/0c83e/93dd5/f9964/398bd/97de9/c296f/923cb"
	if pathKey.PathName != expectPathName {
		t.Errorf("pathname:%s , %s", pathKey.PathName, expectPathName)
	}
	if pathKey.Filename != expectOriginalKey {
		t.Errorf("filename: %s,%s", pathKey.Filename, expectOriginalKey)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "store_dir"
	data := []byte("test data")
	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
		t.Fatalf("failed to write stream: %v", err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("food_%d", i)

		data := []byte("test data")
		if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
			t.Fatalf("failed to write stream: %v", err)
		}
		r, err := s.Read(key)
		if err != nil {
			t.Fatalf("failed to read: %v", err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read all: %v", err)
		}
		r.Close()

		assert.True(t, s.Has(key), "Expected key to exist")

		if !bytes.Equal(b, data) {
			t.Errorf("expected data %q, got %q", data, b)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); ok {
			t.Errorf("expected to not have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
