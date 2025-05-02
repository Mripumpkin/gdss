package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/jekki/gdss/gcrypto"
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
	id := gcrypto.GenerateID()
	if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(id, key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	id := gcrypto.GenerateID()

	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("food_%d", i)

		data := []byte("test data")
		if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		if ok := s.Has(id, key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		_, r, err := s.Read(id, key)
		if err != nil {
			t.Error(err)
		}

		b, _ := ioutil.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(id, key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); ok {
			t.Errorf("expected to NOT have key %s", key)
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
