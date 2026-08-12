package testutils

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// TestSimulatorHandshakeResponses verifies the simulator answers the link-layer
// handshake deterministically: Reset Link Stations → ACK, Request Link Status →
// Link Status, addressed from the outstation to the master (DNP3-006/007/036).
func TestSimulatorHandshakeResponses(t *testing.T) {
	sim := NewMVPOutstationSimulator(1024, 0xFFFF)

	// Reset Link Stations.
	reset := &frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncResetLinkStations},
		DestAddr: 1024, SrcAddr: 0xFFFF,
	}
	resetRaw, _ := frame.Encode(reset)
	if err := sim.Send(resetRaw); err != nil {
		t.Fatalf("send reset: %v", err)
	}
	ackRaw, err := sim.Receive()
	if err != nil {
		t.Fatalf("receive ACK: %v", err)
	}
	ack, err := frame.Decode(ackRaw)
	if err != nil {
		t.Fatalf("decode ACK: %v", err)
	}
	if ack.Control.DIR || ack.Control.PRM {
		t.Fatalf("ACK DIR/PRM = %v/%v, want false/false", ack.Control.DIR, ack.Control.PRM)
	}
	if ack.Control.FuncCode != frame.FuncAck {
		t.Fatalf("ACK func = %d, want %d", ack.Control.FuncCode, frame.FuncAck)
	}
	if ack.SrcAddr != 1024 || ack.DestAddr != 0xFFFF {
		t.Fatalf("ACK src/dst = %d/%d, want 1024/0xFFFF", ack.SrcAddr, ack.DestAddr)
	}

	// Request Link Status.
	rls := &frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncRequestLinkStatus},
		DestAddr: 1024, SrcAddr: 0xFFFF,
	}
	rlsRaw, _ := frame.Encode(rls)
	if err := sim.Send(rlsRaw); err != nil {
		t.Fatalf("send RLS: %v", err)
	}
	lsRaw, err := sim.Receive()
	if err != nil {
		t.Fatalf("receive LinkStatus: %v", err)
	}
	ls, err := frame.Decode(lsRaw)
	if err != nil {
		t.Fatalf("decode LinkStatus: %v", err)
	}
	if ls.Control.FuncCode != frame.FuncLinkStatus {
		t.Fatalf("LinkStatus func = %d, want %d", ls.Control.FuncCode, frame.FuncLinkStatus)
	}
	if ls.SrcAddr != 1024 || ls.DestAddr != 0xFFFF {
		t.Fatalf("LinkStatus src/dst = %d/%d, want 1024/0xFFFF", ls.SrcAddr, ls.DestAddr)
	}
}

// TestSimulatorReadResponseEncodesGoldenData verifies the simulator's Read
// response carries golden G1/G20/G30 data with count8 qualifiers and the
// request's application SEQ (DNP3-010/036).
func TestSimulatorReadResponseEncodesGoldenData(t *testing.T) {
	sim := NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetBinaryInputs([]*types.BinaryInput{{Index: 0, Value: true, Quality: types.QualityOnline}})
	sim.SetAnalogInputs([]*types.AnalogInput{{Index: 0, Value: 42, Quality: types.QualityOnline}})
	sim.SetCounters([]*types.Counter{{Index: 0, Value: 100, Quality: types.QualityOnline}})

	for _, tc := range []struct {
		name  string
		group uint8
	}{
		{"G1", 1}, {"G20", 20}, {"G30", 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := buildSimReadRequest(tc.group, 7)
			raw, _ := frame.Encode(req)
			if err := sim.Send(raw); err != nil {
				t.Fatalf("send: %v", err)
			}
			respRaw, err := sim.Receive()
			if err != nil {
				t.Fatalf("receive: %v", err)
			}
			resp, err := frame.Decode(respRaw)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Control.FuncCode != frame.FuncConfirmedUserDataR {
				t.Fatalf("DLL func = %d, want %d", resp.Control.FuncCode, frame.FuncConfirmedUserDataR)
			}
			frag, err := tl.DecodeFragment(resp.Data)
			if err != nil {
				t.Fatalf("decode TL: %v", err)
			}
			apdu, err := al.DecodeResponse(frag.Data)
			if err != nil {
				t.Fatalf("decode APDU: %v", err)
			}
			if apdu.Header.Control.Seq != 7 {
				t.Fatalf("SEQ = %d, want 7 (echo)", apdu.Header.Control.Seq)
			}
			if len(apdu.Data) == 0 {
				t.Fatalf("no object data for %s", tc.name)
			}
			if apdu.Data[0] != tc.group {
				t.Fatalf("object group = %d, want %d", apdu.Data[0], tc.group)
			}
		})
	}
}

