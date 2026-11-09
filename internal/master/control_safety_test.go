package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"io"

	"dnp3/internal/al"
)

// These tests lock the VEC-01 B4 bounded offline Go control safety correction.
// They use deterministic transport doubles only (no sockets/listeners), so they
// run entirely offline. Each double records the number of Send calls so a test
// can prove a non-idempotent control was transmitted AT MOST ONCE (F1).

// failReceiveTransport accepts Send (recording it) and fails Receive
// immediately with a fixed error. It models a control whose response is
// missing, so the remote outcome cannot be positively established.
type failReceiveTransport struct {
	mu    sync.Mutex
	sends int
	err   error
}

func (t *failReceiveTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return nil
}

func (t *failReceiveTransport) SetTimeout(ms int) {}

func (t *failReceiveTransport) Receive() ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}
	return nil, errors.New("no response")
}

func (t *failReceiveTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// timeoutTransport accepts Send (recording it) and, on Receive, sleeps for the
// configured transport timeout before returning a timeout error. It models the
// blocking response wait the operation deadline must bound (F2).
type timeoutTransport struct {
	mu        sync.Mutex
	sends     int
	timeoutMs int
}

func (t *timeoutTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return nil
}

func (t *timeoutTransport) SetTimeout(ms int) {
	t.mu.Lock()
	t.timeoutMs = ms
	t.mu.Unlock()
}

func (t *timeoutTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	ms := t.timeoutMs
	t.mu.Unlock()
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	return nil, errors.New("timeout")
}

func (t *timeoutTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

func newControlMaster(t *testing.T, tr TransportHandler) *Master {
	t.Helper()
	m := NewMaster(DefaultConfig())
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	return m
}

// TestControlSingleShotNoRetryOnMissingResponse (F1) proves a DirectOperate
// whose response never arrives is transmitted exactly ONCE — there is no hidden
// retry of a non-idempotent control — and fails closed as UNKNOWN.
func TestControlSingleShotNoRetryOnMissingResponse(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err == nil {
		t.Fatal("expected fail-closed error for a control with no response")
	}
	if !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control sent %d times, want exactly 1 (no autonomous retry)", got)
	}
	if !m.IsQuarantined(1) {
		t.Fatal("outstation not quarantined after unknown control outcome")
	}
}

// TestControlSingleShotNoRetryOnDisconnect (F1/F3) proves a control whose link
// drops after the send is sent exactly once and fails closed as UNKNOWN.
func TestControlSingleShotNoRetryOnDisconnect(t *testing.T) {
	tr := &failReceiveTransport{err: io.EOF}
	m := newControlMaster(t, tr)

	_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err == nil {
		t.Fatal("expected fail-closed error on disconnect")
	}
	if !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control sent %d times, want exactly 1", got)
	}
	if !m.IsQuarantined(1) {
		t.Fatal("outstation not quarantined after disconnect during control")
	}
}

// TestControlDeadlineBoundsBlockingWait (F2) proves the whole control completes
// within its finite budget instead of overshooting to the (much larger) default
// response timeout. The transport honors the capping the master applies.
func TestControlDeadlineBoundsBlockingWait(t *testing.T) {
	tr := &timeoutTransport{}
	m := newControlMaster(t, tr)

	const budget = 80 * time.Millisecond
	start := time.Now()
	status, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(budget))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected fail-closed error, got status %d", status)
	}
	if !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	// The default response timeout is 5000ms; a bounded control must return far
	// sooner. Allow generous CI slack while still proving the budget is finite.
	if elapsed > time.Second {
		t.Fatalf("control took %v, want it bounded near the %v budget", elapsed, budget)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control sent %d times, want exactly 1", got)
	}
}

