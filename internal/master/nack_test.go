package master

import (
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
)

// buildSecondaryNACK builds a link-layer NACK frame (PRM=0, FuncNack=1) from
// the outstation (src) to the master (dst). Used for DNP3-058 NACK injection.
func buildSecondaryNACK(t *testing.T, dst, src uint16) []byte {
	t.Helper()
	dll := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncNack},
		DestAddr: dst,
		SrcAddr:  src,
		Data:     nil,
	}
	raw, err := frame.Encode(dll)
	if err != nil {
		t.Fatalf("encode NACK frame: %v", err)
	}
	return raw
}

// nackAlwaysTransport responds to every Receive with a link-layer NACK and
// records the number of Send calls. Used to verify the master retries a NACKed
// request up to the NACK retry budget, then fails with ErrLinkNACK (DNP3-058).
type nackAlwaysTransport struct {
	sends int
}

func (t *nackAlwaysTransport) Send(data []byte) error {
	t.sends++
	return nil
}

func (t *nackAlwaysTransport) Receive() ([]byte, error) {
	return buildSecondaryNACK(&testing.T{}, 1, 2), nil
}

func (t *nackAlwaysTransport) SetTimeout(ms int) {}

// nackThenOKTransport NACKs the first N response(s), then returns a valid
// application response echoing the request SEQ. Proves NACK retry recovers.
type nackThenOKTransport struct {
	sent    [][]byte
	nacks   int
	served  int
}

func (t *nackThenOKTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *nackThenOKTransport) Receive() ([]byte, error) {
	if t.served < t.nacks {
		t.served++
		return buildSecondaryNACK(&testing.T{}, 1, 2), nil
	}
	// Echo the most recent request's SEQ in a valid response.
	last := extractRequestSeqForBuild(t.sent[len(t.sent)-1])
	return buildResponseFrameWithSeq(last), nil
}

func (t *nackThenOKTransport) SetTimeout(ms int) {}

// TestNACKRetriedThenFails verifies DNP3-058: a request that is always NACKed
// is retried up to the NACK retry budget (NACKRetries) and then surfaced as an
// error wrapping ErrLinkNACK.
func TestNACKRetriedThenFails(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 50, MaxRetries: 3, RetryDelay: 0}
	m := NewMaster(cfg)
	tr := &nackAlwaysTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	err := m.sendWithRetry(req, 2)
	if err == nil {
		t.Fatal("expected error when every response is a NACK")
	}
	if !errors.Is(err, ErrLinkNACK) {
		t.Fatalf("error does not wrap ErrLinkNACK: %v", err)
	}
	// NACKRetries = MaxRetries = 3 → the request is attempted exactly 3 times.
	if tr.sends != 3 {
		t.Fatalf("NACKed request sent %d times, want 3 (NACKRetries budget)", tr.sends)
	}
}

// TestNACKRetriedThenRecovers verifies DNP3-058: after a NACK, the master
// retries and a subsequent valid response succeeds (no terminal error).
func TestNACKRetriedThenRecovers(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 50, MaxRetries: 3, RetryDelay: 0}
	m := NewMaster(cfg)
	tr := &nackThenOKTransport{nacks: 1}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	if err := m.sendWithRetry(req, 2); err != nil {
		t.Fatalf("sendWithRetry after NACK-recovery failed: %v", err)
	}
	// One NACKed attempt + one successful attempt.
	if len(tr.sent) != 2 {
		t.Fatalf("sent %d frames, want 2 (1 NACKed + 1 OK)", len(tr.sent))
	}
}

// TestResetLinkNACKRetriedThenFails verifies DNP3-058 on the handshake path:
// a secondary NACK in response to Reset Link Stations is retried up to the
// NACK budget and then surfaced as ErrLinkNACK.
func TestResetLinkNACKRetriedThenFails(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 50, MaxRetries: 2, RetryDelay: 0}
	m := NewMaster(cfg)
	tr := &nackAlwaysTransport{}
	m.SetTransport(tr)
	m.AddOutstation(2, "RTU-1")

	err := m.sendResetLink()
	if err == nil {
		t.Fatal("expected error when reset link is always NACKed")
	}
	if !errors.Is(err, ErrLinkNACK) {
		t.Fatalf("reset link error does not wrap ErrLinkNACK: %v", err)
	}
	// NACKRetries = 2 → the reset frame is attempted exactly 2 times.
	if tr.sends != 2 {
		t.Fatalf("NACKed reset link sent %d times, want 2 (NACKRetries budget)", tr.sends)
	}
}

// TestResetLinkNACKThenACK verifies DNP3-058 recovery on the handshake path:
// after a NACKed Reset Link Stations, a subsequent ACK succeeds.
func TestResetLinkNACKThenACK(t *testing.T) {
	// A transport that NACKs the first reset-link response, then returns a
	// valid ACK.
	tr := &resetLinkNackThenACKTransport{nacks: 1}
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 3, RetryDelay: 0})
	m.SetTransport(tr)
	m.AddOutstation(2, "RTU-1")

	if err := m.sendResetLink(); err != nil {
		t.Fatalf("sendResetLink after NACK-recovery failed: %v", err)
	}
	if tr.sends != 2 {
		t.Fatalf("reset link sent %d times, want 2 (1 NACKed + 1 ACK)", tr.sends)
	}
}

// resetLinkNackThenACKTransport NACKs the first Receive, then returns a valid
// reset-link ACK. It distinguishes the reset-link phase by counting Sends.
type resetLinkNackThenACKTransport struct {
	sends int
	nacks int
}

func (t *resetLinkNackThenACKTransport) Send(data []byte) error {
	t.sends++
	return nil
}

func (t *resetLinkNackThenACKTransport) Receive() ([]byte, error) {
	if t.sends <= t.nacks {
		return buildSecondaryNACK(&testing.T{}, 1, 2), nil
	}
	// Valid secondary ACK for Reset Link Stations (PRM=0, FuncAck, correct addrs).
	dll := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncAck},
		DestAddr: 1, SrcAddr: 2, Data: nil,
	}
	raw, _ := frame.Encode(dll)
	return raw, nil
}

func (t *resetLinkNackThenACKTransport) SetTimeout(ms int) {}
