package store

import (
	"bytes"
	"io"
	"testing"
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

func TestStore(t *testing.T) {

	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "store_dir"
	data := []byte("test data")
	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
		t.Fatalf("failed to write stream: %v", err)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	defer r.Close() // 确保关闭 io.ReadCloser

	// 读取所有内容
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	// 比较数据
	if !bytes.Equal(b, data) {
		t.Errorf("expected data %q, got %q", data, b)
	}
}