// TestControlQuarantineBlocksFurtherControls (F3) proves that once an outcome
// is UNKNOWN, no new control may be issued (fail closed) and that the quarantine
// is PERSISTENT across reconnect and even explicit teardown (VEC-01 B4 F2/F3
// corrective addendum): it is never released automatically.
func TestControlQuarantineBlocksFurtherControls(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("first control error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("after first control, sends = %d, want 1", got)
	}

	// A second control must be refused before any transmission.
	_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if !errors.Is(err, ErrControlQuarantined) {
		t.Fatalf("second control error = %v, want ErrControlQuarantined", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("quarantined control was transmitted: sends = %d, want 1", got)
	}

	// Reconnect (a state reset) must NOT clear the quarantine.
	m.SetState(StateInitialized)
	if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlQuarantined) {
		t.Fatalf("control after reconnect = %v, want ErrControlQuarantined", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control after reconnect was transmitted: sends = %d, want 1", got)
	}

	// Explicit teardown must ALSO not release an UNKNOWN outcome: releasing it
	// without a separately approved reconciliation protocol would silently
	// discard a possibly-executed control.
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !m.IsQuarantined(1) {
		t.Fatal("Disconnect must NOT clear the control quarantine (persistent UNKNOWN)")
	}
	// A control after Disconnect is still refused (fail closed).
	m.SetState(StateInitialized)
	if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlQuarantined) {
		t.Fatalf("control after Disconnect = %v, want ErrControlQuarantined", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control after Disconnect was transmitted: sends = %d, want 1", got)
	}
}

// TestControlPreSendFailureDoesNotQuarantine (F3) proves a failure positively
// established BEFORE any transmission (here: a cancelled context) does not
// quarantine the outstation and sends nothing.
func TestControlPreSendFailureDoesNotQuarantine(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := m.OperateWithStatusContext(ctx, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second))
	if err == nil {
		t.Fatal("expected error for already-cancelled context")
	}
	if !errors.Is(err, ErrOperationDeadlineExceeded) {
		t.Fatalf("error = %v, want ErrOperationDeadlineExceeded", err)
	}
	if got := tr.sendCount(); got != 0 {
		t.Fatalf("nothing should be sent for a cancelled control; sends = %d", got)
	}
	if m.IsQuarantined(1) {
		t.Fatal("a definite pre-send failure must not quarantine the outstation")
	}
}

// TestControlExplicitFailureStatusNotUnknown proves a valid response carrying
// an explicit non-success command status is a positively classified outcome: it
// is returned as-is (not ErrControlOutcomeUnknown) and never quarantines.
func TestControlExplicitFailureStatusNotUnknown(t *testing.T) {
	// statusEchoTransport is defined in command_status_test.go.
	tr := &statusEchoTransport{commandStatus: uint8(CommandStatusNotSupported)}
	m := newControlMaster(t, tr)

	status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err != nil {
		t.Fatalf("explicit failure status returned error: %v", err)
	}
	if status != CommandStatusNotSupported {
		t.Fatalf("status = %d, want CommandStatusNotSupported", status)
	}
	if m.IsQuarantined(1) {
		t.Fatal("an explicit remote failure status must not quarantine")
	}
}

// TestControlSingleShotUnderDefaultRetries (F1) proves the single-shot control
// is not retried even when the configured retry policy allows retries (the
// default MaxRetries=3). A missing response must produce exactly one send.
func TestControlSingleShotUnderDefaultRetries(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)
	// Assert the default policy would permit retries, so the single-send result
	// is attributable to the single-shot control path, not a retry-disabled policy.
	if p := m.RetryPolicy(); p == nil || p.TimeoutRetries < 1 {
		t.Fatalf("test requires a retrying policy; got %+v", p)
	}

	_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("control sent %d times under default retries, want exactly 1", got)
	}
}

// TestReadBoundedByDeadline (F2) proves the read path also honors a finite
// budget: a blocking response wait is bounded by the deadline instead of the
// full per-step timeout, and the whole read stays near the budget.
func TestReadBoundedByDeadline(t *testing.T) {
	tr := &timeoutTransport{}
	m := newControlMaster(t, tr)
	m.SetState(StateActive)
	m.AddOutstation(1, "test")

	const budget = 80 * time.Millisecond
	start := time.Now()
	_, err := m.SendRequestWithRetryAndGetResponseDeadline(al.NewRequest(0, al.FuncRead), 1, time.Now().Add(budget))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a bounded failure for an unresponsive read")
	}
	if elapsed > time.Second {
		t.Fatalf("read took %v, want it bounded near the %v budget", elapsed, budget)
	}
}

// countingStatusTransport echoes a valid G12V1 status response carrying the
// given per-point command status for the request's SEQ, and records how many
// times Send was called. It proves a positively acknowledged control is
// transmitted exactly once and never quarantines.
type countingStatusTransport struct {
	mu      sync.Mutex
	sends   int
	lastSeq uint8
	status  byte
}

func (t *countingStatusTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.lastSeq = extractRequestSeqForBuild(data)
	t.mu.Unlock()
	return nil
}

func (t *countingStatusTransport) SetTimeout(ms int) {}

