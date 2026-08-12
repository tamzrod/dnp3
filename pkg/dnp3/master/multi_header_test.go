package master

import (
	"context"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// buildMultiHeaderResponse builds a valid DLL+TL+APDU response frame carrying
// all three MVP static groups in ONE APDU (G1V1 packed, G20V1 counter, G30V1
// analog), each count8, mirroring how a real outstation may answer a Class-0
// read with a single multi-object-header response (MEXT-014 / R3).
func buildMultiHeaderResponse(seq uint8) []byte {
	obj := []byte{
		// G1V1 packed binary input, count8, 2 points -> 1 packed byte (0x03)
		0x01, 0x01, 0x07, 0x02,
		0x03,
		// G20V1 counter, count8, 1 point, 5 octets (uint32 + flags)
		0x14, 0x01, 0x07, 0x01,
		0x64, 0x00, 0x00, 0x00, // value 100
		0x01, // flags online
		// G30V1 analog input, count8, 1 point, 5 octets (int32 + flags)
		0x1E, 0x01, 0x07, 0x01,
		0x2A, 0x00, 0x00, 0x00, // value 42
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

// multiHeaderTransport echoes the request SEQ with a single multi-header
// response (G1 + G20 + G30 in one APDU).
type multiHeaderTransport struct{ lastSeq uint8 }

func (t *multiHeaderTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}
func (t *multiHeaderTransport) SetTimeout(ms int) {}
func (t *multiHeaderTransport) Receive() ([]byte, error) {
	return buildMultiHeaderResponse(t.lastSeq), nil
}

// TestReadMultiHeaderReturnsAllGroups asserts the public Read parses ALL
// MVP groups from a single multi-object-header response (G1+G20+G30 in one
// APDU) with no point loss (MEXT-014, fixing R3).
func TestReadMultiHeaderReturnsAllGroups(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &multiHeaderTransport{})
	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if len(resp.BinaryInputs) != 2 {
		t.Fatalf("BinaryInputs = %d, want 2 (multi-header parse lost G1 points)", len(resp.BinaryInputs))
	}
	if len(resp.Counters) != 1 {
		t.Fatalf("Counters = %d, want 1 (multi-header parse lost G20 points)", len(resp.Counters))
	}
	if len(resp.Counters) == 1 && resp.Counters[0].Value != 100 {
		t.Fatalf("Counters[0].Value = %v, want 100", resp.Counters[0].Value)
	}
	if len(resp.AnalogInputs) != 1 {
		t.Fatalf("AnalogInputs = %d, want 1 (multi-header parse lost G30 points)", len(resp.AnalogInputs))
	}
	if len(resp.AnalogInputs) == 1 && resp.AnalogInputs[0].Value != 42 {
		t.Fatalf("AnalogInputs[0].Value = %v, want 42", resp.AnalogInputs[0].Value)
	}
}
