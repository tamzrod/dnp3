---
title: "240 - Object Groups"
owner: object-groups
---

# What are DNP3 Object Groups?

## Purpose

DNP3 Object Groups categorize data types into logical categories. Each group represents a specific type of data (binary input, analog input, counter, etc.).

## Group Summary

### Binary Data Groups

| Group | Name | Description |
|-------|------|-------------|
| 1 | Binary Input | Current binary (on/off) status |
| 2 | Binary Input Event | Binary change events |
| 3 | Double-Bit Binary Input | Two-bit binary status |
| 4 | Double-Bit Binary Event | Double-bit change events |
| 10 | Binary Output | Controllable binary outputs |
| 11 | Binary Output Event | Output change events |
| 12 | Binary Output Command | Control command parameters |

### Counter Groups

| Group | Name | Description |
|-------|------|-------------|
| 20 | Binary Counter | Counting values (pulses) |
| 21 | Frozen Counter | Counter frozen at point in time |
| 22 | Counter Event | Counter change events |
| 23 | Frozen Counter Event | Frozen counter change events |

### Analog Groups

| Group | Name | Description |
|-------|------|-------------|
| 30 | Analog Input | Current analog values |
| 31 | Analog Input Event | Analog change events |
| 32 | Frozen Analog Input | Analog frozen at point in time |
| 33 | Frozen Analog Event | Frozen analog change events |
| 40 | Analog Output | Controllable analog values |
| 41 | Analog Output Command | Setpoint command parameters |
| 43 | Analog Output Status | Current analog output value |

### Time Groups

| Group | Name | Description |
|-------|------|-------------|
| 50 | Time and Date | Current time value |
| 51 | Time and Date Last Event | Time of last event |
| 52 | Time and Date CTO Cohort | Unsynchronized time |

### Control/Status Groups

| Group | Name | Description |
|-------|------|-------------|
| 60 | Class Data | Class assignments |
| 61 | Analog Input Deadband | Deadband configuration |
| 70 | Control Status | Internal device status |
| 80 | Device Attributes | Device identification |

### File/Transfer Groups

| Group | Name | Description |
|-------|------|-------------|
| 90 | File Control | File transfer operations |
| 91 | File Identifier | File identification |
| 92 | File Specification | File specification |

## Binary Input Groups (1-4)

### Group 1: Binary Input

Current on/off status of binary points.

**Typical use**: Breaker positions, switch status, alarm states.

### Group 2: Binary Input Event

Events generated when binary inputs change.

**Typical use**: Change-of-state notifications.

### Group 3: Double-Bit Binary Input

Two-bit status with intermediate state.

**Values**: 
- 00 = Intermediate (transitioning)
- 01 = Determinate Open
- 10 = Determinate Closed
- 11 = Indeterminate

**Typical use**: Breaker positions with transitional indication.

## Counter Groups (20-23)

### Group 20: Binary Counter

Accumulating counter values.

**Typical use**: Energy pulses, event counts, accumulated values.

**Variations**:
- 16-bit, 32-bit
- With/without overflow flag
- Delta values

### Group 21: Frozen Counter

Counter value frozen at a point in time.

**Typical use**: Snapshot for billing, synchronization.

## Analog Groups (30-33)

### Group 30: Analog Input

Current continuous values.

**Typical use**: Voltage, current, power, temperature.

**Variations**:
- 16-bit integer
- 32-bit integer
- 32-bit float
- 64-bit double

### Group 31: Analog Input Event

Events generated when analog values change beyond deadband.

**Typical use**: Threshold violations, measurement changes.

## Control Groups (10-12, 40-43)

### Group 10: Binary Output

Controllable binary outputs with status.

**Typical use**: Controllable relays, indicators.

### Group 12: Binary Output Command

Parameters for binary control operations.

**Contains**:
- Control code (latch on/off, pulse)
- On/Off time for pulses
- Queue priority

### Group 40: Analog Output

Current analog output values.

**Typical use**: Setpoints, target values.

### Group 43: Analog Output Command

Parameters for analog control (setpoint) operations.

## Time Groups (50-52)

### Group 50: Time and Date

Current time with date.

**Format**: Milliseconds since midnight + days since 1980.

### Group 51: Time and Date Last Event

Timestamp of most recent event.

### Group 52: Unsynchronized Time

Time with indication of unsynchronization gap.

## Class Group (60)

### Group 60: Class Data

Used for class-based polling and configuration.

| Variation | Description |
|-----------|-------------|
| 1 | All classes (0-3) |
| 2 | Class 0 data |
| 3 | Binary events by class |
| 4 | Analog events by class |
| 5 | Counter events by class |

## System/Status Groups (70-89)

### Group 70: Internal Indication

Device status (IIN) flags.

### Group 80: Device Attributes

Device identification and configuration.

| Variation | Description |
|-----------|-------------|
| 1 | Device serial number |
| 2 | Device model |
| 3 | Device hardware version |
| 4 | Device software version |
| 5 | Protocol version |
| 6 | Device location |
| 7 | Device ID |

## Relationships

- **Parent**: [090-object-model.md](090-object-model.md)
- **Related**: [100-measurements.md](100-measurements.md), [250-variations.md](250-variations.md)

## References

- IEEE 1815-2012 Annex A: Object Types
- IEEE 1815-2012 Tables A-1 through A-10
