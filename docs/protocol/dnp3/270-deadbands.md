---
title: "270 - Deadbands"
owner: deadbands
---

# What are DNP3 Deadbands?

## Purpose

DNP3 Deadbands define the **threshold** for analog event generation. Events are only generated when analog values change by more than the configured deadband.

## Problem Being Solved

Every analog change generating an event would overwhelm systems:

1. **Noisy signals** - Small fluctuations would create many events
2. **Bandwidth waste** - Unnecessary event traffic
3. **Processing load** - Master overwhelmed with data

Deadbands solve this by filtering minor fluctuations.

## Deadband Concept

### How Deadbands Work

```
                    ┌────── Deadband ──────┐
                    │                      │
                    ▼                      ▼
Value ────────────► ─────────► ──────────►
                   │        ││
                   │        ││
                   ▼        ▼▼
                 Event    No Event
```

A deadband creates a **hysteresis zone** where changes don't generate events.

### Deadband Types

| Type | Description |
|------|-------------|
| Absolute | Fixed value threshold |
| Percent | Percentage of value |
| Scaled | Based on data range |

## Deadband Definition

### Group 61: Analog Input Deadband

Deadbands are configured per point:

```
┌──────────────────────────────────────────────────────────────────┐
│              ANALOG INPUT DEADBAND (Group 61)                     │
├──────────────────────────────────────────────────────────────────┤
│  Point Index │ Deadband Value │ (Variation-specific)             │
└──────────────────────────────────────────────────────────────────┘
```

### Deadband Encoding

| Variation | Encoding | Size |
|-----------|----------|------|
| 1 | 16-bit integer | 2 bytes |
| 2 | 32-bit integer | 4 bytes |
| 3 | 32-bit float | 4 bytes |

## Event Triggering

### Trigger Condition

An event is generated when:

```
|new_value - last_event_value| > deadband
```

### Example

```
Current value: 100.0
Last event value: 100.0
Deadband: 5.0

New value: 102.0 → |102 - 100| = 2 < 5 → No event
New value: 106.0 → |106 - 100| = 6 > 5 → Event generated
```

## Deadband Behavior

### Positive Change

```
Last event: 100.0, Deadband: 5.0

New value: 106.0 → Event (106.0 - 100.0 = 6 > 5)
New last event value: 106.0

New value: 107.0 → No event (107.0 - 106.0 = 1 < 5)
New value: 111.0 → Event (111.0 - 106.0 = 5, equals deadband)
```

### Negative Change

```
Last event: 100.0, Deadband: 5.0

New value: 94.0 → Event (100.0 - 94.0 = 6 > 5)
New last event value: 94.0

New value: 93.0 → No event (94.0 - 93.0 = 1 < 5)
New value: 89.0 → Event (94.0 - 89.0 = 5, equals deadband)
```

## Deadband Value Selection

### Factors

| Factor | Consideration |
|--------|--------------|
| Signal noise | Higher deadband for noisy signals |
| Required accuracy | Lower deadband for precise tracking |
| Event rate | Balance between responsiveness and traffic |
| System capacity | Deadband affects event buffer usage |

### Example Values

| Application | Typical Deadband |
|-------------|-----------------|
| Voltage | ±1-5 V |
| Current | ±5-10 A |
| Power | ±10-50 kW |
| Temperature | ±0.5-2 °C |

## Configuration

### WRITE Deadbands

Masters configure deadbands using WRITE:

```
Function: WRITE
Objects:
  - Group 61, Variation 1, Qualifier 0x05
    - Point 0: Deadband = 5.0
    - Point 1: Deadband = 10.0
```

### READ Deadbands

Masters read current deadband configuration:

```
Function: READ
Objects:
  - Group 61, Variation 1, Qualifier 0x00 (all)
```

## Deadband vs. Event Reporting

### Relationship

```
Data Point Configuration:
├─ Value: Current measurement
├─ Deadband: Event threshold
├─ Class: Event priority
└─ Event Enable: On/Off

Event Generation:
├─ Value changes
├─ Check deadband
├─ If exceeded → Generate event
└─ Store new value as reference
```

### Event Buffer Impact

| Deadband Setting | Event Rate | Buffer Usage |
|-----------------|------------|--------------|
| Very small | High | High |
| Small | Medium | Medium |
| Large | Low | Low |

## Common Mistakes

### Mistake 1: Zero Deadband

**Problem**: Every change generates event (no filtering).

**Fix**: Set reasonable deadband for your application.

### Mistake 2: Too Large Deadband

**Problem**: Significant changes don't generate events.

**Fix**: Balance deadband for required accuracy.

### Mistake 3: Deadband in Wrong Units

**Problem**: Deadband value incompatible with measurement.

**Fix**: Match deadband units to measurement units.

### Mistake 4: Not Configuring Deadbands

**Problem**: Using default deadbands may be inappropriate.

**Fix**: Configure deadbands per application requirements.

## Engineering Notes

### Performance Considerations

| Aspect | Impact |
|--------|--------|
| Small deadband | More events, more bandwidth |
| Large deadband | Fewer events, potential missed changes |
| Adaptive deadband | More complex but efficient |

### Implementation Notes

1. **Store last event value**: For comparison
2. **Check absolute change**: |new - old| > deadband
3. **Apply to appropriate units**: Deadband in measurement units
4. **Consider relative deadband**: Percentage-based for varying ranges

## Relationships

- **Parent**: [150-events.md](150-events.md)
- **Related**: [100-measurements.md](100-measurements.md), [260-quality-flags.md](260-quality-flags.md)

## References

- IEEE 1815-2012 Group 61: Analog Output Deadband
- IEEE 1815-2012 Section 7.8: Event Data
