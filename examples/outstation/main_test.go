package main

import (
	"testing"
)

// TestOutstationExampleCompiles locks the DNP3-095 acceptance criterion: the
// minimal outstation example must compile against the public v0 MVP API.
// (A build-only test; running it requires binding port 20000 and is out of
// scope for the compile gate.)
func TestOutstationExampleCompiles(t *testing.T) {
	t.Log("examples/outstation package is compiled by the build/vet gate")
}