func (t *countingStatusTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	seq, st := t.lastSeq, t.status
	t.mu.Unlock()
	return buildG12V1StatusResponse(seq, st), nil
}

func (t *countingStatusTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestControlSuccessIsSingleShotNotQuarantined (F1/F3) proves a positively
// acknowledged DirectOperate is transmitted exactly once, returns the real
// success status, and does NOT quarantine the outstation (a known-good outcome
// must not block future controls).
func TestControlSuccessIsSingleShotNotQuarantined(t *testing.T) {
	tr := &countingStatusTransport{status: uint8(CommandStatusSuccess)}
	m := newControlMaster(t, tr)

	status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err != nil {
		t.Fatalf("acknowledged control returned error: %v", err)
	}
	if status != CommandStatusSuccess {
		t.Fatalf("status = %d, want CommandStatusSuccess", status)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("acknowledged control sent %d times, want exactly 1", got)
	}
	if m.IsQuarantined(1) {
		t.Fatal("a positively acknowledged control must not quarantine the outstation")
	}
}

// TestControlNoLateReplayAfterOutcome (F3) proves the single-shot dispatch does
// not replay the control after it has returned an UNKNOWN outcome: the send
// count is stable once OperateWithStatus returns, and the outstation stays
// quarantined rather than retrying in the background.
func TestControlNoLateReplayAfterOutcome(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	base := tr.sendCount()
	if base != 1 {
		t.Fatalf("control sent %d times, want exactly 1", base)
	}
	time.Sleep(50 * time.Millisecond)
	if got := tr.sendCount(); got != base {
		t.Fatalf("control replayed after returning: sends %d -> %d", base, got)
	}
	if !m.IsQuarantined(1) {
		t.Fatal("outstation must remain quarantined after an UNKNOWN outcome")
	}
}

// blockingSendTransport blocks inside Send until releaseCh is closed, modelling
// a transport whose send cannot be aborted (an uncooperative custom transport).
// It records that a send was entered so a test can prove the worker is blocked
// mid-send and that the fail-closed gate stays held.
type blockingSendTransport struct {
	mu        sync.Mutex
	sends     int
	entered   chan struct{}
	releaseCh chan struct{}
	once      sync.Once
}

func newBlockingSendTransport() *blockingSendTransport {
	return &blockingSendTransport{entered: make(chan struct{}), releaseCh: make(chan struct{})}
}

func (t *blockingSendTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	t.once.Do(func() { close(t.entered) })
	<-t.releaseCh
	return nil
}
func (t *blockingSendTransport) SetTimeout(ms int)        {}
func (t *blockingSendTransport) Receive() ([]byte, error) { return nil, errors.New("no response") }
func (t *blockingSendTransport) release()                 { close(t.releaseCh) }
func (t *blockingSendTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestControlWorkerGateHeldWhileBlockedSend (F3) proves the control gate stays
// held while a worker is blocked in an uncooperative send: a second control is
// refused with ErrControlInProgress and no second transmission is attempted.
// When the send finally returns, the worker terminates and the gate is released.
func TestControlWorkerGateHeldWhileBlockedSend(t *testing.T) {
	tr := newBlockingSendTransport()
	m := newControlMaster(t, tr)

	type res struct {
		status CommandStatus
		err    error
	}
	done := make(chan res, 1)
	go func() {
		s, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
		done <- res{s, err}
	}()

	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	// The worker is blocked mid-send: the gate must be held and a second control
	// must be refused with ErrControlInProgress (not sent).
	if !m.HasActiveControl(1) {
		t.Fatal("control gate not held while a worker is blocked in Send")
	}
	if _, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second)); !errors.Is(err, ErrControlInProgress) {
		t.Fatalf("second control while blocked error = %v, want ErrControlInProgress", err)
	}

	// Release the send so the worker can terminate.
	tr.release()
	r := <-done
	if !errors.Is(r.err, ErrControlOutcomeUnknown) {
		t.Fatalf("worker result err = %v, want ErrControlOutcomeUnknown", r.err)
	}
	// After termination the gate is released (but the outstation stays
	// quarantined because the outcome was UNKNOWN).
	deadline := time.Now().Add(time.Second)
	for m.HasActiveControl(1) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if m.HasActiveControl(1) {
		t.Fatal("control gate still held after the worker terminated")
	}
	if !m.IsQuarantined(1) {
		t.Fatal("UNKNOWN outcome must persist as a quarantine after worker termination")
	}
}

