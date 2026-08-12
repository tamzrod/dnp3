package master

import (
	"testing"
	"time"
)

// TestIdleMonitorClosesSessionToDisconnected verifies DNP3-042: when an idle
// timeout is configured and no activity occurs, the monitor transitions the
// session to Disconnected.
func TestIdleMonitorClosesSessionToDisconnected(t *testing.T) {
	m := NewMaster(&Config{IdleTimeout: 40}) // 40ms idle threshold
	m.SetState(StateActive)
	m.startIdleMonitor()
	defer m.stopIdleMonitor() // safety in case the assertion fails

	// Wait long enough for the monitor (ticks at ~20ms) to observe the idle
	// threshold and close the session.
	deadline := time.After(500 * time.Millisecond)
	for {
		if m.State() == StateDisconnected {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("state = %s, want Disconnected", m.State())
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// TestIdleMonitorActivityPreventsClose verifies DNP3-042: periodic activity
// (recordActivity) resets the idle timer so an otherwise-idle session is NOT
// closed.
func TestIdleMonitorActivityPreventsClose(t *testing.T) {
	m := NewMaster(&Config{IdleTimeout: 40}) // 40ms idle threshold
	m.SetState(StateActive)
	m.startIdleMonitor()
	defer m.stopIdleMonitor()

	// Tick activity faster than the idle threshold for a while.
	stop := time.After(200 * time.Millisecond)
loop:
	for {
		select {
		case <-stop:
			break loop
		default:
			m.recordActivity()
			time.Sleep(5 * time.Millisecond)
		}
	}
	if m.State() != StateActive {
		t.Fatalf("state = %s, want Active (activity should reset idle timer)", m.State())
	}
}

// TestIdleMonitorDisabledByDefault verifies DNP3-042: with no IdleTimeout
// configured, no monitor runs and the session stays Active indefinitely.
func TestIdleMonitorDisabledByDefault(t *testing.T) {
	m := NewMaster(nil) // default config: IdleTimeout == 0
	m.SetState(StateActive)
	m.startIdleMonitor() // no-op
	// No goroutine should be running.
	m.mu.RLock()
	stop := m.idleStop
	m.mu.RUnlock()
	if stop != nil {
		t.Fatal("idle monitor started despite IdleTimeout <= 0")
	}
	// Sleep past where a 0-timeout monitor would have fired; state unchanged.
	time.Sleep(30 * time.Millisecond)
	if m.State() != StateActive {
		t.Fatalf("state = %s, want Active (monitor disabled)", m.State())
	}
}

// TestIdleMonitorStopIsIdempotent verifies stopIdleMonitor can be called when no
// monitor is running without blocking or panicking (DNP3-042).
func TestIdleMonitorStopIsIdempotent(t *testing.T) {
	m := NewMaster(nil)
	m.stopIdleMonitor() // no-op, must not block
	m.stopIdleMonitor() // repeat
}

// TestIdleMonitorRestartStopsPrevious verifies re-Connecting stops the prior
// monitor before starting a new one (no leaked goroutines / double-close).
func TestIdleMonitorRestartStopsPrevious(t *testing.T) {
	m := NewMaster(&Config{IdleTimeout: 1000}) // long; won't fire during test
	m.SetState(StateActive)
	m.startIdleMonitor()
	m.mu.RLock()
	first := m.idleStop
	m.mu.RUnlock()
	m.startIdleMonitor() // re-Connect: must stop first
	m.mu.RLock()
	second := m.idleStop
	m.mu.RUnlock()
	if first == second {
		t.Fatal("re-Connect did not create a new monitor channel")
	}
	// The first channel must be closed.
	select {
	case <-first:
	default:
		t.Fatal("previous monitor channel not closed on restart")
	}
	m.stopIdleMonitor()
}
