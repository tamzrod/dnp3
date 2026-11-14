package outstation

import (
	"sort"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
)

const (
	unsolDefaultConfirmWait = 2 * time.Second
	unsolMaxRetries         = 2
)

// ValidateEvent checks class, index, and absolute-time bounds.
func ValidateEvent(ev Event) error {
	if ev.Class != Class1 && ev.Class != Class2 && ev.Class != Class3 {
		return ErrEventClass
	}
	if ev.Index > 255 {
		return ErrEventIndex
	}
	if ev.HasTime {
		if _, err := al.EncodeDNPTime(ev.Time); err != nil {
			return err
		}
	}
	if ev.Kind != EventKindBinary && ev.Kind != EventKindAnalog && ev.Kind != EventKindCounter {
		return ErrEventClass
	}
	return nil
}

// PublishEvent queues one event and, when unsolicited reporting is enabled,
// sends it. The caller timestamp is stored on the event.
func (o *Outstation) PublishEvent(ev Event) error {
	if err := ValidateEvent(ev); err != nil {
		return err
	}
	if !o.eventQueue.Add(ev) {
		o.mu.Lock()
		o.refreshEventIINLocked()
		o.mu.Unlock()
		return ErrEventBufferFull
	}
	o.NotifyEvents()
	return nil
}

// UseEventQueue attaches a queue that outlives this connection.
func (o *Outstation) UseEventQueue(q *EventQueue) {
	if q == nil {
		return
	}
	o.mu.Lock()
	o.eventQueue = q
	o.refreshEventIINLocked()
	o.mu.Unlock()
}

// NotifyEvents refreshes class IIN bits and sends an unsolicited fragment
// when at least one class is enabled and no confirm is outstanding.
func (o *Outstation) NotifyEvents() {
	o.mu.Lock()
	o.refreshEventIINLocked()
	o.mu.Unlock()
	o.maybeSendUnsolicited()
}

func (o *Outstation) refreshEventIINLocked() {
	o.iin.Class1Events = o.eventQueue.CountClass(Class1) > 0
	o.iin.Class2Events = o.eventQueue.CountClass(Class2) > 0
	o.iin.Class3Events = o.eventQueue.CountClass(Class3) > 0
	o.iin.BufferOverflow = o.eventQueue.IsFull()
}

// CountClass returns the number of queued events in one class.
func (q *EventQueue) CountClass(class EventClass) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.buffers[class])
}

// SnapshotClass returns a time-ordered copy of one class.
func (q *EventQueue) SnapshotClass(class EventClass) []Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := append([]Event(nil), q.buffers[class]...)
	sortEvents(out)
	return out
}

// RemoveSeqs deletes events with the given insertion sequences.
func (q *EventQueue) RemoveSeqs(ids []uint64) {
	if len(ids) == 0 {
		return
	}
	drop := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		drop[id] = struct{}{}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for class, evs := range q.buffers {
		kept := evs[:0]
		for _, ev := range evs {
			if _, ok := drop[ev.Seq]; ok {
				continue
			}
			kept = append(kept, ev)
		}
		q.buffers[class] = kept
	}
}

// Restore re-inserts one event that was removed from the queue but whose
// response could not be delivered. It keeps the event's original insertion
// sequence number and class buffer, so a later read re-encodes exactly the
// event that was lost. The caller must ensure the buffer has room; a
// just-removed event leaves the buffer no fuller than before.
func (q *EventQueue) Restore(ev Event) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.buffers[ev.Class]) >= q.maxSize/3 {
		return false
	}
	q.buffers[ev.Class] = append(q.buffers[ev.Class], ev)
	if ev.Seq >= q.nextSeq {
		q.nextSeq = ev.Seq + 1
	}
	return true
}

func sortEvents(evs []Event) {
	sort.SliceStable(evs, func(i, j int) bool {
		a, b := evs[i], evs[j]
		if a.HasTime && b.HasTime && !a.Time.Equal(b.Time) {
			return a.Time.Before(b.Time)
		}
		return a.Seq < b.Seq
	})
}