// TestControlCancelledBeforeSendNotSent (F3) proves a control cancelled between
// admission and the transmission point is a definite pre-send NOT_SENT failure:
// nothing is transmitted and the outstation is not quarantined.
func TestControlCancelledBeforeSendNotSent(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.OperateWithStatusContext(ctx, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second))
	if err == nil {
		t.Fatal("expected an error for a pre-send cancelled control")
	}
	if !errors.Is(err, ErrOperationDeadlineExceeded) {
		t.Fatalf("error = %v, want ErrOperationDeadlineExceeded (definite pre-send)", err)
	}
	if got := tr.sendCount(); got != 0 {
		t.Fatalf("nothing should be sent for a pre-send cancelled control; sends = %d", got)
	}
	if m.IsQuarantined(1) {
		t.Fatal("a definite pre-send failure must not quarantine the outstation")
	}
}

// TestControlSessionChangeBlocksLateSend (F3) proves a control worker admitted
// in one session cannot transmit after the session generation advances (a
// Disconnect between admission and send): the late send is a definite NOT_SENT
// failure and nothing is transmitted.
func TestControlSessionChangeBlocksLateSend(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	gen, release, err := m.AcquireControlGate(1)
	if err != nil {
		t.Fatalf("AcquireControlGate: %v", err)
	}
	defer release()

	// Tear down the session before the control transmits: the worker's session
	// generation is now stale.
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	m.SetState(StateInitialized)

	_, err = m.OperateWithStatusGated(context.Background(), gen, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second))
	if err == nil {
		t.Fatal("expected a stale-session control to fail closed without sending")
	}
	if got := tr.sendCount(); got != 0 {
		t.Fatalf("a stale-session control must not transmit; sends = %d", got)
	}
	if m.IsQuarantined(1) {
		t.Fatal("a definite pre-send (stale-session) failure must not quarantine")
	}
}

// errorSendTransport fails Send with a non-disconnect error, modelling an
// ambiguous/partial write: a send was attempted but its outcome cannot be
// positively established. The control must fail closed as UNKNOWN and
// quarantine the outstation.
type errorSendTransport struct {
	mu    sync.Mutex
	sends int
}

func (t *errorSendTransport) Send(data []byte) error {
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return errors.New("write failed after partial transmission")
}
func (t *errorSendTransport) SetTimeout(ms int)        {}
func (t *errorSendTransport) Receive() ([]byte, error) { return nil, errors.New("no response") }
func (t *errorSendTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestControlAmbiguousSendFailureQuarantines (F3) proves an ambiguous send
// failure (a send was attempted but the outcome is unknown) fails closed as
// UNKNOWN and quarantines the outstation, without a second transmission.
func TestControlAmbiguousSendFailureQuarantines(t *testing.T) {
	tr := &errorSendTransport{}
	m := newControlMaster(t, tr)

	_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("ambiguous send attempted %d times, want exactly 1", got)
	}
	if !m.IsQuarantined(1) {
		t.Fatal("an ambiguous send failure must quarantine the outstation")
	}
}

// TestControlDisconnectDuringBlockedSendFencesOutstation (F3-R2A/R2B) proves the
// send-admission/teardown linearization: while a worker is blocked mid-Send, a
// Disconnect must (a) fence the outstation as unresolved (the in-flight send is
// ambiguous), (b) quarantine it, and (c) keep a fresh control refused — so no
// late or replacement send can occur. After the worker terminates the unresolved
// fence is lifted, but the quarantine persists.
func TestControlDisconnectDuringBlockedSendFencesOutstation(t *testing.T) {
	tr := newBlockingSendTransport()
	m := newControlMaster(t, tr)

	type res struct {
		status CommandStatus
		err    error
	}
	done := make(chan res, 1)
	go func() {
		s, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
		done <- res{s, err}
	}()

	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	// The send is in flight: tear the session down. The outstation must be fenced
	// as unresolved and quarantined because the send outcome is now ambiguous.
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !m.sessionUnresolvedFor(1) {
		t.Fatal("Disconnect during an in-flight send must fence the outstation as unresolved")
	}
	if !m.IsQuarantined(1) {
		t.Fatal("Disconnect during an in-flight send must quarantine the outstation")
	}
	if m.anySessionUnresolved() == false {
		t.Fatal("a fresh connect must be refused while a worker is unresolved")
	}
	m.SetState(StateInitialized)

	// A fresh control is refused before any transmission (still gated).
	if _, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second)); err == nil {
		t.Fatal("control after teardown-during-send must be refused")
	}
	if got := tr.sendCount(); got != 1 {
		t.Fatalf("no second transmission is allowed; sends = %d, want 1", got)
	}

	// Let the blocked send complete; the worker terminates, lifting the unresolved
	// fence (but never the quarantine).
	tr.release()
	if r := <-done; !errors.Is(r.err, ErrControlOutcomeUnknown) {
		t.Fatalf("worker result err = %v, want ErrControlOutcomeUnknown", r.err)
	}
	deadline := time.Now().Add(time.Second)
	for m.sessionUnresolvedFor(1) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if m.sessionUnresolvedFor(1) {
		t.Fatal("unresolved fence must be lifted once the worker terminates")
	}
	if !m.IsQuarantined(1) {
		t.Fatal("the UNKNOWN quarantine must persist after the worker terminates")
	}
}

