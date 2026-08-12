package master

import (
	"errors"
	"testing"

	"dnp3/internal/al"
)

// MEXT-018 — Application SEQ + CON solicited-path audit.
//
// The unit-level SEQ/CON invariants are covered across master_test.go,
// app_confirm_test.go, confirm_timeout_test.go and fcb_test.go (stream 0-15
// wrap; advance only on successful send; no advance on send failure;
// processResponse SEQ match/mismatch; waitForConfirmation match/mismatch/
// timeout; CON=1 response triggers an application confirm). This file fills the
// remaining end-to-end gaps: the CON=1 solicited confirm+response combination
// through sendWithRetryAndGetResponse, and a characterization of the retry-SEQ
// behavior under a response-failure (advance-on-send) so the spec-continuity
// behavior is locked or ticketed.

// scriptedSeqTransport records the application SEQ of every sent request and
// serves a queued sequence of raw response frames from Receive (one per call).
// A nil queue entry surfaces a transport error. The first sent SEQ is exposed
// so tests can build matching/mismatching confirms and responses.
type scriptedSeqTransport struct {
	sent      [][]byte
	firstSeq  uint8
	responses [][]byte
	idx       int
}

func (t *scriptedSeqTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	if len(t.sent) == 1 {
		t.firstSeq = extractRequestSeqForBuild(data)
	}
	return nil
}

func (t *scriptedSeqTransport) Receive() ([]byte, error) {
	if t.idx >= len(t.responses) {
		return nil, errReceiveTimeout
	}
	r := t.responses[t.idx]
	t.idx++
	if r == nil {
		return nil, errReceiveTimeout
	}
	return r, nil
}

func (t *scriptedSeqTransport) SetTimeout(ms int) {}

// terminalRetryPolicy disables every retry class so a mismatch/timeout is
// surfaced immediately (wrapped only by the caller, not retried away). Used by
// the audit tests to assert the exact terminal error.
func terminalRetryPolicy() *RetryPolicy {
	return &RetryPolicy{}
}

// TestSolicitedCONConfirmAndResponseMatchingSeq succeeds the full CON=1
// solicited path: a dedicated confirm with the matching SEQ, then a response
// with the matching SEQ. The outstation sequence advances exactly once for the
// completed transaction (DNP3-008/009/010).
func TestSolicitedCONConfirmAndResponseMatchingSeq(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	tr := &scriptedSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	// The outstation's SEQ stream starts at 0, so the request carries SEQ 0.
	// Queue a dedicated confirm (IIN-only) then a response, both with SEQ 0.
	tr.responses = [][]byte{
		buildConfirmFrame(t, 0),
		buildResponseFrameWithSeq(0),
	}

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	req.Control.CON = true

	before := m.currentSequence(2)
	if _, err := m.sendWithRetryAndGetResponse(req, 2); err != nil {
		t.Fatalf("CON solicited path failed: %v", err)
	}

	// The confirm must echo the request SEQ (DNP3-009); the request SEQ was 0.
	if tr.firstSeq != 0 {
		t.Fatalf("request SEQ = %d, want 0", tr.firstSeq)
	}
	if len(tr.sent) < 1 {
		t.Fatalf("expected at least one sent frame, got %d", len(tr.sent))
	}

	// Sequence advances exactly once on the completed transaction.
	if got := m.currentSequence(2); got != (before+1)%16 {
		t.Fatalf("currentSequence = %d, want %d (single advance)", got, (before+1)%16)
	}
}

// TestSolicitedCONConfirmWrongSeqFails proves a CON=1 request whose dedicated
// confirm carries the wrong SEQ fails the solicited path end-to-end
// (ErrConfirmSeqMismatch), not silently accepting a stale confirm (DNP3-009).
func TestSolicitedCONConfirmWrongSeqFails(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 0, RetryDelay: 0})
	m.SetRetryPolicy(terminalRetryPolicy())
	tr := &scriptedSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	req.Control.CON = true

	// Queue a dedicated confirm with a deliberately mismatched SEQ.
	wrongSeq := byte(0x0F)
	tr.responses = [][]byte{buildConfirmFrame(t, wrongSeq)}

	_, err := m.sendWithRetryAndGetResponse(req, 2)
	if err == nil {
		t.Fatal("expected failure on mismatched confirm SEQ, got nil")
	}
	if !errors.Is(err, ErrConfirmSeqMismatch) {
		t.Fatalf("expected ErrConfirmSeqMismatch, got %v", err)
	}
}

