package main

import (
	"fmt"
	"time"

	"github.com/jekki/gdss/common"
	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
	"github.com/jekki/gdss/p2p"
	"github.com/jekki/gdss/pkg"
	"github.com/jekki/gdss/server"
	"github.com/jekki/gdss/store"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s:\n%s", err, pkg.GetCurrentGoroutineStack())
			return
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

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: onPeer func
	}
	tcptTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		StorageRoot:       "gdss_test",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcptTransport,
		TraceID:           pkg.GenerateTraceID(),
	}

	s := server.NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 5)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Error(err)
	}
	// tcpOpts := p2p.TCPTransportOpts{
	// 	ListenAddress: listenAddr,
	// 	HandshakeFunc: p2p.NOPHandshakeFunc,
	// 	Decoder:       &p2p.DefaultDecoder{},
	// 	OnPeer:        OnPeer,
	// 	TraceID:       pkg.GenerateTraceID(),
	// }
	// tr := p2p.NewTCPTransport(tcpOpts)

	// go func() {
	// 	for msg := range tr.Consume() {
	// 		log.WithFields(log.Fields{
	// 			"from":    msg.From,
	// 			"payload": string(msg.Payload),
	// 		}).Infof("Received message")
	// 		fmt.Printf("Received message: %s:%s\n", msg.From, msg.Payload)
	// 	}
	// }()

	// if err := tr.ListenAndAccept(); err != nil {
	// 	log.WithFields(log.Fields{
	// 		"address": listenAddr,
	// 	}).Fatalf("Failed to start TCP transport: %v", err)
	// }

	// select {}
}
