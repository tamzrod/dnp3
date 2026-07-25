// Package transport provides network transport implementations for DNP3.
package transport

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	// ErrNotConnected indicates the transport is not connected.
	ErrNotConnected = errors.New("not connected")
	// ErrTimeout indicates the operation timed out.
	ErrTimeout = errors.New("timeout")
	// ErrClosed indicates the transport has been closed.
	ErrClosed = errors.New("transport closed")
)

// Handler defines the interface for sending and receiving data.
type Handler interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	SetTimeout(ms int)
}

// TCPConfig holds TCP transport configuration.
type TCPConfig struct {
	// Address to connect to (client) or listen on (server)
	Address string
	// Port to connect to or listen on
	Port int
	// Timeout for connect operation (ms)
	ConnectTimeout int
	// Timeout for receive operation (ms)
	ReceiveTimeout int
	// Server mode (listen for connections)
	Server bool
	// Keepalive interval (ms, 0 = disabled)
	KeepAlive int
}

// DefaultTCPConfig returns default TCP configuration.
func DefaultTCPConfig() *TCPConfig {
	return &TCPConfig{
		ConnectTimeout: 5000,
		ReceiveTimeout: 5000,
		KeepAlive:      30000,
	}
}

// TCPTransport implements Handler for TCP connections.
type TCPTransport struct {
	conn    net.Conn
	config  *TCPConfig
	mu      sync.RWMutex
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewTCPTransport creates a new TCP transport.
func NewTCPTransport(config *TCPConfig) *TCPTransport {
	if config == nil {
		config = DefaultTCPConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPTransport{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect establishes a TCP connection.
func (t *TCPTransport) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}
	if t.conn != nil {
		return nil // Already connected
	}

	addr := fmt.Sprintf("%s:%d", t.config.Address, t.config.Port)
	timeout := time.Duration(t.config.ConnectTimeout) * time.Millisecond

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(t.ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP connect failed: %w", err)
	}

	// Set keepalive if configured
	if t.config.KeepAlive > 0 {
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Duration(t.config.KeepAlive) * time.Millisecond)
		}
	}

	t.conn = conn
	return nil
}

// Accept waits for an incoming TCP connection (server mode).
func (t *TCPTransport) Accept() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}
	if t.conn != nil {
		return nil // Already connected
	}

	addr := fmt.Sprintf("%s:%d", t.config.Address, t.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP listen failed: %w", err)
	}

	timeout := time.Duration(t.config.ConnectTimeout) * time.Millisecond
	conn, err := listener.Accept()
	listener.Close()
	if err != nil {
		return fmt.Errorf("TCP accept failed: %w", err)
	}
	conn.SetReadDeadline(time.Now().Add(timeout))

	t.conn = conn
	return nil
}

// Send writes data to the connection.
func (t *TCPTransport) Send(data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return ErrClosed
	}
	if t.conn == nil {
		return ErrNotConnected
	}

	// DNP3 over TCP uses length prefix
	length := uint16(len(data))
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, length)

	_, err := t.conn.Write(header)
	if err != nil {
		return fmt.Errorf("TCP send header failed: %w", err)
	}

	_, err = t.conn.Write(data)
	if err != nil {
		return fmt.Errorf("TCP send data failed: %w", err)
	}

	return nil
}

// Receive reads data from the connection.
func (t *TCPTransport) Receive() ([]byte, error) {
	t.mu.RLock()
	if t.closed {
		t.mu.RUnlock()
		return nil, ErrClosed
	}
	if t.conn == nil {
		t.mu.RUnlock()
		return nil, ErrNotConnected
	}
	conn := t.conn
	t.mu.RUnlock()

	// Set receive timeout
	timeout := time.Duration(t.config.ReceiveTimeout) * time.Millisecond
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Read length prefix (2 bytes)
	header := make([]byte, 2)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrClosed
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("TCP receive header failed: %w", err)
	}

	length := binary.BigEndian.Uint16(header)
	if length == 0 || length > 65535 {
		return nil, fmt.Errorf("invalid length: %d", length)
	}

	// Read data
	data := make([]byte, length)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrClosed
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("TCP receive data failed: %w", err)
	}

	return data, nil
}

// SetTimeout sets the receive timeout.
func (t *TCPTransport) SetTimeout(ms int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config.ReceiveTimeout = ms
}

// Close closes the TCP connection.
func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	t.cancel()

	if t.conn == nil {
		return nil
	}

	err := t.conn.Close()
	t.conn = nil
	return err
}

// LocalAddr returns the local network address.
func (t *TCPTransport) LocalAddr() net.Addr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return nil
	}
	return t.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (t *TCPTransport) RemoteAddr() net.Addr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return nil
	}
	return t.conn.RemoteAddr()
}

// IsConnected returns true if the transport is connected.
func (t *TCPTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conn != nil && !t.closed
}