// TestSolicitedResponseSeqMismatchFailsEndToEnd proves a response whose SEQ
// does not match the outstanding request is rejected end-to-end through the
// solicited path (ErrResponseSeqMismatch), with no application data surfaced
// (DNP3-010).
func TestSolicitedResponseSeqMismatchFailsEndToEnd(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 0, RetryDelay: 0})
	m.SetRetryPolicy(terminalRetryPolicy())
	tr := &scriptedSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})

	// Serve a response with a SEQ that cannot match the request (request SEQ
	// is the outstation's current stream value, here 0; respond with 0x0F).
	tr.responses = [][]byte{buildResponseFrameWithSeq(0x0F)}

	data, err := m.sendWithRetryAndGetResponse(req, 2)
	if err == nil {
		t.Fatal("expected ErrResponseSeqMismatch, got nil")
	}
	if data != nil {
		t.Fatalf("expected no data on mismatch, got %d bytes", len(data))
	}
	if !errors.Is(err, ErrResponseSeqMismatch) {
		t.Fatalf("expected ErrResponseSeqMismatch, got %v", err)
	}
}

// TestSolicitedRetryReusesOrAdvancesSeq characterizes the retry-SEQ behavior
// when the first response fails and the request is retried. The master
// advances the application SEQ at send time (DNP3-008), so a retry of the same
// logical request after a response failure is observed with an INCREMENTED
// SEQ. This test locks the actual behavior; the documented "retries reuse the
// same value" comment is ticketed as a spec-continuity follow-up (see handoff
// MEXT-018 discovery).
func TestSolicitedRetryReusesOrAdvancesSeq(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	// Allow exactly one retry of an Other-class error (ErrResponseSeqMismatch
	// is ClassOther). OtherRetries=2 permits attempt 1 to retry to attempt 2.
	m.SetRetryPolicy(&RetryPolicy{OtherRetries: 2})
	// retryEchoSeqTransport serves a fixed first response (mismatched SEQ) on
	// the first Receive, then echoes the most-recently-sent request SEQ on
	// every subsequent Receive so the retry succeeds regardless of the SEQ
	// the master assigns to the retry attempt.
	tr := &retryEchoSeqTransport{firstResp: buildResponseFrameWithSeq(0x0F)}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})

	if _, err := m.sendWithRetryAndGetResponse(req, 2); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}

	// Two attempts => two sent request frames.
	if len(tr.sent) < 2 {
		t.Fatalf("expected >=2 sent frames (retry), got %d", len(tr.sent))
	}
	firstSeq := extractRequestSeqForBuild(tr.sent[0])
	retrySeq := extractRequestSeqForBuild(tr.sent[1])

	// Characterization: the retry carries an INCREMENTED SEQ because the
	// master advances at send time. (If the implementation is later changed
	// to reuse the SEQ on retry, this assertion flips to equality — update
	// this test and close the ticketed follow-up together.)
	wantRetry := (firstSeq + 1) % 16
	if retrySeq != wantRetry {
		t.Fatalf("retry SEQ = %d, want %d (advance-on-send characterization)", retrySeq, wantRetry)
	}
}

// retryEchoSeqTransport serves firstResp on the first Receive, then echoes the
// most-recently-sent request SEQ on every subsequent Receive.
type retryEchoSeqTransport struct {
	sent      [][]byte
	firstResp []byte
	served    bool
}

func (t *retryEchoSeqTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *retryEchoSeqTransport) Receive() ([]byte, error) {
	if !t.served {
		t.served = true
		return t.firstResp, nil
	}
	last := byte(0)
	if len(t.sent) > 0 {
		last = extractRequestSeqForBuild(t.sent[len(t.sent)-1])
	}
	return buildResponseFrameWithSeq(last), nil
}

func (t *retryEchoSeqTransport) SetTimeout(ms int) {}
