package master

import (
	"context"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/master"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// pubStatusEchoTransport records the SEQ of the most recent request and returns
// a DirectOperate response carrying a G12V1 object whose per-point command
// status byte is the given value. Used to drive the public client.Operate path.
type pubStatusEchoTransport struct {
	lastSeq       uint8
	commandStatus byte
}

func (t *pubStatusEchoTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubStatusEchoTransport) SetTimeout(ms int) {}

func (t *pubStatusEchoTransport) Receive() ([]byte, error) {
	return buildPubG12V1StatusResponse(t.lastSeq, t.commandStatus), nil
}

func extractPubRequestSeq(raw []byte) uint8 {
	f, err := frame.Decode(raw)
	if err != nil {
		return 0
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		return 0
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		return 0
	}
	return apdu.Control.Seq
}

func buildPubG12V1StatusResponse(seq uint8, commandStatus byte) []byte {
	obj := []byte{
		0x0C, 0x01, 0x00, 0x01,
		0x00, 0x00,
		0x07, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		commandStatus,
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

// newConnectedClientWithTransport builds a public client and substitutes its
// internal transport, then marks it connected so Operate can run without a
// real network.
func newConnectedClientWithTransport(t *testing.T, tr master.TransportHandler) *client {
	t.Helper()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.internalMaster.SetTransport(tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()
	return cc
}

// TestPublicOperateSurfacesStatus asserts the public OperateResponse carries
// the real per-point command status parsed from the response (DNP3-021): a
// success byte surfaces ControlSuccess, and a rejected byte surfaces the
// matching failure code — never ControlSuccess on a failure.
func TestPublicOperateSurfacesStatus(t *testing.T) {
	cases := []struct {
		name string
		byte byte
		want types.ControlStatus
	}{
		{"success", 0, types.ControlSuccess},
		{"not_supported", 4, types.ControlNotSupported},
		{"blocked", 6, types.ControlBlocked},
		{"no_select", 2, types.ControlNoSelect},
		{"timeout", 1, types.ControlTimeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: tc.byte})
			resp, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: 12, Variation: 1, Index: 0,
				CommandType: types.DirectOperate,
				Value:       &types.BinaryCommandValue{Value: true},
			})
			if err != nil {
				t.Fatalf("Operate error: %v", err)
			}
			if resp.Status != tc.want {
				t.Fatalf("Status = %v, want %v", resp.Status, tc.want)
			}
			if tc.byte != 0 && resp.Status == types.ControlSuccess {
				t.Fatalf("failed command surfaced as ControlSuccess")
			}
		})
	}
}

// TestPublicOperateMissingStatusNotSuccess confirms that a response with no
// G12V1 status byte (e.g. legacy empty DirectOperate response) is NOT surfaced
// as ControlSuccess.
func TestPublicOperateMissingStatusNotSuccess(t *testing.T) {
	// Build a transport returning an IIN-only response (no G12V1 object).
	cc := newConnectedClientWithTransport(t, &pubIINOnlyTransport{})
	resp, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("Operate error: %v", err)
	}
	if resp.Status == types.ControlSuccess {
		t.Fatalf("missing status byte surfaced as ControlSuccess; want non-success")
	}
}

// pubIINOnlyTransport echoes the SEQ with an IIN-only response (no object data).
type pubIINOnlyTransport struct{ lastSeq uint8 }

func (t *pubIINOnlyTransport) Send(data []byte) error  { t.lastSeq = extractPubRequestSeq(data); return nil }
func (t *pubIINOnlyTransport) SetTimeout(ms int)       {}
func (t *pubIINOnlyTransport) Receive() ([]byte, error) {
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00},
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw, nil
}
