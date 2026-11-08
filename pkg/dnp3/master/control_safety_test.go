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

// pubControlTransport is a scriptable transport.Handler for public Operate
// tests: it counts Sends and returns either a valid G12V1 status response or a
// missing-response error. It opens no socket, so tests stay offline.
type pubControlTransport struct {
	mu      sync.Mutex
	sends   int
	lastSeq uint8
	status  byte
	respond bool
}

func (t *pubControlTransport) Listen() error     { return nil }
func (t *pubControlTransport) Accept() error     { return nil }
func (t *pubControlTransport) Connect() error    { return nil }
func (t *pubControlTransport) Close() error      { return nil }
func (t *pubControlTransport) SetTimeout(ms int) {}

func (t *pubControlTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.lastSeq = extractPubRequestSeq(data)
	t.mu.Unlock()
	return nil
}

func (t *pubControlTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	seq, st, respond := t.lastSeq, t.status, t.respond
	t.mu.Unlock()
	if !respond {
		return nil, errors.New("no response")
	}
	return buildPubG12V1StatusResponse(seq, st), nil
}

func (t *pubControlTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

func connectedOperateClient(t *testing.T, tr *pubControlTransport) *client {
	t.Helper()
	cc := clientWithTransport(t, tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()
	return cc
}

// TestOperateSingleShotUnderDefaultRetries (VEC-01 B4 F1) proves the PUBLIC
// Operate path is single-shot even under the default retry policy
// (RetryCount=3): a control whose response is missing is transmitted exactly
// once and fails closed as UNKNOWN with the outstation quarantined.
func TestOperateSingleShotUnderDefaultRetries(t *testing.T) {
	tr := &pubControlTransport{respond: false}
	cc := connectedOperateClient(t, tr)
	if p := cc.internalMaster.RetryPolicy(); p == nil || p.TimeoutRetries < 1 {
		t.Fatalf("test requires a retrying policy; got %+v", p)
	}

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil {
		t.Fatal("expected fail-closed error for a missing response")
	}
	if !errors.Is(err, master.ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("public control sent %d times under default retries, want exactly 1", got)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("outstation not quarantined after unknown public control outcome")
	}
}

// TestOperateAcknowledgedSingleShotNotQuarantined (VEC-01 B4 F1/F3) proves a
// positively acknowledged public control is transmitted exactly once, returns a
// response, and does NOT quarantine the outstation.
func TestOperateAcknowledgedSingleShotNotQuarantined(t *testing.T) {
	tr := &pubControlTransport{respond: true, status: 0}
	cc := connectedOperateClient(t, tr)

	resp, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("acknowledged public control returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("acknowledged public control returned nil response")
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("acknowledged public control sent %d times, want exactly 1", got)
	}
	if cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("a positively acknowledged control must not quarantine the outstation")
	}
}

// TestOperateSingleShotUnderWithRetryZero (VEC-01 B4 F1) proves the public
// control is single-shot under an explicit WithRetry(0, ...) configuration
// (zero is coerced to the default retry policy, so it must not be relied upon
// as a one-shot fix): still exactly one send, still fail-closed + quarantined.
func TestOperateSingleShotUnderWithRetryZero(t *testing.T) {
	tr := &pubControlTransport{respond: false}
	c, err := NewClient(NewConfig(WithRetry(0, 0)))
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

	_, err = cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if !errors.Is(err, master.ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control under WithRetry(0) sent %d times, want exactly 1", got)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("outstation not quarantined under WithRetry(0) unknown outcome")
	}
}

// malformedControlTransport returns bytes that are not a decodable DNP3 frame,
// modelling a malformed/garbled control response. The control must fail closed
// as UNKNOWN and quarantine the outstation (never a fabricated success).
type malformedControlTransport struct {
	mu    sync.Mutex
	sends int
}

func (t *malformedControlTransport) Listen() error     { return nil }
func (t *malformedControlTransport) Accept() error     { return nil }
func (t *malformedControlTransport) Connect() error    { return nil }
func (t *malformedControlTransport) Close() error      { return nil }
func (t *malformedControlTransport) SetTimeout(ms int) {}
func (t *malformedControlTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return nil
}
func (t *malformedControlTransport) Receive() ([]byte, error) {
	return []byte{0xDE, 0xAD, 0xBE, 0xEF}, nil
}
func (t *malformedControlTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestOperateMalformedResponseFailsClosed (VEC-01 B4 F1/F3) proves a malformed
// control response fails closed as UNKNOWN, is transmitted exactly once, and
// quarantines the outstation.
func TestOperateMalformedResponseFailsClosed(t *testing.T) {
	tr := &malformedControlTransport{}
	cc := clientWithTransport(t, tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil {
		t.Fatal("malformed control response must not be reported as success")
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control sent %d times, want exactly 1", got)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("outstation not quarantined after malformed control response")
	}
}
