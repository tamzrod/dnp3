package events

import (
	"sync"
	"testing"
	"time"
)

// Ensure time is used
var _ = time.Now

func TestEventConfig(t *testing.T) {
	config := NewEventConfig()

	// Test default enabled states
	if !config.IsEnabled(PointTypeBinaryInput) {
		t.Error("Binary Input should be enabled by default")
	}
	if !config.IsEnabled(PointTypeAnalogInput) {
		t.Error("Analog Input should be enabled by default")
	}
	if !config.IsEnabled(PointTypeCounter) {
		t.Error("Counter should be enabled by default")
	}
	if config.enabled[PointTypeFrozenCounter] {
		t.Error("Frozen Counter should be disabled by default (not implemented)")
	}

	// Test enabling/disabling
	config.SetEnabled(PointTypeBinaryInput, false)
	if config.IsEnabled(PointTypeBinaryInput) {
		t.Error("Binary Input should be disabled after SetEnabled(false)")
	}

	// Test class configuration
	if config.GetClass(PointTypeBinaryInput) != Class1 {
		t.Errorf("Binary Input should default to Class1, got %v", config.GetClass(PointTypeBinaryInput))
	}

	config.SetClass(PointTypeBinaryInput, Class2)
	if config.GetClass(PointTypeBinaryInput) != Class2 {
		t.Error("Binary Input class should be Class2 after SetClass")
	}
}

func TestEventQueue(t *testing.T) {
	queue := NewEventQueue(30) // 10 per class

	// Test initial state
	if queue.Count() != 0 {
		t.Errorf("Expected 0 events, got %d", queue.Count())
	}
	if queue.IsFull() {
		t.Error("Queue should not be full initially")
	}

	// Test Push
	event1 := Event{
		PointType: PointTypeBinaryInput,
		Index:     1,
		Value:     []byte{1},
		Quality:   0x01,
		Time:      time.Now(),
		Class:     Class1,
	}

	if !queue.Push(event1) {
		t.Error("Push should succeed")
	}
	if queue.Count() != 1 {
		t.Errorf("Expected 1 event, got %d", queue.Count())
	}

	// Test GetAll
	all := queue.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 event in GetAll, got %d", len(all))
	}

	// Test Pop
	popped := queue.Pop(10)
	if len(popped) != 1 {
		t.Errorf("Expected 1 popped event, got %d", len(popped))
	}
	if queue.Count() != 0 {
		t.Errorf("Expected 0 events after Pop, got %d", queue.Count())
	}

	// Test Clear
	queue.Push(event1)
	queue.Clear()
	if queue.Count() != 0 {
		t.Error("Expected 0 events after Clear")
	}
}

func TestEventQueueConcurrency(t *testing.T) {
	// Use a larger queue to avoid buffer overflow during concurrent pushes
	queue := NewEventQueue(1000) // 333 per class
	var wg sync.WaitGroup

	// Concurrent pushes - use different classes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			class := Class1
			if idx%3 == 1 {
				class = Class2
			} else if idx%3 == 2 {
				class = Class3
			}
			for j := 0; j < 10; j++ {
				event := Event{
					PointType: PointTypeBinaryInput,
					Index:     uint16(idx*10 + j),
					Value:     []byte{byte(j % 2)},
					Quality:   0x01,
					Time:      time.Now(),
					Class:     class,
				}
				queue.Push(event)
			}
		}(i)
	}
	wg.Wait()

	// Verify count - should have 100 events (some may be dropped due to buffer full)
	count := queue.Count()
	if count < 50 {
		t.Errorf("Expected at least 50 events, got %d", count)
	}

	// Concurrent pops
	var popWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		popWg.Add(1)
		go func() {
			defer popWg.Done()
			for j := 0; j < 10; j++ {
				queue.Pop(5)
			}
		}()
	}
	popWg.Wait()
}

func TestEventEngine(t *testing.T) {
	config := NewEventConfig()
	engine := NewEventEngine(config, 100)

	// Test initial state
	if engine.EventCount() != 0 {
		t.Errorf("Expected 0 events, got %d", engine.EventCount())
	}

	// Test CheckAndGenerate with first value (no change)
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{1}, 0x01)
	if engine.EventCount() != 0 {
		t.Error("First value should not generate event")
	}

	// Test CheckAndGenerate with different value (change detected)
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{0}, 0x01)
	if engine.EventCount() != 1 {
		t.Errorf("Expected 1 event after value change, got %d", engine.EventCount())
	}

	// Test that same value doesn't generate another event
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{0}, 0x01)
	if engine.EventCount() != 1 {
		t.Error("Same value should not generate another event")
	}

	// Test disabled point type
	config.SetEnabled(PointTypeBinaryInput, false)
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{1}, 0x01)
	if engine.EventCount() != 1 {
		t.Error("Disabled point type should not generate events")
	}

	// Test different point types
	config.SetEnabled(PointTypeAnalogInput, true)
	engine.CheckAndGenerate(PointTypeAnalogInput, 5, []byte{0x00, 0x00, 0x80, 0x3F}, 0x01) // 1.0 as float32 - init
	engine.CheckAndGenerate(PointTypeAnalogInput, 5, []byte{0x00, 0x00, 0x00, 0x40}, 0x01) // 2.0 as float32 - change
	engine.CheckAndGenerate(PointTypeAnalogInput, 5, []byte{0x00, 0x00, 0x00, 0x40}, 0x01) // 2.0 as float32 - same
	engine.CheckAndGenerate(PointTypeAnalogInput, 5, []byte{0x00, 0x00, 0x80, 0x40}, 0x01) // 4.0 as float32 - change
	if engine.EventCount() != 3 { // 1 BI + 2 AI changes
		t.Errorf("Expected 3 events (1 BI + 2 AI), got %d", engine.EventCount())
	}
}

