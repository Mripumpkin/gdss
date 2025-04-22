// p2p/p2p.go
package p2p

import (
	"fmt"
	"net"
	"sync"

	"github.com/jekki/gdss/log"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	conn     net.Conn
	outbound bool // true if we dialed, false if we accepted
}

// NewTCPPeer creates a new TCPPeer.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

// Conn returns the underlying connection.
func (p *TCPPeer) Conn() net.Conn {
	return p.conn
}

// IsOutbound returns whether the peer is outbound.
func (p *TCPPeer) IsOutbound() bool {
	return p.outbound
}

// close implements the Peers interface.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransport manages TCP listening and connections.
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// NewTCPTransport creates a new TCPTransport.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC), // Buffered channel for RPCs
		// peers:         make(map[net.Addr]Peer), // Initialize peers map
	}
}

// consume implement the Transport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// ListenAndAccept starts listening for incoming connections.
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		fmt.Printf("TCP listen error: %s\n", err)
		return err
	}

	go t.StartAcceptLoop()
	return nil
}

// StartAcceptLoop accepts incoming connections in a loop.
func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		go t.handleConn(conn)
	}
}

// handleConn processes a new incoming connection.
func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, false) // Inbound connection

	if err := t.HandshakeFunc(peer); err != nil {
		log.Errorf("TCP handshake error: %s\n", err)
		conn.Close()
		return
	}

	rpc := RPC{}
	// buf := make([]byte, 1024)
	for {
		//n, err := conn.Read(buf)
		//if err != nil {
		//	fmt.Printf("TCP read error: %s\n", err)
		//	continue
		//}
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
		fmt.Printf("Received message: %s:%s", rpc.From, rpc.Payload)
	}
}
