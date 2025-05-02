package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jekki/gdss/common"
	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/gcrypto"
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
		EncKey:            gcrypto.NewEncryptionKey(),
		StorageRoot:       root,
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcptTransport,
		BootstrapNodes:    nodes,
	}

	s := server.NewFileServer(fileServerOpts)
	tcptTransport.OnPeer = s.OnPeer

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
	root_test := conf.GetString("app.root_test")
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	s1 := makeServer(listenAddr, root_test)
	s2 := makeServer(":7000", root_test)
	s3 := makeServer(":6666", root_test, ":7790", ":7000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.Store(key, data)

		if err := s3.S.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Info(string(b))
	}
}
