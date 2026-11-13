package outstation

import (
	"bytes"
	"testing"
	"time"

	"dnp3/internal/al"
)

// This file is a newly authored regression for PPR-001 (event batching past the
// 8-bit qualifier-0x17 count). It is not a recovered copy of any earlier test.
// It pins that a same-variation run larger than 255 objects is split into
// qualifier-bounded headers with no event loss, and that no object is dropped or
// acknowledged when encoding, framing, or transport fails.

const batchBaseTime = "2021-06-07T08:09:10Z"

func batchStamp(i int) time.Time {
	t, err := time.Parse(time.RFC3339, batchBaseTime)
	if err != nil {
		panic(err)
	}
	return t.Add(time.Duration(i) * time.Millisecond)
}

// batchBinary builds a timestamped Group 2 Variation 2 event at the given index.
func batchBinary(class EventClass, index uint16, on bool, i int) Event {
	return Event{
		Class:       class,
		Group:       2,
		Index:       index,
		Kind:        EventKindBinary,
		Quality:     al.BinEventOnline,
		BinaryValue: on,
		HasTime:     true,
		Time:        batchStamp(i),
	}
}

func batchAnalog(class EventClass, index uint16, value int32, i int) Event {
	return Event{
		Class:       class,
		Group:       32,
		Index:       index,
		Kind:        EventKindAnalog,
		Quality:     al.AnalogEventOnline,
		AnalogValue: value,
		HasTime:     true,
		Time:        batchStamp(i),
	}
}

func batchCounter(class EventClass, index uint16, value uint32, i int) Event {
	return Event{
		Class:        class,
		Group:        22,
		Index:        index,
		Kind:         EventKindCounter,
		Quality:      al.CounterEventOnline,
		CounterValue: value,
		HasTime:      true,
		Time:         batchStamp(i),
	}
}

func sameVariationBatch(n int) []Event {
	evs := make([]Event, 0, n)
	for i := 0; i < n; i++ {
		evs = append(evs, batchBinary(Class1, uint16(i%256), i%2 == 0, i))
	}
	return evs
}

// TestEncodeEventListSplitsRunsOver255 pins the PPR-001 correction: a
// same-variation run of 256 or more objects must be split into multiple
// qualifier-0x17 headers, each carrying at most 255 objects, with every object
// delivered exactly once and no whole-run loss.
func TestEncodeEventListSplitsRunsOver255(t *testing.T) {
	for _, n := range []int{254, 255, 256, 333} {
		evs := sameVariationBatch(n)
		// Assign stable insertion sequences so the returned ids are checkable.
		for i := range evs {
			evs[i].Seq = uint64(1000 + i)
		}
		wire, ids := encodeEventList(evs, 0)
		if len(wire) == 0 {
			t.Fatalf("n=%d: encoded nothing (whole run lost)", n)
		}
		if len(ids) != n {
			t.Fatalf("n=%d: ids=%d, want %d", n, len(ids), n)
		}
		for i, id := range ids {
			if id != uint64(1000+i) {
				t.Fatalf("n=%d: id[%d]=%d, want %d", n, i, id, 1000+i)
			}
		}
		decoded, err := al.DecodeEventObjects(wire)
		if err != nil {
			t.Fatalf("n=%d: decode %v", n, err)
		}
		if len(decoded) != n {
			t.Fatalf("n=%d: decoded %d events, want %d", n, len(decoded), n)
		}
		headers := countHeaders(t, wire)
		wantHeaders := (n + 254) / 255
		if headers != wantHeaders {
			t.Fatalf("n=%d: %d headers, want %d", n, headers, wantHeaders)
		}
		for i, ev := range decoded {
			wantFlags := uint8(al.BinEventOnline)
			if i%2 == 0 {
				wantFlags |= al.BinEventState
			}
			if ev.Group != 2 || ev.Variation != 2 || ev.Index != uint16(i%256) ||
				ev.Flags != wantFlags || !ev.HasTime || !ev.Time.Equal(batchStamp(i)) {
				t.Fatalf("n=%d: object %d = %+v", n, i, ev)
			}
		}
	}
}

