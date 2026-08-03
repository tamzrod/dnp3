// Package events implements the DNP3 Event Subsystem.
// See event.go for package documentation.
package events

import (
	"sync"
)

// DataHandler wraps an existing DataHandler and generates events on point changes.
type DataHandler struct {
	mu           sync.RWMutex
	inner        interface {
		GetBinaryInputs() []interface{ GetValue() (bool, uint8); GetIndex() uint16 }
	}
	eventEngine *EventEngine

	// Current values for comparison
	binaryInputs   map[uint16]PointValue
	analogInputs   map[uint16]PointValue
	counters       map[uint16]PointValue
	binaryOutputs  map[uint16]PointValue
	analogOutputs  map[uint16]PointValue
}

// NewDataHandler creates a new event-generating data handler.
func NewDataHandler(eventEngine *EventEngine) *DataHandler {
	return &DataHandler{
		eventEngine:    eventEngine,
		binaryInputs:  make(map[uint16]PointValue),
		analogInputs:  make(map[uint16]PointValue),
		counters:     make(map[uint16]PointValue),
		binaryOutputs: make(map[uint16]PointValue),
		analogOutputs: make(map[uint16]PointValue),
	}
}

// UpdateBinaryInput updates a binary input and generates an event if changed.
func (h *DataHandler) UpdateBinaryInput(index uint16, value bool, quality uint8) {
	h.mu.Lock()
	defer h.mu.Unlock()

	newValue := []byte{0}
	if value {
		newValue[0] = 1
	}

	// Check if value changed
	prev, exists := h.binaryInputs[index]
	if exists && bytesEqual(prev.Value, newValue) && prev.Quality == quality {
		return // No change
	}

	// Update stored value
	h.binaryInputs[index] = PointValue{Value: newValue, Quality: quality}

	// Generate event if engine is available
	if h.eventEngine != nil {
		h.eventEngine.CheckAndGenerate(PointTypeBinaryInput, index, newValue, quality)
	}
}

// UpdateAnalogInput updates an analog input and generates an event if changed.
func (h *DataHandler) UpdateAnalogInput(index uint16, value []byte, quality uint8) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if value changed
	prev, exists := h.analogInputs[index]
	if exists && bytesEqual(prev.Value, value) && prev.Quality == quality {
		return // No change
	}

	// Update stored value
	h.analogInputs[index] = PointValue{Value: value, Quality: quality}

	// Generate event if engine is available
	if h.eventEngine != nil {
		h.eventEngine.CheckAndGenerate(PointTypeAnalogInput, index, value, quality)
	}
}

// UpdateCounter updates a counter and generates an event if changed.
func (h *DataHandler) UpdateCounter(index uint16, value uint32, quality uint8) {
	h.mu.Lock()
	defer h.mu.Unlock()

	newValue := []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}

	// Check if value changed
	prev, exists := h.counters[index]
	if exists && bytesEqual(prev.Value, newValue) && prev.Quality == quality {
		return // No change
	}

	// Update stored value
	h.counters[index] = PointValue{Value: newValue, Quality: quality}

	// Generate event if engine is available
	if h.eventEngine != nil {
		h.eventEngine.CheckAndGenerate(PointTypeCounter, index, newValue, quality)
	}
}

// UpdateBinaryOutput updates a binary output and generates an event if changed.
func (h *DataHandler) UpdateBinaryOutput(index uint16, value bool, quality uint8) {
	h.mu.Lock()
	defer h.mu.Unlock()

	newValue := []byte{0}
	if value {
		newValue[0] = 1
	}

	// Check if value changed
	prev, exists := h.binaryOutputs[index]
	if exists && bytesEqual(prev.Value, newValue) && prev.Quality == quality {
		return // No change
	}

	// Update stored value
	h.binaryOutputs[index] = PointValue{Value: newValue, Quality: quality}

	// Generate event if engine is available
	if h.eventEngine != nil {
		h.eventEngine.CheckAndGenerate(PointTypeBinaryOutput, index, newValue, quality)
	}
}

// UpdateAnalogOutput updates an analog output and generates an event if changed.
func (h *DataHandler) UpdateAnalogOutput(index uint16, value []byte, quality uint8) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if value changed
	prev, exists := h.analogOutputs[index]
	if exists && bytesEqual(prev.Value, value) && prev.Quality == quality {
		return // No change
	}

	// Update stored value
	h.analogOutputs[index] = PointValue{Value: value, Quality: quality}

	// Generate event if engine is available
	if h.eventEngine != nil {
		h.eventEngine.CheckAndGenerate(PointTypeAnalogOutput, index, value, quality)
	}
}

// bytesEqual compares two byte slices for equality.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
