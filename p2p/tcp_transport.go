package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransport struct {
	listenAdderss string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddress string) *TCPTransport {
	return &TCPTransport{
		listenAdderss: listenAddress,
	}
}

func (t *TCPTransport) ListenAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAdderss)
	if err != nil {
		return err
	}

	go t.acceptloop()
	return nil
}

func (t *TCPTransport) acceptloop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	fmt.Printf("new incoming connectiobn from %+v\n", conn)
}

func (t *TCPTransport) Start() error {
	listener, err := net.Listen("tcp", t.listenAdderss)
	if err != nil {
		return err
	}
	t.listener = listener

	t.peers = make(map[net.Addr]Peer)

	go func() {
		for {
			conn, err := t.listener.Accept()
			if err != nil {
				continue
			}
			t.mu.Lock()
			t.peers[conn.RemoteAddr()] = nil // TODO: create a peer from the connection
			t.mu.Unlock()
		}
	}()

	return nil
}
