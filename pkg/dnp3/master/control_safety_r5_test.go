package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
)

// closeGateDialCountTransport models an uncooperative custom transport whose
// Close blocks until released and which records how many times its dial
// (Connect) and Close were invoked. It opens no socket, so the tests stay
// offline. It is used to prove the F2-R5/F3-R5 atomic teardown-registration
// invariant: a fresh Connect must never dial while an old session's teardown
// fence is held.
type closeGateDialCountTransport struct {
	mu           sync.Mutex
	dials        int
	closes       int
	closeEntered chan struct{}
	closeRelease chan struct{}
	closeOnce    sync.Once
}

func newCloseGateDialCountTransport() *closeGateDialCountTransport {
	return &closeGateDialCountTransport{
		closeEntered: make(chan struct{}),
		closeRelease: make(chan struct{}),
	}
}

func (t *closeGateDialCountTransport) Listen() error          { return nil }
func (t *closeGateDialCountTransport) Accept() error          { return nil }
func (t *closeGateDialCountTransport) Send(data []byte) error { return nil }
func (t *closeGateDialCountTransport) SetTimeout(ms int)      {}
func (t *closeGateDialCountTransport) Receive() ([]byte, error) {
	return nil, errors.New("no response")
}
func (t *closeGateDialCountTransport) Connect() error {
	t.mu.Lock()
	t.dials++
	t.mu.Unlock()
	return nil
}
func (t *closeGateDialCountTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closeEntered) })
	<-t.closeRelease
	t.mu.Lock()
	t.closes++
	t.mu.Unlock()
	return nil
}
func (t *closeGateDialCountTransport) dialCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dials
}
func (t *closeGateDialCountTransport) closeCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closes
}

// TestDisconnectAtomicTeardownFenceNoAdmissionWindow (VEC-01 B4 F2-R5) proves
// the teardown fence is published atomically with the Disconnected transition:
// while the old session's Close is blocked, a fresh Connect is never admitted
// and never dials, even under concurrent admission attempts. Before the atomic
// fix, a Connect could claim the shared transport in the window between the
// Disconnected publication and the teardown registration.
func TestDisconnectAtomicTeardownFenceNoAdmissionWindow(t *testing.T) {
	tr := newCloseGateDialCountTransport()
	cc := clientWithTransport(t, tr)
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	dErr := make(chan error, 1)
	go func() { dErr <- cc.Disconnect(context.Background()) }()

	// Hammer Connect concurrently with the Disconnect admission boundary. Before
	// Disconnect linearizes the state is Active (refused as "already connected");
	// once it publishes Disconnected the fence is already registered, so no
	// attempt may ever be admitted.
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_ = cc.Connect(context.Background())
			}
		}()
	}

	// The teardown must actually enter (and block in) Close.
	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("teardown Close was never entered")
	}

	// Bounded window to let the concurrent attempts run against the boundary.
	time.Sleep(50 * time.Millisecond)

	// No attempt may have dialed: the fence was registered atomically with the
	// Disconnected publication, so there is no admission window.
	if n := tr.dialCount(); n != 0 {
		t.Fatalf("Connect dialed %d time(s) while the old session teardown was blocked", n)
	}

	// From here the state is deterministically Disconnected: every attempt must
	// be refused fail-closed.
	for i := 0; i < 200; i++ {
		if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
			t.Fatalf("Connect during blocked teardown = %v, want ErrSessionUnresolved", err)
		}
	}
	if n := tr.dialCount(); n != 0 {
		t.Fatalf("Connect dialed %d time(s) during blocked teardown", n)
	}

	close(stop)
	wg.Wait()

	// Release the blocked Close: the teardown terminates and the fence clears.
	close(tr.closeRelease)
	if err := <-dErr; err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}
}

// TestMarkTransportDisconnectedAtomicFence (VEC-01 B4 F2-R5) proves the
// transport-drop teardown path publishes its fence atomically too: while the
// dropped session's Close is blocked, a replacement Connect is refused and does
// not dial.
func TestMarkTransportDisconnectedAtomicFence(t *testing.T) {
	tr := newCloseGateDialCountTransport()
	cc := clientWithTransport(t, tr)
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	cc.markTransportDisconnected()

	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}
	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("teardown Close was never entered")
	}
	for i := 0; i < 200; i++ {
		if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
			t.Fatalf("Connect during blocked drop teardown = %v, want ErrSessionUnresolved", err)
		}
	}
	if n := tr.dialCount(); n != 0 {
		t.Fatalf("Connect dialed %d time(s) during blocked drop teardown", n)
	}

	close(tr.closeRelease)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}
}

// TestConcurrentTeardownOwnersSingleClose (VEC-01 B4 F2-R5) proves there is
// exactly ONE teardown owner: concurrent Disconnect calls and the transport-drop
// teardown path join a single registered teardown instead of each running its
// own Close.
func TestConcurrentTeardownOwnersSingleClose(t *testing.T) {
	tr := newCloseGateDialCountTransport()
	cc := clientWithTransport(t, tr)
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	const n = 8
	var started sync.WaitGroup
	started.Add(n + 1)
	barrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n + 1)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			started.Done()
			<-barrier
			_ = cc.Disconnect(context.Background())
		}()
	}
	go func() {
		defer wg.Done()
		started.Done()
		<-barrier
		cc.markTransportDisconnected()
	}()

	started.Wait()
	close(barrier)

	// The single teardown must enter Close and stay blocked until released.
	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("teardown Close was never entered")
	}
	// Bounded window so every participant reaches the shared teardown.
	time.Sleep(50 * time.Millisecond)

	close(tr.closeRelease)
	wg.Wait()

	if got := tr.closeCount(); got != 1 {
		t.Fatalf("Close called %d times, want exactly 1 (single teardown owner)", got)
	}
}

