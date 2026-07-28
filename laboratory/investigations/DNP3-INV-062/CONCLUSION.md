# DNP3-INV-062 Conclusion

**Investigation ID**: DNP3-INV-062
**Status**: APPROVED WITH ADDENDUM
**Date**: 2026-07-27

---

## 1. Summary

### 1.1 Investigation Request

User requested separation of Master and Outstation into dedicated instances, not tabs, panes, or mode switching.

### 1.2 User Requirements Confirmed

| Requirement | Confirmed |
|-------------|------------|
| Single executable | ✅ |
| Dedicated instances | ✅ |
| No shared UI | ✅ |
| Parallel operation | ✅ |
| Startup mode selection | ✅ |

### 1.3 Approved Addendum

| Requirement | Description |
|-------------|-------------|
| File menu with window controls | Minimize, Maximize/Restore, Close buttons |

---

## 2. Approved Solution

### 2.1 Architecture

```
./workbench --mode master     → Master Window
./workbench --mode outstation  → Outstation Window
./workbench                   → Mode Selection Dialog
```

### 2.2 Key Changes

1. Add `--mode` flag to command-line interface
2. Create mode selection dialog for default behavior
3. Separate window layouts for Master and Outstation
4. Dedicated controllers per mode
5. File menu with window controls (Minimize, Maximize/Restore, Close)

### 2.3 File Menu Structure

```
File
├── Minimize        (Alt+N)
├── Maximize        (Alt+M)
├── Restore         (Alt+R)
├── ─────────────
├── Close           (Alt+F4)
├── Exit            (Ctrl+Q)
```

---

## 3. Implementation Status

| Phase | Status |
|-------|--------|
| Investigation | ✅ APPROVED WITH ADDENDUM |
| Specification | ✅ COMPLETE |
| Implementation | READY TO PROCEED |
| Testing | PENDING |
| Deployment | PENDING |

---

## 4. Next Actions

1. [ ] Create experiments for validation
2. [ ] Implement --mode flag parsing
3. [ ] Implement mode selection dialog
4. [ ] Create Master controller and window
5. [ ] Create Outstation controller and window
6. [ ] Add File menu with window controls
7. [ ] Test all modes

---

*Investigation approved: 2026-07-27*
