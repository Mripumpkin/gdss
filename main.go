package main

import (
	"fmt"

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

func makeServer(listenAddr, root string, nodes ...string) *server.FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcptTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		StorageRoot:       root,
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcptTransport,
		BootstrapNodes:    nodes,
		TraceID:           pkg.GenerateTraceID(),
	}

	s := server.NewFileServer(fileServerOpts)
	tcpTransportOpts.OnPeer = s.Onpeer

	return s
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

	s := makeServer(listenAddr, "gdss_test", ":7790", ":7790")

	// go func() {
	// 	time.Sleep(time.Second * 5)
	// 	s.Stop()
	// }()

	go func() {
		if err := s.Start(); err != nil {
			log.Error(err)
		}
	}()
	select {}
}
