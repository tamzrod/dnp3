# DNP3-INV-003: DNP3 Workbench - Outstation & Master Integration

**Status**: IN_PROGRESS
**Created**: 2026-07-27
**Agent**: OpenHands

## Quick Summary

Investigate and implement a working DNP3 workbench with functional outstation and master modes that can communicate with each other.

## Investigation Goals

1. **Two Windows executables** - master and outstation
2. **Outstation with random data** - provides moving data on data points
3. **Master read/write** - can read and write data to outstation
4. **Interoperability** - any DNP3 master can talk to our outstation (and vice versa)

## Files

| File | Purpose |
|------|---------|
| SPEC.md | Investigation specification |
| README.md | This file |
| CONCLUSION.md | Investigation findings (when complete) |

## Investigation Status

- [x] SPEC.md created
- [ ] Gap analysis completed
- [ ] Architecture selected
- [ ] Master integration verified
- [ ] Outstation integration verified
- [ ] Random data simulation implemented
- [ ] Build verified
- [ ] Connection test passed

## Running This Investigation

```bash
# Check current workbench state
cd cmd/workbench && go build .

# Run gap analysis
# (See CONCLUSION.md for findings)

# Build executables
cd cmd/workbench && go build -o workbench-master.exe .
cd cmd/workbench && go build -o workbench-outstation.exe .
```

## Related Documents

- `laboratory/planning/DNP3-ENG-WORKBENCH-001.md` - Original engineering plan
- `cmd/workbench/` - Implementation location

## Key Questions to Answer

1. Is the DNP3 protocol stack complete?
2. What is missing from master/outstation implementation?
3. Architecture: single app vs two executables?
4. What DNP3 conformance level is needed?
