package store

import (
	"io"
	"os"

	"github.com/jekki/gdss/log"
)

type PathTransformFunc func(string) string

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
	Logger            log.Logger
}

type Store struct {
	StoreOpts
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	filename := "somefilename"
	pathAndFilename := pathName + "/" + filename

	f, err := os.Open(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	s.Logger.Info("Wrote %d bytes to %s", n, pathAndFilename)
	return nil
}
