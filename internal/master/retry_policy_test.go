package master

import (
	"errors"
	"io"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
)

// countingTransport records every Send and returns a fixed canned response per
// Receive call (cycling through frames). Used to drive sendWithRetry against
// deterministic NACK / CRC / timeout scenarios (DNP3-034).
type countingTransport struct {
	sent    int
	frames  [][]byte
	receive func() ([]byte, error)
	idx     int
}

func (t *countingTransport) Send(data []byte) error {
	t.sent++
	return nil
}

func (t *countingTransport) SetTimeout(ms int) {}

func (t *countingTransport) Receive() ([]byte, error) {
	if t.receive != nil {
		return t.receive()
	}
	if t.idx >= len(t.frames) {
		return nil, errReceiveTimeout
	}
	f := t.frames[t.idx]
	t.idx++
	return f, nil
}

// seqAwareTransport records sent frames and exposes the last request's
// application SEQ, so recovery tests can build a response that echoes it
// (DNP3-010 compliant) after an initial NACK/CRC failure (DNP3-034).
type seqAwareTransport struct {
	sent     [][]byte
	receive  func() ([]byte, error)
	lastSeq  uint8
}

func (t *seqAwareTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *seqAwareTransport) SetTimeout(ms int) {}

func (t *seqAwareTransport) Receive() ([]byte, error) {
	return t.receive()
}

// buildNACKFrame builds a valid DLL frame carrying a secondary link-layer NACK
// (PRM=0, FuncNack=1) from the outstation (DNP3-034).
func buildNACKFrame() []byte {
	f := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncNack},
		DestAddr: 1,
		SrcAddr:  2,
		Data:     nil,
	}
	b, _ := frame.Encode(f)
	return b
}

// buildBadCRCResponseFrame builds a DLL frame whose header CRC has been
// corrupted so frame.Decode reports a CRC validation failure (DNP3-034). It
// starts from a valid minimal response frame and flips one header byte covered
// by the header CRC (the dest low byte) without altering the length field, so
// the decoder reaches the CRC-validation stage rather than rejecting the frame
// as truncated/oversized.
func buildBadCRCResponseFrame(t *testing.T) []byte {
	t.Helper()
	raw := buildMinimalResponseFrame(t)
	// Header layout: [0:2] sync, [2] length, [3] control, [4:6] dest,
	// [6:8] src, [8:10] header CRC. Corrupt the dest low byte (offset 4),
	// which is inside the header-CRC range [0:8] but does not change the
	// parsed length/size.
	if len(raw) <= 5 {
		t.Fatalf("response frame too short: %d", len(raw))
	}
	raw[4] ^= 0xFF
	return raw
}

// newRetryMaster builds a master wired to tr with a fast retry policy for
// deterministic testing (DNP3-034).
func newRetryMaster(tr TransportHandler, retries int, delay time.Duration) *Master {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: retries, RetryDelay: int(delay / time.Millisecond)})
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")
	m.SetRetryPolicy(&RetryPolicy{
		TimeoutRetries: retries, TimeoutDelay: delay,
		NACKRetries: retries, NACKDelay: delay,
		CRCRetries: retries, CRCDelay: delay,
		OtherRetries: retries, OtherDelay: delay,
	})
	return m
}

// TestRetryClassifiesTimeout verifies a receive timeout is classified as
// ClassTimeout and retried up to the configured count, then surfaces
// ErrMaxRetries wrapping ErrConfirmTimeout/ErrTimeout (DNP3-034).
func TestRetryClassifiesTimeout(t *testing.T) {
	tr := &countingTransport{receive: func() ([]byte, error) { return nil, errReceiveTimeout }}
	m := newRetryMaster(tr, 3, 0)

	err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !errors.Is(err, ErrMaxRetries) {
		t.Fatalf("expected ErrMaxRetries, got %v", err)
	}
	// 1 initial attempt + 2 retries = 3 sends.
	if tr.sent != 3 {
		t.Fatalf("timeout sends = %d, want 3", tr.sent)
	}
}

// TestRetryClassifiesNACK verifies a link-layer NACK is classified as
// ClassNACK, retried up to the configured count, and surfaced as
// ErrLinkNACK (wrapped by ErrMaxRetries) when exhausted (DNP3-034).
func TestRetryClassifiesNACK(t *testing.T) {
	nack := buildNACKFrame()
	tr := &countingTransport{receive: func() ([]byte, error) { return nack, nil }}
	m := newRetryMaster(tr, 3, 0)

	err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2)
	if err == nil {
		t.Fatal("expected error after NACK retries exhausted")
	}
	if !errors.Is(err, ErrLinkNACK) {
		t.Fatalf("expected ErrLinkNACK, got %v", err)
	}
	if !errors.Is(err, ErrMaxRetries) {
		t.Fatalf("expected ErrMaxRetries wrapping NACK, got %v", err)
	}
	if tr.sent != 3 {
		t.Fatalf("NACK sends = %d, want 3", tr.sent)
	}
}

