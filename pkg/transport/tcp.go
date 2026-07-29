// Package transport provides network transport implementations for DNP3.
//
// DNP3 over TCP uses standard IEEE 1815 framing:
// - Frames start with sync bytes (0x05 0x64)
// - Length field at offset 2 (1 byte)
// - Full frame: sync(2) + length(1) + rest + CRCs
//
// Unlike the old implementation, NO external length prefix is added.
// Frames are self-delimiting via sync bytes.
package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// DNP3 sync bytes - start of every frame per IEEE 1815
const (
	SyncByte1 = 0x05
	SyncByte2 = 0x64
)

var (
	// ErrNotConnected indicates the transport is not connected.
	ErrNotConnected = errors.New("not connected")
	// ErrTimeout indicates the operation timed out.
	ErrTimeout = errors.New("timeout")
	// ErrClosed indicates the transport has been closed.
	ErrClosed = errors.New("transport closed")
	// ErrInvalidFrame indicates an invalid DNP3 frame was received.
	ErrInvalidFrame = errors.New("invalid DNP3 frame")
)

// Handler defines the interface for sending and receiving data.
type Handler interface {
	// Connect establishes a connection (client mode)
	Connect() error
	// Accept waits for an incoming connection (server mode)
	Accept() error
	// Close closes the transport
	Close() error
	// Send sends data
	Send(data []byte) error
	// Receive receives data
	Receive() ([]byte, error)
	// SetTimeout sets the receive timeout in milliseconds
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
	conn     net.Conn
	listener net.Listener
	config   *TCPConfig
	mu       sync.RWMutex
	closed   bool
	ctx      context.Context
	cancel   context.CancelFunc
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
// In server mode, the listener is kept open to accept multiple connections.
func (t *TCPTransport) Accept() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}

	// If already connected, return success
	if t.conn != nil {
		return nil
	}

	// Create listener if not already created
	if t.listener == nil {
		addr := fmt.Sprintf("%s:%d", t.config.Address, t.config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP listen failed: %w", err)
		}
		t.listener = listener
	}

	// For deadline support, we need to use TCPListener specifically
	// Cast to *net.TCPListener if possible
	var conn net.Conn
	var err error

	if tcpListener, ok := t.listener.(*net.TCPListener); ok {
		// Set deadline for accepting
		timeout := time.Duration(t.config.ConnectTimeout) * time.Millisecond
		tcpListener.SetDeadline(time.Now().Add(timeout))
		conn, err = tcpListener.Accept()
	} else {
		// Fallback for non-TCP listeners
		conn, err = t.listener.Accept()
	}

	if err != nil {
		// Check if it's a timeout
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return ErrTimeout
		}
		// Check if it's a closed error (context cancellation)
		if t.closed {
			return ErrClosed
		}
		return fmt.Errorf("TCP accept failed: %w", err)
	}

	// Set read deadline on connection
	timeout := time.Duration(t.config.ConnectTimeout) * time.Millisecond
	conn.SetReadDeadline(time.Now().Add(timeout))

	t.conn = conn
	return nil
}

// Send writes data to the connection.
//
// Standard DNP3 over TCP: sends data directly without length prefix.
// The data should be a complete DNP3 DLL frame starting with sync bytes (0x05 0x64).
func (t *TCPTransport) Send(data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return ErrClosed
	}
	if t.conn == nil {
		return ErrNotConnected
	}

	// Standard DNP3: send frame directly (no length prefix)
	// Frame must start with sync bytes 0x05 0x64
	if len(data) >= 2 && data[0] == SyncByte1 && data[1] == SyncByte2 {
		// Valid DNP3 frame - send as-is
	} else if len(data) > 0 {
		// Data doesn't look like standard DNP3 - still send (for flexibility)
		// Caller is responsible for proper framing
	}

	_, err := t.conn.Write(data)
	if err != nil {
		return fmt.Errorf("TCP send failed: %w", err)
	}

	return nil
}

