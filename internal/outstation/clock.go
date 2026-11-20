package outstation

import (
	"sync"
	"time"
)

// ClockSource identifies where a logical DNP3 clock last got its absolute
// time. It does not describe the host operating-system clock, which is never
// read for control or modified by this package.
type ClockSource uint8

const (
	// ClockUnset means no absolute time has been supplied yet. Logical Now()
	// falls back to the host clock (offset 0) for elapsed-time and SOE
	// timestamping, but IIN NEED_TIME stays asserted.
	ClockUnset ClockSource = iota
	// ClockSystem means the clock was seeded from the host clock as an
	// explicit fallback, not from an authoritative master. NEED_TIME is not
	// cleared by this source.
	ClockSystem
	// ClockMaster means the time was set by a master time-sync object (G50V1).
	// This is the only source that clears IIN NEED_TIME.
	ClockMaster
)

func (s ClockSource) String() string {
	switch s {
	case ClockSystem:
		return "SYSTEM"
	case ClockMaster:
		return "MASTER"
	default:
		return "UNSET"
	}
}

// ClockStatus is a diagnostic snapshot of the logical clock. It is copied out
// under lock so callers can inspect it without holding the clock's mutex.
type ClockStatus struct {
	Source       ClockSource
	SourceName   string
	HasTime      bool
	NeedTime     bool
	Expired      bool
	Validity     time.Duration
	Offset       time.Duration
	LastSync     time.Time // host wall time of the last authoritative sync
	LastValue    time.Time // the logical time value applied at that sync
	LastRecorded time.Time // last valid RECORD_CURRENT_TIME value (diagnostic only)
	LastDelayMS  uint16    // last G52 processing delay reported to a master
	LogicalNow   time.Time
	SystemNow    time.Time
	SyncCount    uint64
}

// LogicalClock is a DNP3-only clock. It is deliberately separate from the
// Linux host clock: no method here calls settimeofday or touches NTP. The clock
// advances as host time + an offset learned from the last master time sync, so
// its elapsed-time arithmetic uses Go's monotonic time source and cannot jump
// backwards when a master applies a large offset.
//
// A logical clock may be shared by all connection instances of an outstation
// server so that time sync survives a master reconnect.
type LogicalClock struct {
	mu sync.RWMutex

	offset       time.Duration // logical = host + offset
	source       ClockSource
	needTime     bool
	validity     time.Duration // 0 means "never expires"
	lastSync     time.Time
	lastValue    time.Time
	lastRecorded time.Time
	lastDelayMS  uint16
	syncCount    uint64
}

// NewLogicalClock returns a logical clock. needTime seeds IIN NEED_TIME (the
// normal device-startup condition). validity > 0 makes a master-synced time
// expire after that duration, re-asserting NEED_TIME until the master syncs
// again; validity == 0 disables expiry.
func NewLogicalClock(validity time.Duration, needTime bool) *LogicalClock {
	if validity < 0 {
		validity = 0
	}
	return &LogicalClock{
		source:   ClockUnset,
		needTime: needTime,
		validity: validity,
	}
}

// Now returns the current logical time. The returned time.Time keeps Go's
// monotonic reading from time.Now(), so elapsed-time measurements remain
// monotonic even though the wall-clock component carries a DNP3 offset.
func (c *LogicalClock) Now() time.Time {
	c.mu.RLock()
	offset := c.offset
	c.mu.RUnlock()
	return time.Now().Add(offset)
}

// Set applies an absolute DNP3 time. Only ClockMaster clears IIN NEED_TIME;
// ClockSystem is an explicit fallback seed and leaves NEED_TIME as it was.
func (c *LogicalClock) Set(t time.Time, source ClockSource) {
	if t.IsZero() {
		return
	}
	hostNow := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offset = t.Sub(hostNow)
	// A system fallback seed must not downgrade an authoritative master sync.
	if source == ClockMaster || c.source == ClockUnset {
		c.source = source
	}
	c.lastValue = t
	if source == ClockMaster {
		c.lastSync = hostNow
		c.needTime = false
		c.syncCount++
	}
}

// MarkNeedTime sets or clears the IIN NEED_TIME condition explicitly. It does
// not change the clock's offset or last-sync time.
func (c *LogicalClock) MarkNeedTime(need bool) {
	c.mu.Lock()
	c.needTime = need
	c.mu.Unlock()
}

// NeedsTime reports whether IIN NEED_TIME must be asserted. A configured,
// elapsed validity window re-asserts the bit even after a successful sync.
func (c *LogicalClock) NeedsTime() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.needTime || c.expiredLocked()
}

// Expired reports whether a master-synced time has passed its validity window.
// It never re-asserts NEED_TIME by itself.
func (c *LogicalClock) Expired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.expiredLocked()
}

func (c *LogicalClock) expiredLocked() bool {
	if c.validity <= 0 || c.source != ClockMaster || c.lastSync.IsZero() {
		return false
	}
	return time.Since(c.lastSync) > c.validity
}

// RecordTime stores a valid RECORD_CURRENT_TIME value for diagnostics. It
// deliberately does not change the logical clock or any buffered SOE time:
// reinterpreting an already-captured detection timestamp would break the
// physical-edge vs poll-observation distinction.
func (c *LogicalClock) RecordTime(t time.Time) {
	if t.IsZero() {
		return
	}
	c.mu.Lock()
	c.lastRecorded = t
	c.mu.Unlock()
}

// SetLastDelayMS records the processing delay reported in the last G52 delay
// measurement response for diagnostics.
func (c *LogicalClock) SetLastDelayMS(ms uint16) {
	c.mu.Lock()
	c.lastDelayMS = ms
	c.mu.Unlock()
}

// Status returns a consistent diagnostic snapshot.
func (c *LogicalClock) Status() ClockStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	hostNow := time.Now()
	return ClockStatus{
		Source:       c.source,
		SourceName:   c.source.String(),
		HasTime:      c.source != ClockUnset,
		NeedTime:     c.needTime || c.expiredLocked(),
		Expired:      c.expiredLocked(),
		Validity:     c.validity,
		Offset:       c.offset,
		LastSync:     c.lastSync,
		LastValue:    c.lastValue,
		LastRecorded: c.lastRecorded,
		LastDelayMS:  c.lastDelayMS,
		LogicalNow:   hostNow.Add(c.offset),
		SystemNow:    hostNow,
		SyncCount:    c.syncCount,
	}
}
