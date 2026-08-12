package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/transport"
)

// blockingTransport is a transport.Handler whose Connect blocks until releaseCh
// is closed (simulating a slow/hanging dial). All other methods are no-ops. It
// records whether Close was called so tests can assert no live connection
// survives a cancelled Connect. Close is idempotent (real transports tolerate
// being closed twice during cancel teardown).
type blockingTransport struct {
	connectCalled chan struct{}
	releaseCh     chan struct{}
	closeOnce     sync.Once
	closed        bool
}

func newBlockingTransport() *blockingTransport {
	return &blockingTransport{
		connectCalled: make(chan struct{}, 1),
		releaseCh:     make(chan struct{}),
	}
}

func (t *blockingTransport) Listen() error                   { return nil }
func (t *blockingTransport) Accept() error                   { return nil }
func (t *blockingTransport) Send(data []byte) error         { return nil }
func (t *blockingTransport) Receive() ([]byte, error)      { return nil, errors.New("closed") }
func (t *blockingTransport) SetTimeout(ms int)              {}
func (t *blockingTransport) Connect() error {
	select {
	case t.connectCalled <- struct{}{}:
	default:
	}
	<-t.releaseCh // block until released by Close or test teardown
	return errors.New("connect aborted")
}
func (t *blockingTransport) Close() error {
	t.closeOnce.Do(func() {
		t.closed = true
		close(t.releaseCh)
	})
	return nil
}

// clientWithTransport builds a public client and substitutes its transport with
// the given handler, re-wiring the internal master's transport adapter.
func clientWithTransport(t *testing.T, tr transport.Handler) *client {
	t.Helper()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	return cc
}

// TestConnectAlreadyCancelledContext asserts that an already-cancelled context
// causes Connect to return promptly with ErrContextCanceled and leave the
// client disconnected (no live connection).
func TestConnectAlreadyCancelledContext(t *testing.T) {
	bt := newBlockingTransport()
	cc := clientWithTransport(t, bt)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := cc.Connect(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Connect error = %v, want ErrContextCanceled", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Connect did not return promptly: %v", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}
	if bt.closed {
		t.Fatalf("blocking transport should not have been Close()d for a pre-connect cancel")
	}
}

// TestConnectCancelledMidDial asserts that cancelling the context while the
// transport dial is in progress causes Connect to return promptly with
// ErrContextCanceled and close the transport (no live connection).
func TestConnectCancelledMidDial(t *testing.T) {
	bt := newBlockingTransport()
	cc := clientWithTransport(t, bt)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- cc.Connect(ctx) }()

	// Wait for the blocking dial to start.
	select {
	case <-bt.connectCalled:
	case <-time.After(time.Second):
		t.Fatal("transport.Connect was never called")
	}

	start := time.Now()
	cancel()
	err := <-errCh
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Connect error = %v, want ErrContextCanceled", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Connect did not return promptly after cancel: %v", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected (no live connection)", cc.State())
	}
	if !bt.closed {
		t.Fatalf("transport was not Close()d; a live connection may remain")
	}
}
