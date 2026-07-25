package types

import (
	"testing"
	"time"
)

func TestQualityFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    QualityFlags
		expected string
	}{
		{"Online only", QualityOnline, "ONLINE"},
		{"Online with restart", QualityOnline | QualityRestart, "ONLINE|RESTART"},
		{"Comm lost", QualityCommLost, "COMM_LOST"},
		{"Online and comm lost", QualityOnline | QualityCommLost, "ONLINE|COMM_LOST"},
		{"All flags", QualityOnline | QualityRestart | QualityCommLost | QualityDiscontinuous, "ONLINE|RESTART|COMM_LOST|DISCONTINUOUS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.String(); got != tt.expected {
				t.Errorf("QualityFlags.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestQualityFlags_IsGood(t *testing.T) {
	tests := []struct {
		name     string
		flags    QualityFlags
		expected bool
	}{
		{"Online only", QualityOnline, true},
		{"Online with restart", QualityOnline | QualityRestart, true},
		{"Comm lost", QualityCommLost, false},
		{"Online and comm lost", QualityOnline | QualityCommLost, false},
		{"Zero", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsGood(); got != tt.expected {
				t.Errorf("QualityFlags.IsGood() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBinaryInput(t *testing.T) {
	bi := NewBinaryInput(42, true, QualityOnline)

	if bi.Index != 42 {
		t.Errorf("BinaryInput.Index = %v, want 42", bi.Index)
	}
	if bi.Value != true {
		t.Errorf("BinaryInput.Value = %v, want true", bi.Value)
	}
	if bi.Quality != QualityOnline {
		t.Errorf("BinaryInput.Quality = %v, want QualityOnline", bi.Quality)
	}
}

func TestAnalogInput(t *testing.T) {
	ai := NewAnalogInput(100, 123.456, QualityOnline)

	if ai.Index != 100 {
		t.Errorf("AnalogInput.Index = %v, want 100", ai.Index)
	}
	if ai.Value != 123.456 {
		t.Errorf("AnalogInput.Value = %v, want 123.456", ai.Value)
	}
	if ai.Quality != QualityOnline {
		t.Errorf("AnalogInput.Quality = %v, want QualityOnline", ai.Quality)
	}
}

func TestCounter(t *testing.T) {
	c := NewCounter(5, 1000, QualityOnline)

	if c.Index != 5 {
		t.Errorf("Counter.Index = %v, want 5", c.Index)
	}
	if c.Value != 1000 {
		t.Errorf("Counter.Value = %v, want 1000", c.Value)
	}
	if c.Quality != QualityOnline {
		t.Errorf("Counter.Quality = %v, want QualityOnline", c.Quality)
	}
}

func TestTimestamp(t *testing.T) {
	// Test timestamp from known value
	ts := &Timestamp{Value: 0}

	if !ts.IsNull() {
		t.Error("Timestamp with Value=0 should be null")
	}

	// Test timestamp from time
	now := time.Now()
	duration := now.Sub(DNP3Epoch)
	ms := uint64(duration.Milliseconds())
	ts2 := &Timestamp{Value: ms}

	if ts2.IsNull() {
		t.Error("Timestamp from now() should not be null")
	}

	// Test Time() conversion
	ts3 := &Timestamp{Value: 0}
	_ = ts3.Time() // Should not panic
}

func TestTimestamp_Time(t *testing.T) {
	// Create a timestamp at known time
	expected := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	ts := &Timestamp{Value: 0}

	got := ts.Time()
	if !got.Equal(expected) {
		t.Errorf("Timestamp.Time() = %v, want %v", got, expected)
	}
}

func TestCommandTypes(t *testing.T) {
	tests := []struct {
		cmdType  CommandType
		expected string
	}{
		{SelectThenOperate, "SelectThenOperate"},
		{DirectOperate, "DirectOperate"},
		{DirectOperateNoResponse, "DirectOperateNoResponse"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			var got string
			switch tt.cmdType {
			case SelectThenOperate:
				got = "SelectThenOperate"
			case DirectOperate:
				got = "DirectOperate"
			case DirectOperateNoResponse:
				got = "DirectOperateNoResponse"
			}
			if got != tt.expected {
				t.Errorf("CommandType = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewBinaryControl(t *testing.T) {
	cmd := NewBinaryControl(42, true, DirectOperate)

	if cmd.Group != 12 {
		t.Errorf("BinaryControl.Group = %v, want 12", cmd.Group)
	}
	if cmd.Index != 42 {
		t.Errorf("BinaryControl.Index = %v, want 42", cmd.Index)
	}
	if cmd.CommandType != DirectOperate {
		t.Errorf("BinaryControl.CommandType = %v, want DirectOperate", cmd.CommandType)
	}
}

func TestNewAnalogControl(t *testing.T) {
	cmd := NewAnalogControl(10, 50.5, SelectThenOperate)

	if cmd.Group != 41 {
		t.Errorf("AnalogControl.Group = %v, want 41", cmd.Group)
	}
	if cmd.Index != 10 {
		t.Errorf("AnalogControl.Index = %v, want 10", cmd.Index)
	}
	if cmd.CommandType != SelectThenOperate {
		t.Errorf("AnalogControl.CommandType = %v, want SelectThenOperate", cmd.CommandType)
	}
}

func TestReadRequest(t *testing.T) {
	req := NewReadRequest(
		GroupRequest{Group: 1, Variation: 1},
		GroupRequest{Group: 30, Variation: 1},
	)

	if len(req.Groups) != 2 {
		t.Errorf("ReadRequest.Groups length = %v, want 2", len(req.Groups))
	}
	if req.Groups[0].Group != 1 {
		t.Errorf("ReadRequest.Groups[0].Group = %v, want 1", req.Groups[0].Group)
	}
}

func TestControlStatus(t *testing.T) {
	tests := []struct {
		status   ControlStatus
		expected string
	}{
		{ControlSuccess, "success"},
		{ControlTimeout, "timeout"},
		{ControlNotSupported, "not_supported"},
		{ControlBlocked, "blocked"},
		{ControlSuccess + 99, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("ControlStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