// firstDialGateTransport blocks the FIRST dial until released, lets later dials
// succeed, and blocks Close until released, counting both dials and closes. It
// opens no socket. It lets a test force abortConnect (a Connect whose blocked
// dial exhausts its budget) while concurrently hammering a replacement Connect.
type firstDialGateTransport struct {
	mu           sync.Mutex
	dials        int
	closes       int
	dial1Entered chan struct{}
	dial1Release chan struct{}
	dial1Once    sync.Once
	closeEntered chan struct{}
	closeRelease chan struct{}
	closeOnce    sync.Once
}

func newFirstDialGateTransport() *firstDialGateTransport {
	return &firstDialGateTransport{
		dial1Entered: make(chan struct{}),
		dial1Release: make(chan struct{}),
		closeEntered: make(chan struct{}),
		closeRelease: make(chan struct{}),
	}
}

func (t *firstDialGateTransport) Listen() error            { return nil }
func (t *firstDialGateTransport) Accept() error            { return nil }
func (t *firstDialGateTransport) Send(data []byte) error   { return nil }
func (t *firstDialGateTransport) SetTimeout(ms int)        {}
func (t *firstDialGateTransport) Receive() ([]byte, error) { return nil, errors.New("no response") }
func (t *firstDialGateTransport) Connect() error {
	t.mu.Lock()
	t.dials++
	n := t.dials
	t.mu.Unlock()
	if n == 1 {
		t.dial1Once.Do(func() { close(t.dial1Entered) })
		<-t.dial1Release
		return errors.New("dial aborted")
	}
	return nil
}
func (t *firstDialGateTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closeEntered) })
	<-t.closeRelease
	t.mu.Lock()
	t.closes++
	t.mu.Unlock()
	return nil
}
func (t *firstDialGateTransport) dialCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dials
}
func (t *firstDialGateTransport) closeCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closes
}

// TestAbortConnectAtomicFenceNoAdmissionWindow (VEC-01 B4 F2-R5) proves the
// abortConnect path (a Connect whose blocked dial exhausts the whole-connect
// budget) registers its teardown fence atomically with the Disconnected
// publication: while the old session's Close is blocked, a concurrently
// hammered replacement Connect is never admitted and never dials.
func TestAbortConnectAtomicFenceNoAdmissionWindow(t *testing.T) {
	tr := newFirstDialGateTransport()
	cc := clientWithTransport(t, tr)
	cc.config.ConnectTimeout = 120 * time.Millisecond

	aErr := make(chan error, 1)
	go func() { aErr <- cc.Connect(context.Background()) }()

	// Wait until dial #1 is blocked, i.e. abortConnect is imminent.
	select {
	case <-tr.dial1Entered:
	case <-time.After(time.Second):
		t.Fatal("dial #1 was never entered")
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_ = cc.Connect(context.Background())
			}
		}()
	}

	// Connect A must fail closed on its budget and abortConnect must enter Close.
	if err := <-aErr; !errors.Is(err, dnp3.ErrTimeout) {
		t.Fatalf("Connect A error = %v, want ErrTimeout", err)
	}
	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("abortConnect teardown Close was never entered")
	}
	time.Sleep(50 * time.Millisecond)

	// No replacement may have dialed: only A's original dial exists.
	if n := tr.dialCount(); n != 1 {
		t.Fatalf("replacement Connect dialed while abortConnect teardown was blocked: dials = %d, want 1", n)
	}

	close(stop)
	wg.Wait()

	// Release the blocked Close: the teardown terminates.
	close(tr.closeRelease)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}
	// A's worker is still blocked in its dial, so a replacement is still refused.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect while worker A outstanding = %v, want ErrSessionUnresolved", err)
	}

	// Release A's dial: the worker terminates and a replacement may dial.
	close(tr.dial1Release)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return !cc.connectWorkerOutstanding
	}, time.Second) {
		t.Fatal("connect worker slot did not clear after A's dial returned")
	}
	if err := cc.Connect(context.Background()); errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect after worker A terminated = %v, want it to proceed past the fence", err)
	}
	if got := tr.dialCount(); got != 2 {
		t.Fatalf("replacement dial count = %d, want 2", got)
	}
	// Let B's own failed-handshake cleanup (if any) settle so no goroutine leaks.
	_ = waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil && !cc.connectWorkerOutstanding
	}, time.Second)
}

// TestTeardownCompletionRacingReplacementAdmission (VEC-01 B4 F2-R5) proves a
// replacement Connect is admitted only after the previous teardown has actually
// terminated, and that the completed teardown cannot clobber the replacement.
func TestTeardownCompletionRacingReplacementAdmission(t *testing.T) {
	tr := newCloseGateDialCountTransport()
	cc := clientWithTransport(t, tr)
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	dErr := make(chan error, 1)
	go func() { dErr <- cc.Disconnect(context.Background()) }()

	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("teardown Close was never entered")
	}
	// While teardown A is outstanding, a replacement Connect is refused.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect during outstanding teardown A = %v, want ErrSessionUnresolved", err)
	}

	// Complete teardown A.
	close(tr.closeRelease)
	if err := <-dErr; err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after A completed")
	}

	// Now a replacement Connect is admitted past the fence and dials exactly once.
	if err := cc.Connect(context.Background()); errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect B after teardown A completed = %v, want it to proceed past the fence", err)
	}
	if got := tr.dialCount(); got != 1 {
		t.Fatalf("replacement Connect dials = %d, want 1", got)
	}
}
