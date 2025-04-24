package store

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	// Test the default path transform function
	key := "11111"
	pathname := CASPathTransformFunc(key)
	if pathname != "7b218/48ac9/af35b/e0ddb/2d6b9/fc385/1934d/b8420" {
		t.Error("expected 7b218/48ac9/af35b/e0ddb/2d6b9/fc385/1934d/b8420, got", pathname)
	}
}
func TestStore(t *testing.T) {

	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)
	data := bytes.NewBuffer([]byte("test data"))
	if err := s.writeStream("store_dir", data); err != nil {
		t.Fatalf("failed to write stream: %v", err)
	}
}
