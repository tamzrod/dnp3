# KDE-INV-063: Workbench Master-Outstation Binary Integration Test

**Engine**: KDE-ENGINE-001 (v1.0.0)  
**Seed**: SEED-001 (v1.0.0)  
**Timestamp**: 2026-07-29T08:20:00Z  
**Status**: 🔬 IN_PROGRESS

## Research Question

Can two separate workbench binaries (master and outstation) successfully communicate over TCP and exchange DNP3 data?

## Background

Prior investigation (KDE-INV-061/062) fixed parsing bugs in the DNP3 library. However, testing was done in-process using the same memory space. We need to verify actual TCP communication between two separate binaries.

## Passing Criteria

| ID | Criteria |
|----|----------|
| PC-1 | Outstation binary started and have data |
| PC-2 | Outstation listening on a TCP port |
| PC-3 | Master binary connects to outstation |
| PC-4 | Master can READ data from outstation |
| PC-5 | Master can WRITE data to outstation |

## Scope

- Test `cmd/workbench/workbench-fixed` binary in both modes
- Verify TCP socket communication
- Confirm end-to-end data flow

## Next Actions

1. Create experiment LAB-063
2. Run outstation binary in background
3. Run master binary and verify connection
4. Capture evidence of communication
