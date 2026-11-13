package al

import (
	"bytes"
	"testing"
	"time"
)

func TestDNPTimeRoundTripAndRejects(t *testing.T) {
	stamp := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC)
	enc, err := EncodeDNPTime(stamp)
	if err != nil {
		t.Fatal(err)
	}
	want := [6]byte{0xE8, 0x03, 0x00, 0x00, 0x00, 0x00}
	if enc != want {
		t.Fatalf("encode = % X, want % X", enc, want)
	}
	got, err := DecodeDNPTime(enc[:])
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(stamp) {
		t.Fatalf("decode = %s, want %s", got, stamp)
	}
	if _, err := EncodeDNPTime(time.Time{}); err == nil {
		t.Fatal("zero time must be rejected")
	}
	if _, err := EncodeDNPTime(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("pre-epoch time must be rejected")
	}
	if _, err := EncodeDNPTime(time.Date(30000, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("48-bit overflow must be rejected")
	}
}

// TestDecodeDNPTimePastDurationLimit pins the year-2300 and 48-bit-maximum
// timestamps. time.Duration overflows near year 2290, and the old Add path
// decoded year 2300 as a date in 1715.
func TestDecodeDNPTimePastDurationLimit(t *testing.T) {
	stamp := time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC)
	enc, err := EncodeDNPTime(stamp)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeDNPTime(enc[:])
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(stamp) {
		t.Fatalf("year 2300 decode = %s", got.UTC())
	}
	again, err := EncodeDNPTime(got)
	if err != nil {
		t.Fatal(err)
	}
	if again != enc {
		t.Fatalf("year 2300 re-encode = % X, want % X", again, enc)
	}

	var maxb [6]byte
	u := uint64(DNPTimeMax)
	for i := 0; i < 6; i++ {
		maxb[i] = byte(u)
		u >>= 8
	}
	got, err = DecodeDNPTime(maxb[:])
	if err != nil {
		t.Fatal(err)
	}
	want := time.UnixMilli(DNPEpoch.UnixMilli() + int64(DNPTimeMax)).UTC()
	if !got.Equal(want) {
		t.Fatalf("48-bit max decode = %s, want %s", got.UTC(), want)
	}
	back, err := EncodeDNPTime(got)
	if err != nil {
		t.Fatal(err)
	}
	if back != maxb {
		t.Fatalf("48-bit max re-encode = % X, want % X", back, maxb)
	}
}

func TestEventObjectGoldens(t *testing.T) {
	stamp := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC)
	bin, err := PutBinaryEventObject(BinEventOnline|BinEventState, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	g2v1, err := AppendIndexedEvents(nil, 2, 1, [][]byte{append([]byte{7}, bin...)})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(g2v1, []byte{0x02, 0x01, 0x17, 0x01, 0x07, 0x81}) {
		t.Fatalf("G2V1 = % X", g2v1)
	}

	binT, err := PutBinaryEventObject(0x81, true, stamp)
	if err != nil {
		t.Fatal(err)
	}
	g2v2, err := AppendIndexedEvents(nil, 2, 2, [][]byte{append([]byte{7}, binT...)})
	if err != nil {
		t.Fatal(err)
	}
	want2 := []byte{0x02, 0x02, 0x17, 0x01, 0x07, 0x81, 0xE8, 0x03, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(g2v2, want2) {
		t.Fatalf("G2V2 = % X, want % X", g2v2, want2)
	}

	an, err := PutAnalogEventObject(AnalogEventOnline, -2, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	g32, err := AppendIndexedEvents(nil, 32, 1, [][]byte{append([]byte{3}, an...)})
	if err != nil {
		t.Fatal(err)
	}
	want32 := []byte{0x20, 0x01, 0x17, 0x01, 0x03, 0x01, 0xFE, 0xFF, 0xFF, 0xFF}
	if !bytes.Equal(g32, want32) {
		t.Fatalf("G32V1 = % X, want % X", g32, want32)
	}

	anT, err := PutAnalogEventObject(AnalogEventOnline, -2, true, stamp)
	if err != nil {
		t.Fatal(err)
	}
	g32t, err := AppendIndexedEvents(nil, 32, 3, [][]byte{append([]byte{3}, anT...)})
	if err != nil {
		t.Fatal(err)
	}
	if g32t[0] != 32 || g32t[1] != 3 || len(g32t) != 4+1+11 {
		t.Fatalf("G32V3 = % X", g32t)
	}

	ctr, err := PutCounterEventObject(CounterEventOnline|CounterEventRollover, 0x01020304, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	g22, err := AppendIndexedEvents(nil, 22, 1, [][]byte{append([]byte{4}, ctr...)})
	if err != nil {
		t.Fatal(err)
	}
	want22 := []byte{0x16, 0x01, 0x17, 0x01, 0x04, 0x21, 0x04, 0x03, 0x02, 0x01}
	if !bytes.Equal(g22, want22) {
		t.Fatalf("G22V1 = % X, want % X", g22, want22)
	}

	ctrT, err := PutCounterEventObject(0x21, 0x01020304, true, stamp)
	if err != nil {
		t.Fatal(err)
	}
	g22t, err := AppendIndexedEvents(nil, 22, 5, [][]byte{append([]byte{4}, ctrT...)})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeEventObjects(append(append(append([]byte{}, g2v2...), g32t...), g22t...))
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 3 {
		t.Fatalf("decoded %d events", len(decoded))
	}
	if !decoded[0].HasTime || !decoded[0].Time.Equal(stamp) || decoded[0].Flags != 0x81 || decoded[0].Index != 7 {
		t.Fatalf("binary %+v", decoded[0])
	}
	if decoded[1].Analog != -2 || !decoded[1].Time.Equal(stamp) {
		t.Fatalf("analog %+v", decoded[1])
	}
	if decoded[2].Counter != 0x01020304 || !decoded[2].Time.Equal(stamp) {
		t.Fatalf("counter %+v", decoded[2])
	}
}

func TestDecodeRejectsShortEvent(t *testing.T) {
	if _, err := DecodeEventObjects([]byte{0x02, 0x01, 0x17, 0x01}); err == nil {
		t.Fatal("truncated object must fail")
	}
}

// TestDecodeEventQualifier16 pins OpenDNP3 3.1.2's event layout: qualifier
// 0x28 is a 16-bit count followed by a 16-bit index on each object.
func TestDecodeEventQualifier16(t *testing.T) {
	stamp := time.Date(2021, 6, 7, 8, 9, 10, 0, time.UTC)
	body, err := PutBinaryEventObject(BinEventOnline|BinEventState, true, stamp)
	if err != nil {
		t.Fatal(err)
	}
	// g2v2, count 1, index 300 (0x012C).
	frame := []byte{0x02, 0x02, qualCountIndex16, 0x01, 0x00, 0x2C, 0x01}
	frame = append(frame, body...)
	decoded, err := DecodeEventObjects(frame)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 || decoded[0].Index != 300 || !decoded[0].Time.Equal(stamp) || decoded[0].Flags != 0x81 {
		t.Fatalf("decoded %+v", decoded)
	}
	// A 16-bit header with no room for the index and object must fail.
	if _, err := DecodeEventObjects([]byte{0x02, 0x02, qualCountIndex16, 0x01, 0x00, 0x07}); err == nil {
		t.Fatal("truncated 0x28 object must fail")
	}
	// Count zero is an empty event response, not static data.
	empty, err := DecodeEventObjects([]byte{0x02, 0x02, qualCountIndex16, 0x00, 0x00})
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty 0x28 = %+v err=%v", empty, err)
	}
}
