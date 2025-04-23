package store

import (
	"fmt"
	"testing"

	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
)

func TestStore(t *testing.T) {
	conf, err := config.LoadConfigProvider()
	if err != nil {
		fmt.Println("config init failed!")
		return
	}
	log := log.NewLogger(conf)
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
		Logger:            log,
	}
	s := NewStore(opts)
	if s == nil {
		t.Error("Failed to create store")
		return
	}
}
