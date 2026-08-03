// +build ignore

// This is a demonstration of the DNP3 Event Subsystem.
// Run with: go run test/integration/event_demo.go
package main

import (
	"fmt"
	"unsafe"

	"dnp3/internal/outstation/events"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("DNP3 Event Subsystem Demonstration")
	fmt.Println("===========================================")
	fmt.Println()

	// Create event configuration
	config := events.NewEventConfig()
	fmt.Println("Event Configuration:")
	fmt.Printf("  Binary Input Events Enabled: %v\n", config.IsEnabled(events.PointTypeBinaryInput))
	fmt.Printf("  Analog Input Events Enabled: %v\n", config.IsEnabled(events.PointTypeAnalogInput))
	fmt.Printf("  Counter Events Enabled: %v\n", config.IsEnabled(events.PointTypeCounter))
	fmt.Printf("  Binary Output Events Enabled: %v\n", config.IsEnabled(events.PointTypeBinaryOutput))
	fmt.Printf("  Analog Output Events Enabled: %v\n", config.IsEnabled(events.PointTypeAnalogOutput))
	fmt.Println()

	// Create event engine
	engine := events.NewEventEngine(config, 100)
	fmt.Println("Event Engine created with 100 buffer size")
	fmt.Println()

	// Track events for demonstration
	eventCount := 0
	engine.SetEventCallback(func(event events.Event) {
		eventCount++
		fmt.Printf("  [Callback #%d] Event generated: %s Index=%d Time=%v\n",
			eventCount,
			event.PointType,
			event.Index,
			event.Time.Format("15:04:05.000"))
	})

	// Demonstrate binary input events
	fmt.Println("-------------------------------------------")
	fmt.Println("1. Binary Input Event Generation")
	fmt.Println("-------------------------------------------")

	// First value - should NOT generate event (initialization)
	fmt.Println("\n  Step 1: Initialize Binary Input 0 to TRUE")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 0, []byte{1}, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Same value - should NOT generate event
	fmt.Println("\n  Step 2: Set Binary Input 0 to TRUE again (no change)")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 0, []byte{1}, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Different value - SHOULD generate event
	fmt.Println("\n  Step 3: Change Binary Input 0 to FALSE")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 0, []byte{0}, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Demonstrate analog input events
	fmt.Println("\n-------------------------------------------")
	fmt.Println("2. Analog Input Event Generation")
	fmt.Println("-------------------------------------------")

	// First value - should NOT generate event
	fmt.Println("\n  Step 1: Initialize Analog Input 5 to 100.5")
	value1 := float32ToBytes(100.5)
	engine.CheckAndGenerate(events.PointTypeAnalogInput, 5, value1, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Different value - SHOULD generate event
	fmt.Println("\n  Step 2: Change Analog Input 5 to 200.75")
	value2 := float32ToBytes(200.75)
	engine.CheckAndGenerate(events.PointTypeAnalogInput, 5, value2, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Demonstrate counter events
	fmt.Println("\n-------------------------------------------")
	fmt.Println("3. Counter Event Generation")
	fmt.Println("-------------------------------------------")

	// First value - should NOT generate event
	fmt.Println("\n  Step 1: Initialize Counter 0 to 1000")
	engine.CheckAndGenerate(events.PointTypeCounter, 0, uint32ToBytes(1000), 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Different value - SHOULD generate event
	fmt.Println("\n  Step 2: Change Counter 0 to 1500")
	engine.CheckAndGenerate(events.PointTypeCounter, 0, uint32ToBytes(1500), 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Demonstrate disabling events
	fmt.Println("\n-------------------------------------------")
	fmt.Println("4. Event Configuration - Disabling Events")
	fmt.Println("-------------------------------------------")

	// First, let's use a different point to demonstrate
	// Binary Input 1: TRUE -> FALSE (generates event)
	fmt.Println("\n  Step 1: Initialize Binary Input 1 to TRUE")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 1, []byte{1}, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Disable events for Binary Input
	config.SetEnabled(events.PointTypeBinaryInput, false)
	fmt.Println("\n  Disabled Binary Input events")
	fmt.Println("\n  Step 2: Change Binary Input 1 to FALSE (should NOT generate event)")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 1, []byte{0}, 0x01)
	fmt.Printf("  Events in queue: %d (unchanged because events disabled)\n", engine.EventCount())

	// Re-enable
	config.SetEnabled(events.PointTypeBinaryInput, true)
	fmt.Println("\n  Re-enabled Binary Input events")
	fmt.Println("\n  Step 3: Change Binary Input 1 to TRUE (SHOULD generate event)")
	engine.CheckAndGenerate(events.PointTypeBinaryInput, 1, []byte{1}, 0x01)
	fmt.Printf("  Events in queue: %d\n", engine.EventCount())

	// Demonstrate event retrieval
	fmt.Println("\n-------------------------------------------")
	fmt.Println("5. Event Retrieval and Queue Management")
	fmt.Println("-------------------------------------------")

	eventsList := engine.GetEvents()
	fmt.Printf("\n  Total events in queue: %d\n", len(eventsList))
	for i, ev := range eventsList {
		fmt.Printf("  Event %d: Type=%s Index=%d Group=%d Variation=%d Time=%s\n",
			i+1,
			ev.PointType,
			ev.Index,
			ev.Group,
			ev.Variation,
			ev.Time.Format("15:04:05.000"))
	}

	// Pop events
	fmt.Println("\n  Popping 3 events from queue...")
	popped := engine.PopEvents(3)
	fmt.Printf("  Popped %d events\n", len(popped))
	fmt.Printf("  Remaining events: %d\n", engine.EventCount())

	// Demonstrate event builder
	fmt.Println("\n-------------------------------------------")
	fmt.Println("6. Event Object Encoding")
	fmt.Println("-------------------------------------------")

	remaining := engine.GetEvents()
	builder := events.NewEventBuilder()
	encoded := builder.BuildEventObjects(remaining)
	fmt.Printf("\n  Encoded %d events into %d bytes\n", len(remaining), len(encoded))
	if len(encoded) > 0 {
		fmt.Printf("  First 20 bytes (hex): %x...\n", encoded[:20])
	}

	// Clear events
	fmt.Println("\n  Clearing all events...")
	engine.ClearEvents()
	fmt.Printf("  Events after clear: %d\n", engine.EventCount())

	fmt.Println("\n===========================================")
	fmt.Println("Event Subsystem Demonstration Complete")
	fmt.Println("===========================================")
}

// float32ToBytes converts float32 to big-endian bytes
func float32ToBytes(f float32) []byte {
	bits := float32Bits(f)
	return []byte{
		byte(bits >> 24),
		byte(bits >> 16),
		byte(bits >> 8),
		byte(bits),
	}
}

// float32Bits returns the bit representation of a float32
func float32Bits(f float32) uint32 {
	return *(*uint32)(unsafe.Pointer(&f))
}

// uint32ToBytes converts uint32 to big-endian bytes
func uint32ToBytes(v uint32) []byte {
	return []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}