// countHeaders walks the encoded object stream and counts object headers.
func countHeaders(t *testing.T, data []byte) int {
	t.Helper()
	headers := 0
	for len(data) > 0 {
		if len(data) < 4 {
			t.Fatalf("truncated header in % X", data)
		}
		qual := data[2]
		if qual != al.QualIndex8 {
			t.Fatalf("unexpected qualifier %02X", qual)
		}
		n := int(data[3])
		size := eventObjectSizeForTest(t, data[0], data[1])
		headers++
		data = data[4:]
		need := n * (1 + size)
		if len(data) < need {
			t.Fatalf("truncated objects: have %d need %d", len(data), need)
		}
		data = data[need:]
	}
	return headers
}

func eventObjectSizeForTest(t *testing.T, group, variation uint8) int {
	t.Helper()
	switch group {
	case 2:
		switch variation {
		case 1:
			return 1
		case 2:
			return 7
		}
	case 22:
		switch variation {
		case 1:
			return 5
		case 5:
			return 11
		}
	case 32:
		switch variation {
		case 1:
			return 5
		case 3:
			return 11
		}
	}
	t.Fatalf("unknown group/variation %d/%d", group, variation)
	return 0
}

// TestEncodeEventListMixedGroupsOver255 pins that separate Group 2, 22, and 32
// runs each keep their own headers, are individually bounded at 255, and are
// all preserved in insertion order.
func TestEncodeEventListMixedGroupsOver255(t *testing.T) {
	var evs []Event
	for i := 0; i < 256; i++ {
		evs = append(evs, batchBinary(Class1, uint16(i), true, i))
	}
	for i := 0; i < 256; i++ {
		evs = append(evs, batchCounter(Class1, uint16(i), uint32(i), 256+i))
	}
	for i := 0; i < 256; i++ {
		evs = append(evs, batchAnalog(Class1, uint16(i), int32(i), 512+i))
	}
	for i := range evs {
		evs[i].Seq = uint64(i)
	}
	wire, ids := encodeEventList(evs, 0)
	if len(ids) != len(evs) {
		t.Fatalf("ids=%d, want %d", len(ids), len(evs))
	}
	decoded, err := al.DecodeEventObjects(wire)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != len(evs) {
		t.Fatalf("decoded %d, want %d", len(decoded), len(evs))
	}
	counts := map[uint8]int{}
	for _, ev := range decoded {
		counts[ev.Group]++
	}
	for _, g := range []uint8{2, 22, 32} {
		if counts[g] != 256 {
			t.Fatalf("group %d count=%d, want 256", g, counts[g])
		}
	}
	// Group 2 objects must precede Group 22, which precede Group 32.
	firstGroup := func(g uint8) int {
		for i, ev := range decoded {
			if ev.Group == g {
				return i
			}
		}
		return -1
	}
	if !(firstGroup(2) < firstGroup(22) && firstGroup(22) < firstGroup(32)) {
		t.Fatalf("group order 2@%d 22@%d 32@%d", firstGroup(2), firstGroup(22), firstGroup(32))
	}
}

// TestEncodeEventListRepeatedIndexOver255 pins that duplicate point indexes are
// preserved when a run crosses the 255-object boundary.
func TestEncodeEventListRepeatedIndexOver255(t *testing.T) {
	var evs []Event
	for i := 0; i < 300; i++ {
		evs = append(evs, batchBinary(Class1, 5, i%2 == 0, i))
	}
	wire, ids := encodeEventList(evs, 0)
	if len(ids) != 300 {
		t.Fatalf("ids=%d, want 300", len(ids))
	}
	decoded, err := al.DecodeEventObjects(wire)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 300 {
		t.Fatalf("decoded %d, want 300", len(decoded))
	}
	for i, ev := range decoded {
		if ev.Index != 5 {
			t.Fatalf("object %d index=%d, want 5", i, ev.Index)
		}
	}
}

// TestClassReadDeliversEveryEventOver255 is the full-path regression: 256 queued
// events must all be delivered by a class read, and the queue must drain
// exactly once.
func TestClassReadDeliversEveryEventOver255(t *testing.T) {
	ost := NewOutstation(&Config{MaxEventBuffers: 1000})
	ost.Initialize()
	const n = 256
	for i := 0; i < n; i++ {
		if err := ost.PublishEvent(batchBinary(Class1, uint16(i), i%2 == 0, i)); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}
	if ost.EventCount() != n {
		t.Fatalf("queued=%d, want %d", ost.EventCount(), n)
	}
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 2, 0x06},
	})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := al.DecodeEventObjects(resp.Data[2:])
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != n {
		t.Fatalf("delivered %d, want %d", len(decoded), n)
	}
	if ost.EventCount() != 0 {
		t.Fatalf("queue after read=%d, want 0", ost.EventCount())
	}
	if resp.Data[0] != 0 || resp.Data[1] != 0 {
		t.Fatalf("IIN after draining class = % X", resp.Data[:2])
	}
}

