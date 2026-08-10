package master

import (
	"encoding/hex"
	"testing"
)

func TestBuildCROBRequestGoldenVector(t *testing.T) {
	crob := &CROB{Code: 7, Count: 1, OnTime: 1000, OffTime: 2000, Status: 0}
	got := buildCROBRequest(0x1234, crob)
	want, err := hex.DecodeString("0c01000134120701e8030000d007000000")
	if err != nil { t.Fatal(err) }
	if string(got) != string(want) {
		t.Fatalf("CROB bytes = %X, want %X", got, want)
	}
}
