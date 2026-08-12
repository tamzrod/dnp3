package master

import (
	"testing"
	"time"

	"dnp3/internal/al"
	"errors"
)

// timeoutRecorderTransport records the last SetTimeout value and serves a fixed
// canned frame per Receive (cycling). Used to verify which timeout budget
// each code path applies (DNP3-066).
type timeoutRecorderTransport struct {
	frames    [][]byte
	idx       int
	lastTimeout int
}

func (t *timeoutRecorderTransport) Send(data []byte) error { return nil }

func (t *timeoutRecorderTransport) Receive() ([]byte, error) {
	if len(t.frames) == 0 {
		return nil, ErrTimeout
	}
	f := t.frames[t.idx%len(t.frames)]
	t.idx++
	return f, nil
}

func (t *timeoutRecorderTransport) SetTimeout(ms int) {
	t.lastTimeout = ms
}

// TestDefaultConfirmTimeoutDistinctFromResponse verifies DNP3-066: the default
// ConfirmTimeout is a distinct (shorter) value than the response Timeout, and
// is present on DefaultConfig.
func TestDefaultConfirmTimeoutDistinctFromResponse(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ConfirmTimeout <= 0 {
		t.Fatalf("DefaultConfig ConfirmTimeout = %d, want > 0", cfg.ConfirmTimeout)
	}
	if cfg.ConfirmTimeout == cfg.Timeout {
		t.Fatalf("DefaultConfig ConfirmTimeout (%d) == Timeout (%d); must be distinct (DNP3-066)",
			cfg.ConfirmTimeout, cfg.Timeout)
	}
	if cfg.ConfirmTimeout >= cfg.Timeout {
		t.Fatalf("DefaultConfig ConfirmTimeout (%d) should be shorter than Timeout (%d) (DNP3-066)",
			cfg.ConfirmTimeout, cfg.Timeout)
	}
}

// TestConfirmTimeoutFallsBackToResponseTimeout verifies DNP3-066's documented
// backward-compatible relation: a non-positive ConfirmTimeout falls back to the
// response Timeout.
func TestConfirmTimeoutFallsBackToResponseTimeout(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 7777, MaxRetries: 1, RetryDelay: 0, ConfirmTimeout: 0}
	m := NewMaster(cfg)
	if got := int(m.confirmTimeout / time.Millisecond); got != 7777 {
		t.Fatalf("confirmTimeout fallback = %d ms, want 7777 (Timeout)", got)
	}
}

// TestConfirmTimeoutUsedOnConfirmPath verifies DNP3-066: waitForConfirmation
// applies Config.ConfirmTimeout to the transport, not Config.Timeout.
func TestConfirmTimeoutUsedOnConfirmPath(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 9000, MaxRetries: 1, RetryDelay: 0, ConfirmTimeout: 1234}
	m := NewMaster(cfg)
	tr := &timeoutRecorderTransport{frames: [][]byte{buildConfirmFrame(t, 1)}}
	m.SetTransport(tr)

	if err := m.waitForConfirmation(1); err != nil {
		t.Fatalf("waitForConfirmation: %v", err)
	}
	if tr.lastTimeout != 1234 {
		t.Fatalf("confirm path SetTimeout = %d, want 1234 (ConfirmTimeout)", tr.lastTimeout)
	}
}

// TestResponseTimeoutUsedOnResponsePath verifies DNP3-066: the response path
// (waitForResponse, used by sendWithRetry) applies Config.Timeout, distinct
// from ConfirmTimeout.
func TestResponseTimeoutUsedOnResponsePath(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 9000, MaxRetries: 1, RetryDelay: 0, ConfirmTimeout: 1234}
	m := NewMaster(cfg)
	tr := &timeoutRecorderTransport{frames: [][]byte{buildResponseFrameWithSeq(0)}}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	if err := m.sendWithRetry(req, 2); err != nil {
		t.Fatalf("sendWithRetry: %v", err)
	}
	// The response path must set Timeout (9000), not ConfirmTimeout (1234).
	if tr.lastTimeout != 9000 {
		t.Fatalf("response path SetTimeout = %d, want 9000 (Timeout)", tr.lastTimeout)
	}
}

// TestConfirmTimeoutSurfacesErrConfirmTimeout verifies that when the confirm
// path times out (no frames), the error is ErrConfirmTimeout (DNP3-009/066).
func TestConfirmTimeoutSurfacesErrConfirmTimeout(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 9000, MaxRetries: 1, RetryDelay: 0, ConfirmTimeout: 1234}
	m := NewMaster(cfg)
	tr := &timeoutRecorderTransport{frames: nil}
	m.SetTransport(tr)

	err := m.waitForConfirmation(5)
	if err == nil {
		t.Fatal("expected ErrConfirmTimeout, got nil")
	}
	if !errors.Is(err, ErrConfirmTimeout) {
		t.Fatalf("expected ErrConfirmTimeout, got %v", err)
	}
}
