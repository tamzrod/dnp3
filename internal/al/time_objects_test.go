package al

import (
	"bytes"
	"testing"
	"time"
)

// Independent wire vectors for the time objects. The G50V1 timestamp vector
// is the same independently derived value used by the event-object epoch tests
// (2026-11-19T11:02:56.910Z == 1795086176910 ms == 8E 52 79 F3 A1 01 LSB-first).

func TestBuildG50V1ObjectGolden(t *testing.T) {
	stamp := time.Date(2026, 11, 19, 11, 2, 56, 910_000_000, time.UTC)
	got, err := BuildG50V1Object(stamp)
	if err != nil {
		t.Fatalf("BuildG50V1Object: %v", err)
	}
	want := []byte{0x32, 0x01, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}
	if !bytes.Equal(got, want) {
		t.Fatalf("G50V1 wire mismatch\n got: % X\nwant: % X", got, want)
	}
}

func TestParseG50V1ObjectForms(t *testing.T) {
	stamp := time.Date(2026, 11, 19, 11, 2, 56, 910_000_000, time.UTC)
	// 0x07 count, 0x17 indexed, 0x06 all-objects, 0x28 single-point range.
	cases := [][]byte{
		{0x32, 0x01, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
		{0x32, 0x01, 0x17, 0x01, 0x00, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
		{0x32, 0x01, 0x06, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
		{0x32, 0x01, 0x28, 0x00, 0x00, 0x00, 0x00, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
	}
	for i, data := range cases {
		got, err := ParseG50V1Object(data)
		if err != nil {
			t.Errorf("form %d: ParseG50V1Object: %v", i, err)
			continue
		}
		if !got.Equal(stamp) {
			t.Errorf("form %d: got %v, want %v", i, got, stamp)
		}
	}
}

func TestParseG50V1ObjectRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"wrong group", []byte{0x2A, 0x01, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}},
		{"wrong variation", []byte{0x32, 0x02, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}},
		{"multi-point", []byte{0x32, 0x01, 0x07, 0x02, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}},
		{"truncated time", []byte{0x32, 0x01, 0x07, 0x01, 0x8E, 0x52}},
		{"unsupported qualifier", []byte{0x32, 0x01, 0x00, 0x01, 0x00, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseG50V1Object(tc.data); err == nil {
				t.Fatal("expected rejection of malformed G50V1")
			}
		})
	}
}

func TestG52V1BuildParseGolden(t *testing.T) {
	got := BuildG52V1Object(0x1234)
	want := []byte{0x34, 0x01, 0x07, 0x01, 0x34, 0x12}
	if !bytes.Equal(got, want) {
		t.Fatalf("G52V1 wire mismatch\n got: % X\nwant: % X", got, want)
	}
	parsed, err := ParseG52V1Object(want)
	if err != nil || parsed != 0x1234 {
		t.Fatalf("ParseG52V1Object = %d, %v; want 0x1234", parsed, err)
	}
	if _, err := ParseG52V1Object([]byte{0x34, 0x01, 0x07, 0x01, 0x34}); err == nil {
		t.Fatal("expected rejection of truncated G52V1")
	}
}
