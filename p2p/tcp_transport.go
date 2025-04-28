package p2p

import (
	"errors"
	"net"

	"github.com/jekki/gdss/log"
)

// TCPPeer represents a remote node over a TCP connection.
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

// NewTCPPeer creates a new TCPPeer.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Conn returns the underlying connection.
func (p *TCPPeer) Conn() net.Conn {
	return p.conn
}

// IsOutbound returns whether the peer is outbound.
func (p *TCPPeer) IsOutbound() bool {
	return p.outbound
}

// Close closes the connection.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransportOpts holds configuration options for TCPTransport.
type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport manages TCP listening and connections.
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// NewTCPTransport creates a new TCPTransport.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

func (t *TCPTransport) LocalAddr() string {
	return t.ListenAddress
}

// Consume returns a read-only channel for incoming RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// ListenAndAccept starts listening for incoming connections.
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	logger := log.WithFields(log.Fields{
		"address": t.ListenAddress,
	})
	if err != nil {
		logger.Errorf("TCP listen error: %v", err)
		return err
	}

	logger.Infof("Listening for connections")
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
		go t.handleConn(conn)
	}
}

// handleConn processes a new incoming connection.
func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	peerAddr := conn.RemoteAddr().String()
	logger := log.WithPeerContext(peerAddr, t.ListenAddress)

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	peer := NewTCPPeer(conn, false)

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

	rpc := RPC{}
	for {
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			logger.Errorf("Decode error: %v", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
