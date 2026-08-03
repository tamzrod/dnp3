// Package events implements the DNP3 Event Subsystem.
//
// The Event Subsystem is responsible for:
// - Detecting point changes
// - Generating events with timestamps
// - Managing event queues per class
// - Serving events to the protocol layer
// - Removing confirmed events
//
// Architecture:
//
//	Point Store (current state)
//	    ↓ change detected
//	Event Engine (change detection & event creation)
//	    ↓
//	Event Queue (per-class FIFO)
//	    ↓
//	Protocol Layer (event objects in response)
//	    ↓
//	Master confirms events
//	    ↓
//	Event Queue (remove confirmed)
//
// Reference: IEEE 1815-2012 Section 3.2 (Event Buffer)
package events

import (
	"sync"
	"time"
)

// EventClass represents the priority class of an event.
type EventClass uint8

const (
	// Class0 is for static data (no events in this class)
	Class0 EventClass = 0
	// Class1 is the highest priority event class
	Class1 EventClass = 1
	// Class2 is the medium priority event class
	Class2 EventClass = 2
	// Class3 is the lowest priority event class
	Class3 EventClass = 3
)

// String returns the class name.
func (c EventClass) String() string {
	switch c {
	case Class0:
		return "Class 0"
	case Class1:
		return "Class 1"
	case Class2:
		return "Class 2"
	case Class3:
		return "Class 3"
	default:
		return "Unknown"
	}
}

// Event represents a DNP3 event.
//
// Events are generated when a point value changes. Each event contains:
// - Point type and index
// - New value at time of change
// - Quality flags at time of change
// - Timestamp of the change
// - Event class (1, 2, or 3)
type Event struct {
	// PointType identifies the type of point that generated this event
	PointType PointType
	// Index is the point index within its type
	Index uint16
	// Value contains the encoded value at time of change
	Value []byte
	// Quality contains quality flags at time of change
	Quality uint8
	// Time is when the event was generated
	Time time.Time
	// Class is the event priority class
	Class EventClass
	// Variation is the DNP3 object variation for this event
	Variation uint8
	// Group is the DNP3 object group for this event
	Group uint8
}

// PointType identifies the type of data point.
type PointType uint8

const (
	PointTypeBinaryInput   PointType = 1
	PointTypeDoubleBinary  PointType = 2 // Not implemented
	PointTypeBinaryOutput  PointType = 10
	PointTypeCounter       PointType = 20
	PointTypeFrozenCounter PointType = 21 // Not implemented
	PointTypeAnalogInput   PointType = 30
	PointTypeAnalogOutput  PointType = 40
)

// String returns the point type name.
func (pt PointType) String() string {
	switch pt {
	case PointTypeBinaryInput:
		return "Binary Input"
	case PointTypeDoubleBinary:
		return "Double Binary"
	case PointTypeBinaryOutput:
		return "Binary Output"
	case PointTypeCounter:
		return "Counter"
	case PointTypeFrozenCounter:
		return "Frozen Counter"
	case PointTypeAnalogInput:
		return "Analog Input"
	case PointTypeAnalogOutput:
		return "Analog Output"
	default:
		return "Unknown"
	}
}

// EventConfig enables or disables event generation per point type.
type EventConfig struct {
	mu       sync.RWMutex
	enabled  map[PointType]bool
	classes  map[PointType]EventClass
}

// NewEventConfig creates a new event configuration with defaults.
func NewEventConfig() *EventConfig {
	return &EventConfig{
		enabled: map[PointType]bool{
			PointTypeBinaryInput:  true,  // Events enabled by default
			PointTypeDoubleBinary: false, // Not implemented
			PointTypeBinaryOutput: true,  // Events enabled by default
			PointTypeCounter:      true,  // Events enabled by default
			PointTypeFrozenCounter: false, // Not implemented
			PointTypeAnalogInput:  true,  // Events enabled by default
			PointTypeAnalogOutput: true,  // Events enabled by default
		},
		classes: map[PointType]EventClass{
			PointTypeBinaryInput:   Class1, // Default to Class 1
			PointTypeDoubleBinary:  Class1,
			PointTypeBinaryOutput:  Class1,
			PointTypeCounter:       Class1,
			PointTypeFrozenCounter: Class1,
			PointTypeAnalogInput:   Class1,
			PointTypeAnalogOutput:  Class1,
		},
	}
}

// SetEnabled enables or disables event generation for a point type.
func (c *EventConfig) SetEnabled(pointType PointType, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled[pointType] = enabled
}

// IsEnabled returns true if event generation is enabled for the point type.
func (c *EventConfig) IsEnabled(pointType PointType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled[pointType]
}

// SetClass sets the event class for a point type.
func (c *EventConfig) SetClass(pointType PointType, class EventClass) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.classes[pointType] = class
}

// GetClass returns the event class for a point type.
func (c *EventConfig) GetClass(pointType PointType) EventClass {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.classes[pointType]
}
