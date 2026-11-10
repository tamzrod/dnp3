package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// waitForCond polls cond until it is true or the deadline elapses. It is a
// finite watchdog that only prevents a hung test; it does not introduce
// scheduler-dependent ordering.
func waitForCond(cond func() bool, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}

// gatedSendTransport is a controllable physical send: the first Send is paused
// AFTER admission (beginTransmit) but BEFORE the physical write, then released
// explicitly. It records the number of physical sends so a test can prove a
// fenced old-session worker is not redirected onto a replacement session. It
// opens no socket.
type gatedSendTransport struct {
	mu       sync.Mutex
	sends    int
	entered  chan struct{}
	release  chan struct{}
	once     sync.Once
	closed   bool
	closeMtx sync.Mutex
}

func newGatedSendTransport() *gatedSendTransport {
	return &gatedSendTransport{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (t *gatedSendTransport) Send(data []byte) error {
	t.once.Do(func() { close(t.entered) })
	<-t.release
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return nil
}
func (t *gatedSendTransport) SetTimeout(ms int) {}
func (t *gatedSendTransport) Receive() ([]byte, error) {
	return nil, errors.New("no response")
}
func (t *gatedSendTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}
func (t *gatedSendTransport) releaseSend() {
	t.closeMtx.Lock()
	defer t.closeMtx.Unlock()
	if !t.closed {
		t.closed = true
		close(t.release)
	}
}

// TestControlOldSessionSendNotContaminatingReplacement (VEC-01 B4 F3-R4) proves
// the admitted-send physical-session isolation precisely: a control admitted in
// session A and paused BEFORE its physical Send is fenced as unresolved by a
// Disconnect, so a replacement session is refused and the fenced worker cannot
// be redirected onto one. Because the send had been admitted (outcome not
// provably NOT_SENT), the worker's eventual write is confined to transport A
// (exactly one physical send) and the outstation stays quarantined.
func TestControlOldSessionSendNotContaminatingReplacement(t *testing.T) {
	trA := newGatedSendTransport()
	m := NewMaster(DefaultConfig())
	m.SetTransport(trA)
	m.SetState(StateInitialized)
	m.AddOutstation(1, "os1")

	done := make(chan error, 1)
	go func() {
		_, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
		done <- err
	}()

	// Wait until the worker is admitted and paused at the physical-send boundary.
	select {
	case <-trA.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	// Tear down session A while the send is admitted but paused pre-write. The
	// worker is fenced as unresolved; the outstation is quarantined because the
	// send had been admitted (outcome not provably NOT_SENT on the wire).
	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !m.sessionUnresolvedFor(1) {
		t.Fatal("Disconnect must fence an admitted pre-write send as unresolved")
	}

	// A replacement session B must be refused while the fence holds: no old
	// worker can be redirected onto B.
	if err := m.ConnectContext(context.Background(), time.Second); !errors.Is(err, ErrSessionUnresolved) {
		t.Fatalf("ConnectContext while unresolved = %v, want ErrSessionUnresolved", err)
	}

	// Release the paused send: the worker completes on transport A, terminates,
	// and the fence lifts.
	trA.releaseSend()
	if err := <-done; err == nil {
		t.Fatal("a fenced pre-write control must not report success")
	}
	if !waitForCond(func() bool { return !m.anySessionUnresolved() }, time.Second) {
		t.Fatal("unresolved fence must lift once the worker terminates")
	}
	if got := trA.sendCount(); got != 1 {
		t.Fatalf("the old worker's physical send on transport A = %d, want 1", got)
	}

	// The UNKNOWN quarantine persists after the fence lifts (distinct concepts).
	if !m.IsQuarantined(1) {
		t.Fatal("the UNKNOWN quarantine must persist independently of the lifted fence")
	}
}

// TestControlQuarantineAdmissionBothOrdersNoSecondSend (VEC-01 B4 F3-R4) forces
// both quarantine/control-admission linearization orders deterministically and
// proves no admitted control can bypass a quarantine established after its
// admission (and that a quarantined outstation refuses a fresh admission),
// with at most one physical send in either order.
func TestControlQuarantineAdmissionBothOrdersNoSecondSend(t *testing.T) {
	// Order 1: admission THEN quarantine → the pre-send safety check refuses the
	// admitted control, zero sends.
	t.Run("admission_then_quarantine", func(t *testing.T) {
		tr := &failReceiveTransport{}
		m := newControlMaster(t, tr)
		gen, release, err := m.AcquireControlGate(1)
		if err != nil {
			t.Fatalf("AcquireControlGate: %v", err)
		}
		defer release()
		m.quarantineOutstation(1)
		if _, err := m.OperateWithStatusGated(context.Background(), gen, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second)); !errors.Is(err, ErrControlQuarantined) {
			t.Fatalf("admitted-then-quarantined control = %v, want ErrControlQuarantined", err)
		}
		if got := tr.sendCount(); got != 0 {
			t.Fatalf("admitted-then-quarantined control transmitted; sends = %d, want 0", got)
		}
	})

	// Order 2: quarantine THEN admission → a fresh admission is refused, zero
	// sends.
	t.Run("quarantine_then_admission", func(t *testing.T) {
		tr := &failReceiveTransport{}
		m := newControlMaster(t, tr)
		m.quarantineOutstation(1)
		if _, _, err := m.AcquireControlGate(1); !errors.Is(err, ErrControlQuarantined) {
			t.Fatalf("admission after quarantine = %v, want ErrControlQuarantined", err)
		}
		if _, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(time.Second)); !errors.Is(err, ErrControlQuarantined) {
			t.Fatalf("control after quarantine = %v, want ErrControlQuarantined", err)
		}
		if got := tr.sendCount(); got != 0 {
			t.Fatalf("quarantine-then-admission transmitted; sends = %d, want 0", got)
		}
	})
}

// TestControlDisconnectBlockedSendReplacementRefused (VEC-01 B4 F3-R4) proves a
// blocked physical Send plus a concurrent Disconnect leaves the unresolved fence
// (and persistent UNKNOWN) correctly distinct: a replacement Connect is refused
// while the send is blocked, and the fence lifts only after the worker
// terminates.
func TestControlDisconnectBlockedSendReplacementRefused(t *testing.T) {
	tr := newBlockingSendTransport()
	m := newControlMaster(t, tr)

	done := make(chan error, 1)
	go func() {
		_, err := m.OperateWithStatusContext(context.Background(), 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
		done <- err
	}()
	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !m.sessionUnresolvedFor(1) {
		t.Fatal("Disconnect during a blocked send must fence the outstation unresolved")
	}
	if !m.IsQuarantined(1) {
		t.Fatal("Disconnect during a blocked send must quarantine (UNKNOWN) the outstation")
	}
	if err := m.ConnectContext(context.Background(), time.Second); !errors.Is(err, ErrSessionUnresolved) {
		t.Fatalf("replacement Connect while send blocked = %v, want ErrSessionUnresolved", err)
	}

	tr.release()
	if r := <-done; !errors.Is(r, ErrControlOutcomeUnknown) {
		t.Fatalf("worker result err = %v, want ErrControlOutcomeUnknown", r)
	}
	if !waitForCond(func() bool { return !m.anySessionUnresolved() }, time.Second) {
		t.Fatal("unresolved fence must lift once the worker terminates")
	}
	if !m.IsQuarantined(1) {
		t.Fatal("the UNKNOWN quarantine must persist after the fence lifts")
	}
}
