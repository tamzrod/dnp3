package outstation

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
)

type captureTransport struct {
	mu   sync.Mutex
	sent [][]byte
}

func (c *captureTransport) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, append([]byte(nil), data...))
	return nil
}

func (c *captureTransport) Receive() ([]byte, error) { return nil, errClosed }
func (c *captureTransport) SetTimeout(int)           {}

func (c *captureTransport) frames() [][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([][]byte, len(c.sent))
	copy(out, c.sent)
	return out
}

var errClosed = errString("closed")

type errString string

func (e errString) Error() string { return string(e) }

func testEvent(class EventClass, index uint16, on bool, at time.Time) Event {
	ev := Event{
		Class:       class,
		Group:       2,
		Index:       index,
		Kind:        EventKindBinary,
		Quality:     al.BinEventOnline,
		BinaryValue: on,
	}
	if !at.IsZero() {
		ev.HasTime = true
		ev.Time = at
	}
	return ev
}

func TestClassReadEncodesQueuedEventsAndClearsThem(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	stamp := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC)
	if err := ost.PublishEvent(testEvent(Class1, 7, true, stamp)); err != nil {
		t.Fatal(err)
	}
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 4},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 2, 0x06},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ost.EventCount() != 0 {
		t.Fatalf("events left = %d", ost.EventCount())
	}
	obj := resp.Data[2:]
	evs, err := al.DecodeEventObjects(obj)
	if err != nil {
		t.Fatalf("decode % X: %v", obj, err)
	}
	if len(evs) != 1 || evs[0].Index != 7 || evs[0].Flags != 0x81 || !evs[0].Time.Equal(stamp) || evs[0].Variation != 2 {
		t.Fatalf("events %+v", evs)
	}
	if resp.Data[0] != 0 || resp.Data[1] != 0 {
		t.Fatalf("IIN after draining class = % X", resp.Data[:2])
	}
}

func TestCountZeroEventReadDoesNotDumpStatic(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	if err := ost.PublishEvent(testEvent(Class1, 1, true, time.Time{})); err != nil {
		t.Fatal(err)
	}
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{2, 1, 0x07, 0x00},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("count 0 returned %d object bytes % X", len(resp.Data)-2, resp.Data)
	}
	if ost.EventCount() != 1 {
		t.Fatal("count 0 must not remove events")
	}
	if !ost.IIN().Class1Events {
		t.Fatal("class 1 IIN should remain set")
	}
}

func TestPublishRejectsBadTimeAndIndex(t *testing.T) {
	ost := NewOutstation(nil)
	if err := ost.PublishEvent(testEvent(Class1, 256, true, time.Time{})); err != ErrEventIndex {
		t.Fatalf("index error = %v", err)
	}
	ev := testEvent(Class2, 1, false, time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC))
	if err := ost.PublishEvent(ev); err == nil {
		t.Fatal("pre-epoch time accepted")
	}
}

func TestEventOrderByTimestamp(t *testing.T) {
	q := NewEventQueue(30)
	later := time.Date(2000, 1, 1, 0, 0, 2, 0, time.UTC)
	earlier := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC)
	if !q.Add(testEvent(Class1, 1, true, later)) || !q.Add(testEvent(Class1, 2, false, earlier)) {
		t.Fatal("add")
	}
	got := q.SnapshotClass(Class1)
	if got[0].Index != 2 || got[1].Index != 1 {
		t.Fatalf("order = %d then %d", got[0].Index, got[1].Index)
	}
}