func (o *Outstation) encodeClassRead(class EventClass) ([]byte, []uint64) {
	evs := o.eventQueue.SnapshotClass(class)
	return encodeEventList(evs, 0)
}

func (o *Outstation) encodeDirectRead(group, variation uint8, hdr []byte) ([]byte, []uint64) {
	if len(hdr) < 3 {
		return nil, nil
	}
	qual := hdr[2]
	var limit int
	switch qual {
	case 0x06:
		limit = -1
	case 0x07, 0x17:
		if len(hdr) < 4 || hdr[3] == 0 {
			return nil, nil
		}
		limit = int(hdr[3])
	case 0x27:
		if len(hdr) < 5 {
			return nil, nil
		}
		n := int(hdr[3]) | int(hdr[4])<<8
		if n == 0 {
			return nil, nil
		}
		limit = n
	default:
		return nil, nil
	}
	var matched []Event
	for _, class := range []EventClass{Class1, Class2, Class3} {
		for _, ev := range o.eventQueue.SnapshotClass(class) {
			if ev.Group == group || kindGroup(ev.Kind) == group {
				matched = append(matched, ev)
			}
		}
	}
	sortEvents(matched)
	if limit >= 0 && limit < len(matched) {
		matched = matched[:limit]
	}
	return encodeEventList(matched, variation)
}

func kindGroup(kind uint8) uint8 {
	switch kind {
	case EventKindBinary:
		return 2
	case EventKindCounter:
		return 22
	case EventKindAnalog:
		return 32
	default:
		return 0
	}
}

type encodedObj struct {
	id        uint64
	group     uint8
	variation uint8
	body      []byte
}

