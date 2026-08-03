// Package events implements the DNP3 Event Subsystem.
// See event.go for package documentation.
package events

import (
	"sync"
	"time"
)

// PointValue represents a point's current value for comparison.
type PointValue struct {
	Value   []byte
	Quality uint8
}

// EventEngine manages event generation based on point changes.
type EventEngine struct {
	mu      sync.RWMutex
	config  *EventConfig
	queue   *EventQueue

	// Previous values for change detection
	// Key: "pointType-index"
	prevValues map[string]PointValue

	// Change callback - called when an event is generated
	onEventGenerated func(event Event)
}

// NewEventEngine creates a new EventEngine.
func NewEventEngine(config *EventConfig, maxBufferSize int) *EventEngine {
	if config == nil {
		config = NewEventConfig()
	}
	if maxBufferSize <= 0 {
		maxBufferSize = 1000
	}

	return &EventEngine{
		config:      config,
		queue:       NewEventQueue(maxBufferSize),
		prevValues:  make(map[string]PointValue),
	}
}

// SetEventCallback sets a callback to be called when an event is generated.
// This is useful for triggering unsolicited responses or logging.
func (e *EventEngine) SetEventCallback(callback func(event Event)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onEventGenerated = callback
}

// pointKey generates a unique key for a point.
func pointKey(pointType PointType, index uint16) string {
	return string(rune(pointType)) + "-" + string(rune(index>>8)) + string(rune(index&0xFF))
}

// CheckAndGenerate checks if a point value has changed and generates an event if needed.
func (e *EventEngine) CheckAndGenerate(pointType PointType, index uint16, newValue []byte, quality uint8) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if events are enabled for this point type
	if !e.config.enabled[pointType] {
		return false
	}

	key := pointKey(pointType, index)

	// Get previous value
	prev, exists := e.prevValues[key]

	// Check if value has changed
	valueChanged := false
	if !exists {
		// First time seeing this point - initialize but don't generate event
		e.prevValues[key] = PointValue{
			Value:   copyBytes(newValue),
			Quality: quality,
		}
		return false // No event for first value
	} else {
		// Compare values
		if len(prev.Value) != len(newValue) {
			valueChanged = true
		} else {
			for i := range prev.Value {
				if prev.Value[i] != newValue[i] {
					valueChanged = true
					break
				}
			}
		}
		// Also generate event if quality changed
		valueChanged = valueChanged || (prev.Quality != quality)
	}

	if !valueChanged {
		return false
	}

	// Generate event
	event := Event{
		PointType: pointType,
		Index:     index,
		Value:     copyBytes(newValue),
		Quality:   quality,
		Time:      time.Now(),
		Class:     e.config.classes[pointType],
	}

	// Set group and variation based on point type
	event.Group, event.Variation = GetEventGroupVariation(pointType)

	// Add to queue
	e.queue.Push(event)

	// Update previous value
	e.prevValues[key] = PointValue{
		Value:   copyBytes(newValue),
		Quality: quality,
	}

	// Call callback if set (outside lock to avoid deadlock)
	if e.onEventGenerated != nil {
		go func(ev Event) {
			e.mu.RLock()
			cb := e.onEventGenerated
			e.mu.RUnlock()
			if cb != nil {
				cb(ev)
			}
		}(event)
	}

	return true
}

// GetEventGroupVariation returns the DNP3 group and variation for event objects.
func GetEventGroupVariation(pointType PointType) (group, variation uint8) {
	// Default to variation 1 (with time)
	switch pointType {
	case PointTypeBinaryInput:
		return 2, 1 // Binary Input Event with Time
	case PointTypeDoubleBinary:
		return 3, 1 // Double-bit Binary Input Event with Time
	case PointTypeBinaryOutput:
		return 11, 1 // Binary Output Event with Time
	case PointTypeCounter:
		return 22, 1 // Counter Event with Time
	case PointTypeFrozenCounter:
		return 23, 1 // Frozen Counter Event with Time
	case PointTypeAnalogInput:
		return 32, 1 // Analog Input Event with Time
	case PointTypeAnalogOutput:
		return 42, 1 // Analog Output Event with Time
	default:
		return 0, 0
	}
}

// copyBytes creates a copy of a byte slice.
func copyBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// EventCount returns the total number of events in the queue.
func (e *EventEngine) EventCount() int {
	return e.queue.Count()
}

// HasEvents returns true if there are events in the queue.
func (e *EventEngine) HasEvents() bool {
	return e.queue.Count() > 0
}

// GetEvents returns all events, optionally filtered by class.
func (e *EventEngine) GetEvents(classes ...EventClass) []Event {
	if len(classes) == 0 {
		return e.queue.GetAll()
	}
	return e.queue.GetByClasses(classes...)
}

// PopEvents removes and returns events from the queue.
func (e *EventEngine) PopEvents(maxCount int, classes ...EventClass) []Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.queue.Pop(maxCount, classes...)
}

// ClearEvents removes all events from the queue.
func (e *EventEngine) ClearEvents() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.queue.Clear()
	e.prevValues = make(map[string]PointValue)
}

// ClearClass removes all events of a specific class.
func (e *EventEngine) ClearClass(class EventClass) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.queue.ClearClass(class)
}

// Config returns the event configuration.
func (e *EventEngine) Config() *EventConfig {
	return e.config
}