// TestRetryClassifiesCRC verifies a corrupted received frame (CRC failure) is
// classified as ClassCRC, retried up to the configured count, and surfaced as
// ErrCRCError when exhausted (DNP3-034).
func TestRetryClassifiesCRC(t *testing.T) {
	bad := buildBadCRCResponseFrame(t)
	tr := &countingTransport{receive: func() ([]byte, error) { return bad, nil }}
	m := newRetryMaster(tr, 3, 0)

	err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2)
	if err == nil {
		t.Fatal("expected error after CRC retries exhausted")
	}
	if !errors.Is(err, ErrCRCError) {
		t.Fatalf("expected ErrCRCError, got %v", err)
	}
	if !errors.Is(err, ErrMaxRetries) {
		t.Fatalf("expected ErrMaxRetries wrapping CRC, got %v", err)
	}
	if tr.sent != 3 {
		t.Fatalf("CRC sends = %d, want 3", tr.sent)
	}
}

// TestRetryNACKRecoversOnSuccess verifies the retry loop retries after a NACK
// and succeeds when a valid response follows (DNP3-034).
func TestRetryNACKRecoversOnSuccess(t *testing.T) {
	nack := buildNACKFrame()
	tr := &seqAwareTransport{}
	first := true
	tr.receive = func() ([]byte, error) {
		if first {
			first = false
			return nack, nil
		}
		return buildResponseFrameWithSeq(tr.lastSeq), nil
	}
	m := newRetryMaster(tr, 3, 0)

	if err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2); err != nil {
		t.Fatalf("expected recovery after NACK, got %v", err)
	}
	if len(tr.sent) != 2 {
		t.Fatalf("sends = %d, want 2 (1 NACK + 1 success)", len(tr.sent))
	}
}

// TestRetryCRCRecoversOnSuccess verifies the retry loop retries after a CRC
// error and succeeds when a valid response follows (DNP3-034).
func TestRetryCRCRecoversOnSuccess(t *testing.T) {
	bad := buildBadCRCResponseFrame(t)
	tr := &seqAwareTransport{}
	first := true
	tr.receive = func() ([]byte, error) {
		if first {
			first = false
			return bad, nil
		}
		return buildResponseFrameWithSeq(tr.lastSeq), nil
	}
	m := newRetryMaster(tr, 3, 0)

	if err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2); err != nil {
		t.Fatalf("expected recovery after CRC, got %v", err)
	}
	if len(tr.sent) != 2 {
		t.Fatalf("sends = %d, want 2 (1 CRC + 1 success)", len(tr.sent))
	}
}

// TestRetryDisconnectNotRetried verifies a transport disconnect is NOT retried
// regardless of the retry policy (the link is dead — DNP3-031/034).
func TestRetryDisconnectNotRetried(t *testing.T) {
	tr := &countingTransport{receive: func() ([]byte, error) { return nil, io.EOF }}
	m := newRetryMaster(tr, 3, 0)

	err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2)
	if err == nil {
		t.Fatal("expected disconnect error")
	}
	if !errors.Is(err, ErrTransportDisconnected) {
		t.Fatalf("expected ErrTransportDisconnected, got %v", err)
	}
	if tr.sent != 1 {
		t.Fatalf("disconnect sends = %d, want 1 (no retry)", tr.sent)
	}
}

// TestRetryPerClassCounts verifies distinct per-class retry counts: a policy
// allowing 2 NACK retries but only 1 CRC retry yields the right send counts
// (DNP3-034).
func TestRetryPerClassCounts(t *testing.T) {
	// NACK with NACKRetries=2 → 2 sends.
	nack := buildNACKFrame()
	trN := &countingTransport{receive: func() ([]byte, error) { return nack, nil }}
	mN := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 3, RetryDelay: 0})
	mN.SetTransport(trN)
	mN.SetState(StateInitialized)
	mN.AddOutstation(2, "RTU-1")
	mN.SetRetryPolicy(&RetryPolicy{
		TimeoutRetries: 3, TimeoutDelay: 0,
		NACKRetries:    2, NACKDelay: 0,
		CRCRetries:     3, CRCDelay: 0,
		OtherRetries:   3, OtherDelay: 0,
	})
	if err := mN.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2); err == nil {
		t.Fatal("expected NACK exhaustion")
	}
	if trN.sent != 2 {
		t.Fatalf("NACK sends = %d, want 2", trN.sent)
	}

	// CRC with CRCRetries=1 → 1 send.
	bad := buildBadCRCResponseFrame(t)
	trC := &countingTransport{receive: func() ([]byte, error) { return bad, nil }}
	mC := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 3, RetryDelay: 0})
	mC.SetTransport(trC)
	mC.SetState(StateInitialized)
	mC.AddOutstation(2, "RTU-1")
	mC.SetRetryPolicy(&RetryPolicy{
		TimeoutRetries: 3, TimeoutDelay: 0,
		NACKRetries:    3, NACKDelay: 0,
		CRCRetries:     1, CRCDelay: 0,
		OtherRetries:   3, OtherDelay: 0,
	})
	if err := mC.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2); err == nil {
		t.Fatal("expected CRC exhaustion")
	}
	if trC.sent != 1 {
		t.Fatalf("CRC sends = %d, want 1", trC.sent)
	}
}

