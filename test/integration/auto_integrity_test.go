package integration

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/testutils"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

// DeviceRestart is IIN1.7 (bit 0x01 of the first IIN octet).
var deviceRestartIIN = [2]byte{0x01, 0x00}

// TestAutoIntegrityOnRestart verifies DNP3-053: when AutoIntegrityOnRestart is
// enabled, a response carrying the DeviceRestart IIN bit automatically triggers
// a Class-0 integrity poll (G1, G20, G30) after the triggering Read completes.
func TestAutoIntegrityOnRestart(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 7, Quality: types.QualityOnline},
	})
	sim.SetCounters([]*types.Counter{
		{Index: 0, Value: 3, Quality: types.QualityOnline},
	})
	// The FIRST application response carries DeviceRestart; later responses are
	// clean (one-shot injection).
	sim.SetNextResponseIIN(deviceRestartIIN)

	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
			master.WithRetry(1, 0),
			master.WithAutoIntegrityOnRestart(),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// A single Read of G1 — its response carries DeviceRestart, so the master
	// must auto-poll integrity (G1, G20, G30) afterward.
	if _, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1})); err != nil {
		t.Fatalf("trigger Read: %v", err)
	}

	groups := sim.ReadGroups()
	// Expected sequence: trigger Read G1, then auto integrity G1, G20, G30.
	want := []uint8{1, 1, 20, 30}
	if len(groups) != len(want) {
		t.Fatalf("ReadGroups = %v, want %v", groups, want)
	}
	for i, g := range groups {
		if g != want[i] {
			t.Fatalf("ReadGroups = %v, want %v (mismatch at %d)", groups, want, i)
		}
	}

	// The auto-poll refreshed the master's stored IIN; after it, LastIIN must be
	// clean (the auto-poll responses carried no DeviceRestart) and no further
	// integrity poll is pending.
	if got := client.LastIIN(); got != [2]byte{0, 0} {
		t.Fatalf("LastIIN after auto-integrity = % X, want 0000", got)
	}
}

// TestAutoIntegrityOnRestartDisabled verifies that with the default (disabled)
// policy, a DeviceRestart IIN does NOT trigger an automatic integrity poll —
// the caller is responsible for polling manually. This locks the opt-in
// behavior of the DNP3-053 config flag.
func TestAutoIntegrityOnRestartDisabled(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetNextResponseIIN(deviceRestartIIN)

	// Default config: AutoIntegrityOnRestart is false.
	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
			master.WithRetry(1, 0),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	if _, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1})); err != nil {
		t.Fatalf("trigger Read: %v", err)
	}

	// Only the single trigger Read should have been handled — no follow-on
	// integrity poll.
	groups := sim.ReadGroups()
	if len(groups) != 1 || groups[0] != 1 {
		t.Fatalf("ReadGroups = %v, want [1] (no auto integrity when disabled)", groups)
	}
}

// TestAutoIntegrityOnRestartFromIntegrityPoll verifies the recursion guard
// (DNP3-053): when an explicit IntegrityPoll's group reads see DeviceRestart,
// it does NOT spawn a nested auto integrity poll. Only the three integrity
// group reads are observed.
func TestAutoIntegrityOnRestartFromIntegrityPoll(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	// DeviceRestart lands on the FIRST group read of the integrity poll.
	sim.SetNextResponseIIN(deviceRestartIIN)

	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
			master.WithRetry(1, 0),
			master.WithAutoIntegrityOnRestart(),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	if _, err := client.IntegrityPoll(ctx); err != nil {
		t.Fatalf("IntegrityPoll: %v", err)
	}

	// Exactly the three integrity group reads — no nested auto-poll.
	groups := sim.ReadGroups()
	want := []uint8{1, 20, 30}
	if len(groups) != len(want) {
		t.Fatalf("ReadGroups = %v, want %v (no nested auto-integrity)", groups, want)
	}
	for i, g := range groups {
		if g != want[i] {
			t.Fatalf("ReadGroups = %v, want %v (mismatch at %d)", groups, want, i)
		}
	}
}
