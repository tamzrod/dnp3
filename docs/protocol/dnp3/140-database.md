---
title: "140 - Database"
owner: database
---

# What is the DNP3 Database?

## Purpose

The DNP3 **Database** is the logical organization of all data points within an outstation. It provides a standardized way to address, access, and manage measurement and control data.

## Problem Being Solved

SCADA systems need to:

1. **Organize data** - Many points need structured addressing
2. **Track changes** - Know what changed and when
3. **Assign priorities** - Classify data by importance
4. **Configure behavior** - Set deadbands, event options per point

The database concept solves all of these.

## Database Concept

### What the Database Is NOT

The DNP3 database is a **logical concept**, not:

- A SQL database
- A specific file format
- A required implementation

### What the Database IS

A logical structure containing:

```
┌──────────────────────────────────────────────────────────────────┐
│                      LOGICAL DATABASE                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ POINTS (All data in the outstation)                        │ │
│  │                                                              │ │
│  │  Binary Inputs    │ Points 0-N  │ Group 1                   │ │
│  │  Binary Outputs   │ Points 0-N  │ Group 10                  │ │
│  │  Analog Inputs    │ Points 0-N  │ Group 30                  │ │
│  │  Analog Outputs   │ Points 0-N  │ Group 40                  │ │
│  │  Counters         │ Points 0-N  │ Group 20                  │ │
│  │  Frozen Values    │ Points 0-N  │ Groups 21, 32             │ │
│  │  Time Values      │ Points 0-N  │ Group 50                  │ │
│  │                                                              │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  Each point has:                                                  │
│  ├─ Current value                                                │
│  ├─ Quality flags                                                │
│  ├─ Timestamp (for events)                                        │
│  ├─ Class assignment                                             │
│  ├─ Event configuration                                          │
│  └─ Deadband (for analog)                                        │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

## Point Structure

### Common Point Attributes

Every database point has:

| Attribute | Description | Applies To |
|-----------|-------------|------------|
| Index | Point number within group | All |
| Value | Current data value | All |
| Flags | Quality/status indicators | All |
| Class | Event priority (0-3) | All |
| Timestamp | Time of last change | Events |
| Deadband | Change threshold | Analog |
| Config | Event enable, default variation | All |

### Binary Point Structure

```
┌──────────────────────────────────────────────────────────────────┐
│                    BINARY INPUT POINT                              │
├──────────────────────────────────────────────────────────────────┤
│  Index: 5                                                         │
│  Value: 1 (ON)                                                    │
│  Flags: ONLINE                                                    │
│  Class: 1 (High priority)                                         │
│  Event Enable: Yes                                               │
└──────────────────────────────────────────────────────────────────┘
```

### Analog Point Structure

```
┌──────────────────────────────────────────────────────────────────┐
│                    ANALOG INPUT POINT                              │
├──────────────────────────────────────────────────────────────────┤
│  Index: 3                                                         │
│  Value: 1234.56                                                   │
│  Flags: ONLINE                                                    │
│  Class: 2 (Medium priority)                                       │
│  Deadband: 10.0                                                   │
│  Event Enable: Yes                                               │
│  Default Variation: 3 (32-bit float)                              │
└──────────────────────────────────────────────────────────────────┘
```

## Database Organization

### Group-Based Organization

DNP3 organizes points by **group**:

| Group | Data Type | Typical Points |
|-------|-----------|----------------|
| 1 | Binary Input | Breaker status, alarms |
| 2 | Binary Input Event | Event tracking |
| 3 | Double-Bit Binary Input | Positions |
| 10 | Binary Output | Control outputs |
| 12 | Binary Output Command | Control commands |
| 20 | Binary Counter | Energy, pulses |
| 21 | Frozen Counter | Frozen at time |
| 30 | Analog Input | Voltage, current |
| 32 | Frozen Analog Input | Frozen at time |
| 40 | Analog Output | Setpoints |
| 43 | Analog Output Command | Setpoint commands |
| 50 | Time and Date | System time |

### Index Scope

Indexes are **per-group**:

```
Group 1: Point 0, 1, 2, ... N (Binary inputs)
Group 30: Point 0, 1, 2, ... M (Analog inputs)
Group 20: Point 0, 1, 2, ... K (Counters)

Point 5 in Group 1 ≠ Point 5 in Group 30
```

## Class Assignment

### Class Overview

Classes assign priority to points:

| Class | Priority | Typical Use |
|-------|----------|-------------|
| 0 | Static data | Current values (no events) |
| 1 | Highest | Critical alarms, trips |
| 2 | Medium | Status changes, measurements |
| 3 | Low | Non-critical data |

### Class Assignment Command

Masters assign classes using Function 48:

```
Function: ASSIGN_CLASS
Objects:
  - Group 1, Variation 1, Qualifier: Points 0-9 → Class 1
  - Group 30, Variation 1, Qualifier: Points 0-9 → Class 2
