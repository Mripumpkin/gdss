package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jekki/gdss/gcrypto"
	"github.com/jekki/gdss/log"
	"github.com/jekki/gdss/p2p"
	"github.com/jekki/gdss/store"
)

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	S        *store.Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.ID) == 0 {
		opts.ID = gcrypto.GenerateID()
	}

	return &FileServer{
		FileServerOpts: opts,
		S:              store.NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	logger := log.WithServerContext(s.Transport.Addr(), s.ID)

	if s.S.Has(s.ID, key) {
		logger.Infof("serving file (%s) from local disk\n", key)
		_, r, err := s.S.Read(s.ID, key)
		return r, err
	}

	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: gcrypto.HashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	timeout := time.After(time.Second * 5)
	responseCh := make(chan io.Reader)
	errCh := make(chan error)

	go func() {
		var fileReader io.Reader

		for _, peer := range s.peers {
			var fileSize int64
			if err := binary.Read(peer, binary.LittleEndian, &fileSize); err != nil {
				errCh <- err
				return
			}

			reader := io.LimitReader(peer, fileSize)

			n, err := s.S.WriteDecrypt(s.EncKey, s.ID, key, reader)
			if err != nil {
				errCh <- err
				return
			}

			logger.Infof("received (%d) bytes over the network from (%s)", n, peer.RemoteAddr())

			peer.CloseStream()
		}

		_, fileReader, err := s.S.Read(s.ID, key)
		if err != nil {
			errCh <- err
			return
		}
		responseCh <- fileReader
	}()

	select {
	case r := <-responseCh:
		return r, nil
	case err := <-errCh:
		return nil, err
	case <-timeout:
		return nil, fmt.Errorf("timeout while waiting for file from peers")
	}
}

func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	logger := log.WithServerContext(s.Transport.Addr(), s.ID)

	size, err := s.S.Write(s.ID, key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			ID:   s.ID,
			Key:  gcrypto.HashKey(key),
			Size: size + 16,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	timeout := time.After(time.Second * 5)
	responseCh := make(chan error)

	go func() {
		time.Sleep(time.Millisecond * 500)

		peers := []io.Writer{}
		for _, peer := range s.peers {
			peers = append(peers, peer)
		}

		mw := io.MultiWriter(peers...)
		mw.Write([]byte{p2p.IncomingStream})

		n, err := gcrypto.CopyEncrypt(s.EncKey, fileBuffer, mw)
		if err != nil {
			responseCh <- err
			return
		}

		logger.Infof("received and written (%d) bytes to disk\n", n)
		responseCh <- nil
	}()

	select {
	case err := <-responseCh:
		if err != nil {
			return fmt.Errorf("failed to send file to peers: %w", err)
		}
	case <-timeout:
		return fmt.Errorf("timeout while waiting for file transfer to peers")
	}

	return nil
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	logger := log.WithServerContext(s.Transport.Addr(), s.ID)
	logger.Infof("connected with remote %s", p.RemoteAddr())
	return nil
}

func (s *FileServer) loop() {
	logger := log.WithServerContext(s.Transport.Addr(), s.ID)

	defer func() {
		logger.Info("file server stopped due to user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Infoln("decoding error: ", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				logger.Infoln("handle message error: ", err)
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	logger := log.WithServerContext(s.Transport.Addr(), s.ID)
	if !s.S.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}

	logger.Infof("serving file (%s) over the network\n", msg.Key)

	fileSize, r, err := s.S.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		logger.Infof("closing readCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	logger.Infof("written (%d) bytes over the network to %s\n", n, from)

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	logger := log.WithServerContext(s.Transport.Addr(), s.ID)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	n, err := s.S.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	logger.Infof("written %d bytes to disk\n", n)

	peer.CloseStream()

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Errorf("dial error: %s", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNodes) != 0 {
		s.bootstrapNetwork()
	}

	s.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
