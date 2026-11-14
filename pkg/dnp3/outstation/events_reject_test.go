package outstation

import (
	"errors"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/pkg/dnp3"
)

// TestOutstationEventsRejectedByDefault pins S1 on the outstation: with the
// default (non-experimental) configuration the public event/unsolicited entry
// points must reject with ErrUnsupportedOption and must not silently register a
// handler. The published v0 supported profile does not advertise events or
// unsolicited delivery.
func TestOutstationEventsRejectedByDefault(t *testing.T) {
	srv, err := NewServer(NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0),
	))
	if err != nil {
		t.Fatal(err)
	}
	s := srv.(*server)

	stamp := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := srv.PublishBinaryEvent(1, true, al.BinEventOnline, EventClass1, stamp); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("PublishBinaryEvent error = %v, want ErrUnsupportedOption", err)
	}
	if err := srv.PublishAnalogEvent(1, 42, al.AnalogEventOnline, EventClass2, stamp); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("PublishAnalogEvent error = %v, want ErrUnsupportedOption", err)
	}
	if err := srv.PublishCounterEvent(1, 7, al.CounterEventOnline, EventClass3, stamp); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("PublishCounterEvent error = %v, want ErrUnsupportedOption", err)
	}

	// No error return; it must refuse to register the handler rather than
	// silently activating a prohibited capability.
	srv.SetUnsolicitedHandler(noopUnsolicitedHandler{})
	if s.unsolicitedHandler != nil {
		t.Fatal("unsolicited handler was registered by default")
	}
}

// TestOutstationUnsolicitedModeRejectedByDefault pins S1: WithUnsolicitedMode
// without the explicit experimental switch is rejected by NewServer.
func TestOutstationUnsolicitedModeRejectedByDefault(t *testing.T) {
	_, err := NewServer(NewConfig(WithUnsolicitedMode(true)))
	if err == nil {
		t.Fatal("unsolicited mode must be rejected without the experimental switch")
	}
	var ce *dnp3.ConfigurationError
	if !errors.As(err, &ce) || ce.Field != "UnsolicitedMode" {
		t.Fatalf("error = %v, want ConfigurationError on UnsolicitedMode", err)
	}
}

// TestOutstationExperimentalPathEnablesEvents confirms the explicit,
// non-default experimental switch opens the EVT-001 outstation path: publishing
// succeeds, unsolicited mode validates, and a handler is registered.
func TestOutstationExperimentalPathEnablesEvents(t *testing.T) {
	srv, err := NewServer(NewConfig(
		WithAddress(1024),
		WithMasterAddress(1),
		WithTransport(dnp3.TCP, "127.0.0.1", 0),
		WithUnsolicitedMode(true),
		WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatalf("NewServer with experimental switch: %v", err)
	}
	s := srv.(*server)
	if err := srv.PublishBinaryEvent(1, true, al.BinEventOnline, EventClass1, time.Time{}); err != nil {
		t.Fatalf("PublishBinaryEvent: %v", err)
	}
	srv.SetUnsolicitedHandler(noopUnsolicitedHandler{})
	if s.unsolicitedHandler == nil {
		t.Fatal("handler was not registered on the experimental path")
	}
}

type noopUnsolicitedHandler struct{}

func (noopUnsolicitedHandler) OnUnsolicitedEnabled()  {}
func (noopUnsolicitedHandler) OnUnsolicitedDisabled() {}
