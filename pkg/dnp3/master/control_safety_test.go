package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// TestOperateCancelQuarantinesOutstation (VEC-01 B4 F3) asserts that cancelling
// the context while a non-idempotent DirectOperate is outstanding returns
// promptly with ErrContextCanceled AND fails closed by quarantining the
// outstation, because the control may already have been transmitted. The
// quarantine must refuse a subsequent control before any wire traffic.
func TestOperateCancelQuarantinesOutstation(t *testing.T) {
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
	errCh := make(chan error, 1)
	go func() {
		_, err := cc.Operate(ctx, &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		errCh <- err
	}()

	// Wait for the request to be sent so the control is genuinely outstanding.
	select {
	case <-tr.sent:
	case <-time.After(time.Second):
		t.Fatal("request was never sent")
	}

	start := time.Now()
	cancel()
	err = <-errCh
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Operate did not return promptly after cancel: %v", elapsed)
	}

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Operate error = %v, want ErrContextCanceled", err)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("outstation not quarantined after cancelled non-idempotent control")
	}

	// A subsequent control must be refused by the quarantine before any send.
	_, err2 := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if !errors.Is(err2, master.ErrControlQuarantined) {
		t.Fatalf("second Operate error = %v, want ErrControlQuarantined", err2)
	}
}

// sleepReceiveTransport is a transport.Handler whose Receive sleeps for the
// most recently configured timeout (SetTimeout) before failing. It models a
// peer that never answers the link handshake, so the connect budget must bound
// the whole handshake rather than each blocking wait.
type sleepReceiveTransport struct {
	mu        sync.Mutex
	timeoutMs int
}

func (t *sleepReceiveTransport) Listen() error          { return nil }
func (t *sleepReceiveTransport) Accept() error          { return nil }
func (t *sleepReceiveTransport) Connect() error         { return nil }
func (t *sleepReceiveTransport) Close() error           { return nil }
func (t *sleepReceiveTransport) Send(data []byte) error { return nil }
func (t *sleepReceiveTransport) SetTimeout(ms int) {
	t.mu.Lock()
	t.timeoutMs = ms
	t.mu.Unlock()
}
func (t *sleepReceiveTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	ms := t.timeoutMs
	t.mu.Unlock()
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	return nil, errors.New("no handshake response")
}

// TestConnectContextRespectsBudget (VEC-01 B4 F2) asserts that a connect whose
// link handshake never completes fails closed near the budget with
// ErrOperationDeadlineExceeded instead of hanging for the full per-step timeout,
// and leaves the master in a non-active (failed) state.
func TestConnectContextRespectsBudget(t *testing.T) {
	tr := &sleepReceiveTransport{}
	cc := clientWithTransport(t, tr)
	// Register an outstation so the link handshake (Reset Link Stations +
	// Request Link Status) actually runs and blocks on the unresponsive peer.
	cc.internalMaster.AddOutstation(1024, "test")

	const budget = 150 * time.Millisecond
	start := time.Now()
	err := cc.internalMaster.ConnectContext(context.Background(), budget)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected connect to fail closed when the budget is exceeded")
	}
	if !errors.Is(err, master.ErrOperationDeadlineExceeded) {
		t.Fatalf("error = %v, want ErrOperationDeadlineExceeded", err)
	}
	if elapsed > time.Second {
		t.Fatalf("ConnectContext took %v, want it bounded near the %v budget", elapsed, budget)
	}
	if cc.internalMaster.State() == master.StateActive {
		t.Fatal("a failed connect must not report an active session")
	}
}

// TestClientConnectRespectsConfiguredBudget (VEC-01 B4 F2) proves the public
// ConnectTimeout config is plumbed through to the link-handshake budget: a
// connect whose handshake never completes fails closed near the configured
// budget rather than hanging for the full per-step timeout.
func TestClientConnectRespectsConfiguredBudget(t *testing.T) {
	tr := &sleepReceiveTransport{}
	c, err := NewClient(NewConfig(WithConnectTimeout(150 * time.Millisecond)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})

	start := time.Now()
	err = cc.Connect(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected connect to fail closed when the budget is exceeded")
	}
	if elapsed > time.Second {
		t.Fatalf("Connect took %v, want it bounded near the 150ms budget", elapsed)
	}
}

// TestIntegrityPollSharesBudget (VEC-01 B4 F2) proves the whole integrity poll
// — primary multi-group read plus the per-group fallback — shares ONE finite
// budget instead of resetting it per exchange. With a 150ms read budget and an
// unresponsive peer, the poll must return near 150ms, not ~4x that.
func TestIntegrityPollSharesBudget(t *testing.T) {
	tr := &sleepReceiveTransport{}
	c, err := NewClient(NewConfig(WithReadTimeout(150 * time.Millisecond)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.internalMaster.SetState(master.StateActive)
	cc.internalMaster.AddOutstation(cc.config.OutstationAddress, "test")
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	start := time.Now()
	_, err = cc.IntegrityPoll(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected the integrity poll to fail against an unresponsive peer")
	}
	if elapsed > 400*time.Millisecond {
		t.Fatalf("integrity poll took %v; the per-group fallback must share the 150ms budget", elapsed)
	}
}
