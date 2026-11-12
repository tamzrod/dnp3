package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
)

// AO-002 F1 — fail closed when the configured data handler does not implement
// DataWriter. Previously handleDirectOperate logged and continued, constructing a
// success response; for G41V1 the master reads the echoed command-status octet,
// so that would be reported as a false success despite no execution. F1 makes
// the path stop with a classified error, never a SUCCESS echo, without calling a
// handler or mutating output state.

// readOnlyDataHandler implements DataHandler but deliberately NOT DataWriter
// (no WriteBinaryOutput / WriteAnalogOutput), so o.data.(DataWriter) fails.
type readOnlyDataHandler struct{}

func (readOnlyDataHandler) GetBinaryInputs() []BinaryInput   { return nil }
func (readOnlyDataHandler) GetAnalogInputs() []AnalogInput   { return nil }
func (readOnlyDataHandler) GetCounters() []Counter           { return nil }
func (readOnlyDataHandler) GetBinaryOutputs() []BinaryOutput { return nil }
func (readOnlyDataHandler) GetAnalogOutputs() []AnalogOutput { return nil }
func (readOnlyDataHandler) GetFrozenCounters() []Counter     { return nil }
func (readOnlyDataHandler) FreezeCounters(clear bool) error  { return nil }

func newReadOnlyOutstation() *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(readOnlyDataHandler{})
	return o
}

// TestOutstationG41V1NoDataWriterFailsClosed asserts a G41V1 DirectOperate to an
// outstation without a DataWriter fails closed: an error is returned, no success
// response object is constructed, and there is no side effect.
func TestOutstationG41V1NoDataWriterFailsClosed(t *testing.T) {
	o := newReadOnlyOutstation()
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := o.ProcessRequest(req)
	if err == nil {
		t.Fatalf("expected fail-closed error for a missing DataWriter, got resp=%v", resp)
	}
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("error = %v, want ErrNotSupported", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
	}
}

// TestOutstationG12V1NoDataWriterFailsClosed documents the shared fail-closed
// behavior introduced for F1: a G12V1 DirectOperate without a DataWriter also
// fails closed rather than reporting success. This is the minimal safe solution;
// the verified G12V1 success path (with a DataWriter) is covered by
// TestOutstationG41V1DecodeGolden's positive control and the existing CROB tests.
func TestOutstationG12V1NoDataWriterFailsClosed(t *testing.T) {
	o := newReadOnlyOutstation()
	// G12V1 CROB: header(5) + index(2) + CROB(11)
	crob := []byte{0x0C, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     crob,
	}
	resp, err := o.ProcessRequest(req)
	if err == nil {
		t.Fatalf("expected fail-closed error for a missing DataWriter, got resp=%v", resp)
	}
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("error = %v, want ErrNotSupported", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
}

// TestOutstationG41V1DataWriterPositiveControl confirms the positive control: a
// valid DataWriter still executes the G41V1 command and returns a SUCCESS echo,
// i.e. F1 did not break the working path.
func TestOutstationG41V1DataWriterPositiveControl(t *testing.T) {
	h := &g41RecordingHandler{}
	o := newG41Outstation(h)
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := o.ProcessRequest(req)
	if err != nil {
		t.Fatalf("positive control failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	// IIN(2) + G41V1 echo object with SUCCESS status octet.
	want := []byte{0x00, 0x00, 0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("G41V1 echo = %X, want %X", resp.Data, want)
	}
}
