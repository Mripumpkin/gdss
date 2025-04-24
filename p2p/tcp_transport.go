package p2p

import (
	"fmt"
	"net"
	"time"

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
	NodeID        string // Unique identifier for the node
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

// Consume returns a read-only channel for incoming RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// ListenAndAccept starts listening for incoming connections.
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"node_id": t.NodeID,
			"address": t.ListenAddress,
		}).Errorf("TCP listen error: %v", err)
		return err
	}

	log.WithFields(log.Fields{
		"node_id": t.NodeID,
		"address": t.ListenAddress,
	}).Infof("Listening for connections")
	go t.StartAcceptLoop()
	return nil
}

// StartAcceptLoop accepts incoming connections in a loop.
func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"node_id": t.NodeID,
				"address": t.ListenAddress,
			}).Errorf("TCP accept error: %v", err)
			continue
		}
		go t.handleConn(conn)
	}
}

// withPeerContext creates a logger with peer-specific fields.
func withPeerContext(nodeID, peerAddr, localAddr string, traceID ...string) log.Logger {
	fields := log.Fields{
		"node_id":    nodeID,
		"peer":       peerAddr,
		"local_addr": localAddr,
	}
	if len(traceID) > 0 && traceID[0] != "" {
		fields["trace_id"] = traceID[0]
	}
	return log.WithFields(fields)
}

// handleConn processes a new incoming connection.
func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	peerAddr := conn.RemoteAddr().String()
	traceID := generateTraceID()
	logger := withPeerContext(t.NodeID, peerAddr, t.ListenAddress, traceID)

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

// generateTraceID generates a unique trace ID for distributed tracing (placeholder).
func generateTraceID() string {
	return "trace-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
