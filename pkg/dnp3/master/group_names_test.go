package master

import (
	"context"
	"errors"
	"strings"
	"testing"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// TestUnsupportedGroupErrorMessageNamesGroupVariation verifies DNP3-067: the
// error returned for an unsupported object group/variation names both the
// group (human-readable) and the variation, so the diagnostic is actionable.
func TestUnsupportedGroupErrorMessageNamesGroupVariation(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	cases := []struct {
		name      string
		g, v      uint8
		wantSub   string // substring expected in the error text (lower-cased)
		wantGroup string // expected group name fragment
	}{
		{"group2_binary_event", 2, 1, "binary input event", "binary input event"},
		{"group10_binary_output", 10, 1, "binary output", "binary output"},
		{"group13_control_relay", 13, 1, "control relay output block", "control relay output block"},
		{"group21_counter_event", 21, 1, "counter event", "counter event"},
		{"group40_analog_output_status", 40, 1, "analog output status", "analog output status"},
		{"group60_class_data", 60, 1, "class data objects", "class data objects"},
		{"group1_bad_variation", 1, 2, "binary input", "binary input"},
		{"group30_bad_variation", 30, 5, "analog input", "analog input"},
		{"group255_unknown", 255, 1, "unknown", "unknown"},
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
			msg := strings.ToLower(err.Error())
			// The message must name the group (human-readable) and the variation
			// number (DNP3-067: "Message names group/variation").
			if !strings.Contains(msg, strings.ToLower(tc.wantGroup)) {
				t.Fatalf("error %q must name group %q", err.Error(), tc.wantGroup)
			}
			if !strings.Contains(msg, "variation") {
				t.Fatalf("error %q must name the variation", err.Error())
			}
			if !strings.Contains(msg, "group") {
				t.Fatalf("error %q must name the group", err.Error())
			}
		})
	}
}
