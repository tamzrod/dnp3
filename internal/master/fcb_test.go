package master

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// fcbRecorderTransport records every frame the master sends and returns a
// minimal valid response per request so sendWithRetryAndGetResponse completes.
type fcbRecorderTransport struct {
	sent    [][]byte
	lastSeq uint8
}

func (t *fcbRecorderTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	if f, err := frame.Decode(data); err == nil {
		if frag, err := tl.DecodeFragment(f.Data); err == nil {
			if apdu, err := al.Decode(frag.Data); err == nil {
				t.lastSeq = apdu.Control.Seq
			}
		}
	}
	return nil
}

func (t *fcbRecorderTransport) Receive() ([]byte, error) {
	return buildResponseFrameWithSeq(t.lastSeq), nil
}

func (t *fcbRecorderTransport) SetTimeout(ms int) {}

// confirmedDataFrames filters the recorded sent frames down to the master's
// confirmed-user-data frames (PRM=1, FuncConfirmedUserData), returning each
// one's control (FCB/FCV) for pattern assertions.
func confirmedDataFrames(t *testing.T, frames [][]byte) []frame.Control {
	t.Helper()
	var out []frame.Control
	for _, raw := range frames {
		f, err := frame.Decode(raw)
		if err != nil {
			continue
		}
		if f.Control.PRM && f.Control.FuncCode == frame.FuncConfirmedUserData {
			out = append(out, f.Control)
		}
	}
	return out
}

// TestFCBTogglesAcrossConfirmedData verifies DNP3-057: the master sets FCV=1
// on every confirmed-user-data frame and the FCB alternates 0,1,0,1 across
// committed transactions (toggling only on success, not per retry attempt).
func TestFCBTogglesAcrossConfirmedData(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	tr := &fcbRecorderTransport{}
	m.SetTransport(tr)
	// Register the outstation so its FCB state is tracked.
	m.AddOutstation(2, "os")

	readReq := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true},
		FuncCode: al.FuncRead,
		Data:     []byte{0x01, 0x00, 0x06, 0x00}, // G1V0 all-objects
	}

	for i := 0; i < 4; i++ {
		if _, err := m.SendRequestWithRetryAndGetResponse(readReq, 2); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}

	controls := confirmedDataFrames(t, tr.sent)
	if len(controls) < 4 {
		t.Fatalf("expected >=4 confirmed-data frames, got %d", len(controls))
	}
	// Take the first 4 confirmed-data frames (one per request transaction).
	want := []bool{false, true, false, true}
	for i, c := range controls[:4] {
		if !c.FCV {
			t.Errorf("frame %d: FCV = false, want true", i)
		}
		if c.FCB != want[i] {
			t.Errorf("frame %d: FCB = %v, want %v", i, c.FCB, want[i])
		}
	}
}

// TestFCBResetAfterResetLink verifies DNP3-057: a Reset Link Stations exchange
// resets the primary-side FCB to false. Register an outstation, advance the
// FCB by sending one confirmed-data request, assert the next FCB is true,
// then call resetFCB (as sendResetLink does) and assert the FCB returns to
// false.
func TestFCBResetAfterResetLink(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	tr := &fcbRecorderTransport{}
	m.SetTransport(tr)
	m.AddOutstation(2, "os")

	readReq := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true},
		FuncCode: al.FuncRead,
		Data:     []byte{0x01, 0x00, 0x06, 0x00},
	}
	if _, err := m.SendRequestWithRetryAndGetResponse(readReq, 2); err != nil {
		t.Fatalf("request: %v", err)
	}
	if got := m.outstationFCB(2); got != true {
		t.Fatalf("after first transaction, FCB = %v, want true", got)
	}
	// Simulate the Reset Link Stations effect on FCB.
	o, ok := m.GetOutstation(2)
	if !ok {
		t.Fatal("outstation not found")
	}
	o.resetFCB()
	if got := m.outstationFCB(2); got != false {
		t.Fatalf("after reset, FCB = %v, want false", got)
	}
}

// TestFCBStableAcrossRetries verifies DNP3-057: the FCB does not toggle on a
// failed/retried attempt — it advances only when the transaction commits. The
// first response carries a mismatched application sequence (forcing a retry);
// both confirmed-data attempts must carry the SAME FCB, and FCV must be set.
func TestFCBStableAcrossRetries(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 2, RetryDelay: 0})
	tr := &seqMismatchThenOKTransport{}
	m.SetTransport(tr)
	m.AddOutstation(2, "os")

	readReq := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true},
		FuncCode: al.FuncRead,
		Data:     []byte{0x01, 0x00, 0x06, 0x00},
	}
	if _, err := m.SendRequestWithRetryAndGetResponse(readReq, 2); err != nil {
		t.Fatalf("request: %v", err)
	}
	controls := confirmedDataFrames(t, tr.sent)
	if len(controls) < 2 {
		t.Fatalf("expected >=2 confirmed-data attempts (a retry), got %d", len(controls))
	}
	if controls[0].FCB != controls[1].FCB {
		t.Errorf("FCB changed across retry: attempt0=%v attempt1=%v (should stay equal)", controls[0].FCB, controls[1].FCB)
	}
	if !controls[0].FCV {
		t.Errorf("attempt0: FCV = false, want true")
	}
}

// seqMismatchThenOKTransport records sends and returns a response with a
// mismatched application sequence on the first Receive (forcing a retry),
// then echoes the request seq on subsequent receives.
type seqMismatchThenOKTransport struct {
	sent    [][]byte
	lastSeq uint8
	calls   int
}

func (t *seqMismatchThenOKTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	if f, err := frame.Decode(data); err == nil {
		if frag, err := tl.DecodeFragment(f.Data); err == nil {
			if apdu, err := al.Decode(frag.Data); err == nil {
				t.lastSeq = apdu.Control.Seq
			}
		}
	}
	return nil
}

func (t *seqMismatchThenOKTransport) Receive() ([]byte, error) {
	t.calls++
	if t.calls == 1 {
		// Mismatched seq (XOR the low nibble) to force ErrResponseSeqMismatch.
		mismatch := t.lastSeq ^ 0x0F
		return buildResponseFrameWithSeq(mismatch), nil
	}
	return buildResponseFrameWithSeq(t.lastSeq), nil
}

func (t *seqMismatchThenOKTransport) SetTimeout(ms int) {}
