package master_test

import (
"testing"
"time"

"dnp3/cmd/workbench/internal/logger"
"dnp3/cmd/workbench/internal/master"
)

func TestAutoPollToggle(t *testing.T) {
log := logger.New()
ctrl := master.NewController(log)

// Start controller
if err := ctrl.Start(); err != nil {
t.Fatalf("Start error: %v", err)
}
defer ctrl.Stop()

// Test auto-poll toggle ON
ctrl.EnableAutoPoll(true)
if !ctrl.IsAutoPollEnabled() {
t.Fatal("Auto-poll should be enabled")
}

// Test auto-poll toggle OFF
ctrl.EnableAutoPoll(false)
if ctrl.IsAutoPollEnabled() {
t.Fatal("Auto-poll should be disabled")
}

// Check State includes AutoPollEnabled
state := ctrl.State()
if state.AutoPollEnabled {
t.Fatal("AutoPollEnabled in State should be false")
}

t.Log("✅ Auto-poll toggle works correctly")
}

func TestAutoPollDisableOnDisconnect(t *testing.T) {
log := logger.New()
ctrl := master.NewController(log)

if err := ctrl.Start(); err != nil {
t.Fatalf("Start error: %v", err)
}
defer ctrl.Stop()

// Enable auto-poll
ctrl.EnableAutoPoll(true)

// Disconnect should disable auto-poll
ctrl.Disconnect()

// Small delay for goroutine to complete
time.Sleep(10 * time.Millisecond)

state := ctrl.State()
if state.AutoPollEnabled {
t.Fatal("AutoPollEnabled should be false after disconnect")
}

t.Log("✅ Auto-poll disabled on disconnect")
}
