package outstation

import (
	"bytes"
	"encoding/hex"
	"testing"

	"dnp3/internal/al"
)

// captureDataHandler reproduces the point profile carried in the captured
// Windows-master Class 0+1+2+3 response on 2026-11-19: eight binary inputs,
// four analog inputs, four binary output statuses (DO3 on), no counters and
// no analog outputs.
type captureDataHandler struct{}

func (captureDataHandler) GetBinaryInputs() []BinaryInput {
	out := make([]BinaryInput, 8)
	for i := range out {
		out[i] = BinaryInput{Quality: BinaryQualityOnline}
	}
	return out
}

func (captureDataHandler) GetAnalogInputs() []AnalogInput {
	out := make([]AnalogInput, 4)
	for i := range out {
		out[i] = AnalogInput{Quality: AnalogQualityOnline}
	}
	return out
}

func (captureDataHandler) GetCounters() []Counter { return nil }
func (captureDataHandler) GetBinaryOutputs() []BinaryOutput {
	out := make([]BinaryOutput, 4)
	for i := range out {
		out[i] = BinaryOutput{Quality: BinaryQualityOnline}
	}
	out[2].Value = true // DO3 was ON in the captured response
	return out
}
func (captureDataHandler) GetAnalogOutputs() []AnalogOutput { return nil }
func (captureDataHandler) GetFrozenCounters() []Counter     { return nil }
func (captureDataHandler) FreezeCounters(bool) error        { return nil }

func newCaptureOutstation(t *testing.T) *Outstation {
	t.Helper()
	o := NewOutstation(nil)
	o.Initialize()
	o.data = captureDataHandler{}
	return o
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	return b
}

// TestCapturedClass0123ResponseObjectBytes is a byte-exact regression for the
// Integrity poll captured from the Windows third-party master. Its request was
// G60V2/V3/V4/V1, qualifier 0x06 (all objects), in that order. The old encoder
// emitted four flags octets for G10V1, leaving three surplus octets that the
// master parsed as a further object header and reported as "object header at
// offset 38: app: truncated object data".
func TestCapturedClass0123ResponseObjectBytes(t *testing.T) {
	request := mustHex(t, "3c02063c03063c04063c0106")
	want := mustHex(t,
		"0101070800"+ // G1V1, count 8, packed bits
			"1e010704"+ // G30V1, count 4
			"0100000000"+"0100000000"+"0100000000"+"0100000000"+ // four analog values
			"0a010704"+ // G10V1, count 4
			"04") // packed binary output status: DO3 (index 2) on

	o := newCaptureOutstation(t)
	got := o.buildReadResponse(request)
	if !bytes.Equal(got, want) {
		t.Fatalf("response objects mismatch\n got = % X\n want= % X", got, want)
	}

	// The response must be walkable with the IEEE 1815 object sizes used by
	// the independent master: G1V1 and G10V1 are one bit each, G30V1 is 40
	// bits. A trailing header is what produced the reported truncation.
	if err := walkCapturedObjects(want); err != nil {
		t.Fatalf("strict object walk failed: %v", err)
	}
}

// walkCapturedObjects walks the captured response the way a strict master
// does and rejects any surplus octets after the last object.
func walkCapturedObjects(b []byte) error {
	const (
		g1v1Bits  = 1
		g10v1Bits = 1
		g30v1Bits = 40
	)
	for off := 0; off < len(b); {
		if len(b)-off < 4 {
			return &walkError{"short object header", off}
		}
		group, variation, qual, count := b[off], b[off+1], b[off+2], int(b[off+3])
		off += 4
		if qual != 0x07 {
			return &walkError{"unexpected qualifier", off - 1}
		}
		var bits int
		switch {
		case group == 1 && variation == 1:
			bits = g1v1Bits
		case group == 10 && variation == 1:
			bits = g10v1Bits
		case group == 30 && variation == 1:
			bits = g30v1Bits
		default:
			return &walkError{"unexpected object group", off - 4}
		}
		n := (count*bits + 7) / 8
		if off+n > len(b) {
			return &walkError{"truncated object data", off}
		}
		off += n
	}
	return nil
}

type walkError struct {
	reason string
	offset int
}

func (e *walkError) Error() string {
	return e.reason + " at offset " + itoa(e.offset)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// TestCapturedClass123EmptyEvents verifies that an explicit event-only class
// read returns an object-less response rather than falling back to static
// data or emitting a spurious header.
func TestCapturedClass123EmptyEvents(t *testing.T) {
	o := newCaptureOutstation(t)
	got := o.buildReadResponse(mustHex(t, "3c02063c03063c0406"))
	if len(got) != 0 {
		t.Fatalf("empty event classes must produce no objects, got % X", got)
	}
}

// TestG80V1WriteDoesNotPolluteLaterIIN reproduces the master's startup
// WRITE of G80V1 (clear DEVICE_RESTART) followed by an Integrity READ. The
// WRITE used to set a sticky IIN2.2 PARAMETER_ERROR that appeared in every
// later response.
func TestG80V1WriteDoesNotPolluteLaterIIN(t *testing.T) {
	o := newCaptureOutstation(t)
	o.iin.DeviceRestart = true

	writeResp, err := o.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncWrite,
		Data:     mustHex(t, "500100070700"), // G80V1, start=7 stop=7, value=0
	})
	if err != nil {
		t.Fatalf("G80V1 WRITE: %v", err)
	}
	if writeResp == nil {
		t.Fatal("G80V1 WRITE produced no response")
	}
	if o.iin.DeviceRestart {
		t.Fatal("G80V1 WRITE did not clear DEVICE_RESTART")
	}

	readResp, err := o.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncRead,
		Data:     mustHex(t, "3c02063c03063c04063c0106"),
	})
	if err != nil {
		t.Fatalf("READ after WRITE: %v", err)
	}
	if readResp == nil || len(readResp.Data) < 2 {
		t.Fatal("READ produced no response")
	}
	if readResp.Data[0] != 0 || readResp.Data[1] != 0 {
		t.Fatalf("IIN = % X, want 00 00", readResp.Data[:2])
	}
}

// TestStrictObjectWalkRejectsMalformedData pins the negative cases that a
// strict master must reject: a short object header, a packed block whose
// declared count exceeds the available octets, an unexpected qualifier, and
// surplus octets after the last complete object.
func TestStrictObjectWalkRejectsMalformedData(t *testing.T) {
	validG1V1 := mustHex(t, "0101070800")
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"short header", []byte{1, 1, 0x07}, "short object header"},
		{"truncated G1V1", mustHex(t, "01010702"), "truncated object data"},
		{"truncated G30V1", append(mustHex(t, "1e010704"), make([]byte, 19)...), "truncated object data"},
		{"unexpected qualifier", mustHex(t, "0101170800"), "unexpected qualifier"},
		{"surplus trailing byte", append(append([]byte{}, validG1V1...), 0x00), "short object header"},
		{"old defective G10V1", mustHex(t, "01010708001e01070401000000000100000000010000000001000000000a01070405058505"), "short object header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := walkCapturedObjects(tt.data)
			if err == nil {
				t.Fatalf("walk succeeded on malformed data % X", tt.data)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tt.want)) {
				t.Fatalf("error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}
