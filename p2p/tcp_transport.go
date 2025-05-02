package p2p

import (
	"errors"
	"net"
	"sync"

	"github.com/jekki/gdss/log"
)

// TCPPeer represents a remote node over a TCP connection.
type TCPPeer struct {
	net.Conn
	outbound bool
	wg       *sync.WaitGroup
}

// NewTCPPeer creates a new TCPPeer.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// // Conn returns the underlying connection.
// func (p *TCPPeer) Conn() net.Conn {
// 	return p.conn
// }

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// Close the connection.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// NewTCPTransport creates a new TCPTransport.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddress
}

// Consume returns a read-only channel for incoming RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// close implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

// ListenAndAccept starts listening for incoming connections.
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	logger := log.WithFields(log.Fields{
		"listenaddr": t.ListenAddress,
	})
	if err != nil {
		logger.Errorf("TCP listen error: %v", err)
		return err
	}

	go t.StartAcceptLoop()

	return nil
}

// StartAcceptLoop accepts incoming connections in a loop.
func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.WithFields(log.Fields{
				"address": t.ListenAddress,
			}).Errorf("TCP accept error: %v", err)
			continue
		}
		go t.handleConn(conn, false)
	}
}

// handleConn processes a new incoming connection.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	peerAddr := conn.RemoteAddr().String()
	logger := log.WithPeerContext(peerAddr, t.ListenAddress)

	defer func() {
		if err != nil {
			log.Error("dropping peer connection: %s", err)
			conn.Close()
		}
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		logger.Errorf("TCP handshake error: %v", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			logger.Errorf("OnPeer callback error: %v", err)
			return
		}
	}

	for {
		rpc := RPC{}
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			logger.Errorf("Decode error: %v", err)
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			logger.Info("incoming stream, waiting...")
			peer.wg.Wait()
			logger.Info("stream closed, resuming read loop")
			continue
		}
		t.rpcch <- rpc
	}
}
