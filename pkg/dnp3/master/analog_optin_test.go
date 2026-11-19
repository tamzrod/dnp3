package master

import (
	"context"
	"errors"
	"testing"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// newConnectedStableAnalogClientWithTransport is
// newConnectedExperimentalAnalogClientWithTransport but using the stable
// WithAnalogControl() opt-in name, so the tests pin the stable path itself
// rather than only the legacy experimental alias.
func newConnectedStableAnalogClientWithTransport(t *testing.T, tr master.TransportHandler) *client {
	t.Helper()
	c, err := NewClient(NewConfig(WithAnalogControl()))
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

// TestWithAnalogControlEnablesG41 asserts the stable opt-in name admits the
// typed G41V1-V4 controls to the wire, exactly like the experimental alias.
func TestWithAnalogControlEnablesG41(t *testing.T) {
	cases := []struct {
		name string
		cmd  *types.ControlOutput
	}{
		{"G41V1 int32", types.NewAnalogInt32Control(0, 2300, types.DirectOperate)},
		{"G41V2 int16", types.NewAnalogInt16Control(0, 42, types.DirectOperate)},
		{"G41V3 float32", types.NewAnalogFloat32Control(0, 1.5, types.DirectOperate)},
		{"G41V4 float64", types.NewAnalogFloat64Control(0, 1.5, types.DirectOperate)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedStableAnalogClientWithTransport(t, tr)
			_, _ = cc.Operate(context.Background(), tc.cmd)
			if !tr.sent {
				t.Fatalf("%s with WithAnalogControl() was not transmitted", tc.name)
			}
		})
	}
}

// TestWithAnalogControlDefaultOff asserts that WithAnalogControl is off by
// default: a default client still refuses G41 with ErrUnsupportedOption and
// sends no wire traffic.
func TestWithAnalogControlDefaultOff(t *testing.T) {
	tr := &recordingPubTransport{}
	cc := newConnectedClientWithTransport(t, tr)
	_, err := cc.Operate(context.Background(), types.NewAnalogInt32Control(0, 1, types.DirectOperate))
	if !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("error = %v, want ErrUnsupportedOption", err)
	}
	if tr.sent {
		t.Fatalf("default client transmitted a G41 control")
	}
}

// TestLegacyAnalogConstructorRejectedEvenWhenOptedIn asserts the ambiguous
// NewAnalogControl(float64) stays rejected at Group 41 Variation 1 even with the
// opt-in on, and that nothing is transmitted. This pins the explicit
// carrier-type rejection so a future float->int32 coercion cannot silently admit
// a truncated legacy value.
func TestLegacyAnalogConstructorRejectedEvenWhenOptedIn(t *testing.T) {
	for _, tc := range []struct {
		name string
		tr   master.TransportHandler
		cc   func(*testing.T, master.TransportHandler) *client
	}{
		{"stable opt-in", &recordingPubTransport{}, newConnectedStableAnalogClientWithTransport},
		{"experimental opt-in", &recordingPubTransport{}, newConnectedExperimentalAnalogClientWithTransport},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := tc.tr.(*recordingPubTransport)
			cc := tc.cc(t, tr)
			resp, err := cc.Operate(context.Background(), types.NewAnalogControl(0, 1.5, types.DirectOperate))
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
			if resp != nil && resp.Status == types.ControlSuccess {
				t.Fatalf("legacy constructor surfaced as ControlSuccess")
			}
			if tr.sent {
				t.Fatalf("legacy ambiguous constructor was transmitted despite opt-in")
			}
		})
	}
}

// TestG41V4StillAdmitsGenericFloat64Carrier asserts the explicit V1 legacy
// rejection does not over-reach: the generic AnalogCommandValue{float64} remains
// a documented, exactly-representable source for V4 and is still transmitted.
func TestG41V4StillAdmitsGenericFloat64Carrier(t *testing.T) {
	tr := &recordingPubTransport{}
	cc := newConnectedStableAnalogClientWithTransport(t, tr)
	_, _ = cc.Operate(context.Background(), &types.ControlOutput{
		Group: 41, Variation: 4, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogCommandValue{Value: 1.0},
	})
	if !tr.sent {
		t.Fatalf("exactly-representable generic float64 G41V4 value was not transmitted")
	}
}

// TestWithAnalogControlConfigFlag asserts both option names set the same
// underlying config flag and that the default is false.
func TestWithAnalogControlConfigFlag(t *testing.T) {
	if DefaultConfig().ExperimentalAnalogControl {
		t.Fatalf("analog control must be off by default")
	}
	if !NewConfig(WithAnalogControl()).ExperimentalAnalogControl {
		t.Fatalf("WithAnalogControl did not enable the analog-control flag")
	}
	if !NewConfig(WithExperimentalAnalogControl()).ExperimentalAnalogControl {
		t.Fatalf("WithExperimentalAnalogControl did not enable the analog-control flag")
	}
}
