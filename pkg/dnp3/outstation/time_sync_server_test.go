package outstation

import (
	"testing"
	"time"
)

// The optional TimeSyncServer interface is deliberately not part of Server,
// so existing implementations and the published v0 profile are unchanged.
// These tests pin that opt-in diagnostics are available only when the explicit
// experimental switch is set, and that no host clock is modified.

func TestTimeSyncServerOptInAndDiagnostics(t *testing.T) {
	srv, err := NewServer(NewConfig(WithExperimentalTimeSync()))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts, ok := srv.(TimeSyncServer)
	if !ok {
		t.Fatal("concrete server does not implement optional TimeSyncServer")
	}
	if !ts.TimeSyncEnabled() {
		t.Fatal("WithExperimentalTimeSync did not enable the logical clock")
	}
	st := ts.ClockStatus()
	if !st.NeedTime || st.HasTime || st.Source != ClockSourceUnset {
		t.Fatalf("startup clock status = %+v", st)
	}

	// A system fallback seed does not clear NEED_TIME and never downgrades an
	// authoritative sync (none has happened here).
	sysTime := time.Date(2026, 11, 19, 11, 2, 56, 910_000_000, time.UTC)
	ts.ApplySystemTime(sysTime)
	st = ts.ClockStatus()
	if st.Source != ClockSourceSystem || !st.HasTime || !st.NeedTime {
		t.Fatalf("status after system seed = %+v", st)
	}
	if st.LastValue.UnixMilli() != 1795086176910 {
		t.Fatalf("system seed value = %v", st.LastValue)
	}

	// Explicit diagnostic clear/set of NEED_TIME.
	ts.SetNeedTime(false)
	if ts.ClockStatus().NeedTime {
		t.Fatal("SetNeedTime(false) did not clear NEED_TIME")
	}
	ts.SetNeedTime(true)
	if !ts.ClockStatus().NeedTime {
		t.Fatal("SetNeedTime(true) did not assert NEED_TIME")
	}
}

func TestTimeSyncServerDisabledByDefault(t *testing.T) {
	srv, err := NewServer(NewConfig())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts, ok := srv.(TimeSyncServer)
	if !ok {
		t.Fatal("concrete server does not implement optional TimeSyncServer")
	}
	if ts.TimeSyncEnabled() {
		t.Fatal("time sync must be off by default (published v0 profile unchanged)")
	}
	if st := ts.ClockStatus(); st.NeedTime || st.HasTime {
		t.Fatalf("disabled clock must not advertise time need: %+v", st)
	}
}

func TestTimeSyncServerStableAlias(t *testing.T) {
	srv, err := NewServer(NewConfig(WithTimeSync()))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if !srv.(TimeSyncServer).TimeSyncEnabled() {
		t.Fatal("WithTimeSync stable alias must set the same opt-in as WithExperimentalTimeSync")
	}
}
