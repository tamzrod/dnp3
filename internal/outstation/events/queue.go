// Package events implements the DNP3 Event Subsystem.
// See event.go for package documentation.
package events

import (
	"sync"
)

// EventQueue manages event buffers per class.
// It is a FIFO queue that supports:
// - Adding events
// - Retrieving events (with optional class filter)
// - Removing confirmed events
// - Thread-safe operations
type EventQueue struct {
	mu       sync.Mutex
	buffers  map[EventClass][]Event
	maxSize  int
}

// NewEventQueue creates a new event queue with the specified maximum size.
func NewEventQueue(maxSize int) *EventQueue {
	if maxSize <= 0 {
		maxSize = 1000
	}
	// Divide max size by 3 for each class
	classSize := maxSize / 3
	if classSize < 10 {
		classSize = 10
	}

	return &EventQueue{
		buffers: map[EventClass][]Event{
			Class1: make([]Event, 0, classSize),
			Class2: make([]Event, 0, classSize),
			Class3: make([]Event, 0, classSize),
		},
		maxSize: maxSize,
	}
}

// Push adds an event to the queue.
// Returns true if added, false if buffer is full for that class.
func (q *EventQueue) Push(event Event) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	classEvents := q.buffers[event.Class]
	classMax := q.maxSize / 3

	if len(classEvents) >= classMax {
		return false // Buffer full for this class
	}

	q.buffers[event.Class] = append(classEvents, event)
	return true
}

// GetAll returns all events from all classes.
func (q *EventQueue) GetAll() []Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	var all []Event
	for _, events := range q.buffers {
		all = append(all, events...)
	}
	return all
}

// GetByClasses returns events filtered by the specified classes.
func (q *EventQueue) GetByClasses(classes ...EventClass) []Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	var result []Event
	for _, class := range classes {
		if events, ok := q.buffers[class]; ok {
			result = append(result, events...)
		}
	}
	return result
}

// Pop removes and returns events from the queue.
// maxCount limits how many events to return total.
// classes filters by event class (if empty, all classes are included).
func (q *EventQueue) Pop(maxCount int, classes ...EventClass) []Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if maxCount <= 0 {
		maxCount = q.maxSize
	}

	var result []Event

	// Determine which classes to process
	var classList []EventClass
	if len(classes) == 0 {
		classList = []EventClass{Class1, Class2, Class3}
	} else {
		classList = classes
	}

	// Pop events in priority order (Class1 first)
	for _, class := range classList {
		if len(result) >= maxCount {
			break
		}

		events := q.buffers[class]
		if len(events) == 0 {
			continue
		}

		// Calculate how many we can take from this class
		remaining := maxCount - len(result)
		if len(events) <= remaining {
			// Take all from this class
			result = append(result, events...)
			q.buffers[class] = q.buffers[class][:0]
		} else {
			// Take only what we need
			result = append(result, events[:remaining]...)
			q.buffers[class] = events[remaining:]
		}
	}

	return result
}

// Count returns the total number of events.
func (q *EventQueue) Count() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, events := range q.buffers {
		count += len(events)
	}
	return count
}

// CountByClass returns the number of events in a specific class.
func (q *EventQueue) CountByClass(class EventClass) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	if events, ok := q.buffers[class]; ok {
		return len(events)
	}
	return 0
}

// IsFull returns true if any event class buffer is full.
func (q *EventQueue) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	classMax := q.maxSize / 3
	for _, events := range q.buffers {
		if len(events) >= classMax {
			return true
		}
	}
	return false
}

// IsFullClass returns true if a specific class buffer is full.
func (q *EventQueue) IsFullClass(class EventClass) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	classMax := q.maxSize / 3
	if events, ok := q.buffers[class]; ok {
		return len(events) >= classMax
	}
	return false
}

// Clear removes all events from all classes.
func (q *EventQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	classSize := q.maxSize / 3
	if classSize < 10 {
		classSize = 10
	}

	for class := range q.buffers {
		q.buffers[class] = make([]Event, 0, classSize)
	}
}

// ClearClass removes all events of a specific class.
func (q *EventQueue) ClearClass(class EventClass) {
	q.mu.Lock()
	defer q.mu.Unlock()

	classSize := q.maxSize / 3
	if classSize < 10 {
		classSize = 10
	}
	q.buffers[class] = make([]Event, 0, classSize)
}

// Remove removes a specific event by index within a class.
// Returns true if the event was found and removed.
func (q *EventQueue) Remove(class EventClass, index int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	events := q.buffers[class]
	if index < 0 || index >= len(events) {
		return false
	}

	// Remove element by shifting
	q.buffers[class] = append(events[:index], events[index+1:]...)
	return true
}

// MaxSize returns the configured maximum buffer size.
func (q *EventQueue) MaxSize() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.maxSize
}
