package master

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
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

// blockedDialCloseTransport models an uncooperative custom transport whose dial
// (Connect) and teardown (Close) both block until released. It opens no socket,
// so tests stay offline.
type blockedDialCloseTransport struct {
	dialEntered  chan struct{}
	dialRelease  chan struct{}
	closeEntered chan struct{}
	closeRelease chan struct{}
	dialOnce     sync.Once
	closeOnce    sync.Once
}

func newBlockedDialCloseTransport() *blockedDialCloseTransport {
	return &blockedDialCloseTransport{
		dialEntered:  make(chan struct{}),
		dialRelease:  make(chan struct{}),
		closeEntered: make(chan struct{}),
		closeRelease: make(chan struct{}),
	}
}

func (t *blockedDialCloseTransport) Listen() error          { return nil }
func (t *blockedDialCloseTransport) Accept() error          { return nil }
func (t *blockedDialCloseTransport) Send(data []byte) error { return nil }
func (t *blockedDialCloseTransport) Receive() ([]byte, error) {
	return nil, errors.New("no response")
}
func (t *blockedDialCloseTransport) SetTimeout(ms int) {}
func (t *blockedDialCloseTransport) Connect() error {
	t.dialOnce.Do(func() { close(t.dialEntered) })
	<-t.dialRelease
	return errors.New("dial aborted")
}
func (t *blockedDialCloseTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closeEntered) })
	<-t.closeRelease
	return nil
}

// TestConnectBudgetBlockedDialBlockedClosePrompt (VEC-01 B4 F2-R4) proves the
// public Connect returns promptly within the whole-connect budget even when an
// uncooperative transport blocks indefinitely in BOTH the dial and the cleanup
// Close, and that the outstanding worker/teardown state is accurately reported:
// a replacement Connect is refused fail-closed until the worker terminates and
// the teardown completes.
func TestConnectBudgetBlockedDialBlockedClosePrompt(t *testing.T) {
	tr := newBlockedDialCloseTransport()
	c, err := NewClient(NewConfig(WithConnectTimeout(120 * time.Millisecond)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})

	start := time.Now()
	err = cc.Connect(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected connect to fail closed when the budget is exceeded")
	}
	if !errors.Is(err, dnp3.ErrTimeout) {
		t.Fatalf("Connect error = %v, want ErrTimeout (budget exhausted)", err)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("Connect took %v; a blocked dial AND blocked Close must not wedge the public call", elapsed)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}

	// The asynchronous cleanup must actually enter (and block in) Close.
	select {
	case <-tr.closeEntered:
	case <-time.After(time.Second):
		t.Fatal("cleanup Close was never entered")
	}
	cc.mu.RLock()
	teardownOutstanding := cc.teardown != nil
	cc.mu.RUnlock()
	if !teardownOutstanding {
		t.Fatal("teardown must be reported outstanding while Close is blocked")
	}

	// A replacement Connect is refused fail-closed while the cleanup is blocked.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect during blocked cleanup = %v, want ErrSessionUnresolved", err)
	}

	// Release the cleanup: the teardown fence must clear.
	close(tr.closeRelease)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}

	// The connect worker is still blocked in its dial, so a replacement Connect is
	// still refused.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect while connect worker outstanding = %v, want ErrSessionUnresolved", err)
	}

	// Release the dial so the worker terminates and the fence clears.
	close(tr.dialRelease)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return !cc.connectWorkerOutstanding
	}, time.Second) {
		t.Fatal("connect worker slot did not clear after the dial returned")
	}
}

// slowCloseTransport already exists in operate_disconnect_cancel_test.go and
// blocks in Close until releaseCh is closed.

// TestPreCancelledDisconnectFencesReplacementConnect (VEC-01 B4 F3-R4) proves a
// pre-cancelled public Disconnect still tears the session down (so no live
// session survives) and fences a replacement Connect fail-closed until the
// teardown terminates, after which Connect is no longer refused by the fence.
func TestPreCancelledDisconnectFencesReplacementConnect(t *testing.T) {
	tr := newSlowCloseTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err = cc.Disconnect(ctx)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("pre-cancelled Disconnect did not return promptly: %v", elapsed)
	}
	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Disconnect error = %v, want ErrContextCanceled", err)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}

	// The teardown is outstanding (Close is blocked): a replacement Connect is
	// refused fail-closed.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect during outstanding teardown = %v, want ErrSessionUnresolved", err)
	}

	// Release the blocked Close: the teardown terminates and the fence clears.
	close(tr.releaseCh)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}
	if err := cc.Connect(context.Background()); errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect after teardown termination = %v, want it to proceed past the fence", err)
	}
}

// TestDisconnectCancellationRacingTeardown (VEC-01 B4 F3-R4) proves that
// cancelling Disconnect while the teardown is already running returns promptly
// without publishing a reusable session: the replacement Connect stays refused
// until the teardown actually terminates.
func TestDisconnectCancellationRacingTeardown(t *testing.T) {
	tr := newSlowCloseTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.mu.Lock()
	cc.state = dnp3.StateActive
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- cc.Disconnect(ctx) }()

	// Let the teardown enter the blocking Close, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	start := time.Now()
	err = <-errCh
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Disconnect did not return promptly after cancel: %v", elapsed)
	}
	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Disconnect error = %v, want ErrContextCanceled", err)
	}
	if cc.State() != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", cc.State())
	}
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect during cancelled-but-ongoing teardown = %v, want ErrSessionUnresolved", err)
	}

	close(tr.releaseCh)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return cc.teardown == nil
	}, time.Second) {
		t.Fatal("teardown fence did not clear after Close returned")
	}
}

