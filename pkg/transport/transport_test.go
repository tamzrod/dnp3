// Package transport provides tests for network transport implementations.
package transport

import (
	"bytes"
	"net"
	"testing"
)

func TestTCPReceiveGoldenFrame(t *testing.T) {
	peer, local := net.Pipe()
	defer peer.Close()
	transport := NewTCPTransport(&TCPConfig{ReceiveTimeout: 1000})
	transport.SetConn(local)
	defer transport.Close()

	want := []byte{0x05, 0x64, 0x0B, 0xC4, 0x04, 0x00, 0x03, 0x00, 0xE4, 0x2B, 0xE5, 0xC0, 0x01, 0x02, 0x00, 0x06, 0x98, 0x5C}
	writeDone := make(chan error, 1)
	go func() {
		_, err := peer.Write(want)
		peer.Close()
		writeDone <- err
	}()

	got, err := transport.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("received frame = %x, want %x", got, want)
	}
	if err := <-writeDone; err != nil {
		t.Fatalf("write golden frame: %v", err)
	}
}

func TestTCPReceiveConcatenatedFrames(t *testing.T) {
	peer, local := net.Pipe()
	defer peer.Close()
	transport := NewTCPTransport(&TCPConfig{ReceiveTimeout: 1000})
	transport.SetConn(local)
	defer transport.Close()

	frame := []byte{0x05, 0x64, 0x0B, 0xC4, 0x04, 0x00, 0x03, 0x00, 0xE4, 0x2B, 0xE5, 0xC0, 0x01, 0x02, 0x00, 0x06, 0x98, 0x5C}
	want := append(append([]byte(nil), frame...), frame...)
	go func() { _, _ = peer.Write(want); _ = peer.Close() }()

	first, err := transport.Receive()
	if err != nil { t.Fatalf("first Receive: %v", err) }
	if !bytes.Equal(first, frame) { t.Fatalf("first frame = %x, want %x", first, frame) }
	second, err := transport.Receive()
	if err != nil { t.Fatalf("second Receive: %v", err) }
	if !bytes.Equal(second, frame) { t.Fatalf("second frame = %x, want %x", second, frame) }
}

func TestTCPReceiveFragmentedWrite(t *testing.T) {
	peer, local := net.Pipe()
	defer peer.Close()
	transport := NewTCPTransport(&TCPConfig{ReceiveTimeout: 1000})
	transport.SetConn(local)
	defer transport.Close()

	want := []byte{0x05, 0x64, 0x0B, 0xC4, 0x04, 0x00, 0x03, 0x00, 0xE4, 0x2B, 0xE5, 0xC0, 0x01, 0x02, 0x00, 0x06, 0x98, 0x5C}
	go func() {
		_, _ = peer.Write(want[:3])
		_, _ = peer.Write(want[3:10])
		_, _ = peer.Write(want[10:])
		_ = peer.Close()
	}()

	got, err := transport.Receive()
	if err != nil { t.Fatalf("Receive: %v", err) }
	if !bytes.Equal(got, want) { t.Fatalf("received frame = %x, want %x", got, want) }
}

// TestTCPConfig tests TCP configuration defaults.
func TestTCPConfig(t *testing.T) {
	cfg := DefaultTCPConfig()
	if cfg.ConnectTimeout != 5000 {
		t.Errorf("Default connect timeout = %d, want 5000", cfg.ConnectTimeout)
	}
	if cfg.ReceiveTimeout != 5000 {
		t.Errorf("Default receive timeout = %d, want 5000", cfg.ReceiveTimeout)
	}
	if cfg.KeepAlive != 30000 {
		t.Errorf("Default keepalive = %d, want 30000", cfg.KeepAlive)
	}
}

// TestTLSConfig tests TLS configuration defaults.
func TestTLSConfig(t *testing.T) {
	cfg := DefaultTLSConfig()
	if cfg.ConnectTimeout != 5000 {
		t.Errorf("Default connect timeout = %d, want 5000", cfg.ConnectTimeout)
	}
	if cfg.MinVersion != 0x0303 { // TLS 1.2
		t.Errorf("Default min version = 0x%04X, want 0x0303", cfg.MinVersion)
	}
}

// TestTCPTransportClose tests that Close can be called multiple times.
func TestTCPTransportClose(t *testing.T) {
	transport := NewTCPTransport(nil)
	
	// Close when not connected
	err := transport.Close()
	if err != nil {
		t.Errorf("Close on unconnected transport failed: %v", err)
	}
	
	// Close again
	err = transport.Close()
	if err != nil {
		t.Errorf("Double close failed: %v", err)
	}
}

// TestTLSTransportClose tests that TLS transport Close works.
func TestTLSTransportClose(t *testing.T) {
	transport, err := NewTLSTransport(nil)
	if err != nil {
		t.Fatalf("NewTLSTransport failed: %v", err)
	}
	
	// Close when not connected
	err = transport.Close()
	if err != nil {
		t.Errorf("Close on unconnected transport failed: %v", err)
	}
}

// TestTCPTransportState tests transport state methods.
func TestTCPTransportState(t *testing.T) {
	transport := NewTCPTransport(nil)
	
	if transport.IsConnected() {
		t.Error("New transport should not be connected")
	}
	
	if transport.LocalAddr() != nil {
		t.Error("Unconnected transport should have nil local address")
	}
	
	if transport.RemoteAddr() != nil {
		t.Error("Unconnected transport should have nil remote address")
	}
}

// TestTLSTransportState tests TLS transport state methods.
func TestTLSTransportState(t *testing.T) {
	transport, err := NewTLSTransport(nil)
	if err != nil {
		t.Fatalf("NewTLSTransport failed: %v", err)
	}
	
	if transport.IsConnected() {
		t.Error("New TLS transport should not be connected")
	}
	
	_, ok := transport.ConnectionState()
	if ok {
		t.Error("Unconnected TLS transport should have no connection state")
	}
}

// TestTCPTransportSetTimeout tests timeout setting.
func TestTCPTransportSetTimeout(t *testing.T) {
	transport := NewTCPTransport(nil)
	
	transport.SetTimeout(10000)
	transport.SetTimeout(0)
	transport.SetTimeout(-1)
}

// TestTLSTransportSetTimeout tests TLS timeout setting.
func TestTLSTransportSetTimeout(t *testing.T) {
	transport, err := NewTLSTransport(nil)
	if err != nil {
		t.Fatalf("NewTLSTransport failed: %v", err)
	}
	
	transport.SetTimeout(10000)
	transport.SetTimeout(0)
}

// TestTCPClientServer is a placeholder for TCP client/server integration tests.
// Real integration tests would require proper port management and network setup.
func TestTCPClientServer(t *testing.T) {
	// Skip in unit tests - would require network infrastructure
	t.Skip("Integration test - run manually with network setup")
}