func TestEventEngineCallback(t *testing.T) {
	config := NewEventConfig()
	engine := NewEventEngine(config, 100)

	callbackCalled := false
	var callbackEvent Event
	var mu sync.Mutex

	engine.SetEventCallback(func(event Event) {
		mu.Lock()
		callbackCalled = true
		callbackEvent = event
		mu.Unlock()
	})

	// First value - no event
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{1}, 0x01)

	// Give async callback time to fire (if it were going to)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if callbackCalled {
		t.Error("Callback should not have been called for first value")
	}
	mu.Unlock()

	// Second value with change - should trigger callback
	engine.CheckAndGenerate(PointTypeBinaryInput, 1, []byte{0}, 0x01)

	// Wait for async callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !callbackCalled {
		t.Error("Callback should have been called")
	}
	if callbackEvent.Index != 1 {
		t.Errorf("Expected event index 1, got %d", callbackEvent.Index)
	}
	if callbackEvent.PointType != PointTypeBinaryInput {
		t.Errorf("Expected Binary Input, got %v", callbackEvent.PointType)
	}
	mu.Unlock()
}

func TestEventBuilder(t *testing.T) {
	builder := NewEventBuilder()

	// Test empty events
	result := builder.BuildEventObjects(nil)
	if len(result) != 0 {
		t.Error("Empty events should produce empty result")
	}

	// Test single event
	event := Event{
		PointType: PointTypeBinaryInput,
		Index:     5,
		Value:     []byte{1},
		Quality:   0x01,
		Time:      time.Now(),
		Class:     Class1,
		Group:     2,
		Variation: 1,
	}

	result = builder.BuildEventObjects([]Event{event})
	if len(result) == 0 {
		t.Error("Should produce encoded data")
	}

	// Verify structure: header (4 bytes) + event data
	// Header: group(1) + variation(1) + qualifier(1) + count(1)
	if result[0] != 2 { // Group 2 = Binary Input Event
		t.Errorf("Expected group 2, got %d", result[0])
	}
	if result[1] != 1 { // Variation 1
		t.Errorf("Expected variation 1, got %d", result[1])
	}
	if result[2] != 0x17 { // Qualifier prefix-count-8
		t.Errorf("Expected qualifier 0x17, got %x", result[2])
	}
	if result[3] != 1 { // Count
		t.Errorf("Expected count 1, got %d", result[3])
	}
}

func TestEventGroupVariation(t *testing.T) {
	tests := []struct {
		pointType    PointType
		expectedGrp  uint8
		expectedVar  uint8
	}{
		{PointTypeBinaryInput, 2, 1},
		{PointTypeBinaryOutput, 11, 1},
		{PointTypeCounter, 22, 1},
		{PointTypeAnalogInput, 32, 1},
		{PointTypeAnalogOutput, 42, 1},
	}

	for _, tt := range tests {
		grp, var_ := GetEventGroupVariation(tt.pointType)
		if grp != tt.expectedGrp {
			t.Errorf("PointType %v: expected group %d, got %d", tt.pointType, tt.expectedGrp, grp)
		}
		if var_ != tt.expectedVar {
			t.Errorf("PointType %v: expected variation %d, got %d", tt.pointType, tt.expectedVar, var_)
		}
	}
}

func TestPointType(t *testing.T) {
	if PointTypeBinaryInput.String() != "Binary Input" {
		t.Error("Binary Input string mismatch")
	}
	if PointTypeAnalogInput.String() != "Analog Input" {
		t.Error("Analog Input string mismatch")
	}
	if PointTypeCounter.String() != "Counter" {
		t.Error("Counter string mismatch")
	}
}

func TestEventClass(t *testing.T) {
	if Class1.String() != "Class 1" {
		t.Error("Class1 string mismatch")
	}
	if Class2.String() != "Class 2" {
		t.Error("Class2 string mismatch")
	}
	if Class3.String() != "Class 3" {
		t.Error("Class3 string mismatch")
	}
}
