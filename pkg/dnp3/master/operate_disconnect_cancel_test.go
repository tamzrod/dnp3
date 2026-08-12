package master

import (
	"context"
	"errors"
	"testing"
	"time"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// TestOperateAlreadyCancelledContext asserts that an already-cancelled context
// causes Operate to return promptly with ErrContextCanceled and no response.
func TestOperateAlreadyCancelledContext(t *testing.T) {
	tr := newBlockingReceiveTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	resp, err := cc.Operate(ctx, &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Operate error = %v, want ErrContextCanceled", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response on cancel, got %+v", resp)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Operate did not return promptly: %v", elapsed)
	}
}

// TestOperateCancelledMidRequest asserts that cancelling the context while an
// operate request is outstanding causes Operate to return promptly with
// ErrContextCanceled and no response.
func TestOperateCancelledMidRequest(t *testing.T) {
	tr := newBlockingReceiveTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	type res struct {
		resp *OperateResponse
		err  error
	}
	errCh := make(chan res, 1)
	go func() {
		resp, err := cc.Operate(ctx, &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		errCh <- res{resp, err}
	}()

	// Wait for the request to be sent so the operate is genuinely outstanding.
	select {
	case <-tr.sent:
	case <-time.After(time.Second):
		t.Fatal("request was never sent")
	}

	start := time.Now()
	cancel()
	r := <-errCh
	elapsed := time.Since(start)

	if !errors.Is(r.err, dnp3.ErrContextCanceled) {
		t.Fatalf("Operate error = %v, want ErrContextCanceled", r.err)
	}
	if r.resp != nil {
		t.Fatalf("expected nil response on cancel; got %+v", r.resp)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Operate did not return promptly after cancel: %v", elapsed)
	}
}

// slowCloseTransport is a transport.Handler whose Close blocks until an
// external caller closes releaseCh, simulating a slow teardown. Reads from a
// closed channel return immediately, so a repeated Close after release is a
// no-op.
type slowCloseTransport struct {
	releaseCh chan struct{}
}

func newSlowCloseTransport() *slowCloseTransport {
	return &slowCloseTransport{releaseCh: make(chan struct{})}
}

func (t *slowCloseTransport) Listen() error                   { return nil }
func (t *slowCloseTransport) Accept() error                   { return nil }
func (t *slowCloseTransport) Connect() error                  { return nil }
func (t *slowCloseTransport) Send(data []byte) error          { return nil }
func (t *slowCloseTransport) Receive() ([]byte, error)      { return nil, errors.New("closed") }
func (t *slowCloseTransport) SetTimeout(ms int)              {}
func (t *slowCloseTransport) Close() error {
	<-t.releaseCh
	return nil
}

// TestDisconnectAlreadyCancelledContext asserts that an already-cancelled
// context causes Disconnect to return promptly with ErrContextCanceled and
// still resets the client state to Disconnected.
func TestDisconnectAlreadyCancelledContext(t *testing.T) {
	tr := newSlowCloseTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err = cc.Disconnect(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Disconnect error = %v, want ErrContextCanceled", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Disconnect did not return promptly: %v", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}
	// Release the background Close so the test goroutine does not leak.
	close(tr.releaseCh)
}

// TestDisconnectCancelledMidTeardown asserts that cancelling while the
// transport Close is in progress causes Disconnect to return promptly with
// ErrContextCanceled and resets state to Disconnected.
func TestDisconnectCancelledMidTeardown(t *testing.T) {
	tr := newSlowCloseTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- cc.Disconnect(ctx) }()

	// Give the teardown a moment to enter the blocking Close.
	time.Sleep(20 * time.Millisecond)

	start := time.Now()
	cancel()
	err = <-errCh
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Disconnect error = %v, want ErrContextCanceled", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Disconnect did not return promptly after cancel: %v", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}
	close(tr.releaseCh)
}
