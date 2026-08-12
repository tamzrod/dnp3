package master

import (
	"errors"
	"sync"
	"testing"
	"time"

	"dnp3/internal/al"
)

// blockingTransport holds the first Receive call open until released, so a
// request stays "outstanding" and a concurrent same-outstation request can be
// observed. Send is a no-op; SetTimeout is ignored.
type blockingTransport struct {
	mu       sync.Mutex
	released bool
	hold     chan struct{}
}

func newBlockingTransport() *blockingTransport {
	return &blockingTransport{hold: make(chan struct{})}
}

func (t *blockingTransport) Send(data []byte) error { return nil }
func (t *blockingTransport) SetTimeout(ms int)     {}

func (t *blockingTransport) Receive() ([]byte, error) {
	// Block until the test releases the held request.
	<-t.hold
	return nil, errReleased
}

// release unblocks the in-flight Receive so the outstanding request completes.
func (t *blockingTransport) release() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.released {
		t.released = true
		close(t.hold)
	}
}

var errReleased = errors.New("released")

// TestConcurrentSameOutstationRejected verifies DNP3-040: a concurrent request
// to the SAME outstation while one is already outstanding is rejected with
// ErrRequestOutstanding instead of blocking behind reqMu.
func TestConcurrentSameOutstationRejected(t *testing.T) {
	m := NewMaster(nil)
	tr := newBlockingTransport()
	m.SetTransport(tr)
	// Force an operational state so the readiness guards pass.
	m.SetState(StateInitialized)
	m.AddOutstation(0x0400, "rtu")

	started := make(chan struct{})
	firstDone := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(started)
		// This request blocks in Receive until released.
		req := buildRequest(0, al.FuncRead, []byte{1, 0, 0x06, 0x00})
		_, err := m.sendWithRetryAndGetResponse(req, 0x0400)
		firstDone <- err
	}()
	<-started

	// The first request is now outstanding (marker set).
	if !m.HasOutstandingRequest(0x0400) {
		t.Fatal("expected outstanding marker for outstation before concurrent attempt")
	}

	// A concurrent same-outstation request must be rejected immediately.
	second := make(chan error, 1)
	go func() {
		req := buildRequest(0, al.FuncRead, []byte{1, 0, 0x06, 0x00})
		_, err := m.sendWithRetryAndGetResponse(req, 0x0400)
		second <- err
	}()
	select {
	case err := <-second:
		if !errors.Is(err, ErrRequestOutstanding) {
			t.Fatalf("concurrent same-outstation: err = %v, want ErrRequestOutstanding", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent same-outstation request blocked instead of being rejected")
	}

	// Release the held request so it completes and clears the marker.
	tr.release()
	wg.Wait()
	<-firstDone // drain the first request's result; the marker is cleared by defer.
	if m.HasOutstandingRequest(0x0400) {
		t.Fatal("outstanding marker not cleared after request completed")
	}

	// A subsequent (sequential) request to the same outstation now succeeds
	// the begin phase (no ErrRequestOutstanding).
	if err := m.beginRequest(0x0400); err != nil {
		t.Fatalf("sequential request after completion rejected: %v", err)
	}
	m.endRequest(0x0400)
}

// TestDistinctOutstationsNotRejectedByOutstandingTracking verifies the
// outstanding marker is per-outstation: a request to a different outstation
// while one is outstanding is NOT rejected by the tracker (it may queue on
// reqMu, but that is separate from DNP3-040's per-outstation rule).
func TestDistinctOutstationsNotRejectedByOutstandingTracking(t *testing.T) {
	m := NewMaster(nil)
	if err := m.beginRequest(0x0001); err != nil {
		t.Fatalf("beginRequest outstation 1: %v", err)
	}
	defer m.endRequest(0x0001)
	// A different outstation's marker is independent.
	if err := m.beginRequest(0x0002); err != nil {
		t.Fatalf("beginRequest distinct outstation 2 rejected: %v", err)
	}
	m.endRequest(0x0002)
	// The original marker survives.
	if !m.HasOutstandingRequest(0x0001) {
		t.Fatal("outstanding marker for outstation 1 lost while tracking outstation 2")
	}
}

// TestBeginRequestIdempotentGuard verifies endRequest is idempotent and that a
// repeated beginRequest for the same outstation is rejected while one is held.
func TestBeginRequestIdempotentGuard(t *testing.T) {
	m := NewMaster(nil)
	if err := m.beginRequest(0x0001); err != nil {
		t.Fatalf("first beginRequest: %v", err)
	}
	if err := m.beginRequest(0x0001); !errors.Is(err, ErrRequestOutstanding) {
		t.Fatalf("second beginRequest: err = %v, want ErrRequestOutstanding", err)
	}
	m.endRequest(0x0001)
	m.endRequest(0x0001) // idempotent: must not panic
	if m.HasOutstandingRequest(0x0001) {
		t.Fatal("marker present after endRequest")
	}
}
