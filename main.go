package main

import (
	"fmt"

	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
	"github.com/jekki/gdss/p2p"
)

func main() {
	conf, err := config.LoadConfigProvider()
	if err != nil {
		fmt.Println("config init failed!")
		return
	}
	log := log.NewLogger(conf)

	tr := p2p.NewTCPTransport(":4344")
	if err := tr.ListenAndAccept(); err != nil {
		log.Error("failed to start TCP transport", "error", err)
		return
	}
	select {}
}
