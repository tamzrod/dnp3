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

// countingBlockingTransport records each Send and blocks Receive until
// releaseCh is closed. It models a control whose response never arrives and
// lets a test count transmissions across a cancellation and late completion.
type countingBlockingTransport struct {
	mu        sync.Mutex
	sends     int
	sent      chan struct{}
	releaseCh chan struct{}
	closeOnce sync.Once
}

func newCountingBlockingTransport() *countingBlockingTransport {
	return &countingBlockingTransport{
		sent:      make(chan struct{}, 1),
		releaseCh: make(chan struct{}),
	}
}

func (t *countingBlockingTransport) Listen() error  { return nil }
func (t *countingBlockingTransport) Accept() error  { return nil }
func (t *countingBlockingTransport) Connect() error { return nil }
func (t *countingBlockingTransport) Close() error {
	t.closeOnce.Do(func() { close(t.releaseCh) })
	return nil
}
func (t *countingBlockingTransport) SetTimeout(ms int) {}
func (t *countingBlockingTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	select {
	case t.sent <- struct{}{}:
	default:
	}
	return nil
}
func (t *countingBlockingTransport) Receive() ([]byte, error) {
	<-t.releaseCh
	return nil, errors.New("receive aborted")
}
func (t *countingBlockingTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestOperateUnknownPersistsAcrossDisconnectAndLateCompletion (VEC-01 B4 F2/F3)
// proves an UNKNOWN control outcome is not silently cleared: after cancellation
// quarantines the outstation, a late worker completion and an explicit client
// Disconnect both leave the quarantine in place, and a further control is
// refused before any transmission.
func TestOperateUnknownPersistsAcrossDisconnectAndLateCompletion(t *testing.T) {
	tr := newCountingBlockingTransport()
	cc := clientWithTransport(t, tr)
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

	select {
	case <-tr.sent:
	case <-time.After(time.Second):
		t.Fatal("control request was never sent")
	}
	cancel()
	if err := <-errCh; !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Operate error = %v, want ErrContextCanceled", err)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("outstation not quarantined after cancelled control")
	}

	// Let the late worker complete (Receive returns) and give it a moment.
	_ = tr.Close()
	time.Sleep(50 * time.Millisecond)
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("late worker completion must not clear the UNKNOWN quarantine")
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control transmitted %d times; a late worker must not resend", got)
	}

	// Explicit client Disconnect must NOT clear an UNKNOWN quarantine.
	if err := cc.Disconnect(context.Background()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("Disconnect must not clear the UNKNOWN quarantine")
	}

	// A further control is refused before any transmission.
	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if !errors.Is(err, master.ErrControlQuarantined) {
		t.Fatalf("control after Disconnect = %v, want ErrControlQuarantined", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("refused control was transmitted: sends = %d, want 1", got)
	}
}

// slowDialStallTransport's Connect blocks for dialDelay then succeeds; Receive
// then sleeps for the configured timeout and fails, stalling the link handshake.
// It proves the public Connect shares ONE budget across the dial and the
// handshake instead of granting the handshake a fresh full budget after the
// dial.
type slowDialStallTransport struct {
	mu        sync.Mutex
	dialDelay time.Duration
	timeoutMs int
}

func (t *slowDialStallTransport) Listen() error { return nil }
func (t *slowDialStallTransport) Accept() error { return nil }
func (t *slowDialStallTransport) Close() error  { return nil }
func (t *slowDialStallTransport) Send(data []byte) error {
	return nil
}
func (t *slowDialStallTransport) SetTimeout(ms int) {
	t.mu.Lock()
	t.timeoutMs = ms
	t.mu.Unlock()
}
func (t *slowDialStallTransport) Connect() error {
	time.Sleep(t.dialDelay)
	return nil
}
func (t *slowDialStallTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	ms := t.timeoutMs
	t.mu.Unlock()
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	return nil, errors.New("no handshake response")
}

// TestConnectSlowDialSharesOneBudget (VEC-01 B4 F2) proves the dial and the
// link handshake share ONE whole-connect budget: with a 200ms budget, a 120ms
// dial leaves only ~80ms for the handshake, so Connect fails near 200ms rather
// than dial + a fresh full handshake budget.
func TestConnectSlowDialSharesOneBudget(t *testing.T) {
	tr := &slowDialStallTransport{dialDelay: 120 * time.Millisecond}
	c, err := NewClient(NewConfig(WithConnectTimeout(200 * time.Millisecond)))
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
	if elapsed > 600*time.Millisecond {
		t.Fatalf("Connect took %v; the dial and handshake must share the 200ms budget", elapsed)
	}
}

// lateDialTransport blocks in Connect until releaseCh is closed, then reports a
// successful dial; Receive then stalls. It models an uncooperative transport
// whose dial exceeds the whole-connect budget and only completes afterwards, so
// a test can prove the public Connect fails timely under a deadline even with
// context.Background() and that the late worker cannot commit Active.
type lateDialTransport struct {
	entered   chan struct{}
	releaseCh chan struct{}
	once      sync.Once
}

func newLateDialTransport() *lateDialTransport {
	return &lateDialTransport{entered: make(chan struct{}), releaseCh: make(chan struct{})}
}
func (t *lateDialTransport) Listen() error          { return nil }
func (t *lateDialTransport) Accept() error          { return nil }
func (t *lateDialTransport) Close() error           { return nil }
func (t *lateDialTransport) Send(data []byte) error { return nil }
func (t *lateDialTransport) SetTimeout(ms int)      {}
func (t *lateDialTransport) Connect() error {
	t.once.Do(func() { close(t.entered) })
	<-t.releaseCh
	return nil
}
func (t *lateDialTransport) Receive() ([]byte, error) {
	time.Sleep(50 * time.Millisecond)
	return nil, errors.New("no handshake response")
}
func (t *lateDialTransport) release() { close(t.releaseCh) }

