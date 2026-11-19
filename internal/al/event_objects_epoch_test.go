package al

import (
	"bytes"
	"testing"
	"time"
)

// Independent IEEE 1815 vectors. The expected octets below were derived outside
// the library from the 1970-01-01T00:00:00Z absolute-time origin:
//
//	2026-11-19T11:02:56.910Z -> 1_795_086_176_910 ms -> 8E 52 79 F3 A1 01
//
// They are not produced by this package's encoder or decoder.
const (
	vectorUTC   = "2026-11-19T11:02:56.910Z"
	vectorMS    = uint64(1795086176910)
	vectorBytes = "\x8e\x52\x79\xf3\xa1\x01"
)

func vectorTime(t *testing.T) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339Nano, vectorUTC)
	if err != nil {
		t.Fatal(err)
	}
	return v.UTC()
}

func TestDNPTimeKnownAbsoluteVectors(t *testing.T) {
	stamp := vectorTime(t)

	if got := uint64(stamp.UnixMilli()); got != vectorMS {
		t.Fatalf("UnixMilli vector = %d, want %d", got, vectorMS)
	}

	enc, err := EncodeDNPTime(stamp)
	if err != nil {
		t.Fatal(err)
	}
	want := [6]byte{0x8e, 0x52, 0x79, 0xf3, 0xa1, 0x01}
	if enc != want {
		t.Fatalf("EncodeDNPTime(%s) = % X, want % X (IEEE 1815 1970 epoch)", vectorUTC, enc, want)
	}
	if !bytes.Equal(enc[:], []byte(vectorBytes)) {
		t.Fatalf("encode bytes = % X, want % X", enc, vectorBytes)
	}

	got, err := DecodeDNPTime([]byte(vectorBytes))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(stamp) {
		t.Fatalf("DecodeDNPTime(% X) = %s, want %s", vectorBytes, got.UTC(), stamp)
	}
}

func TestEventKnownAbsoluteWireVectors(t *testing.T) {
	stamp := vectorTime(t)

	t.Run("G2V2", func(t *testing.T) {
		obj, err := PutBinaryEventObject(BinEventOnline|BinEventState, true, stamp)
		if err != nil {
			t.Fatal(err)
		}
		wire, err := AppendIndexedEvents(nil, 2, 2, [][]byte{append([]byte{7}, obj...)})
		if err != nil {
			t.Fatal(err)
		}
		want := []byte{0x02, 0x02, 0x17, 0x01, 0x07, 0x81, 0x8e, 0x52, 0x79, 0xf3, 0xa1, 0x01}
		if !bytes.Equal(wire, want) {
			t.Fatalf("G2V2 wire = % X, want % X", wire, want)
		}
		decoded, err := DecodeEventObjects(wire)
		if err != nil {
			t.Fatal(err)
		}
		assertEventTime(t, decoded, 2, 2, 7, stamp)
	})

	t.Run("G32V3", func(t *testing.T) {
		obj, err := PutAnalogEventObject(AnalogEventOnline, -2, true, stamp)
		if err != nil {
			t.Fatal(err)
		}
		wire, err := AppendIndexedEvents(nil, 32, 3, [][]byte{append([]byte{3}, obj...)})
		if err != nil {
			t.Fatal(err)
		}
		want := []byte{0x20, 0x03, 0x17, 0x01, 0x03, 0x01, 0xfe, 0xff, 0xff, 0xff, 0x8e, 0x52, 0x79, 0xf3, 0xa1, 0x01}
		if !bytes.Equal(wire, want) {
			t.Fatalf("G32V3 wire = % X, want % X", wire, want)
		}
		decoded, err := DecodeEventObjects(wire)
		if err != nil {
			t.Fatal(err)
		}
		assertEventTime(t, decoded, 32, 3, 3, stamp)
	})

	t.Run("G22V5", func(t *testing.T) {
		obj, err := PutCounterEventObject(CounterEventOnline, 0x01020304, true, stamp)
		if err != nil {
			t.Fatal(err)
		}
		wire, err := AppendIndexedEvents(nil, 22, 5, [][]byte{append([]byte{9}, obj...)})
		if err != nil {
			t.Fatal(err)
		}
		want := []byte{0x16, 0x05, 0x17, 0x01, 0x09, 0x01, 0x04, 0x03, 0x02, 0x01, 0x8e, 0x52, 0x79, 0xf3, 0xa1, 0x01}
		if !bytes.Equal(wire, want) {
			t.Fatalf("G22V5 wire = % X, want % X", wire, want)
		}
		decoded, err := DecodeEventObjects(wire)
		if err != nil {
			t.Fatal(err)
		}
		assertEventTime(t, decoded, 22, 5, 9, stamp)
	})
}

func assertEventTime(t *testing.T, evs []EventObject, group, variation uint8, index uint16, want time.Time) {
	t.Helper()
	if len(evs) != 1 {
		t.Fatalf("decoded %d events, want 1", len(evs))
	}
	ev := evs[0]
	if ev.Group != group || ev.Variation != variation || ev.Index != index {
		t.Fatalf("decoded g%dv%d index=%d, want g%dv%d index=%d", ev.Group, ev.Variation, ev.Index, group, variation, index)
	}
	if !ev.HasTime || !ev.Time.Equal(want) {
		t.Fatalf("decoded time = %s (has=%v), want %s", ev.Time.UTC(), ev.HasTime, want)
	}
}
