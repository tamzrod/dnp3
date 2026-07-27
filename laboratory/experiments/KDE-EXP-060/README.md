# KDE-EXP-060: Data Table UI Prototype

**Experiment**: KDE-EXP-060  
**Date**: 2026-07-27  
**Status**: in_progress  
**Investigation**: KDE-INV-060  
**Decision**: KDE-DEC-060  

---

## Hypothesis

A table-based data view with checkboxes will provide better UX for monitoring and controlling DNP3 data points compared to simple labels.

## Variables

### Independent Variables
- Data display format (table vs labels)
- Selection method (checkboxes vs multi-select)
- Control location (separate panel vs modal dialogs)

### Dependent Variables
- Time to find a specific point value
- Time to perform an operate action
- User error rate in point selection

## Method

### 1. Mock Data Generation

Create mock DNP3 data points for testing:
- 50 Binary Inputs (0-49)
- 50 Binary Outputs (0-49)
- 50 Analog Inputs (0-49)
- 50 Analog Outputs (0-49)
- 50 Counters (0-49)

### 2. Table Implementation

Implement Fyne Table widget with:
- Custom cell factory for checkboxes
- Virtual scrolling for performance
- Sortable columns
- Real-time value updates via binding

### 3. Control Panel Implementation

Add control panel that:
- Shows selected output points
- Provides ON/OFF buttons for DO
- Provides value input for AO
- Sends operate commands via controller

## Test Scenarios

### Scenario 1: Find Binary Input Value
1. Open table with 250 points
2. Find index 23, check if DI value matches expected

### Scenario 2: Operate Binary Output
1. Select DO at index 5
2. Click "Operate ON"
3. Verify command sent

### Scenario 3: Monitor Multiple Points
1. Select DI 10, DI 20, DI 30
2. Click "Read Selected"
3. Verify only selected points polled

## Expected Results

| Metric | Labels (Current) | Table (Proposed) |
|--------|------------------|------------------|
| Find point time | ~5s | ~1s |
| Operate action | ~10s (dialog) | ~3s (inline) |
| Error rate | 15% | 2% |

## Implementation Files

| File | Purpose |
|------|---------|
| `cmd/workbench/internal/ui/panels/datatable.go` | Table panel with mock data |
| `cmd/workbench/internal/ui/panels/control.go` | Control panel |
| `test_mock_data.go` | Mock data generation |

## Status

- [x] Experiment created
- [x] Mock data implemented
- [x] Table widget implemented
- [x] Control panel implemented
- [ ] Results documented

## Notes

This experiment validates the UI approach before full implementation.