// TestConnectDeadlineBoundsBlockingDialNoContext (VEC-01 B4 F2-R2) proves the
// public Connect is bounded by the whole-connect deadline even with
// context.Background() and an uncooperative dial that blocks past the budget:
// it returns near the budget with a timeout sentinel rather than hanging, and
// leaves the client Disconnected. A late worker completion must not commit
// Active.
func TestConnectDeadlineBoundsBlockingDialNoContext(t *testing.T) {
	tr := newLateDialTransport()
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
	if !errors.Is(err, dnp3.ErrTimeout) {
		t.Fatalf("Connect error = %v, want ErrTimeout (budget exhausted)", err)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("Connect took %v; a blocking dial must be bounded by the 150ms budget", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected after a bounded connect", cc.State())
	}

	// Let the late worker finish; it must not commit Active.
	tr.release()
	time.Sleep(100 * time.Millisecond)
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v after late worker completion, want Disconnected (fenced)", cc.State())
	}
}

// TestConnectRefusedWhileSupersededWorkerOutstanding (VEC-01 B4 F2-R3) proves
// the public Connect refuses a replacement attempt while a superseded connect
// worker has not terminated. A late-dial transport blocks past the whole-connect
// budget; the first Connect times out and returns promptly, but the worker keeps
// holding the shared transport, so a fresh Connect is refused fail-closed with
// ErrSessionUnresolved until the worker terminates. Once released, a later
// Connect may proceed.
func TestConnectRefusedWhileSupersededWorkerOutstanding(t *testing.T) {
	tr := newLateDialTransport()
	c, err := NewClient(NewConfig(WithConnectTimeout(100 * time.Millisecond)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})

	// Attempt 1: the dial blocks past the budget, so Connect returns promptly.
	start := time.Now()
	err = cc.Connect(context.Background())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected Connect to fail closed when the budget is exceeded")
	}
	if !errors.Is(err, dnp3.ErrTimeout) {
		t.Fatalf("Connect error = %v, want ErrTimeout", err)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("Connect took %v; a blocking dial must be bounded by the 100ms budget", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected after a bounded connect", cc.State())
	}

	// Attempt 2 while the superseded worker is still blocked in its dial: refused
	// before dialing because the old worker may still hold the shared transport.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect while worker outstanding = %v, want ErrSessionUnresolved", err)
	}

	// Release the late dial so the worker terminates; the slot frees up.
	tr.release()
	deadline := time.Now().Add(time.Second)
	for {
		cc.mu.RLock()
		outstanding := cc.connectWorkerOutstanding
		cc.mu.RUnlock()
		if !outstanding || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if err := cc.Connect(context.Background()); errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect after worker termination = %v, want it to proceed past the fence", err)
	}
}

// pubBlockingSendTransport blocks inside Send until releaseCh is closed,
// modelling a transport whose send cannot be aborted. It records that a send was
// entered so a test can prove the control worker is blocked mid-send. It opens no
// socket, so the test stays offline.
type pubBlockingSendTransport struct {
	mu        sync.Mutex
	sends     int
	entered   chan struct{}
	releaseCh chan struct{}
	once      sync.Once
}

func newPubBlockingSendTransport() *pubBlockingSendTransport {
	return &pubBlockingSendTransport{entered: make(chan struct{}), releaseCh: make(chan struct{})}
}

func (t *pubBlockingSendTransport) Listen() error     { return nil }
func (t *pubBlockingSendTransport) Accept() error     { return nil }
func (t *pubBlockingSendTransport) Connect() error    { return nil }
func (t *pubBlockingSendTransport) Close() error      { return nil }
func (t *pubBlockingSendTransport) SetTimeout(ms int) {}
func (t *pubBlockingSendTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	t.once.Do(func() { close(t.entered) })
	<-t.releaseCh
	return nil
}
func (t *pubBlockingSendTransport) Receive() ([]byte, error) { return nil, errors.New("no response") }
func (t *pubBlockingSendTransport) release()                 { close(t.releaseCh) }

// TestConnectRefusedWhileControlSendUnresolved (VEC-01 B4 F3-R3) proves the
// public Connect refuses a fresh attempt while an admitted control send from a
// torn-down session remains outstanding (its physical Send may still be on the
// wire on the original transport). This closes the F3-R3 gap at the public
// Connect entry, before any dial.
func TestConnectRefusedWhileControlSendUnresolved(t *testing.T) {
	tr := newPubBlockingSendTransport()
	cc := clientWithTransport(t, tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	// Start a control that blocks inside Send.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		_, err := cc.Operate(ctx, &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		errCh <- err
	}()
	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	// Tear the session down while the send is in flight: the outstation is now
	// fenced as unresolved.
	if err := cc.internalMaster.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	cc.mu.Lock()
	cc.state = dnp3.StateDisconnected
	cc.mu.Unlock()

	if !cc.internalMaster.HasUnresolvedSession() {
		t.Fatal("a Disconnect during an in-flight control send must fence the session unresolved")
	}

	// A fresh public Connect is refused before dialing.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect while control send unresolved = %v, want ErrSessionUnresolved", err)
	}

	// Release the blocked send so the worker terminates.
	tr.release()
	if err := <-errCh; err == nil {
		t.Fatal("the interrupted control should not report success")
	}
	deadline := time.Now().Add(time.Second)
	for cc.internalMaster.HasUnresolvedSession() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if cc.internalMaster.HasUnresolvedSession() {
		t.Fatal("the unresolved fence must clear once the control worker terminates")
	}
}