// TestControlConnectRefusedWhileUnresolved (F3-R2B) proves a fresh Connect is
// refused fail-closed while a control worker from a torn-down session remains
// outstanding, so a late send cannot contaminate a replacement session.
func TestControlConnectRefusedWhileUnresolved(t *testing.T) {
	tr := newBlockingSendTransport()
	m := newControlMaster(t, tr)

	done := make(chan error, 1)
	go func() {
		_, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
		done <- err
	}()
	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}

	// A fresh connect is refused while the worker is unresolved.
	if err := m.ConnectContext(context.Background(), time.Second); !errors.Is(err, ErrSessionUnresolved) {
		t.Fatalf("ConnectContext while unresolved = %v, want ErrSessionUnresolved", err)
	}

	tr.release()
	<-done
	deadline := time.Now().Add(time.Second)
	for m.anySessionUnresolved() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if m.anySessionUnresolved() {
		t.Fatal("unresolved fence must clear after the worker terminates")
	}
}

// TestControlDisconnectBetweenAdmissionAndSendBlocksLateStart (F3-R2A) proves a
// worker admitted in one session that has not yet transmitted is a definite
// NOT_SENT after a Disconnect (the linearization point refuses the send) and
// does not quarantine the outstation (no send occurred, so the outcome is not
// ambiguous).
func TestControlDisconnectBetweenAdmissionAndSendBlocksLateStart(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	gen, release, err := m.AcquireControlGate(1)
	if err != nil {
		t.Fatalf("AcquireControlGate: %v", err)
	}
	defer release()

	// The worker has not transmitted yet. Tear the session down first.
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	m.SetState(StateInitialized)

	if _, err := m.OperateWithStatusGated(context.Background(), gen, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second)); err == nil {
		t.Fatal("a stale-session control must be refused at the admission point")
	}
	if got := tr.sendCount(); got != 0 {
		t.Fatalf("a stale-session control must not transmit; sends = %d", got)
	}
	if m.IsQuarantined(1) {
		t.Fatal("a definite pre-send (not-yet-transmitted) failure must not quarantine")
	}
}

// TestControlQuarantineAdmissionAtomic (F3-R2B) proves UNKNOWN admission is
// atomic: a quarantine established concurrently with worker admission can never
// allow a second send. Once quarantined, a fresh admission is refused before any
// transmission even under concurrent attempts.
func TestControlQuarantineAdmissionAtomic(t *testing.T) {
	tr := &failReceiveTransport{}
	m := newControlMaster(t, tr)

	// Establish the UNKNOWN quarantine.
	if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlOutcomeUnknown) {
		t.Fatalf("first control = %v, want ErrControlOutcomeUnknown", err)
	}
	base := tr.sendCount()
	if base != 1 {
		t.Fatalf("first control sends = %d, want 1", base)
	}

	// Concurrently attempt many further controls: all must be refused by the
	// quarantine and none may transmit.
	const n = 32
	var wg sync.WaitGroup
	wg.Add(n)
	refused := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
			refused <- err
		}()
	}
	wg.Wait()
	close(refused)
	for err := range refused {
		if !errors.Is(err, ErrControlQuarantined) {
			t.Fatalf("concurrent control error = %v, want ErrControlQuarantined", err)
		}
	}
	if got := tr.sendCount(); got != base {
		t.Fatalf("a quarantined outstation transmitted %d extra controls; sends = %d, want %d", got-base, got, base)
	}
}
