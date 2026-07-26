package outstation

import (
	"testing"
	"time"

	"dnp3/internal/al"
)

// TestSBOSelectThenOperate tests the SBO select-then-operate flow.
func TestSBOSelectThenOperate(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create a SELECT request
	selectReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncSelect,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01}, // Group 12, Var 1, index 0
	}

	// Process SELECT
	resp, err := ost.ProcessRequest(selectReq)
	if err != nil {
		t.Fatalf("handleSelect failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response to SELECT")
	}

	// Verify pending select was registered
	if ost.PendingSelectCount() != 1 {
		t.Errorf("PendingSelectCount = %d, want 1", ost.PendingSelectCount())
	}

	// Create OPERATE request with same parameters
	operateReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncOperate,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01}, // Same as select
	}

	// Process OPERATE
	resp, err = ost.ProcessRequest(operateReq)
	if err != nil {
		t.Fatalf("handleOperate failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response to OPERATE")
	}

	// Verify pending select was cleared
	if ost.PendingSelectCount() != 0 {
		t.Errorf("PendingSelectCount = %d, want 0 after operate", ost.PendingSelectCount())
	}
}

// TestSBOOperateWithoutSelect tests that operate without prior select fails.
func TestSBOOperateWithoutSelect(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create OPERATE request without prior SELECT
	operateReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncOperate,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01},
	}

	// Process OPERATE - should fail
	resp, err := ost.ProcessRequest(operateReq)
	if err == nil {
		t.Fatal("Expected error for operate without select")
	}
	if err != ErrNoSelectPending {
		t.Errorf("Got error %v, want ErrNoSelectPending", err)
	}
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

// TestSBOSelectTimeout tests that select times out.
func TestSBOSelectTimeout(t *testing.T) {
	config := &Config{
		SBOTimeout: 100, // 100ms timeout for testing
	}
	ost := NewOutstation(config)
	ost.Initialize()
	ost.Start()

	// Create a SELECT request
	selectReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncSelect,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01},
	}

	// Process SELECT
	_, err := ost.ProcessRequest(selectReq)
	if err != nil {
		t.Fatalf("handleSelect failed: %v", err)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Create OPERATE request - should fail due to timeout
	operateReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncOperate,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01},
	}

	_, err = ost.ProcessRequest(operateReq)
	if err == nil {
		t.Fatal("Expected error for operate after select timeout")
	}
	if err != ErrSelectTimeout {
		t.Errorf("Got error %v, want ErrSelectTimeout", err)
	}
}

// TestSBOClearPendingSelects tests clearing pending selects.
func TestSBOClearPendingSelects(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create a SELECT request
	selectReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncSelect,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01},
	}
	ost.ProcessRequest(selectReq)

	if ost.PendingSelectCount() != 1 {
		t.Fatalf("PendingSelectCount = %d, want 1", ost.PendingSelectCount())
	}

	// Clear pending selects
	ost.ClearPendingSelects()

	if ost.PendingSelectCount() != 0 {
		t.Errorf("PendingSelectCount = %d, want 0 after clear", ost.PendingSelectCount())
	}
}

// TestEventQueueAdd tests adding events to the queue.
func TestEventQueueAdd(t *testing.T) {
	queue := NewEventQueue(100)

	// Add an event
	event := Event{
		Class:   Class1,
		Group:   2,
		Index:   0,
		Value:   []byte{0x01},
		Quality: BinaryQualityOnline,
		Time:    time.Now(),
	}

	if !queue.Add(event) {
		t.Error("Expected Add to succeed")
	}

	if queue.Count() != 1 {
		t.Errorf("Count = %d, want 1", queue.Count())
	}
}

// TestEventQueueOverflow tests buffer overflow behavior.
func TestEventQueueOverflow(t *testing.T) {
	queue := NewEventQueue(6) // Small buffer for testing

	// Fill the buffer
	for i := 0; i < 3; i++ {
		event := Event{
			Class:   Class1,
			Group:   2,
			Index:   uint16(i),
			Value:   []byte{byte(i)},
			Quality: BinaryQualityOnline,
			Time:    time.Now(),
		}
		queue.Add(event)
	}

	// Next add should fail
	event := Event{
		Class:   Class1,
		Group:   2,
		Index:   99,
		Value:   []byte{0x01},
		Quality: BinaryQualityOnline,
		Time:    time.Now(),
	}

	if queue.Add(event) {
		t.Error("Expected Add to fail on full buffer")
	}

	// IsFull should return true
	if !queue.IsFull() {
		t.Error("Expected IsFull to return true")
	}
}

// TestEventQueueClear tests clearing events.
func TestEventQueueClear(t *testing.T) {
	queue := NewEventQueue(100)

	// Add some events
	for i := 0; i < 3; i++ {
		event := Event{
			Class:   Class1,
			Group:   2,
			Index:   uint16(i),
			Value:   []byte{byte(i)},
			Quality: BinaryQualityOnline,
			Time:    time.Now(),
		}
		queue.Add(event)
	}

	if queue.Count() != 3 {
		t.Fatalf("Count = %d, want 3", queue.Count())
	}

	// Clear
	queue.Clear()

	if queue.Count() != 0 {
		t.Errorf("Count = %d, want 0 after clear", queue.Count())
	}
}

