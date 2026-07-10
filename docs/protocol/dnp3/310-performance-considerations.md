---
title: "310 - Performance Considerations"
owner: performance
---

# What are DNP3 Performance Considerations?

## Purpose

DNP3 Performance Considerations guide optimization of DNP3 implementations and deployments. Proper optimization ensures efficient, reliable communication.

## Problem Being Solved

Poor performance causes:

1. **Data loss** - Events dropped under load
2. **Latency** - Delayed responses
3. **Bandwidth waste** - Excessive traffic
4. **Scalability limits** - Can't support many devices

Understanding performance helps avoid these issues.

## Performance Metrics

### Latency

| Operation | Typical Latency |
|-----------|-----------------|
| Local round-trip | < 10 ms |
| LAN round-trip | 1-50 ms |
| WAN round-trip | 50-500 ms |
| Serial round-trip | 10-100 ms |

### Throughput

| Metric | Typical Value |
|--------|---------------|
| Frames/second | 100-10,000 |
| Data rate (TCP) | 1-50 Mbps |
| Event rate | Device-dependent |

### Capacity

| Resource | Typical Limit |
|----------|---------------|
| Points per outstation | 100-10,000 |
| Connections per master | 10-10,000 |
| Event buffer | 100-10,000 events |

## Bandwidth Optimization

### Efficient Data Encoding

| Variation | Bytes | Efficiency |
|-----------|-------|-------------|
| Binary | 1-2 | Very high |
| 16-bit int | 2-3 | High |
| 32-bit int | 4-5 | Medium |
| 32-bit float | 4-5 | Medium |
| 64-bit float | 8-9 | Low |

**Tip**: Use smallest variation that meets accuracy needs.

### Event Reporting vs. Polling

| Scenario | Recommended Method |
|----------|-------------------|
| Low change rate | Polling |
| High change rate | Events |
| Critical data | Events (Class 1) |
| Non-critical data | Polling |

### Class-Based Polling

Efficient polling uses classes:

```
Without classes:
  Poll interval: 10 seconds
  All data transmitted every poll

With classes:
  Class 1: 2 seconds (100 bytes)
  Class 2: 10 seconds (200 bytes)
  Class 0: 1 hour (1000 bytes)
  
  Average: 50 bytes/sec vs 100 bytes/sec
```

## Latency Optimization

### Reducing Round-Trip Time

| Technique | Impact |
|-----------|--------|
| TCP keepalive off | Minor |
| Small MTU | Moderate |
| Local network | Major |
| Eliminate retries | Major |

### Response Processing

```
┌──────────────────────────────────────────────────────────────────┐
│                    RESPONSE LATENCY BREAKDOWN                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Frame received ──► Parse ──► Validate ──► Process ──► Respond │
│                                                                   │
│  Typical breakdown:                                                │
│  ├─ Network: 1-50 ms                                             │
│  ├─ Parse: 0.1-1 ms                                             │
│  ├─ Validate: 0.1-1 ms                                          │
│  ├─ Process: 0.1-10 ms                                          │
│  └─ Respond: 0.1-1 ms                                           │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Timeout Selection

| Network | Recommended Timeout |
|---------|--------------------|
| LAN | 1-5 seconds |
| WAN | 5-30 seconds |
| Serial | 5-30 seconds |

## Scalability

### Master Scaling

| Outstations | Recommended Architecture |
|-------------|------------------------|
| 1-50 | Single master |
| 50-200 | Multiple channels |
| 200-1000 | Hierarchical |
| 1000+ | Concentrators |

### Event Rate Management

```
┌──────────────────────────────────────────────────────────────────┐
│                    EVENT RATE MANAGEMENT                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Burst event rate:                                                │
│  ├─ Limit queue depth                                            │
│  ├─ Prioritize by class                                          │
│  └─ Drop lowest priority if full                                 │
│                                                                   │
│  Sustained event rate:                                            │
│  ├─ Increase poll frequency                                       │
│  ├─ Adjust deadbands                                             │
│  └─ Consider aggregation                                          │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

## Memory Optimization

### Buffer Sizing

| Buffer | Minimum Size | Recommended |
|--------|--------------|-------------|
| Frame buffer | 302 bytes | 302 bytes |
| Fragment buffer | 4 KB | 10 KB |
| Event buffer | 100 events | 1000 events |

### Memory Usage

```
Typical outstation memory requirements:

┌──────────────────────────────────────────────────────────────────┐
│                    MEMORY BUDGET                                   │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Database: 10 KB (1000 points × 10 bytes avg)                    │
│  Event buffer: 100 KB (1000 events × 100 bytes avg)              │
│  Fragment buffer: 10 KB                                          │
│  Stack/protocol: 50 KB                                           │
│  ─────────────────────────────────────────────────────────────────│
│  Total: ~170 KB minimum                                          │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

## Common Performance Issues

### Issue 1: Buffer Too Small

**Problem**: Fragmentation fails, events dropped.

**Solution**: Size buffers for maximum expected message.

### Issue 2: Excessive Polling

**Problem**: Unnecessary bandwidth usage.

**Solution**: Use events, increase poll intervals.

### Issue 3: Small Deadbands

**Problem**: Event storm, buffer overflow.

**Solution**: Increase deadbands appropriately.

### Issue 4: No Connection Pooling

**Problem**: Connection overhead.

**Solution**: Reuse connections.

## Optimization Techniques

### 1. Batch Reads

```
Inefficient:
  READ G1 → READ G30 → READ G20

Efficient:
  READ G1, G30, G20 in single request
```

### 2. Unsolicited for High-Rate Data

```
Polling:
  Poll 10x/sec for 100 points = 1000 bytes/sec

Unsolicited:
  Average 10 events/sec = 100 bytes/sec
```

### 3. Appropriate Variation

```
Need: ±1 accuracy
Inefficient: 32-bit float (4 bytes)
Efficient: 16-bit int (2 bytes)
```

### 4. Class Separation

```
All classes together:
  Poll 10 sec = 1000 bytes

Class-separated:
  Class 1: 2 sec = 200 bytes
  Class 2: 10 sec = 200 bytes
  Class 0: 1 hour = 100 bytes/sec average
```

## Benchmarking

### Key Benchmarks

| Metric | Test Method |
|--------|-------------|
| Frame encode | 1M frames, measure time |
| Frame decode | 1M frames, measure time |
| Round-trip latency | Ping-pong, measure RTT |
| Event throughput | Generate N events, measure time |
| Memory usage | Profile during operation |

### Tools

- go test -bench
- pprof for profiling
- Wireshark for network analysis

## Engineering Notes

### Deployment Planning

1. **Estimate event rate** - Based on process dynamics
2. **Size buffers** - For peak load
3. **Configure deadbands** - Balance sensitivity vs. traffic
4. **Plan polling** - Based on data criticality

### Monitoring

Monitor in production:
- Event rate
- Buffer utilization
- Response latency
- Error rates

## Relationships

- **Related**: [150-events.md](150-events.md), [270-deadbands.md](270-deadbands.md), [160-class-polling.md](160-class-polling.md)

## References

- IEEE 1815-2012 Section 9: Performance Requirements
- DNP3 Users Group Technical Guidelines
