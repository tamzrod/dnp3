# Migration Guide: Grandfathered to KDE_RUNTIME

**Document Version**: 1.0.0  
**Date**: 2026-07-28  
**Source**: INV-081 (Caveman/ENZO Evolution Analysis)  
**Status**: APPROVED

---

## Purpose

This guide helps investigations migrate from the grandfathered `KDE_RUNTIME_AUTHENTICITY: GENERIC_AI_WITH_KDE_FORMAT` format to the new `EXECUTION_MODE: KDE_RUNTIME` format required by Rule 8.

---

## Background

### Why Migration is Needed

The Caveman/ENZO evolution series (INV-055-073) was stopped because investigations were using the old HTML comment format:

```html
<!-- KDE_RUNTIME_AUTHENTICITY: GENERIC_AI_WITH_KDE_FORMAT -->
```

This format:
1. Was a grandfathered exemption (pre-Rule 8)
2. Did not require actual KDE_RUNTIME execution
3. Classified investigations as GENERIC_AI
4. Has authenticity score of only 15%

### New Format Benefits

| Aspect | Old Format | New Format |
|--------|------------|------------|
| Authenticity Score | 15% | 100% |
| Engine Recognition | Limited | Full |
| Evolution Support | Blocked | Enabled |
| Governance | Legacy | Current |

---

## Migration Steps

### Step 1: Add EXECUTION_MODE Header

Add a YAML frontmatter to your investigation's README.md:

```yaml
---
EXECUTION_MODE: KDE_RUNTIME
AUTHENTICITY_SCORE: 100%
RUNTIME_AUTHORITY: Verified
BOOTSTRAP_VERIFIED: YES
---
```

**Note**: The `AUTHENTICITY_SCORE` should be 100% if you're actually running with KDE_RUNTIME.

### Step 2: Remove Old Format

Remove the old HTML comment:
```html
<!-- Remove this line -->
<!-- KDE_RUNTIME_AUTHENTICITY: GENERIC_AI_WITH_KDE_FORMAT -->
```

### Step 3: Run Bootstrap Verification

Verify your environment is ready:

```bash
cd /workspace/project/kde
python3 .kde/bootstrap/gates.py
```

Expected output:
```
============================================================
KDE BOOTSTRAP GATES
============================================================

  [PASS] Gate 1: Directory structure exists
  [PASS] Gate 2: Runtime initialized
  [PASS] Gate 3: ECU available
  [PASS] Gate 4: Engine loaded
  [PASS] Gate 5: Governance rules loaded
  [PASS] Gate 6: Seed loaded

[OK] Bootstrap verification complete
```

### Step 4: Run Pre-Flight Check

```bash
python3 .kde/commands/check.py
```

Expected output:
```
============================================================
KDE PRE-FLIGHT CHECK
============================================================

  [PASS] Bootstrap Gates: 6/6 checks passed
  [PASS] Runtime State: initialized
  [PASS] ECU Enforcement: ECU check skipped

[OK] Ready for KDE_RUNTIME investigation
```

### Step 5: Validate ECU Enforcement

```python
from runtime.ecu import create_ecu
ecu = create_ecu()
print(f"ECU Status: {ecu.state}")
```

### Step 6: Run Verification

```bash
python3 .kde/verification/compliance.py
```

Should show 0 errors for your investigation.

---

## Common Issues

### Issue: "Missing EXECUTION_MODE"

**Cause**: README.md doesn't have the header.

**Fix**: Add the YAML frontmatter as shown in Step 1.

### Issue: "Invalid EXECUTION_MODE value"

**Cause**: EXECUTION_MODE has wrong value.

**Fix**: Use one of: `KDE_RUNTIME`, `GENERIC_AI`, `HYBRID`

### Issue: "Missing AUTHENTICITY_SCORE"

**Cause**: Required for GENERIC_AI or HYBRID modes.

**Fix**: Add `AUTHENTICITY_SCORE: XX%` to header.

### Issue: "Grandfathered investigation"

**Cause**: Old format still detected.

**Fix**: Remove the `<!-- KDE_RUNTIME_AUTHENTICITY: -->` comment entirely.

---

## Validation Checklist

| Check | Command | Expected |
|-------|---------|----------|
| Bootstrap | `python3 .kde/bootstrap/gates.py` | 6/6 PASS |
| Pre-flight | `python3 .kde/commands/check.py` | READY |
| Verification | `python3 .kde/verification/compliance.py` | 0 errors |
| ECU | `python3 -c "from runtime.ecu import create_ecu; print('OK')"` | OK |

---

## Rollback

If migration causes issues, you can temporarily use:

```yaml
---
EXECUTION_MODE: GENERIC_AI
AUTHENTICITY_SCORE: 50%
---
```

This will:
- Allow investigation to proceed
- Mark it as partially authentic
- Require gradual improvement

---

## Support

For issues during migration:
1. Run pre-flight check: `python3 .kde/commands/check.py`
2. Check verification: `python3 .kde/verification/compliance.py --strict`
3. Review this guide's Common Issues section
4. Consult INV-081 for detailed troubleshooting

---

**Document Status**: APPROVED  
**Human Review Required**: No (automated migration)  
**Evidence**: INV-081/REC-003
