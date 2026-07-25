// Package transport provides network transport implementations for DNP3.
package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// TLSConfig holds TLS transport configuration.
type TLSConfig struct {
	// TCP configuration
	TCPConfig
	// TLS certificate file (PEM)
	CertFile string
	// TLS key file (PEM)
	KeyFile string
	// CA certificate file for client verification (PEM)
	CAFile string
	// Server mode (require client certificates)
	Server bool
	// Minimum TLS version (1.0, 1.1, 1.2, 1.3)
	MinVersion uint16
	// Insecure skip verify (DO NOT use in production)
	InsecureSkipVerify bool
}

// DefaultTLSConfig returns default TLS configuration.
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		TCPConfig: *DefaultTCPConfig(),
		MinVersion: tls.VersionTLS12,
	}
}

// TLSTransport implements Handler for TLS connections.
type TLSTransport struct {
	conn    *tls.Conn
	config  *TLSConfig
	tlsConfig *tls.Config
	mu      sync.RWMutex
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewTLSTransport creates a new TLS transport.
func NewTLSTransport(config *TLSConfig) (*TLSTransport, error) {
	if config == nil {
		config = DefaultTLSConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	t := &TLSTransport{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Build TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:         config.MinVersion,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS certificate failed: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if config.CAFile != "" {
		caCert, err := loadCAFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("load CA certificate failed: %w", err)
		}
		tlsConfig.RootCAs = caCert
		tlsConfig.ClientCAs = caCert
	}

	if config.Server {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	t.tlsConfig = tlsConfig
	return t, nil
}

// loadCAFile loads a CA certificate file.
func loadCAFile(filename string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	data, err := loadFile(filename)
	if err != nil {
		return nil, err
	}
	if !pool.AppendCertsFromPEM(data) {
		return nil, errors.New("failed to parse CA certificate")
	}
	return pool, nil
}

// loadFile reads a file.
func loadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %w", filename, err)
	}
	return data, nil
}

// Connect establishes a TLS connection.
func (t *TLSTransport) Connect() error {
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

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, t.tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connect failed: %w", err)
	}

	t.conn = conn
	return nil
}

// Accept waits for an incoming TLS connection (server mode).
func (t *TLSTransport) Accept() error {
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
		return fmt.Errorf("TLS listen failed: %w", err)
	}

	timeout := time.Duration(t.config.ConnectTimeout) * time.Millisecond
	conn, err := listener.Accept()
	listener.Close()
	if err != nil {
		return fmt.Errorf("TLS accept failed: %w", err)
	}

	// Upgrade to TLS
	tlsConn := tls.Server(conn, t.tlsConfig)
	tlsConn.SetReadDeadline(time.Now().Add(timeout))

	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return fmt.Errorf("TLS handshake failed: %w", err)
	}

	t.conn = tlsConn
	return nil
}

// Send writes data to the TLS connection.
func (t *TLSTransport) Send(data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return ErrClosed
	}
	if t.conn == nil {
		return ErrNotConnected
	}

	// DNP3 over TLS uses length prefix
	length := uint16(len(data))
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, length)

	_, err := t.conn.Write(header)
	if err != nil {
		return fmt.Errorf("TLS send header failed: %w", err)
	}

	_, err = t.conn.Write(data)
	if err != nil {
		return fmt.Errorf("TLS send data failed: %w", err)
	}

	return nil
}

// Receive reads data from the TLS connection.
func (t *TLSTransport) Receive() ([]byte, error) {
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
		return nil, fmt.Errorf("TLS receive header failed: %w", err)
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
		return nil, fmt.Errorf("TLS receive data failed: %w", err)
	}

	return data, nil
}

// SetTimeout sets the receive timeout.
func (t *TLSTransport) SetTimeout(ms int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config.ReceiveTimeout = ms
}

// Close closes the TLS connection.
func (t *TLSTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	if t.cancel != nil {
		t.cancel()
	}

	if t.conn == nil {
		return nil
	}

	err := t.conn.Close()
	t.conn = nil
	return err
}

// LocalAddr returns the local network address.
func (t *TLSTransport) LocalAddr() net.Addr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return nil
	}
	return t.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (t *TLSTransport) RemoteAddr() net.Addr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return nil
	}
	return t.conn.RemoteAddr()
}

// IsConnected returns true if the transport is connected.
func (t *TLSTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conn != nil && !t.closed
}

// ConnectionState returns the TLS connection state.
func (t *TLSTransport) ConnectionState() (tls.ConnectionState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return tls.ConnectionState{}, false
	}
	return t.conn.ConnectionState(), true
}