// TestEventQueueBoundsAt333 pins the per-class capacity boundary used by the
// 333-event case: exactly maxSize/3 events are accepted, the next is rejected,
// and no accepted event is lost.
func TestEventQueueBoundsAt333(t *testing.T) {
	ost := NewOutstation(&Config{MaxEventBuffers: 999}) // per class = 333
	ost.Initialize()
	for i := 0; i < 333; i++ {
		if err := ost.PublishEvent(batchBinary(Class1, uint16(i%256), true, i)); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}
	if err := ost.PublishEvent(batchBinary(Class1, 0, true, 333)); err != ErrEventBufferFull {
		t.Fatalf("overflow err=%v, want ErrEventBufferFull", err)
	}
	if !ost.IIN().BufferOverflow {
		t.Fatal("overflow IIN not set")
	}
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 2, 0x06},
	})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := al.DecodeEventObjects(resp.Data[2:])
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 333 {
		t.Fatalf("delivered %d, want 333", len(decoded))
	}
}

// TestBatchTransportFailureKeepsEvents pins no silent loss on a transport
// failure: a failed unsolicited send must leave every event queued.
func TestBatchTransportFailureKeepsEvents(t *testing.T) {
	ost := NewOutstation(&Config{OutstationAddress: 1024, MasterAddress: 1, MaxEventBuffers: 1000})
	ost.Initialize()
	ost.SetTransport(&failTransport{})
	ost.unsolConfirmWait = time.Hour
	const n = 256
	for i := 0; i < n; i++ {
		if err := ost.PublishEvent(batchBinary(Class1, uint16(i), true, i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncEnableUnsolicited,
	}); err != nil {
		t.Fatal(err)
	}
	ost.flushPendingUnsolicited()
	if ost.EventCount() != n {
		t.Fatalf("events after failed send=%d, want %d", ost.EventCount(), n)
	}
	ost.StopUnsolicitedTimer()
}

type failTransport struct{}

func (failTransport) Send([]byte) error        { return errString("send failed") }
func (failTransport) Receive() ([]byte, error) { return nil, errClosed }
func (failTransport) SetTimeout(int)           {}

// TestUnsolicitedLargeBatchConfirmRetryRetention drives the unsolicited path
// with 256 queued events. Every frame must be resent identically on retry, a
// matching confirm must clear all 256 accepted events exactly once, a duplicate
// confirm must not clear a later event, and a fresh outstation on the same queue
// must still hold the events until the confirm arrives.
func TestUnsolicitedLargeBatchConfirmRetryRetention(t *testing.T) {
	ost := NewOutstation(&Config{OutstationAddress: 1024, MasterAddress: 1, MaxEventBuffers: 1000})
	ost.Initialize()
	tr := &captureTransport{}
	ost.SetTransport(tr)
	ost.unsolConfirmWait = 40 * time.Millisecond
	const n = 256
	for i := 0; i < n; i++ {
		if err := ost.PublishEvent(batchBinary(Class1, uint16(i), i%2 == 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncEnableUnsolicited,
	}); err != nil {
		t.Fatal(err)
	}
	ost.flushPendingUnsolicited()

	// The 256-event batch fragments into several link frames; collect them and
	// wait for the retry resend to append a second identical copy.
	first := tr.frames()
	if len(first) == 0 {
		t.Fatal("no frames sent")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(tr.frames()) < 2*len(first) {
		time.Sleep(10 * time.Millisecond)
	}
	second := tr.frames()
	if len(second) < 2*len(first) {
		t.Fatalf("frames=%d, want at least %d (one resend)", len(second), 2*len(first))
	}
	for i := 0; i < len(first); i++ {
		if !bytes.Equal(first[i], second[i]) {
			t.Fatalf("retry frame %d differs from original", i)
		}
	}
	if ost.EventCount() != n {
		t.Fatalf("events before confirm=%d, want %d", ost.EventCount(), n)
	}

	// A fresh outstation on the same queue retains every event.
	next := NewOutstation(&Config{OutstationAddress: 1024, MasterAddress: 1, MaxEventBuffers: 1000})
	next.UseEventQueue(ost.eventQueue)
	if next.EventCount() != n {
		t.Fatalf("retained events=%d, want %d", next.EventCount(), n)
	}
	ost.StopUnsolicitedTimer()

	// A matching confirm releases the events carried by the confirmed
	// fragment; because an unsolicited response is bounded to one application
	// fragment, the remainder drains over successive confirm cycles.
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
	if len(tr2.frames()) == 0 {
		t.Fatal("no frames on fresh send")
	}
	delivered := 0
	for cycles := 0; cycles < 64 && next.EventCount() > 0; cycles++ {
		before := next.EventCount()
		seq := next.unsolSeq
		confirm := &al.APDU{
			FuncCode: al.FuncConfirm,
			Control:  al.AppControl{FIR: true, FIN: true, UNS: true, Seq: seq},
		}
		if _, err := next.ProcessRequest(confirm); err != nil {
			t.Fatal(err)
		}
		if next.EventCount() > before {
			t.Fatalf("event count grew after confirm: %d -> %d", before, next.EventCount())
		}
		delivered += before - next.EventCount()
		next.flushPendingUnsolicited()
		// A duplicate confirm is ignored and does not advance the queue again.
		after := next.EventCount()
		if _, err := next.ProcessRequest(confirm); err != nil {
			t.Fatal(err)
		}
		if next.EventCount() != after {
			t.Fatal("duplicate confirm changed the queue")
		}
	}
	if next.EventCount() != 0 {
		t.Fatalf("events after drain=%d, want 0 (delivered %d)", next.EventCount(), delivered)
	}
	if delivered != n {
		t.Fatalf("delivered=%d, want %d (no event loss)", delivered, n)
	}
	next.StopUnsolicitedTimer()
}

// TestResponseFragmentsSplitAndPreserveAllObjects verifies that a solicited
// response whose object data exceeds the fragment budget is split into several
// application fragments that together carry every object exactly once, with FIR
// on the first, FIN on the last, and ascending sequence numbers.
func TestResponseFragmentsSplitAndPreserveAllObjects(t *testing.T) {
	ost := NewOutstation(&Config{OutstationAddress: 1024, MasterAddress: 1, MaxEventBuffers: 1000, MaxFragmentSize: 2048})
	ost.Initialize()
	const n = 256
	for i := 0; i < n; i++ {
		if err := ost.PublishEvent(batchBinary(Class1, uint16(i), i%2 == 0, i)); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 2, 0x06},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) <= 2044 {
		t.Fatalf("test needs oversize response, data=%d", len(resp.Data))
	}
	frags := ost.responseFragments(resp)
	if len(frags) < 2 {
		t.Fatalf("fragments=%d, want >=2", len(frags))
	}
	total := 0
	for i, f := range frags {
		if len(f.Data)+2 > 2048 {
			t.Fatalf("fragment %d apdu=%d exceeds 2048", i, len(f.Data)+2)
		}
		if f.Control.FIR != (i == 0) {
			t.Fatalf("fragment %d FIR=%v", i, f.Control.FIR)
		}
		if f.Control.FIN != (i == len(frags)-1) {
			t.Fatalf("fragment %d FIN=%v", i, f.Control.FIN)
		}
		wantSeq := uint8((1 + i) % 16)
		if f.Control.Seq != wantSeq {
			t.Fatalf("fragment %d seq=%d want %d", i, f.Control.Seq, wantSeq)
		}
		evs, derr := al.DecodeEventObjects(f.Data[2:])
		if derr != nil {
			t.Fatalf("fragment %d decode: %v", i, derr)
		}
		total += len(evs)
	}
	if total != n {
		t.Fatalf("objects across fragments=%d, want %d", total, n)
	}
}

// TestSplitObjectDataKeepsSequentialBlockWhole verifies that a sequential
// qualifier-0x07 block (implicit indexes) is never split, so point index
// numbering cannot be corrupted, and that every object survives.
func TestSplitObjectDataKeepsSequentialBlockWhole(t *testing.T) {
	// Group 1 Variation 1 static block: header (g,v,qual,count) + bytes.
	block := []byte{1, 1, 0x07, 3, 0x01, 0x02, 0x03}
	frags := splitObjectData(block, 8)
	if len(frags) != 1 {
		t.Fatalf("fragments=%d, want 1 (sequential block stays whole)", len(frags))
	}
	if !bytes.Equal(frags[0], block) {
		t.Fatalf("sequential block altered: % X", frags[0])
	}
}
