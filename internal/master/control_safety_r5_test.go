package master

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// enterSendTransport blocks inside the physical Send until released and records
// the number of physical sends. It opens no socket, so the test stays offline.
type enterSendTransport struct {
	mu      sync.Mutex
	sends   int
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newEnterSendTransport() *enterSendTransport {
	return &enterSendTransport{entered: make(chan struct{}), release: make(chan struct{})}
}
func (t *enterSendTransport) Send(data []byte) error {
	t.once.Do(func() { close(t.entered) })
	<-t.release
	t.mu.Lock()
	t.sends++
	t.mu.Unlock()
	return nil
}
func (t *enterSendTransport) SetTimeout(ms int)        {}
func (t *enterSendTransport) Receive() ([]byte, error) { return nil, errors.New("no response") }
func (t *enterSendTransport) sendCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sends
}

// TestControlQuarantineAdmissionOverlapNoSecondSend (VEC-01 B4 F3-R5) forces a
// GENUINE overlap between quarantine establishment and control admission using a
// start barrier rather than sequentially arranged calls: one goroutine admits a
// control and drives it into the physical Send, while another establishes the
// UNKNOWN quarantine concurrently. Whichever order linearizes at safetyMu, at
// most one physical send occurs, and a later control can never transmit once the
// quarantine holds. The loop exercises both linearization orders.
func TestControlQuarantineAdmissionOverlapNoSecondSend(t *testing.T) {
	const iters = 60
	var admittedFirst, quarantinedFirst int
	for i := 0; i < iters; i++ {
		tr := newEnterSendTransport()
		m := newControlMaster(t, tr)

		barrier := make(chan struct{})
		var admitted int32
		var gateErr error
		workerDone := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer close(workerDone)
			<-barrier
			gen, release, err := m.AcquireControlGate(1)
			if err != nil {
				gateErr = err
				return
			}
			atomic.StoreInt32(&admitted, 1)
			_, _ = m.OperateWithStatusGated(context.Background(), gen, 1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn), time.Now().Add(2*time.Second))
			release()
		}()
		go func() {
			defer wg.Done()
			<-barrier
			m.quarantineOutstation(1)
		}()
		close(barrier)

		// If the worker was admitted it is now blocked in Send; release it so the
		// test cannot hang. If it was refused (quarantine linearized first) the
		// worker completes without ever sending.
		select {
		case <-tr.entered:
			close(tr.release)
		case <-workerDone:
		case <-time.After(2 * time.Second):
			t.Fatalf("iter %d: neither a send nor worker completion was observed", i)
		}
		wg.Wait()

		if !m.IsQuarantined(1) {
			t.Fatalf("iter %d: quarantine was not established", i)
		}
		got := tr.sendCount()
		if got > 1 {
			t.Fatalf("iter %d: control transmitted %d times; at most one physical send is permitted", i, got)
		}
		if atomic.LoadInt32(&admitted) == 1 {
			// Admission linearized first. The control either reached the wire
			// before the quarantine (1 send) or was refused by the pre-send
			// safety check (0 sends) — both are fail-closed and safe.
			admittedFirst++
		} else {
			quarantinedFirst++
			if !errors.Is(gateErr, ErrControlQuarantined) {
				t.Fatalf("iter %d: refused admission err = %v, want ErrControlQuarantined", i, gateErr)
			}
			if got != 0 {
				t.Fatalf("iter %d: a refused control transmitted %d time(s)", i, got)
			}
		}

		// No SECOND unauthorized transmission may occur once the quarantine holds.
		if _, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn)); !errors.Is(err, ErrControlQuarantined) {
			t.Fatalf("iter %d: post-quarantine control = %v, want ErrControlQuarantined", i, err)
		}
		if got2 := tr.sendCount(); got2 != got {
			t.Fatalf("iter %d: a second control transmitted (sends %d -> %d)", i, got, got2)
		}
	}
	t.Logf("linearization orders observed: admission-first=%d, quarantine-first=%d", admittedFirst, quarantinedFirst)
}