// TestSimulatorOperateResponseEchoesStatus verifies the simulator echoes the
// request's G12V1 object with the configured command status replacing the
// CROB status byte (DNP3-020/021/036).
func TestSimulatorOperateResponseEchoesStatus(t *testing.T) {
	sim := NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetCommandStatus(types.ControlBlocked)

	// Build a DirectOperate request: G12V1, qualifier 0x00, count 1, index 0,
	// 11-byte CROB (status byte last).
	crob := []byte{
		0x0C, 0x01, 0x00, 0x01, // group, var, qualifier, count
		0x00, 0x00, // index
		0x03, 0x01, // code, count
		0x00, 0x00, 0x00, 0x00, // onTime
		0x00, 0x00, 0x00, 0x00, // offTime
		0x00, // request status (will be replaced)
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 3},
		FuncCode: al.FuncDirectOperate,
		Data:     crob,
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	req := &frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
		DestAddr: 1024, SrcAddr: 0xFFFF, Data: tlData,
	}
	raw, _ := frame.Encode(req)
	if err := sim.Send(raw); err != nil {
		t.Fatalf("send: %v", err)
	}
	respRaw, err := sim.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	resp, _ := frame.Decode(respRaw)
	rfrag, _ := tl.DecodeFragment(resp.Data)
	rApdu, _ := al.DecodeResponse(rfrag.Data)

	if rApdu.Header.Control.Seq != 3 {
		t.Fatalf("SEQ = %d, want 3 (echo)", rApdu.Header.Control.Seq)
	}
	// Status byte is header(4) + index(2) + 10 within the object data.
	statusIdx := 4 + 2 + 10
	if statusIdx >= len(rApdu.Data) {
		t.Fatalf("response too short: %d bytes", len(rApdu.Data))
	}
	if got := rApdu.Data[statusIdx]; got != byte(types.ControlBlocked) {
		t.Fatalf("status byte = %d, want %d (Blocked)", got, byte(types.ControlBlocked))
	}
}

// TestSimulatorRecordsSentFrames verifies the simulator records the frames the
// master sent for test assertions (DNP3-036).
func TestSimulatorRecordsSentFrames(t *testing.T) {
	sim := NewMVPOutstationSimulator(1024, 0xFFFF)
	reset := &frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncResetLinkStations},
		DestAddr: 1024, SrcAddr: 0xFFFF,
	}
	raw, _ := frame.Encode(reset)
	_ = sim.Send(raw)

	sent := sim.SentFrames()
	if len(sent) != 1 {
		t.Fatalf("SentFrames = %d, want 1", len(sent))
	}
	if sent[0].Control.FuncCode != frame.FuncResetLinkStations {
		t.Fatalf("recorded func = %d, want ResetLinkStations", sent[0].Control.FuncCode)
	}
}

// buildSimReadRequest builds a Confirmed User Data frame carrying a Read
// request for the given group with the all-objects (0x06) qualifier and SEQ.
func buildSimReadRequest(group uint8, seq uint8) *frame.Frame {
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncRead,
		Data:     []byte{group, 0x00, 0x06, 0x00}, // group, variation 0, all-objects
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	return &frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
		DestAddr: 1024, SrcAddr: 0xFFFF, Data: tlData,
	}
}