func TestUnsolicitedConfirmRetryAndReconnectQueue(t *testing.T) {
	ost := NewOutstation(&Config{
		OutstationAddress: 1024,
		MasterAddress:     1,
		MaxEventBuffers:   30,
	})
	ost.Initialize()
	ost.Start()
	tr := &captureTransport{}
	ost.SetTransport(tr)
	ost.unsolConfirmWait = 40 * time.Millisecond
	stamp := time.Date(2010, 5, 6, 7, 8, 9, 0, time.UTC)
	if err := ost.PublishEvent(testEvent(Class1, 4, true, stamp)); err != nil {
		t.Fatal(err)
	}
	if _, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncEnableUnsolicited,
	}); err != nil {
		t.Fatal(err)
	}
	ost.flushPendingUnsolicited()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(tr.frames()) < 3 {
		time.Sleep(15 * time.Millisecond)
	}
	frames := tr.frames()
	if len(frames) < 3 {
		t.Fatalf("sends = %d, want 3 identical retries", len(frames))
	}
	if !bytes.Equal(frames[0], frames[1]) || !bytes.Equal(frames[1], frames[2]) {
		t.Fatal("retries did not resend the original frame")
	}
	decoded, err := frame.Decode(frames[0])
	if err != nil {
		t.Fatal(err)
	}
	if decoded.DestAddr != 1 || decoded.SrcAddr != 1024 {
		t.Fatalf("addresses dest=%d src=%d", decoded.DestAddr, decoded.SrcAddr)
	}
	time.Sleep(80 * time.Millisecond)
	if len(tr.frames()) != 3 {
		t.Fatalf("sends after bound = %d, want 3", len(tr.frames()))
	}
	if ost.EventCount() != 1 {
		t.Fatalf("events after abandoned confirm = %d", ost.EventCount())
	}

	// A new outstation on the same queue still has the event.
	next := NewOutstation(&Config{OutstationAddress: 1024, MasterAddress: 1, MaxEventBuffers: 30})
	next.UseEventQueue(ost.eventQueue)
	if next.EventCount() != 1 {
		t.Fatalf("retained events = %d", next.EventCount())
	}

	// Matching confirm on a fresh send clears the buffer. Wrong sequence does not.
	tr2 := &captureTransport{}
	next.SetTransport(tr2)
	next.unsolConfirmWait = time.Hour
	if _, err := next.ProcessRequest(&al.APDU{
		FuncCode: al.FuncEnableUnsolicited,
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 2},
	}); err != nil {
		t.Fatal(err)
	}
	next.flushPendingUnsolicited()
	if len(tr2.frames()) != 1 {
		t.Fatalf("resend = %d", len(tr2.frames()))
	}
	if _, err := next.ProcessRequest(&al.APDU{
		FuncCode: al.FuncConfirm,
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 9},
	}); err != nil {
		t.Fatal(err)
	}
	if next.EventCount() != 1 {
		t.Fatal("wrong confirm cleared the event")
	}
	if _, err := next.ProcessRequest(&al.APDU{
		FuncCode: al.FuncConfirm,
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
	}); err != nil {
		t.Fatal(err)
	}
	if next.EventCount() != 0 {
		t.Fatalf("events after confirm = %d", next.EventCount())
	}
	// Duplicate confirm is ignored.
	if _, err := next.ProcessRequest(&al.APDU{
		FuncCode: al.FuncConfirm,
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
	}); err != nil {
		t.Fatal(err)
	}
	next.StopUnsolicitedTimer()
	ost.StopUnsolicitedTimer()
}

func TestMultipleEventsSamePointAndOverflow(t *testing.T) {
	ost := NewOutstation(&Config{MaxEventBuffers: 9})
	ost.Initialize()
	for i := 0; i < 3; i++ {
		if err := ost.PublishEvent(testEvent(Class1, 1, i%2 == 0, time.Time{})); err != nil {
			t.Fatal(err)
		}
	}
	if err := ost.PublishEvent(testEvent(Class1, 1, true, time.Time{})); err != ErrEventBufferFull {
		t.Fatalf("overflow = %v", err)
	}
	if !ost.IIN().BufferOverflow {
		t.Fatal("overflow IIN not set")
	}
	if ost.EventCount() != 3 {
		t.Fatalf("count = %d, want the older 3 events", ost.EventCount())
	}
}
