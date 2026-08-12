package outstation

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
)

// TestMaxConnectionsDefaultIsOne verifies DNP3-084: the v0 MVP profile is
// single-master by default.
func TestMaxConnectionsDefaultIsOne(t *testing.T) {
	cfg := NewConfig()
	if cfg.MaxConnections != 1 {
		t.Fatalf("default MaxConnections = %d, want 1 (MVP single-master)", cfg.MaxConnections)
	}
}

// TestWithMaxConnections verifies the functional option applies and validation
// rejects non-positive values.
func TestWithMaxConnections(t *testing.T) {
	cfg := NewConfig(WithMaxConnections(3))
	if cfg.MaxConnections != 3 {
		t.Fatalf("MaxConnections = %d, want 3", cfg.MaxConnections)
	}
	bad := NewConfig(WithMaxConnections(0))
	if err := bad.Validate(); err == nil {
		t.Fatal("expected Validate error for MaxConnections < 1")
	}
}

// TestSecondConnectionRejected verifies DNP3-084: with MaxConnections=1 (MVP
// single), a second concurrent connection is rejected (closed by the server),
// and ActiveConnections never exceeds the limit.
func TestSecondConnectionRejected(t *testing.T) {
	cfg := NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0), // ephemeral port
		WithMaxConnections(1),
	)
	srvI, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv := srvI.(*server)
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		_ = srv.Stop(ctx)
	}()

	addr := srv.listenerAddr()
	if addr == nil {
		t.Fatal("listener not bound")
	}

	// First connection: accepted.
	c1, err := net.DialTimeout("tcp", addr.String(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial first: %v", err)
	}
	defer c1.Close()

	// Wait until the server registers the first connection.
	if !waitFor(time.Second, func() bool { return srv.ActiveConnections() == 1 }) {
		t.Fatalf("first connection not accepted: ActiveConnections=%d", srv.ActiveConnections())
	}

	// Second connection: rejected. Dial succeeds (TCP accept), then the server
	// closes it immediately. A subsequent Read must surface an EOF/error.
	c2, err := net.DialTimeout("tcp", addr.String(), 2*time.Second)
	if err != nil {
		// If the second dial itself fails (e.g. backlog/accept rejection),
		// that also satisfies "rejected" — but we still assert the limit held.
		t.Logf("second dial failed (acceptable rejection): %v", err)
	} else {
		defer c2.Close()
		c2.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 16)
		if _, rerr := c2.Read(buf); rerr == nil {
			t.Fatal("second connection read succeeded — expected server to close it")
		} else if rerr != io.EOF {
			t.Logf("second connection read returned %v (acceptable)", rerr)
		}
	}

	// The active connection count must never exceed MaxConnections.
	if n := srv.ActiveConnections(); n > 1 {
		t.Fatalf("ActiveConnections=%d, must not exceed MaxConnections=1", n)
	}

	// The first connection is still active (the rejection did not drop it).
	if n := srv.ActiveConnections(); n != 1 {
		t.Fatalf("ActiveConnections=%d after rejection, want 1", n)
	}
}

// waitFor polls cond every 5ms up to timeout, returning true once cond() is true.
func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}