// TestRetryDelayApplied verifies the policy delay is slept before a retry
// (DNP3-034). It uses a measurable delay and a wall-clock bound.
func TestRetryDelayApplied(t *testing.T) {
	tr := &countingTransport{receive: func() ([]byte, error) { return nil, errReceiveTimeout }}
	delay := 25 * time.Millisecond
	m := newRetryMaster(tr, 2, delay)

	start := time.Now()
	err := m.sendWithRetry(buildRequest(0, al.FuncRead, []byte{0x00}), 2)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	// 2 attempts → 1 retry delay between them.
	if elapsed < delay {
		t.Fatalf("elapsed = %v, want >= %v (retry delay not applied)", elapsed, delay)
	}
	if tr.sent != 2 {
		t.Fatalf("sends = %d, want 2", tr.sent)
	}
}

// TestClassifyRetryError verifies the error → RetryClass mapping (DNP3-034).
func TestClassifyRetryError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want RetryClass
	}{
		{"nil", nil, ClassOther},
		{"disconnect-EOF", io.EOF, ClassDisconnect},
		{"disconnect-sentinel", ErrTransportDisconnected, ClassDisconnect},
		{"nack", ErrLinkNACK, ClassNACK},
		{"crc", ErrCRCError, ClassCRC},
		{"confirm-timeout", ErrConfirmTimeout, ClassTimeout},
		{"timeout", ErrTimeout, ClassTimeout},
		{"other", ErrResponseSeqMismatch, ClassOther},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyRetryError(tc.err); got != tc.want {
				t.Fatalf("classifyRetryError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestDefaultRetryPolicy verifies the default retry table is derived from the
// config (DNP3-034).
func TestDefaultRetryPolicy(t *testing.T) {
	cfg := &Config{MasterAddress: 1, Timeout: 50, MaxRetries: 4, RetryDelay: 250}
	p := DefaultRetryPolicy(cfg)
	if p.TimeoutRetries != 4 || p.NACKRetries != 4 || p.CRCRetries != 4 || p.OtherRetries != 4 {
		t.Fatalf("retry counts mismatch: %+v", p)
	}
	wantDelay := 250 * time.Millisecond
	if p.TimeoutDelay != wantDelay || p.NACKDelay != wantDelay || p.CRCDelay != wantDelay || p.OtherDelay != wantDelay {
		t.Fatalf("retry delays mismatch: %+v", p)
	}

	// A zero/underflow MaxRetries falls back to MaxRetries (3).
	p2 := DefaultRetryPolicy(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 0, RetryDelay: 0})
	if p2.TimeoutRetries != MaxRetries {
		t.Fatalf("zero MaxRetries should fall back to %d, got %d", MaxRetries, p2.TimeoutRetries)
	}
}

// TestProcessReceivedBytesSurfacesNACK verifies processReceivedBytes surfaces a
// secondary NACK frame as ErrLinkNACK (DNP3-034).
func TestProcessReceivedBytesSurfacesNACK(t *testing.T) {
	m := NewMaster(DefaultConfig())
	_, err := m.processReceivedBytes(buildNACKFrame())
	if err == nil {
		t.Fatal("expected ErrLinkNACK")
	}
	if !errors.Is(err, ErrLinkNACK) {
		t.Fatalf("expected ErrLinkNACK, got %v", err)
	}
}

// TestProcessReceivedBytesSurfacesCRC verifies processReceivedBytes surfaces a
// corrupted frame as ErrCRCError (DNP3-034).
func TestProcessReceivedBytesSurfacesCRC(t *testing.T) {
	m := NewMaster(DefaultConfig())
	_, err := m.processReceivedBytes(buildBadCRCResponseFrame(t))
	if err == nil {
		t.Fatal("expected ErrCRCError")
	}
	if !errors.Is(err, ErrCRCError) {
		t.Fatalf("expected ErrCRCError, got %v", err)
	}
}
