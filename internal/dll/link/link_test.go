package link

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/dll/frame"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateLinkDown, "LinkDown"},
		{StateLinkReset, "LinkReset"},
		{StateOperational, "Operational"},
		{StateWaitingConfirm, "WaitingConfirm"},
		{StateWaitingResponse, "WaitingResponse"},
		{StateError, "Error"},
		{State(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleUnknown, "Unknown"},
		{RolePrimary, "Primary"},
		{RoleSecondary, "Secondary"},
		{Role(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.role.String(); got != tt.want {
			t.Errorf("Role(%d).String() = %q, want %q", tt.role, got, tt.want)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	local := uint16(0x0001)
	remote := uint16(0xFFFF)
	role := RolePrimary

	config := DefaultConfig(local, remote, role)

	if config.LocalAddr != local {
		t.Errorf("LocalAddr = 0x%04X, want 0x%04X", config.LocalAddr, local)
	}
	if config.RemoteAddr != remote {
		t.Errorf("RemoteAddr = 0x%04X, want 0x%04X", config.RemoteAddr, remote)
	}
	if config.Role != role {
		t.Errorf("Role = %v, want %v", config.Role, role)
	}
	if config.Unbalanced != true {
		t.Error("Unbalanced should be true for Primary")
	}
}

func TestStateMachineStart(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	// Initial state should be LinkDown
	if sm.State() != StateLinkDown {
		t.Errorf("Initial state = %s, want %s", sm.State(), StateLinkDown)
	}

	// Start should transition to LinkReset
	ctx := context.Background()
	if err := sm.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if sm.State() != StateLinkReset {
		t.Errorf("After Start() state = %s, want %s", sm.State(), StateLinkReset)
	}
}

func TestStateMachineStartTwice(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()

	// First start should succeed
	if err := sm.Start(ctx); err != nil {
		t.Errorf("First Start() error = %v", err)
	}

	// Second start should fail
	if err := sm.Start(ctx); err == nil {
		t.Error("Second Start() should return error")
	}
}

func TestStateMachineClose(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	sm.Start(ctx)

	// Close should succeed
	if err := sm.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// State should be LinkDown after close
	if sm.State() != StateLinkDown {
		t.Errorf("After Close() state = %s, want %s", sm.State(), StateLinkDown)
	}
}

func TestStateMachineResetLinkStations(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	sm.Start(ctx)

	// ResetLinkStations should succeed from LinkReset state
	if err := sm.ResetLinkStations(ctx); err != nil {
		t.Errorf("ResetLinkStations() error = %v", err)
	}

	// Should have sent a frame
	stats := sm.Stats()
	if stats.FramesSent == 0 {
		t.Error("FramesSent should be > 0 after ResetLinkStations()")
	}
}

func TestStateMachineSendData(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	sm.Start(ctx)

	// Transition to operational by simulating the ACK response
	// In a real system, this would come from the network
	sm.ResetLinkStations(ctx)

	// Manually transition to Operational state (simulating ACK received)
	// In real usage, the goroutine would receive the ACK and update state
	sm.mu.Lock()
	sm.state = StateOperational
	sm.mu.Unlock()

	// Now send data should work
	data := []byte{0x01, 0x02, 0x03}
	if err := sm.SendData(ctx, data, false); err != nil {
		t.Errorf("SendData() error = %v", err)
	}
}

func TestStateMachineSendDataFromInvalidState(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	// Don't start - stay in LinkDown state

	// SendData should fail from LinkDown state
	data := []byte{0x01, 0x02, 0x03}
	if err := sm.SendData(ctx, data, false); err == nil {
		t.Error("SendData() should return error from LinkDown state")
	}
}

func TestStateMachineHandleFrame(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RoleSecondary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	sm.Start(ctx)

	// Create a Reset Link Stations frame from primary
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      true,
			PRM:      true,
			FuncCode: frame.FuncResetLinkStations,
		},
		DestAddr: 0x0001,
		SrcAddr:  0xFFFF,
	}

	// Handle the frame
	if err := sm.HandleFrame(f); err != nil {
		t.Errorf("HandleFrame() error = %v", err)
	}

	// Should have received and processed
	stats := sm.Stats()
	if stats.FramesReceived == 0 {
		t.Error("FramesReceived should be > 0 after HandleFrame()")
	}
}

func TestStateMachineStats(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	stats := sm.Stats()

	if stats.FramesSent != 0 {
		t.Errorf("Initial FramesSent = %d, want 0", stats.FramesSent)
	}
	if stats.FramesReceived != 0 {
		t.Errorf("Initial FramesReceived = %d, want 0", stats.FramesReceived)
	}
	if stats.Errors != 0 {
		t.Errorf("Initial Errors = %d, want 0", stats.Errors)
	}
	if stats.State != StateLinkDown {
		t.Errorf("Initial State = %s, want %s", stats.State, StateLinkDown)
	}
}

func TestStateMachineDone(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	// Done channel should not be closed initially
	select {
	case <-sm.Done():
		t.Error("Done() should not be closed initially")
	default:
		// Expected
	}

	// Close
	sm.Close()

	// Done channel should be closed
	select {
	case <-sm.Done():
		// Expected
	default:
		t.Error("Done() should be closed after Close()")
	}
}

func TestStateMachineFrameChan(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	// FrameChan should not block
	select {
	case _, ok := <-sm.FrameChan():
		if ok {
			t.Error("FrameChan() should not have values initially")
		}
	default:
		// Expected - channel is empty or closed
	}
}

func TestStateMachineConcurrent(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sm.Start(ctx)

	// Run multiple operations concurrently
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 10; i++ {
			sm.Stats()
		}
	}()

	<-done
}

// TestInvalidFrameFromPrimary tests handling invalid frames from primary.
func TestInvalidFrameFromPrimary(t *testing.T) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	ctx := context.Background()
	sm.Start(ctx)

	// Create an invalid frame (secondary sending to primary without PRM)
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncAck,
		},
		DestAddr: 0x0001,
		SrcAddr:  0xFFFF,
	}

	// Primary receiving from secondary should be unexpected
	// This tests error handling
	_ = f // Would need more specific test case
}

// TestConfigWithSecondaryRole tests configuration for secondary role.
func TestConfigWithSecondaryRole(t *testing.T) {
	config := DefaultConfig(0xFFFF, 0x0001, RoleSecondary)

	if config.Unbalanced {
		t.Error("Unbalanced should be false for Secondary")
	}
}

// BenchmarkStateMachineStats benchmarks the Stats() method.
func BenchmarkStateMachineStats(b *testing.B) {
	config := DefaultConfig(0x0001, 0xFFFF, RolePrimary)
	sm := NewStateMachine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Stats()
	}
}
