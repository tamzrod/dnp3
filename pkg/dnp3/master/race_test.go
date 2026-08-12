package master

import (
	"context"
	"errors"
	"sync"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/master"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// pubReadEchoTransport records the SEQ of the most recent request and returns a
// minimal valid Read response (IIN only) echoing that SEQ. Send/Receive are
// guarded by a mutex so the recorded lastSeq is consistent under concurrent
// callers — the master serializes requests (DNP3-025), but the transport itself
// must also be concurrency-safe.
type pubReadEchoTransport struct {
	mu      sync.Mutex
	lastSeq uint8
}

func (t *pubReadEchoTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubReadEchoTransport) SetTimeout(ms int) {}

func (t *pubReadEchoTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	seq := t.lastSeq
	t.mu.Unlock()
	return buildPubReadResponse(seq), nil
}

// buildPubReadResponse builds a valid DLL+TL+APDU response frame carrying only
// IIN (no object data), with the given application sequence number.
func buildPubReadResponse(seq uint8) []byte {
	iin := al.IIN{}
	iinBytes := iin.Bytes()
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     iinBytes[:],
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// TestConcurrentReadsRaceFree asserts that concurrent Read calls on a single
// client do not trigger the race detector. The master serializes the
// request/reassembly path (DNP3-025) and the public sequence counter is
// guarded by the client mutex, so concurrent Reads must be race-free.
//
// DNP3-040: at most one request may be outstanding per outstation at a time,
// so concurrent same-outstation Reads are rejected with ErrRequestOutstanding
// rather than all succeeding. This is the defined concurrency behavior and is
// not a failure; the test verifies the race detector stays quiet and that no
// error other than ErrRequestOutstanding is returned.
func TestConcurrentReadsRaceFree(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := cc.Read(context.Background(), &types.ReadRequest{
				Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
			})
			if err != nil && !errors.Is(err, master.ErrRequestOutstanding) {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Read failed: %v", err)
	}
}

// TestConcurrentOperateRaceFree asserts concurrent Operate calls are race-free.
// DNP3-040: concurrent same-outstation Operates are rejected with
// ErrRequestOutstanding (defined behavior); only that error is tolerated.
func TestConcurrentOperateRaceFree(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: 0})

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: 12, Variation: 1, Index: 0,
				CommandType: types.DirectOperate,
				Value:       &types.BinaryCommandValue{Value: true},
			})
			if err != nil && !errors.Is(err, master.ErrRequestOutstanding) {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Operate failed: %v", err)
	}
}