// encodeEventListBounded is like encodeEventList but stops adding objects once
// the encoded payload would exceed maxData octets. It is used for unsolicited
// responses, which must fit the peer's transport receive buffer as a single
// application fragment. Always encodes at least one object so a caller cannot
// stall; the remaining events stay queued.
func encodeEventListBounded(evs []Event, maxData int) ([]byte, []uint64) {
	if maxData < 8 {
		maxData = 8
	}
	var objs []encodedObj
	for _, ev := range evs {
		obj, ok := encodeOne(ev, 0)
		if !ok {
			continue
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil, nil
	}
	var dst []byte
	var ids []uint64
	i := 0
	for i < len(objs) {
		j := i + 1
		for j < len(objs) && objs[j].group == objs[i].group && objs[j].variation == objs[i].variation {
			j++
		}
		start := i
		for start < j {
			per := len(objs[start].body)
			avail := maxData - len(dst) - appFragmentOverhead
			n := avail / per
			if n < 1 {
				if len(dst) == 0 {
					// Guarantee forward progress with a single object.
					n = 1
				} else {
					return dst, ids
				}
			}
			if n > j-start {
				n = j - start
			}
			if n > maxObjectsPerHeader {
				n = maxObjectsPerHeader
			}
			bodies := make([][]byte, 0, n)
			for _, obj := range objs[start : start+n] {
				bodies = append(bodies, obj.body)
			}
			var err error
			dst, err = al.AppendIndexedEvents(dst, objs[i].group, objs[i].variation, bodies)
			if err != nil {
				return dst, ids
			}
			for _, obj := range objs[start : start+n] {
				ids = append(ids, obj.id)
			}
			start += n
		}
		i = j
	}
	return dst, ids
}

// maxObjectsPerHeader is the largest object count an 8-bit qualifier (0x17)
// can carry in one object header (IEEE 1815 UINT8_CNT_UINT8_INDEX).
const maxObjectsPerHeader = 255

func encodeEventList(evs []Event, forceVar uint8) ([]byte, []uint64) {
	var objs []encodedObj
	for _, ev := range evs {
		obj, ok := encodeOne(ev, forceVar)
		if !ok {
			continue
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil, nil
	}
	var dst []byte
	var ids []uint64
	i := 0
	for i < len(objs) {
		j := i + 1
		for j < len(objs) && objs[j].group == objs[i].group && objs[j].variation == objs[i].variation {
			j++
		}
		// A same-variation run may exceed the 255-object limit of one
		// qualifier-0x17 header. Emit it as consecutive headers rather than
		// dropping the run. The transport layer fragments the resulting APDU
		// into link frames, so the header split is the only bound needed here.
		for start := i; start < j; start += maxObjectsPerHeader {
			end := start + maxObjectsPerHeader
			if end > j {
				end = j
			}
			bodies := make([][]byte, 0, end-start)
			for _, obj := range objs[start:end] {
				bodies = append(bodies, obj.body)
			}
			var err error
			dst, err = al.AppendIndexedEvents(dst, objs[i].group, objs[i].variation, bodies)
			if err != nil {
				// Return what was already encoded. The caller removes only
				// those ids, so no undelivered event is dropped or acknowledged.
				return dst, ids
			}
			for _, obj := range objs[start:end] {
				ids = append(ids, obj.id)
			}
		}
		i = j
	}
	return dst, ids
}

func encodeOne(ev Event, forceVar uint8) (encodedObj, bool) {
	if ev.Index > 255 {
		return encodedObj{}, false
	}
	withTime := ev.HasTime
	group := kindGroup(ev.Kind)
	if group == 0 {
		group = ev.Group
	}
	var variation uint8
	switch ev.Kind {
	case EventKindBinary:
		variation = 1
		if withTime {
			variation = 2
		}
	case EventKindAnalog:
		variation = 1
		if withTime {
			variation = 3
		}
	case EventKindCounter:
		variation = 1
		if withTime {
			variation = 5
		}
	default:
		return encodedObj{}, false
	}
	if forceVar != 0 {
		if !variationCompatible(ev.Kind, forceVar, ev.HasTime) {
			return encodedObj{}, false
		}
		variation = forceVar
		withTime = variationHasTime(ev.Kind, forceVar)
	}
	var raw []byte
	var err error
	switch ev.Kind {
	case EventKindBinary:
		flags := ev.Quality &^ al.BinEventState
		if ev.BinaryValue {
			flags |= al.BinEventState
		}
		raw, err = al.PutBinaryEventObject(flags, withTime, ev.Time)
	case EventKindAnalog:
		raw, err = al.PutAnalogEventObject(ev.Quality, ev.AnalogValue, withTime, ev.Time)
	case EventKindCounter:
		raw, err = al.PutCounterEventObject(ev.Quality, ev.CounterValue, withTime, ev.Time)
	}
	if err != nil {
		return encodedObj{}, false
	}
	body := append([]byte{byte(ev.Index)}, raw...)
	return encodedObj{id: ev.Seq, group: group, variation: variation, body: body}, true
}

func variationHasTime(kind, variation uint8) bool {
	switch kind {
	case EventKindBinary:
		return variation == 2
	case EventKindAnalog:
		return variation == 3
	case EventKindCounter:
		return variation == 5
	default:
		return false
	}
}

func variationCompatible(kind, variation uint8, hasTime bool) bool {
	switch kind {
	case EventKindBinary:
		if variation == 1 {
			return true
		}
		return variation == 2 && hasTime
	case EventKindAnalog:
		if variation == 1 {
			return true
		}
		return variation == 3 && hasTime
	case EventKindCounter:
		if variation == 1 {
			return true
		}
		return variation == 5 && hasTime
	default:
		return false
	}
}

func (o *Outstation) applyUnsolicitedMask(data []byte, enable bool) {
	c1, c2, c3, saw := classMask(data)
	if len(data) > 0 && !saw {
		return
	}
	if c1 {
		o.unsolClass[0] = enable
	}
	if c2 {
		o.unsolClass[1] = enable
	}
	if c3 {
		o.unsolClass[2] = enable
	}
	o.unsolicited = o.unsolClass[0] || o.unsolClass[1] || o.unsolClass[2]
	if enable && saw {
		o.unsolSendAfter = true
	}
}

func classMask(data []byte) (c1, c2, c3, saw bool) {
	if len(data) == 0 {
		return true, true, true, true
	}
	off := 0
	for off+3 <= len(data) {
		g, v, q := data[off], data[off+1], data[off+2]
		size := requestObjectHeaderSize(data, off)
		if size == 0 || off+size > len(data) {
			break
		}
		if g == 60 && q == al.QualAllObjects {
			switch v {
			case 2:
				c1, saw = true, true
			case 3:
				c2, saw = true, true
			case 4:
				c3, saw = true, true
			}
		}
		off += size
	}
	return c1, c2, c3, saw
}

func (o *Outstation) flushPendingUnsolicited() {
	o.mu.Lock()
	pending := o.unsolSendAfter
	o.unsolSendAfter = false
	o.mu.Unlock()
	if pending {
		o.maybeSendUnsolicited()
	}
}

func (o *Outstation) maybeSendUnsolicited() {
	o.mu.Lock()
	frames, gen := o.prepareUnsolicitedLocked()
	transport := o.transport
	o.mu.Unlock()
	if len(frames) == 0 || transport == nil {
		return
	}
	if err := sendFrames(transport, frames); err != nil {
		o.mu.Lock()
		if o.unsolGen == gen {
			o.unsolAwait = false
			o.stopUnsolTimerLocked()
		}
		o.mu.Unlock()
	}
}

func (o *Outstation) prepareUnsolicitedLocked() ([][]byte, uint64) {
	if o.unsolAwait || o.transport == nil {
		return nil, 0
	}
	var batch []Event
	for i, on := range o.unsolClass {
		if !on {
			continue
		}
		batch = append(batch, o.eventQueue.SnapshotClass(EventClass(i))...)
	}
	if len(batch) == 0 {
		return nil, 0
	}
	sortEvents(batch)
	// An unsolicited response is a single application fragment: the peer's
	// transport receive buffer bounds it (OpenDNP3 3.1.2 default 2048). Encode
	// only as many events as fit; the rest stay queued and are sent after this
	// fragment is confirmed, mirroring OpenDNP3's per-fragment event loading.
	budget := o.maxAppFragment
	if budget <= appFragmentOverhead {
		budget = defaultMaxAppFragment
	}
	payload, ids := encodeEventListBounded(batch, budget-appFragmentOverhead)
	if len(payload) == 0 {
		return nil, 0
	}
	o.refreshEventIINLocked()
	iin := o.iin.Bytes()
	body := make([]byte, 0, 2+len(payload))
	body = append(body, iin[0], iin[1])
	body = append(body, payload...)
	apdu := al.NewUnsolicited(o.unsolSeq, body)
	frames, err := o.linkFrames(apdu)
	if err != nil || len(frames) == 0 {
		return nil, 0
	}
	o.unsolFrames = frames
	o.unsolIDs = ids
	o.unsolAwait = true
	o.unsolRetries = 0
	o.armUnsolTimerLocked()
	out := make([][]byte, len(frames))
	for i := range frames {
		out[i] = append([]byte(nil), frames[i]...)
	}
	return out, o.unsolGen
}

func (o *Outstation) linkFrames(apdu *al.APDU) ([][]byte, error) {
	parts := o.buildTransportFragments(apdu)
	out := make([][]byte, 0, len(parts))
	for _, frag := range parts {
		dll := &frame.Frame{
			Control: frame.Control{
				DIR:      false,
				PRM:      true,
				FuncCode: frame.FuncUnconfirmedUserData,
			},
			DestAddr: o.config.MasterAddress,
			SrcAddr:  o.config.OutstationAddress,
			Data:     frag,
		}
		enc, err := frame.Encode(dll)
		if err != nil {
			return nil, err
		}
		out = append(out, enc)
	}
	return out, nil
}

func sendFrames(transport TransportHandler, frames [][]byte) error {
	for _, fr := range frames {
		if err := transport.Send(fr); err != nil {
			return err
		}
	}
	return nil
}

func (o *Outstation) confirmWait() time.Duration {
	if o.unsolConfirmWait > 0 {
		return o.unsolConfirmWait
	}
	return unsolDefaultConfirmWait
}

func (o *Outstation) armUnsolTimerLocked() {
	o.unsolGen++
	gen := o.unsolGen
	if o.unsolTimer != nil {
		o.unsolTimer.Stop()
	}
	wait := o.confirmWait()
	o.unsolTimer = time.AfterFunc(wait, func() { o.onUnsolTimeout(gen) })
}

func (o *Outstation) stopUnsolTimerLocked() {
	o.unsolGen++
	if o.unsolTimer != nil {
		o.unsolTimer.Stop()
		o.unsolTimer = nil
	}
}

func (o *Outstation) onUnsolTimeout(gen uint64) {
	o.mu.Lock()
	if !o.unsolAwait || o.unsolGen != gen {
		o.mu.Unlock()
		return
	}
	if o.unsolRetries >= unsolMaxRetries {
		o.unsolAwait = false
		o.unsolRetries = 0
		o.stopUnsolTimerLocked()
		o.mu.Unlock()
		return
	}
	o.unsolRetries++
	frames := make([][]byte, len(o.unsolFrames))
	for i := range o.unsolFrames {
		frames[i] = append([]byte(nil), o.unsolFrames[i]...)
	}
	transport := o.transport
	o.armUnsolTimerLocked()
	o.mu.Unlock()
	if transport == nil {
		return
	}
	if err := sendFrames(transport, frames); err != nil {
		o.mu.Lock()
		o.unsolAwait = false
		o.stopUnsolTimerLocked()
		o.mu.Unlock()
	}
}

func (o *Outstation) acceptConfirmLocked(seq uint8) {
	if !o.unsolAwait || seq != o.unsolSeq {
		return
	}
	o.eventQueue.RemoveSeqs(o.unsolIDs)
	o.unsolAwait = false
	o.unsolFrames = nil
	o.unsolIDs = nil
	o.unsolRetries = 0
	o.unsolSeq = (o.unsolSeq + 1) % 16
	o.stopUnsolTimerLocked()
	o.refreshEventIINLocked()
	o.unsolSendAfter = o.pendingEnabledEventsLocked()
}

func (o *Outstation) pendingEnabledEventsLocked() bool {
	for i, on := range o.unsolClass {
		if on && o.eventQueue.CountClass(EventClass(i)) > 0 {
			return true
		}
	}
	return false
}

func (o *Outstation) cancelInFlightLocked(removed []uint64) {
	if !o.unsolAwait || len(removed) == 0 {
		return
	}
	drop := make(map[uint64]struct{}, len(removed))
	for _, id := range removed {
		drop[id] = struct{}{}
	}
	hit := false
	for _, id := range o.unsolIDs {
		if _, ok := drop[id]; ok {
			hit = true
			break
		}
	}
	if !hit {
		return
	}
	o.unsolAwait = false
	o.unsolFrames = nil
	o.unsolIDs = nil
	o.unsolRetries = 0
	o.unsolSeq = (o.unsolSeq + 1) % 16
	o.stopUnsolTimerLocked()
}

// StopUnsolicitedTimer stops a confirm timer. Called when the outstation stops.
func (o *Outstation) StopUnsolicitedTimer() {
	o.mu.Lock()
	o.unsolAwait = false
	o.stopUnsolTimerLocked()
	o.mu.Unlock()
}