// TestGenerateEvent tests event generation.
func TestGenerateEvent(t *testing.T) {
	// Use a small buffer size to test overflow behavior
	config := &Config{
		MaxEventBuffers: 15, // Small buffer to test overflow
	}
	ost := NewOutstation(config)
	ost.Initialize()
	ost.Start()

	// Generate an event
	if !ost.GenerateEvent(Class1, 2, 0, []byte{0x01}, BinaryQualityOnline) {
		t.Error("Expected GenerateEvent to succeed")
	}

	if ost.EventCount() != 1 {
		t.Errorf("EventCount = %d, want 1", ost.EventCount())
	}

	// Generate more events to fill buffer (15 total, 5 per class)
	for i := 1; i < 15; i++ {
		ost.GenerateEvent(Class1, 2, uint16(i), []byte{byte(i)}, BinaryQualityOnline)
	}

	// Check IIN is set when buffer is full
	iin := ost.IIN()
	if !iin.ByteOver {
		t.Error("Expected IIN.ByteOver to be set when buffer is full")
	}
}

// TestClearEvents tests clearing events.
func TestClearEvents(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Generate some events
	ost.GenerateEvent(Class1, 2, 0, []byte{0x01}, BinaryQualityOnline)
	ost.GenerateEvent(Class2, 32, 0, []byte{0x02}, AnalogQualityOnline)

	if ost.EventCount() != 2 {
		t.Fatalf("EventCount = %d, want 2", ost.EventCount())
	}

	// Clear events
	ost.ClearEvents()

	if ost.EventCount() != 0 {
		t.Errorf("EventCount = %d, want 0 after clear", ost.EventCount())
	}

	// Check IIN is cleared
	iin := ost.IIN()
	if iin.ByteOver {
		t.Error("Expected IIN.ByteOver to be cleared after ClearEvents")
	}
}

// TestFreezeCounters tests the freeze operation.
func TestFreezeCounters(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Get current counters
	counters := ost.data.GetCounters()
	if len(counters) == 0 {
		t.Fatal("No counters available")
	}

	// Freeze counters
	resp, err := ost.handleFreeze(nil)
	if err != nil {
		t.Fatalf("handleFreeze failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response to freeze")
	}

	// Verify frozen counters have values
	frozen := ost.data.GetFrozenCounters()
	if len(frozen) == 0 {
		t.Fatal("No frozen counters")
	}

	// Values should match original counters
	if frozen[0].Value != counters[0].Value {
		t.Errorf("Frozen counter Value = %d, want %d", frozen[0].Value, counters[0].Value)
	}
}

// TestFreezeClearCounters tests the freeze-and-clear operation.
func TestFreezeClearCounters(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Get current counters
	counters := ost.data.GetCounters()
	if len(counters) == 0 {
		t.Fatal("No counters available")
	}
	originalValue := counters[0].Value

	// Freeze and clear counters
	resp, err := ost.handleFreezeClear(nil)
	if err != nil {
		t.Fatalf("handleFreezeClear failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response to freeze clear")
	}

	// Original counters should be zeroed
	counters = ost.data.GetCounters()
	if counters[0].Value != 0 {
		t.Errorf("Counter Value = %d, want 0 after freeze-clear", counters[0].Value)
	}

	// Frozen counters should have original value
	frozen := ost.data.GetFrozenCounters()
	if frozen[0].Value != originalValue {
		t.Errorf("Frozen counter Value = %d, want %d", frozen[0].Value, originalValue)
	}
}

// TestConfirmation tests confirmation generation.
func TestConfirmation(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create a request with CON bit set
	req := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, CON: true, Seq: 5},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, 0x07, 0x00},
	}

	// Process request - in Run, confirmation would be sent automatically
	// Here we just verify it processes correctly
	resp, err := ost.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response")
	}
}

// TestResponseCache tests the response cache.
func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(5 * time.Second)

	// Add a response
	cache.Add(1, []byte{0x01, 0x02, 0x03})

	// Should be retrievable
	resp, found := cache.Get(1)
	if !found {
		t.Error("Expected to find cached response")
	}
	if len(resp) != 3 {
		t.Errorf("Response length = %d, want 3", len(resp))
	}

	// Non-existent key
	_, found = cache.Get(99)
	if found {
		t.Error("Expected not to find non-existent key")
	}
}

// TestResponseCacheExpiry tests cache expiry.
func TestResponseCacheExpiry(t *testing.T) {
	cache := NewResponseCache(50 * time.Millisecond)

	// Add a response
	cache.Add(1, []byte{0x01})

	// Should be retrievable immediately
	_, found := cache.Get(1)
	if !found {
		t.Error("Expected to find cached response immediately")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = cache.Get(1)
	if found {
		t.Error("Expected response to be expired")
	}
}

// TestCleanup tests the cleanup function.
func TestCleanup(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Add some state
	selectReq := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncSelect,
		Data:     []byte{12, 1, 0x00, 0x01, 0x00, 0x00, 0x01},
	}
	ost.ProcessRequest(selectReq)
	ost.GenerateEvent(Class1, 2, 0, []byte{0x01}, BinaryQualityOnline)

	// Verify state exists
	if ost.PendingSelectCount() != 1 {
		t.Fatalf("PendingSelectCount = %d, want 1", ost.PendingSelectCount())
	}
	if ost.EventCount() != 1 {
		t.Fatalf("EventCount = %d, want 1", ost.EventCount())
	}

	// Call cleanup (private, but testable via Stop)
	ost.cleanup()

	// Verify state is cleared
	if ost.PendingSelectCount() != 0 {
		t.Errorf("PendingSelectCount = %d, want 0 after cleanup", ost.PendingSelectCount())
	}
	if ost.EventCount() != 0 {
		t.Errorf("EventCount = %d, want 0 after cleanup", ost.EventCount())
	}

	iin := ost.IIN()
	if iin.ByteOver {
		t.Error("Expected IIN.ByteOver to be cleared after cleanup")
	}
}
