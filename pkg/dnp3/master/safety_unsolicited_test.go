package master

import (
	"context"
	"errors"
	"testing"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
)

// TestUnsolicitedRejectedByDefault pins S1: the public v0 event/unsolicited
// entry points must reject with ErrUnsupportedOption by default, before any
// wire traffic, and must not silently register a handler. The published v0
// supported profile does not advertise events or unsolicited delivery.
func TestUnsolicitedRejectedByDefault(t *testing.T) {
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatal(err)
	}
	cc := c.(*client)

	for name, call := range map[string]func(context.Context) error{
		"enable":  c.EnableUnsolicited,
		"disable": c.DisableUnsolicited,
		"watch":   c.WatchUnsolicited,
	} {
		t.Run(name, func(t *testing.T) {
			err := call(context.Background())
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
		})
	}
	if _, err := c.ReadEvents(context.Background()); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("ReadEvents error = %v, want ErrUnsupportedOption", err)
	}

	// SetUnsolicitedHandler has no error return; it must refuse to register the
	// handler rather than silently enabling a prohibited capability.
	c.SetUnsolicitedHandler(func(*UnsolicitedResponse) {})
	if len(cc.handlers) != 0 {
		t.Fatalf("handlers registered = %d, want 0", len(cc.handlers))
	}
}

// TestUnsolicitedRejectedWhileConnected pins S1 for the connected state: the
// rejection is a profile gate, not merely a "not connected" check. With the
// default config, a connected client still rejects and sends no wire frames.
func TestUnsolicitedRejectedWhileConnected(t *testing.T) {
	tr := &fnCaptureTransport{}
	cc := newConnectedClientWithTransport(t, tr)

	for name, call := range map[string]func(context.Context) error{
		"enable":  cc.EnableUnsolicited,
		"disable": cc.DisableUnsolicited,
		"watch":   cc.WatchUnsolicited,
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(context.Background()); !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
		})
	}
	if _, err := cc.ReadEvents(context.Background()); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("ReadEvents error = %v, want ErrUnsupportedOption", err)
	}
	if len(tr.funcs) != 0 {
		t.Fatalf("wire function codes sent = %v, want none", tr.funcs)
	}
	if len(cc.handlers) != 0 {
		t.Fatalf("handlers registered = %d, want 0", len(cc.handlers))
	}
}

// TestUnsolicitedExperimentalPathReachesWire confirms the explicit, non-default
// experimental switch actually opens the EVT-001 path: enable/disable then
// reach the wire with the canonical function codes, and a handler is registered.
func TestUnsolicitedExperimentalPathReachesWire(t *testing.T) {
	tr := &fnCaptureTransport{}
	c, err := NewClient(NewConfig(WithExperimentalEventDelivery()))
	if err != nil {
		t.Fatal(err)
	}
	cc := c.(*client)
	cc.internalMaster.SetTransport(tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	if err := cc.EnableUnsolicited(context.Background()); err != nil {
		t.Fatalf("EnableUnsolicited: %v", err)
	}
	if err := cc.DisableUnsolicited(context.Background()); err != nil {
		t.Fatalf("DisableUnsolicited: %v", err)
	}
	if len(tr.funcs) != 2 || tr.funcs[0] != 0x14 || tr.funcs[1] != 0x15 {
		t.Fatalf("function codes = %v, want [0x14 0x15]", tr.funcs)
	}
	cc.SetUnsolicitedHandler(func(*UnsolicitedResponse) {})
	if len(cc.handlers) != 1 {
		t.Fatalf("handlers registered = %d, want 1", len(cc.handlers))
	}
}
