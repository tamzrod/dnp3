package master

import (
	"errors"
	"io"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// failConfirmTransport serves a scripted list of raw link frames and fails the
// Send of the nth application confirm (1-based). Non-confirm sends succeed so
// the request itself is delivered; every other send is recorded.
type failConfirmTransport struct {
	frames     [][]byte
	idx        int
	confirms   int
	failOn     int
	sent       [][]byte
	disconnect bool
}

func (t *failConfirmTransport) Send(data []byte) error {
	if isConfirmFrame(data) {
		t.confirms++
		if t.failOn != 0 && t.confirms == t.failOn {
			if t.disconnect {
				return io.EOF
			}
			return errors.New("confirm send failed")
		}
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *failConfirmTransport) Receive() ([]byte, error) {
	if t.idx < len(t.frames) {
		f := t.frames[t.idx]
		t.idx++
		return f, nil
	}
	return nil, ErrTimeout
}

func (t *failConfirmTransport) SetTimeout(int) {}

// isConfirmFrame reports whether raw is a link frame carrying an application
// confirm (FuncCode 0) APDU.
func isConfirmFrame(raw []byte) bool {
	f, err := frame.Decode(raw)
	if err != nil {
		return false
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		return false
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		return false
	}
	return apdu.FuncCode == al.FuncConfirm
}

func multiFragCONFrames(t *testing.T) [][]byte {
	t.Helper()
	frag1 := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: false, CON: true, Seq: 1},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00, 0x02, 0x02, 0x17, 0x01, 0x07, 0x81},
	}
	frag2 := &al.APDU{
		Control:  al.AppControl{FIR: false, FIN: true, CON: true, Seq: 2},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00, 0x02, 0x02, 0x17, 0x01, 0x08, 0x81},
	}
	return [][]byte{
		encodeUserDataFrame(t, frag1, 0),
		encodeUserDataFrame(t, frag2, 1),
	}
}

// TestConfirmSendFailureOnFirstFragmentFailsRead pins PPR-001-F3: when the
// application confirm for the first fragment of a multi-fragment solicited
// response cannot be sent, waitForResponseTimeout must fail (not advance done
// and later report a completed response).
func TestConfirmSendFailureOnFirstFragmentFailsRead(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 100, MaxRetries: 1})
	m.AddOutstation(2, "RTU")
	tr := &failConfirmTransport{frames: multiFragCONFrames(t), failOn: 1}
	m.SetTransport(tr)

	data, err := m.waitForResponseTimeout(100*time.Millisecond, 2)
	if err == nil {
		t.Fatalf("expected error when the first fragment confirm fails, got data=%d bytes", len(data))
	}
	if data != nil {
		t.Fatalf("failed confirm must not return a completed response, got %d bytes", len(data))
	}
	if !errors.Is(err, ErrTransportDisconnected) && tr.disconnect {
		t.Fatalf("disconnect confirm error should classify as ErrTransportDisconnected, got %v", err)
	}
}

// TestConfirmSendFailureOnFinalFragmentFailsRead pins that a failure confirming
// the *final* fragment of a completed multi-fragment response is also surfaced,
// not silently treated as a delivered response.
func TestConfirmSendFailureOnFinalFragmentFailsRead(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 100, MaxRetries: 1})
	m.AddOutstation(2, "RTU")
	tr := &failConfirmTransport{frames: multiFragCONFrames(t), failOn: 2}
	m.SetTransport(tr)

	data, err := m.waitForResponseTimeout(100*time.Millisecond, 2)
	if err == nil {
		t.Fatalf("expected error when the final fragment confirm fails, got data=%d bytes", len(data))
	}
	if data != nil {
		t.Fatalf("failed final confirm must not return a completed response, got %d bytes", len(data))
	}
}

// TestConfirmSendFailureSingleFragmentFailsRead covers the single-fragment CON
// response path (the final confirm runs after the response is complete).
func TestConfirmSendFailureSingleFragmentFailsRead(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 100, MaxRetries: 1})
	m.AddOutstation(2, "RTU")
	one := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, CON: true, Seq: 1},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00},
	}
	tr := &failConfirmTransport{frames: [][]byte{encodeUserDataFrame(t, one, 0)}, failOn: 1}
	m.SetTransport(tr)

	data, err := m.waitForResponseTimeout(100*time.Millisecond, 2)
	if err == nil {
		t.Fatalf("expected error when a single-fragment confirm fails, got data=%d bytes", len(data))
	}
	if data != nil {
		t.Fatalf("failed confirm must not return a completed response, got %d bytes", len(data))
	}
}

// TestConfirmSendDisconnectClassified pins that a disconnected transport during
// the confirm is classified as ErrTransportDisconnected (indeterminate, no
// false success).
func TestConfirmSendDisconnectClassified(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 100, MaxRetries: 1})
	m.AddOutstation(2, "RTU")
	tr := &failConfirmTransport{frames: multiFragCONFrames(t), failOn: 1, disconnect: true}
	m.SetTransport(tr)

	_, err := m.waitForResponseTimeout(100*time.Millisecond, 2)
	if err == nil {
		t.Fatal("expected a disconnect error")
	}
	if !errors.Is(err, ErrTransportDisconnected) {
		t.Fatalf("error = %v, want ErrTransportDisconnected", err)
	}
}
