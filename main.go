package main

import (
	"fmt"

	"github.com/jekki/gdss/common"
	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
	"github.com/jekki/gdss/p2p"
	"github.com/jekki/gdss/pkg"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"stack": pkg.GetCurrentGoroutineStack(),
			}).Fatalf("Panic occurred")
		}
	}()

	go common.Version()

	conf, err := config.LoadConfigProvider()
	if err != nil {
		fmt.Println("config init failed!")
		return
	}

	log.NewLogger(conf)
	log.Infof("Starting gdss node")

	host := conf.GetString("server.host")
	port := conf.GetInt("server.port")
	listenAddr := fmt.Sprintf("%s:%d", host, port)
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
		NodeID:        conf.GetString("node.tcp"),
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for msg := range tr.Consume() {
			log.WithFields(log.Fields{
				"from":    msg.From,
				"payload": string(msg.Payload),
			}).Infof("Received message")
			fmt.Printf("Received message: %s:%s\n", msg.From, msg.Payload)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.WithFields(log.Fields{
			"address": listenAddr,
		}).Fatalf("Failed to start TCP transport: %v", err)
	}

	select {}
}