// Receive reads data from the connection.
//
// Standard DNP3 over TCP: reads self-delimiting frames.
// Uses sync bytes (0x05 0x64) to find frame start, then reads length field.
// Returns the complete DNP3 frame including sync bytes and CRCs.
//
// Frame structure per IEEE 1815:
//   Sync(2) + Length(1) + Control(1) + Dest(2) + Src(2) + Data + CRCs
//
// CRCs are 2 bytes each, covering:
//   - Length + Control (1 pair)
//   - Destination (1 pair)
//   - Source (1 pair)
//   - Data (ceil(dataLen/2) pairs)
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

	// Read first byte to look for sync
	firstByte := make([]byte, 1)
	_, err := io.ReadFull(conn, firstByte)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrClosed
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("TCP receive failed: %w", err)
	}

	// Check if we got sync byte 1
	if firstByte[0] != SyncByte1 {
		// Not sync byte 1 - keep reading until we find it
		// This handles the case where we start mid-frame
		for {
			nextByte := make([]byte, 1)
			_, err := io.ReadFull(conn, nextByte)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
					return nil, ErrClosed
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return nil, ErrTimeout
				}
				return nil, fmt.Errorf("TCP receive failed: %w", err)
			}
			if nextByte[0] == SyncByte1 {
				break
			}
		}
	}

	// Read sync byte 2
	secondByte := make([]byte, 1)
	_, err = io.ReadFull(conn, secondByte)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrClosed
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("TCP receive failed reading sync byte 2: %w", err)
	}

	if secondByte[0] != SyncByte2 {
		return nil, fmt.Errorf("%w: expected 0x%02x, got 0x%02x", ErrInvalidFrame, SyncByte2, secondByte[0])
	}

	// Read length byte (byte at offset 2 in frame)
	lengthByte := make([]byte, 1)
	_, err = io.ReadFull(conn, lengthByte)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrClosed
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("TCP receive failed reading length: %w", err)
	}

	// Length field includes: Control(1) + Dest(2) + Src(2) + Data
	// NOT including the CRCs - they are calculated separately
	frameLength := int(lengthByte[0])
	
	// Calculate data length (length - control - dest - src)
	dataLen := frameLength - 5 // 1 + 2 + 2
	if dataLen < 0 {
		dataLen = 0
	}
	
	// Calculate CRC bytes:
	// - 3 header CRCs (Length+Ctrl, Dest, Src) = 6 bytes
	// - Data CRCs = ceil(dataLen/2) pairs = 2 * ceil(dataLen/2) bytes
	numDataCRCPairs := (dataLen + 1) / 2 // ceil division
	crcBytes := 6 + (numDataCRCPairs * 2)
	
	// Total frame size: sync(2) + length(1) + rest(frameLength) + CRCs
	totalSize := 2 + 1 + frameLength + crcBytes

	// Read the complete frame
	frame := make([]byte, totalSize)
	frame[0] = SyncByte1
	frame[1] = SyncByte2
	frame[2] = lengthByte[0]

	// Read rest of frame (control + dest + src + data + CRCs)
	restSize := totalSize - 3
	if restSize > 0 {
		_, err = io.ReadFull(conn, frame[3:totalSize])
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, ErrClosed
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, ErrTimeout
			}
			return nil, fmt.Errorf("TCP receive failed reading frame data: %w", err)
		}
	}

	return frame, nil
}

// SetTimeout sets the receive timeout.
func (t *TCPTransport) SetTimeout(ms int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config.ReceiveTimeout = ms
}

// Close closes the TCP connection and listener.
func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	t.cancel()

	// Close the connection if it exists
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}

	// Close the listener if it exists
	if t.listener != nil {
		_ = t.listener.Close()
		t.listener = nil
	}

	return nil
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

// SetConn sets the underlying connection directly.
// This is useful for server mode where the connection is accepted
// before the transport is created.
func (t *TCPTransport) SetConn(conn net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.conn = conn
}

// SetListener sets the listener directly.
// This is useful for server mode where the listener is managed externally.
func (t *TCPTransport) SetListener(listener net.Listener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listener = listener
}
