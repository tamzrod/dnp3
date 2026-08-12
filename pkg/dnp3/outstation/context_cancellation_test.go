package outstation

import (
	"context"
	"net"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
)

// TestStartContextCancellationProducesCleanStop verifies DNP3-085: cancelling
// the caller's Start(ctx) context produces a clean stop — the listener is
// closed, the server state transitions to Down, and the accept loop exits.
func TestStartContextCancellationProducesCleanStop(t *testing.T) {
	cfg := NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0), // ephemeral port
	)
	srvI, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv := srvI.(*server)

	ctx, cancel := context.WithCancel(context.Background())
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	addr := srv.listenerAddr()
	if addr == nil {
		t.Fatal("listener not bound")
	}

	// Cancel the Start context → must produce a clean stop.
	cancel()
	if !waitFor(2*time.Second, func() bool { return srv.State() == ServerStateDown }) {
		t.Fatalf("server did not reach Down after ctx cancel: state=%s", srv.State())
	}

	// Listener must be closed: a fresh dial must fail (or be refused).
	// Allow a brief window for the OS to reflect the closed listener.
	deadline := time.Now().Add(2 * time.Second)
	var dialErr error
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr.String(), 200*time.Millisecond)
		if err != nil {
			dialErr = err
			break
		}
		c.Close()
		time.Sleep(20 * time.Millisecond)
	}
	if dialErr == nil {
		t.Fatalf("listener still accepting after ctx cancel (not a clean stop)")
	}
}

// TestStartAlreadyCancelledContextReturnsError verifies DNP3-085: Start with
// an already-cancelled context fails fast rather than starting a doomed loop.
func TestStartAlreadyCancelledContextReturnsError(t *testing.T) {
	cfg := NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0),
	)
	srvI, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv := srvI.(*server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := srv.Start(ctx); err == nil {
		t.Fatal("expected Start to fail on already-cancelled context")
	}
	if got := srv.State(); got != ServerStateDown {
		t.Fatalf("state after failed start = %s, want Down", got)
	}
}

// TestStopWaitsForAcceptLoopExit verifies DNP3-085: Stop returns only after the
// accept loop has fully exited (acceptDone closed), ensuring a clean stop.
func TestStopWaitsForAcceptLoopExit(t *testing.T) {
	cfg := NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0),
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

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer stopCancel()
	if err := srv.Stop(stopCtx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if got := srv.State(); got != ServerStateDown {
		t.Fatalf("state after Stop = %s, want Down", got)
	}
	// acceptDone must be closed after Stop returns.
	srv.mu.RLock()
	ch := srv.acceptDone
	srv.mu.RUnlock()
	if ch == nil {
		t.Fatal("acceptDone channel nil after Stop")
	}
	select {
	case <-ch:
	default:
		t.Fatal("acceptDone not closed after Stop (accept loop still running)")
	}
}
