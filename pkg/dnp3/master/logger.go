package master

import (
	"dnp3/internal/master"
)

// LogLevel is the severity of a public diagnostic event (DNP3-044).
type LogLevel int

const (
	// LogInfo is for routine frame/sequence lifecycle events (send, receive,
	// confirm, state transitions).
	LogInfo LogLevel = iota
	// LogWarn is for recoverable anomalies (retry, sequence mismatch, NACK).
	LogWarn
	// LogError is for failures (timeout, CRC, disconnect, illegal transition).
	LogError
)

// String returns a human-readable name for the level.
func (l LogLevel) String() string {
	switch l {
	case LogInfo:
		return "info"
	case LogWarn:
		return "warn"
	case LogError:
		return "error"
	default:
		return "unknown"
	}
}

// LogEvent is a structured diagnostic event for master frame/sequence
// operations (DNP3-044). It is delivered to the configured [Logger]; the
// default is a no-op logger (silent).
type LogEvent struct {
	// Level is the event severity.
	Level LogLevel
	// Op names the operation, e.g. "send", "receive", "confirm", "retry",
	// "sequence", "state".
	Op string
	// Seq is the application-layer sequence number (0-15) for the event, or
	// the sentinel SeqNA when not applicable.
	Seq uint8
	// Msg is a short human-readable description.
	Msg string
	// Err carries the underlying error for failure events, or nil.
	Err error
}

// Logger is an optional diagnostic sink for master frame/sequence events
// (DNP3-044). Implementations must be safe for concurrent use; the master calls
// Log synchronously but never while holding its own locks, so callbacks may
// safely query client/master state. The default logger is silent (no-op).
type Logger interface {
	Log(e LogEvent)
}

// NopLogger is a no-op Logger that discards every event (the default).
type NopLogger struct{}

// Log implements Logger; it does nothing.
func (NopLogger) Log(LogEvent) {}

// SeqNA marks a LogEvent with no applicable sequence number (mirrors
// master.SeqNA).
const SeqNA uint8 = 0xFF

// loggerFunc adapts an ordinary function to the Logger interface.
type loggerFunc func(LogEvent)

func (f loggerFunc) Log(e LogEvent) { f(e) }

// FuncLogger returns a Logger that forwards each event to the given function.
// Passing nil returns a no-op logger. This is a convenience for tests and
// short handlers.
func FuncLogger(f func(LogEvent)) Logger {
	if f == nil {
		return NopLogger{}
	}
	return loggerFunc(f)
}

// diagAdapter adapts a public Logger to the internal master.DiagHook so events
// raised inside the internal master are surfaced through the public Logger
// (DNP3-044).
func diagAdapter(l Logger) master.DiagHook {
	if l == nil {
		return nil
	}
	return func(e master.DiagEvent) {
		l.Log(LogEvent{
			Level: mapLogLevel(e.Level),
			Op:    e.Op,
			Seq:   e.Seq,
			Msg:   e.Msg,
			Err:   e.Err,
		})
	}
}

// mapLogLevel maps the internal DiagLevel to the public LogLevel.
func mapLogLevel(l master.DiagLevel) LogLevel {
	switch l {
	case master.DiagInfo:
		return LogInfo
	case master.DiagWarn:
		return LogWarn
	case master.DiagError:
		return LogError
	default:
		return LogInfo
	}
}

// emitLog is a nil-safe helper used by the public client to raise events that
// originate above the internal master (e.g. Connect). It avoids a nil-deref
// when no logger is configured (the default).
func (c *client) emitLog(level LogLevel, op, msg string, err error) {
	if c.logger == nil {
		return
	}
	c.logger.Log(LogEvent{Level: level, Op: op, Seq: SeqNA, Msg: msg, Err: err})
}