```

### Default Classes

| Object Type | Default Class |
|-------------|---------------|
| Static values | Class 0 |
| Binary events | Class 1 |
| Analog events | Class 2 |
| Counter events | Class 3 |

## Point Configuration

### Event Configuration

Points can be configured for event generation:

| Setting | Description |
|---------|-------------|
| Event Enable | Generate events on change |
| Class | Event priority class |
| Deadband | Threshold for analog changes |

### Variation Configuration

Points can have default variations:

| Setting | Description |
|---------|-------------|
| Default Static | Variation for static reads |
| Default Event | Variation for event reports |

Masters requesting Variation 0 receive the default.

## Database Operations

### Read Operations

| Request | Response |
|---------|----------|
| READ Group 60, Var 1 | All Class 0 data |
| READ Group 1, Var 0 | All binary inputs (default var) |
| READ Group 30, Var 2 | Specific variation |
| READ Group 30, Var 0 | Specific variation (default) |

### Write Operations

| Request | Effect |
|---------|--------|
| WRITE Group 10 | Update binary outputs |
| WRITE Group 40 | Update analog outputs |
| WRITE Group 60, Var 2 | Configure deadbands |

## Database Size

### Typical Sizes

| Application | Points | Memory |
|-------------|--------|--------|
| Simple RTU | 50-100 | < 10 KB |
| Medium RTU | 100-500 | 10-50 KB |
| Large RTU | 500-2000 | 50-200 KB |
| IED | 50-1000 | Variable |

### Memory Calculation

```
Per binary point: ~8 bytes (value, flags, class, config)
Per analog point: ~16 bytes (value, flags, class, deadband, config)
Per counter point: ~12 bytes (value, flags, class, config)
```

## Database Synchronization

### Master-Outstation Sync

```
Master ──── READ ───────────────────────────────► Outstation
     │       (Get current database state)            │
     │                                                │
     │ ◄─── RESPONSE ─────────────────────────────────
     │       (All point values)
     │                                                │
     Master builds local copy of database
```

### Cold Start vs Warm Start

| Start Type | Description | Data Loss |
|-------------|-------------|-----------|
| Cold Start | Full power loss | All values unknown |
| Warm Start | Application restart | Static values preserved |
| Reboot | Device reboot | Depends on storage |

## Internal Representation

### Possible Implementations

| Implementation | Pros | Cons |
|----------------|------|------|
| Arrays | Fast access | Fixed size |
| Maps/Hash tables | Dynamic size | Slower access |
| Linked lists | Simple | Very slow |
| Database file | Persistent | Overhead |

### Point Lookup

```
Master requests: READ Group 30, Point 5

Internal process:
1. Parse group → Select point array
2. Parse index → Access specific element
3. Encode value → Use configured variation
4. Add flags → Current status
5. Return → Response with objects
```

## Configuration Database

### Configuration Objects

| Group | Purpose |
|-------|---------|
| 80 | Protocol version |
| 81 | Device identification |
| 82 | Device label |
| 83 | Device description |
| 84 | Device serial number |
| 85 | Device configuration |
| 86 | Device model |
| 87 | Device revision |
| 88 | Time synchronization info |

## Common Misconceptions

### Misconception 1: "Database must be SQL"

**Reality**: Database is a logical concept. Implementation can use any structure.

### Misconception 2: "All points must exist"

**Reality**: Only implemented points must exist. Others return error or empty.

### Misconception 3: "Indexes must be contiguous"

**Reality**: Indexes can be sparse. Non-existent indexes return error.

### Misconception 4: "Database is always synchronized"

**Reality**: Master and outstation may have different views. Time sync and event handling manage differences.

## Engineering Notes

### Performance Considerations

| Operation | Time Complexity |
|-----------|-----------------|
| Point lookup | O(1) for array, O(n) for list |
| Range read | O(k) where k = range size |
| Event scan | O(n) for all, O(1) for buffered |
| Class read | O(n) or indexed |

### Implementation Notes

1. **Use arrays**: For fixed-size, fast access
2. **Index validation**: Check bounds on access
3. **Flag updates**: Always update with value
4. **Class tracking**: Maintain per-class lists
5. **Deadband storage**: Store with analog points

## Relationships

- **Parent**: [090-object-model.md](090-object-model.md)
- **Related**: [100-measurements.md](100-measurements.md), [150-events.md](150-events.md), [160-class-polling.md](160-class-polling.md)

## References

- IEEE 1815-2012 Section 7.6: Database Representation
- IEEE 1815-2012 Annex A: Object Types
- DNP3 Users Group Technical Guidelines
