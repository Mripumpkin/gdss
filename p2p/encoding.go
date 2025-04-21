package p2p

import (
	"io"
	"encoding/gob"
)

type Decoder interface {
	Decode(io.Reader, any) error
}

type GOBDecoder struct {}

func (d *GOBDecoder) Decode(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}