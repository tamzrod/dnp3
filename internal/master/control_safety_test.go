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
// is UNKNOWN, no new control may be issued (fail closed) and that the guard
// holds across reconnect (a state reset) — only explicit teardown clears it.
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

	// Explicit teardown is the only event that releases the quarantine.
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if m.IsQuarantined(1) {
		t.Fatal("Disconnect did not clear the control quarantine")
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
