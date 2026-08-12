package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/internal/testutils"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// recordingLogger records every LogEvent it receives. Safe for concurrent use.
type recordingLogger struct {
	mu      sync.Mutex
	events  []LogEvent
}

func (l *recordingLogger) Log(e LogEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
}

func (l *recordingLogger) snapshot() []LogEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LogEvent, len(l.events))
	copy(out, l.events)
	return out
}

// hasOp reports whether any recorded event has the given op.
func (l *recordingLogger) hasOp(op string) bool {
	for _, e := range l.snapshot() {
		if e.Op == op {
			return true
		}
	}
	return false
}

// TestLoggerDefaultSilent asserts that with no Logger configured, the master
// performs a Read without panicking and emits no events (DNP3-044 acceptance:
// default silent).
func TestLoggerDefaultSilent(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})
	// No logger configured; the client.logger is nil.
	cc.logger = nil

	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err != nil {
		t.Fatalf("Read error = %v, want nil (default logger must not affect behavior)", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// A nil logger must never be dereferenced; reaching here proves the default
	// is silent and safe.
}

// TestLoggerHookCalledOnRead verifies a configured Logger receives frame/seq
// events for a successful Read: at least a "send" and a "receive" event with
// the correct application sequence (DNP3-044 acceptance: hook called).
func TestLoggerHookCalledOnRead(t *testing.T) {
	log := &recordingLogger{}
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})
	cc.internalMaster.SetDiagnosticHook(diagAdapter(log))

	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err != nil {
		t.Fatalf("Read error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	events := log.snapshot()
	if len(events) == 0 {
		t.Fatal("expected at least one diagnostic event, got none")
	}
	if !log.hasOp("send") {
		t.Errorf("expected a 'send' event; events = %+v", events)
	}
	if !log.hasOp("receive") {
		t.Errorf("expected a 'receive' event; events = %+v", events)
	}

	// The send/receive events must carry the request's application seq (0 on the
	// first request of a fresh master).
	var sendEvt, recvEvt *LogEvent
	for i := range events {
		switch events[i].Op {
		case "send":
			sendEvt = &events[i]
		case "receive":
			recvEvt = &events[i]
		}
	}
	if sendEvt != nil && sendEvt.Seq != 0 {
		t.Errorf("send event seq = %d, want 0 (first request)", sendEvt.Seq)
	}
	if recvEvt != nil && recvEvt.Seq != 0 {
		t.Errorf("receive event seq = %d, want 0 (first request)", recvEvt.Seq)
	}
	if sendEvt != nil && sendEvt.Level != LogInfo {
		t.Errorf("send event level = %v, want LogInfo", sendEvt.Level)
	}
}

// TestLoggerHookCalledOnStateTransition verifies a configured Logger receives a
// "state" event when the master transitions state (DNP3-044). The public client
// emits a "connect" event on a successful Connect against the in-memory
// simulator transport, and the link handshake emits "state" events.
func TestLoggerHookCalledOnStateTransition(t *testing.T) {
	log := &recordingLogger{}
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	cfg := NewConfig(
		WithLogger(log),
		WithOutstationAddress(1024),
		WithTimeout(2*time.Second),
		WithRetry(1, 0),
	)
	c, err := NewClientWithTransport(cfg, sim)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	cc := c.(*client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := cc.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer cc.Close()

	events := log.snapshot()
	if !log.hasOp("connect") {
		t.Errorf("expected a 'connect' event; events = %+v", events)
	}
	if !log.hasOp("state") {
		t.Errorf("expected a 'state' event from the link handshake; events = %+v", events)
	}
}

// TestLoggerHookCalledOnFailure verifies a failure path raises warn/error
// diagnostic events (DNP3-044). A CRC failure must surface a "receive"/"retry"
// warn event and ultimately leave the logger populated.
func TestLoggerHookCalledOnFailure(t *testing.T) {
	log := &recordingLogger{}
	cc := newConnectedClientWithTransport(t, &badCRCTransport{})
	cc.internalMaster.SetDiagnosticHook(diagAdapter(log))

	_, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err == nil {
		t.Fatal("expected error for CRC failure")
	}
	if !errors.Is(err, dnp3.ErrCRC) {
		t.Fatalf("error = %v, want dnp3.ErrCRC in chain", err)
	}

	events := log.snapshot()
	if len(events) == 0 {
		t.Fatal("expected diagnostic events on CRC failure, got none")
	}
	// A CRC failure triggers a retry path; at least one warn/error-level event
	// must be present.
	sawFailure := false
	for _, e := range events {
		if e.Level == LogWarn || e.Level == LogError {
			sawFailure = true
			break
		}
	}
	if !sawFailure {
		t.Errorf("expected a warn/error event on CRC failure; events = %+v", events)
	}
}

// TestNopLoggerImplementsLogger asserts NopLogger satisfies the Logger
// interface and discards events silently (DNP3-044).
func TestNopLoggerImplementsLogger(t *testing.T) {
	var l Logger = NopLogger{}
	l.Log(LogEvent{Op: "send", Seq: 0}) // must not panic
}

// TestFuncLoggerForwards verifies FuncLogger adapts a function to the Logger
// interface and returns a no-op logger for nil (DNP3-044).
func TestFuncLoggerForwards(t *testing.T) {
	var got []LogEvent
	l := FuncLogger(func(e LogEvent) { got = append(got, e) })
	l.Log(LogEvent{Op: "send", Seq: 3})
	if len(got) != 1 || got[0].Op != "send" || got[0].Seq != 3 {
		t.Fatalf("FuncLogger forwarding = %+v, want one send/seq=3 event", got)
	}

	// nil function -> no-op logger (must not panic).
	nop := FuncLogger(nil)
	nop.Log(LogEvent{Op: "send"})
}

// TestLogLevelString verifies the human-readable names (DNP3-044).
func TestLogLevelString(t *testing.T) {
	cases := []struct {
		level LogLevel
		want  string
	}{
		{LogInfo, "info"},
		{LogWarn, "warn"},
		{LogError, "error"},
		{LogLevel(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}

// TestLoggerDefaultSilentDisconnect ensures the no-logger path is silent on a
// disconnect too (no nil-deref through the diag path).
func TestLoggerDefaultSilentDisconnect(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &peerCloseTransport{})
	cc.logger = nil
	_, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err == nil {
		t.Fatal("expected error on peer close")
	}
	// Reaching here without panicking proves the default no-logger path is safe
	// through the disconnect/retry diag emissions.
}
