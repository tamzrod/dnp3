package master

import (
	"context"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// buildMVPSingleGroupResponse builds a valid DLL+TL+APDU response frame
// carrying exactly one MVP-supported object header, selected by group, so the
// corresponding public parser populates one point (DNP3-035). Using a single
// header per response avoids the legacy skipGroupData helper's packed/sequential
// byte-count assumptions, which are out of scope for this task.
func buildMVPSingleGroupResponse(seq uint8, group uint8) []byte {
	var obj []byte
	switch group {
	case 1: // G1V1 packed binary input, count8, 1 point
		obj = []byte{
			0x01, 0x01, 0x07, 0x01,
			0x01, // point 0 set
		}
	case 30: // G30V1 analog input, count8, 1 point, 5 octets
		obj = []byte{
			0x1E, 0x01, 0x07, 0x01,
			0x2A, 0x00, 0x00, 0x00, // value 42 (int32 LSB)
			0x01, // flags online
		}
	case 20: // G20V1 counter, count8, 1 point, 5 octets
		obj = []byte{
			0x14, 0x01, 0x07, 0x01,
			0x64, 0x00, 0x00, 0x00, // value 100 (uint32 LSB)
			0x01, // flags online
		}
	default:
		panic("unsupported MVP group in test helper")
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, obj...),
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// buildUnsupportedOnlyResponse builds a valid DLL+TL+APDU response frame
// carrying ONLY unsupported object headers (G10V1 binary output, G40V1 analog
// output). It proves the public Read does NOT surface them (DNP3-035).
func buildUnsupportedOnlyResponse(seq uint8) []byte {
	obj := []byte{
		// G10V1 binary output, qualifier 0x00 (index8), count=1, 3 octets/point
		0x0A, 0x01, 0x00, 0x01,
		0x00, 0x00, // index 0
		0x81, // value set (0x80) + online (0x01)
		// G40V1 analog output, qualifier 0x00 (index8), count=1, 7 octets/point
		0x28, 0x01, 0x00, 0x01,
		0x00, 0x00, // index 0
		0x00, 0x00, 0x80, 0x3F, // float32 1.0 (LSB)
		0x01, // flags online
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, obj...),
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// singleGroupTransport returns a canned response carrying one MVP group
// (echoing the request SEQ) for the public Read shape test (DNP3-035).
type singleGroupTransport struct {
	group  uint8
	lastSeq uint8
}

func (t *singleGroupTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *singleGroupTransport) SetTimeout(ms int) {}

func (t *singleGroupTransport) Receive() ([]byte, error) {
	return buildMVPSingleGroupResponse(t.lastSeq, t.group), nil
}

// unsupportedOnlyTransport returns a canned response carrying only
// unsupported object headers (echoing the request SEQ) (DNP3-035).
type unsupportedOnlyTransport struct {
	lastSeq uint8
}

func (t *unsupportedOnlyTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *unsupportedOnlyTransport) SetTimeout(ms int) {}

func (t *unsupportedOnlyTransport) Receive() ([]byte, error) {
	return buildUnsupportedOnlyResponse(t.lastSeq), nil
}

// TestReadResponsePopulatesMVPTypes asserts the public Read populates the
// MVP-supported slices (BinaryInputs, AnalogInputs, Counters) when the
// outstation returns each group's data (DNP3-035).
func TestReadResponsePopulatesMVPTypes(t *testing.T) {
	cases := []struct {
		group uint8
		check func(*testing.T, *ReadResponse)
	}{
		{1, func(t *testing.T, r *ReadResponse) {
			if len(r.BinaryInputs) != 1 {
				t.Fatalf("BinaryInputs = %d, want 1", len(r.BinaryInputs))
			}
		}},
		{30, func(t *testing.T, r *ReadResponse) {
			if len(r.AnalogInputs) != 1 {
				t.Fatalf("AnalogInputs = %d, want 1", len(r.AnalogInputs))
			}
			if r.AnalogInputs[0].Value != 42 {
				t.Fatalf("AnalogInputs[0].Value = %v, want 42", r.AnalogInputs[0].Value)
			}
		}},
		{20, func(t *testing.T, r *ReadResponse) {
			if len(r.Counters) != 1 {
				t.Fatalf("Counters = %d, want 1", len(r.Counters))
			}
			if r.Counters[0].Value != 100 {
				t.Fatalf("Counters[0].Value = %v, want 100", r.Counters[0].Value)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(groupName(tc.group), func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &singleGroupTransport{group: tc.group})
			resp, err := cc.Read(context.Background(), &types.ReadRequest{
				Groups: []types.GroupRequest{{Group: tc.group, Variation: 0}},
			})
			if err != nil {
				t.Fatalf("Read error: %v", err)
			}
			if resp == nil {
				t.Fatal("Read returned nil response")
			}
			tc.check(t, resp)
		})
	}
}

func groupName(g uint8) string {
	switch g {
	case 1:
		return "BinaryInput"
	case 30:
		return "AnalogInput"
	case 20:
		return "Counter"
	default:
		return "Unknown"
	}
}

// TestReadResponseDoesNotSurfaceUnsupportedTypes asserts the public Read leaves
// BinaryOutputs, AnalogOutputs, and FrozenCounters nil even when the
// outstation's response carries unsupported object headers (G10/G40), because
// the v0 MVP Read path does not parse those groups (DNP3-035).
func TestReadResponseDoesNotSurfaceUnsupportedTypes(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &unsupportedOnlyTransport{})

	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if resp == nil {
		t.Fatal("Read returned nil response")
	}

	if resp.BinaryOutputs != nil {
		t.Fatalf("BinaryOutputs = %v, want nil (not an MVP Read type)", resp.BinaryOutputs)
	}
	if resp.AnalogOutputs != nil {
		t.Fatalf("AnalogOutputs = %v, want nil (not an MVP Read type)", resp.AnalogOutputs)
	}
	if resp.FrozenCounters != nil {
		t.Fatalf("FrozenCounters = %v, want nil (not an MVP Read type)", resp.FrozenCounters)
	}
	// The MVP slices are empty (the response carried no MVP data), but must not
	// be populated with bogus points.
	if len(resp.BinaryInputs) != 0 || len(resp.AnalogInputs) != 0 || len(resp.Counters) != 0 {
		t.Fatalf("unsupported response populated MVP types: BI=%d AI=%d C=%d",
			len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))
	}
}

// TestReadResponseEmptyHasNoUnsupportedTypes asserts an empty (IIN-only)
// response leaves every unsupported slice nil and the MVP slices empty
// (DNP3-035).
func TestReadResponseEmptyHasNoUnsupportedTypes(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if resp == nil {
		t.Fatal("Read returned nil response")
	}
	if resp.BinaryOutputs != nil || resp.AnalogOutputs != nil || resp.FrozenCounters != nil {
		t.Fatalf("empty response surfaced unsupported types: BO=%v AO=%v FC=%v",
			resp.BinaryOutputs, resp.AnalogOutputs, resp.FrozenCounters)
	}
	if len(resp.BinaryInputs) != 0 || len(resp.AnalogInputs) != 0 || len(resp.Counters) != 0 {
		t.Fatalf("empty response populated MVP types: BI=%d AI=%d C=%d",
			len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))
	}
}
