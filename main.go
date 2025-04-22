package main

import (
	"fmt"

	"github.com/jekki/gdss/common"
	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
	"github.com/jekki/gdss/p2p"
	"github.com/jekki/gdss/pkg"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[panic] err: %v\nstack: %s\n", err, pkg.GetCurrentGoroutineStack())
		}
	}()

	go common.Version()

	conf, err := config.LoadConfigProvider()
	if err != nil {
		fmt.Println("config init failed!")
		return
	}
	log := log.NewLogger(conf)

	host := conf.GetString("server.host")
	port := conf.GetInt("server.port")
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpts)
	if err := tr.ListenAndAccept(); err != nil {
		log.Error("failed to start TCP transport", "error", err)
		return
	}
	select {}
}
