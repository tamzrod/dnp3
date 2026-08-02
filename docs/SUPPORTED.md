# DNP3 Protocol Support Inventory

This document lists the object groups, variations, and function codes actually implemented in the codebase. Based on code analysis, not marketing claims.

## Application Layer Function Codes

### Master → Outstation (Supported)

| Code | Name | Status |
|------|------|--------|
| 0x02 | READ | ✅ Supported |
| 0x03 | WRITE | ✅ Supported |
| 0x04 | SELECT (implicit via OPERATE) | ⚠️ Partial |
| 0x05 | OPERATE | ✅ Supported |
| 0x06 | DIRECT OPERATE | ✅ Supported |
| 0x07 | DIRECT OPERATE NO RESP | ✅ Supported |
| 0x14 | ENABLE UNSOLICITED | ❌ Not implemented |
| 0x15 | DISABLE UNSOLICITED | ❌ Not implemented |
| 0x08 | AUTHENTICATE | ❌ Not implemented |
| 0x09 | AUTHENTICATION ERROR | ❌ Not implemented |

### Outstation → Master (Supported)

| Code | Name | Status |
|------|------|--------|
| 0x00 | RESPONSE | ✅ Supported |
| 0x01 | UNSOLICITED RESPONSE | ❌ Not implemented |

---

## Object Groups and Variations

### Group 1: Binary Input

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | Binary Input (with flags) | index(2) + flags(1) | ✅ | ✅ | State in bit 7 (0x80), quality in bits 0-6 |
| 2 | Binary Input (without flags) | index(2) + value(1) | ✅ | ✅ | Quality defaults to ONLINE |

### Group 2: Binary Input Event

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | Binary Input Event (with time) | index(2) + flags(1) + time(6) | ❌ | ❌ | Not implemented |
| 2 | Binary Input Event (without time) | index(2) + flags(1) | ❌ | ❌ | Not implemented |
| 3 | Binary Input Event with Relative Time | — | ❌ | ❌ | Not implemented |

### Group 10: Binary Output

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | Binary Output (with flags) | index(2) + flags(1) | ❌ | ✅ | Included in Class 0; State in bit 7 (0x80), quality in bits 0-6 |
| 2 | Binary Output (without flags) | index(2) + value(1) | ❌ | ✅ | Included in Class 0 |

### Group 11: Binary Output Event

| Variation | Name | Status |
|-----------|------|--------|
| 1 | Binary Output Event with Time | ❌ Not implemented |
| 2 | Binary Output Event | ❌ Not implemented |

### Group 20: Counter

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | 32-bit Counter (with flags) | index(2) + value(4) + flags(1) | ✅ | ✅ | Big-endian |
| 2 | 16-bit Counter (with flags) | index(2) + value(2) + flags(1) | ❌ | ❌ | Not implemented |
| 5 | 32-bit Counter (no flags) | index(2) + value(4) | ❌ | ❌ | Not implemented |
| 6 | 16-bit Counter (no flags) | index(2) + value(2) | ❌ | ❌ | Not implemented |

### Group 21: Counter Event

| Variation | Name | Status |
|-----------|------|--------|
| 1 | 32-bit Counter Event with Time | ❌ Not implemented |
| 2 | 16-bit Counter Event with Time | ❌ Not implemented |

### Group 30: Analog Input

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | 32-bit Float (with flags) | index(2) + float32(4) + flags(1) | ✅ | ✅ | IEEE 754 big-endian |
| 2 | 16-bit Integer (with flags) | index(2) + int16(2) + flags(1) | ❌ | ❌ | Not implemented |
| 3 | 32-bit Integer (with flags) | index(2) + int32(4) + flags(1) | ❌ | ❌ | Not implemented |
| 4 | 64-bit Float (with flags) | index(2) + float64(8) + flags(1) | ❌ | ❌ | Not implemented |

### Group 31: Analog Input Event

| Variation | Name | Status |
|-----------|------|--------|
| 1 | 32-bit Float Event with Time | ❌ Not implemented |
| 2 | 16-bit Integer Event with Time | ❌ Not implemented |

### Group 40: Analog Output

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | 32-bit Float (with flags) | index(2) + float32(4) + flags(1) | ❌ | ✅ | Included in Class 0; IEEE 754 big-endian |
| 2 | 16-bit Integer (with flags) | index(2) + int16(2) + flags(1) | ❌ | ✅ | Included in Class 0 |

### Group 50: Time

| Variation | Name | Status |
|-----------|------|--------|
| 1 | Time and Date | ❌ Not implemented |
| 2 | Time and Date CTO | ❌ Not implemented |

### Group 60: Class Data

| Variation | Name | Encoding | Master Parse | Outstation Build | Notes |
|-----------|------|----------|--------------|-----------------|-------|
| 1 | Class 0 Data | — | ✅ | ✅ | Returns all static data |
| 2 | Class 1 Data | — | ❌ | ❌ | Not implemented |
| 3 | Class 2 Data | — | ❌ | ❌ | Not implemented |
| 4 | Class 3 Data | — | ❌ | ❌ | Not implemented |

---

## Quality Flags

| Bit | Name | Status |
|-----|------|--------|
| 0 | ONLINE | ✅ Implemented |
| 1 | RESTART | ✅ Implemented |
| 2 | COMM_LOST | ⚠️ Defined, not actively set |
| 3 | REMOTE_FORCED | ⚠️ Defined, not actively set |
| 4 | LOCAL_FORCED | ⚠️ Defined, not actively set |
| 5 | ROLLOVER | ❌ Not implemented |
| 6 | DISCONTINUITY | ⚠️ Defined, not actively set |
| 7 | OVERRANGE | ❌ Not implemented |

---

## NOT Currently Supported (Future Work)

- Unsolicited responses (Class 0-3 event buffers)
- Time synchronization (Time and Date objects)
- File transfer
- Authentication/Security Association
- DNP3 over TLS (secure authentication)
- Frozen counters (Group 21)
- Device attributes (Group 0)
- Cold restart / warm restart
- Delay measurement
- Command targeting / SBO (select-before-operate)

---

## Protocol Stack Layers

| Layer | Status | Notes |
|-------|--------|-------|
| DLL (Data Link Layer) | ✅ Working | Supports primary/secondary, confirm/not-confirm |
| TL (Transport Layer) | ✅ Working | Handles 0x064 byte, fragmentation |
| AL (Application Layer) | ✅ Working | Function codes, IIN, object parsing |
| Master Controller | ✅ Working | State machine, retry, timeout |
| Outstation Controller | ✅ Working | Request handling, response building |

---

## Spec Compliance Notes

This implementation follows a subset of IEEE 1815-2012 (DNP3). Known deviations:

1. **No unsolicited responses**: Outstation never initiates communication
2. **No security**: No authentication, no secure authentication
3. **No time synchronization**: No Time objects in Class 0
4. **No event buffers**: Static data only; no event tracking
5. **Binary Input state bit**: Uses bit 7 (not bit 0) to avoid conflict with ONLINE flag - see code comment

These limitations are intentional for the current use case (simple master↔outstation demo).