// TestBlockedControlSendDisconnectRefusesConnectUnknownPersists (VEC-01 B4
// F3-R4) proves that a public Disconnect during an in-flight control send keeps
// the unresolved fence (refusing a replacement Connect) and the persistent
// UNKNOWN quarantine distinct: after the blocked worker terminates the fence
// lifts but the UNKNOWN quarantine remains.
func TestBlockedControlSendDisconnectRefusesConnectUnknownPersists(t *testing.T) {
	tr := newPubBlockingSendTransport()
	cc := clientWithTransport(t, tr)
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		_, err := cc.Operate(ctx, &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		errCh <- err
	}()
	select {
	case <-tr.entered:
	case <-time.After(time.Second):
		t.Fatal("control send never entered")
	}

	// Public Disconnect while the send is blocked in flight.
	if err := cc.Disconnect(context.Background()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !cc.internalMaster.HasUnresolvedSession() {
		t.Fatal("Disconnect during an in-flight control send must fence the session unresolved")
	}
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("Disconnect during an in-flight control send must quarantine (UNKNOWN) the outstation")
	}

	// A replacement Connect is refused before dialing while the fence holds.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect while control send unresolved = %v, want ErrSessionUnresolved", err)
	}

	// Release the blocked send so the worker terminates.
	tr.release()
	if err := <-errCh; err == nil {
		t.Fatal("the interrupted control should not report success")
	}
	if !waitForCond(func() bool { return !cc.internalMaster.HasUnresolvedSession() }, time.Second) {
		t.Fatal("the unresolved fence must clear once the control worker terminates")
	}
	// The UNKNOWN quarantine must persist independently of the lifted fence.
	if !cc.internalMaster.IsQuarantined(cc.config.OutstationAddress) {
		t.Fatal("the UNKNOWN quarantine must persist after the unresolved fence lifts")
	}
}

// dialGateTransport is a per-call dial gate: the first Connect blocks until
// released, later Connect calls succeed immediately. Close is counted so a test
// can prove a superseded worker does not tear down a replacement session. It
// opens no socket, so tests stay offline.
type dialGateTransport struct {
	mu       sync.Mutex
	calls    int
	closes   int
	entered1 chan struct{}
	release1 chan struct{}
	once1    sync.Once
}

func newDialGateTransport() *dialGateTransport {
	return &dialGateTransport{entered1: make(chan struct{}), release1: make(chan struct{})}
}

func (t *dialGateTransport) Listen() error          { return nil }
func (t *dialGateTransport) Accept() error          { return nil }
func (t *dialGateTransport) Send(data []byte) error { return nil }
func (t *dialGateTransport) Receive() ([]byte, error) {
	return nil, errors.New("no handshake response")
}
func (t *dialGateTransport) SetTimeout(ms int) {}
func (t *dialGateTransport) Connect() error {
	t.mu.Lock()
	t.calls++
	n := t.calls
	t.mu.Unlock()
	if n == 1 {
		t.once1.Do(func() { close(t.entered1) })
		<-t.release1
	}
	return nil
}
func (t *dialGateTransport) Close() error {
	t.mu.Lock()
	t.closes++
	t.mu.Unlock()
	return nil
}
func (t *dialGateTransport) callCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.calls
}
func (t *dialGateTransport) closeCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closes
}

// TestSupersededConnectWorkerDoesNotTearDownReplacement (VEC-01 B4 F2-R4)
// proves a superseded Connect worker's late completion cannot mutate or tear
// down a replacement session: after Connect A times out, its late dial success
// must not run a cleanup that Closes the shared transport, and a later Connect B
// dials without an extra teardown.
func TestSupersededConnectWorkerDoesNotTearDownReplacement(t *testing.T) {
	tr := newDialGateTransport()
	c, err := NewClient(NewConfig(WithConnectTimeout(120 * time.Millisecond)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})

	// Connect A: dial #1 blocks past the budget, so Connect returns promptly.
	start := time.Now()
	err = cc.Connect(context.Background())
	if err == nil {
		t.Fatal("expected Connect A to fail closed when the budget is exceeded")
	}
	if !errors.Is(err, dnp3.ErrTimeout) {
		t.Fatalf("Connect A error = %v, want ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 600*time.Millisecond {
		t.Fatalf("Connect A took %v; the blocked dial must be bounded by the 120ms budget", elapsed)
	}

	// A replacement Connect is refused while worker A is still blocked in its dial.
	if err := cc.Connect(context.Background()); !errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect B while worker A outstanding = %v, want ErrSessionUnresolved", err)
	}

	// Release A's dial: A's worker resumes, observes it was superseded, and must
	// NOT run a teardown against the shared transport.
	close(tr.release1)
	if !waitForCond(func() bool {
		cc.mu.RLock()
		defer cc.mu.RUnlock()
		return !cc.connectWorkerOutstanding && cc.teardown == nil
	}, time.Second) {
		t.Fatal("connect worker slot / teardown fence did not clear after A's dial returned")
	}
	if got := tr.closeCount(); got != 1 {
		t.Fatalf("superseded worker A performed an extra teardown: closes = %d, want 1 (abortConnect only)", got)
	}

	// A later Connect B is admitted (past the fence) and dials once more.
	if err := cc.Connect(context.Background()); errors.Is(err, master.ErrSessionUnresolved) {
		t.Fatalf("Connect B after worker A terminated = %v, want it to proceed past the fence", err)
	}
	if got := tr.callCount(); got != 2 {
		t.Fatalf("replacement Connect did not dial: calls = %d, want 2", got)
	}
}
