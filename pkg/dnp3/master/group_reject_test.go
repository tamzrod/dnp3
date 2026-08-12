package master

import (
	"context"
	"errors"
	"testing"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// TestReadRejectsUnsupportedGroups asserts the public Read rejects object
// groups/variations outside the v0 profile with ErrUnsupportedGroup before any
// wire traffic is generated (DNP3-029). The v0 read profile is G1, G20, G30
// (variation 0 "any" or 1).
func TestReadRejectsUnsupportedGroups(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	cases := []struct {
		name string
		g    uint8
		v    uint8
	}{
		{"group2_binary_event", 2, 1},
		{"group10_binary_output", 10, 1},
		{"group13_binary_cmd_event", 13, 1},
		{"group21_counter_event", 21, 1},
		{"group32_analog_event", 32, 1},
		{"group40_analog_output", 40, 1},
		{"group60_class_data", 60, 1},
		{"group1_bad_variation", 1, 2},
		{"group30_bad_variation", 30, 5},
		{"group0_invalid", 0, 0},
		{"group255_invalid", 255, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cc.Read(context.Background(), &types.ReadRequest{
				Groups: []types.GroupRequest{{Group: tc.g, Variation: tc.v}},
			})
			if err == nil {
				t.Fatalf("expected error for group %d variation %d", tc.g, tc.v)
			}
			if !errors.Is(err, dnp3.ErrUnsupportedGroup) {
				t.Fatalf("error = %v, want ErrUnsupportedGroup", err)
			}
		})
	}
}

// TestReadAcceptsSupportedGroups asserts the v0-supported groups are NOT
// rejected at the gate (they proceed to the request path). Variation 0 ("any")
// and 1 are both accepted.
func TestReadAcceptsSupportedGroups(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	cases := []struct {
		name string
		g    uint8
		v    uint8
	}{
		{"g1v0", 1, 0},
		{"g1v1", 1, 1},
		{"g20v0", 20, 0},
		{"g20v1", 20, 1},
		{"g30v0", 30, 0},
		{"g30v1", 30, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cc.Read(context.Background(), &types.ReadRequest{
				Groups: []types.GroupRequest{{Group: tc.g, Variation: tc.v}},
			})
			if errors.Is(err, dnp3.ErrUnsupportedGroup) {
				t.Fatalf("group %d variation %d should be supported, got %v", tc.g, tc.v, err)
			}
		})
	}
}
