package master

import (
	"context"
	"errors"
	"testing"
	"time"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// blockingReceiveTransport is a transport.Handler whose Receive blocks until
// releaseCh is closed, simulating an outstation that never responds. Used to
// drive the cancel-mid-read path. Send records the request so the test can
// confirm a request was actually issued.
type blockingReceiveTransport struct {
	sent      chan []byte
	releaseCh chan struct{}
}

func newBlockingReceiveTransport() *blockingReceiveTransport {
	return &blockingReceiveTransport{
		sent:      make(chan []byte, 1),
		releaseCh: make(chan struct{}),
	}
}

func (t *blockingReceiveTransport) Listen() error                   { return nil }
func (t *blockingReceiveTransport) Accept() error                   { return nil }
func (t *blockingReceiveTransport) Connect() error                  { return nil }
func (t *blockingReceiveTransport) Close() error                    { close(t.releaseCh); return nil }
func (t *blockingReceiveTransport) Send(data []byte) error          { t.sent <- data; return nil }
func (t *blockingReceiveTransport) SetTimeout(ms int)              {}
func (t *blockingReceiveTransport) Receive() ([]byte, error) {
	<-t.releaseCh
	return nil, errors.New("receive aborted")
}

// readClient builds a public client wired to a blocking-receive transport,
// marked connected (internal + public state) so Read can proceed.
func readClient(t *testing.T) (*client, *blockingReceiveTransport) {
	t.Helper()
	tr := newBlockingReceiveTransport()
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	cc.transport = tr
	cc.internalMaster.SetTransport(&transportAdapter{Handler: tr})
	cc.internalMaster.SetState(master.StateInitialized)
	cc.mu.Lock()
	cc.state = dnp3.StateConnected
	cc.mu.Unlock()
	return cc, tr
}

// TestReadAlreadyCancelledContext asserts that an already-cancelled context
// causes Read to return promptly with ErrContextCanceled and no response.
func TestReadAlreadyCancelledContext(t *testing.T) {
	cc, _ := readClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	resp, err := cc.Read(ctx, &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	elapsed := time.Since(start)

	if !errors.Is(err, dnp3.ErrContextCanceled) {
		t.Fatalf("Read error = %v, want ErrContextCanceled", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response on cancel, got %+v", resp)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Read did not return promptly: %v", elapsed)
	}
}

// TestReadCancelledMidRequest asserts that cancelling the context while a read
// request is outstanding causes Read to return promptly with ErrContextCanceled
// and no partial points are surfaced.
func TestReadCancelledMidRequest(t *testing.T) {
	cc, tr := readClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan struct{ err error; resp *ReadResponse }, 1)
	go func() {
		resp, err := cc.Read(ctx, &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		errCh <- struct{ err error; resp *ReadResponse }{err, resp}
	}()

	// Wait for the request to be sent so the read is genuinely outstanding.
	select {
	case <-tr.sent:
	case <-time.After(time.Second):
		t.Fatal("request was never sent")
	}

	start := time.Now()
	cancel()
	res := <-errCh
	elapsed := time.Since(start)

	if !errors.Is(res.err, dnp3.ErrContextCanceled) {
		t.Fatalf("Read error = %v, want ErrContextCanceled", res.err)
	}
	if res.resp != nil {
		t.Fatalf("expected nil response on cancel; got partial points: %+v", res.resp)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Read did not return promptly after cancel: %v", elapsed)
	}
}
