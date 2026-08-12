package al

import "testing"

// TestGroupNameKnown checks the human-readable labels for the groups the v0
// profile and common DNP3 data objects reference (DNP3-067).
func TestGroupNameKnown(t *testing.T) {
	cases := []struct {
		g    uint8
		want string
	}{
		{1, "binary input"},
		{20, "counter"},
		{30, "analog input"},
		{13, "control relay output block"},
		{60, "class data objects (integrity scan)"},
	}
	for _, tc := range cases {
		if got := GroupName(tc.g); got != tc.want {
			t.Fatalf("GroupName(%d) = %q, want %q", tc.g, got, tc.want)
		}
	}
}

// TestGroupNameUnknown ensures an unrecognized group yields a stable label.
func TestGroupNameUnknown(t *testing.T) {
	if got := GroupName(255); got != "unknown" {
		t.Fatalf("GroupName(255) = %q, want %q", got, "unknown")
	}
}

// TestDescribeObjectNamesGroupAndVariation verifies DNP3-067: DescribeObject
// names both the group (human-readable) and the variation.
func TestDescribeObjectNamesGroupAndVariation(t *testing.T) {
	got := DescribeObject(1, 2)
	if !contains(got, "group 1") || !contains(got, "binary input") || !contains(got, "variation 2") {
		t.Fatalf("DescribeObject(1,2) = %q must name group 1 (binary input) and variation 2", got)
	}
	// A known variation includes its label.
	got2 := DescribeObject(30, 1)
	if !contains(got2, "32-bit signed") {
		t.Fatalf("DescribeObject(30,1) = %q must name the variation label", got2)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
