# Runtime Configuration

**Directory**: `/governance/runtime/`
**Purpose**: Browser-style Runtime default management
**Authority**: Human Authority

---

## Overview

This directory contains Runtime configuration artifacts that define how KDE initializes and which Engine/Seed is used by default.

---

## Key Principle

**Scientific lifecycle and operational configuration are separate.**

- **Engine Status**: Scientific validation level (Experimental, Candidate, Active, Historical)
- **Runtime Default**: Human-configured operational setting (YES/NO)

---

## Documents

| Document | Purpose |
|----------|---------|
| [defaults.yaml](./defaults.yaml) | Runtime default configuration (single source of truth) |
| [RUNTIME-STARTUP.md](./RUNTIME-STARTUP.md) | Runtime initialization sequence |
| [SESSION-OVERRIDE.md](./SESSION-OVERRIDE.md) | Session override behavior |
| [MIGRATION-NOTES.md](./MIGRATION-NOTES.md) | Migration to browser-style model |

---

## Current Configuration

| Field | Value |
|-------|-------|
| Default Engine | KDE-ENGINE-002 (Beta) |
| Default Seed | SEED-003 (Bootstrap Validation) |
| Default Status | YES |

---

## Browser-Style Default Model

| Browser | KDE Runtime |
|---------|-------------|
| User sets home page | Human sets default Engine |
| Extensions can override temporarily | Sessions can override temporarily |
| User always has final authority | Human always has final authority |

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [Engine registry](../../engines/current.md) | Verified repository-relative engine registry (status versus default) |
| [KDE governance hierarchy](../GOVERNANCE-HIERARCHY.md) | Verified human-approval and authority definitions; former `/workspace/project/kde/laboratory/LABORATORY-RULES.md` is absent and must not be inferred |

---

**Authority**: Human Authority
**Immutable by KDE**: YES

## Documentation precedence (GOV-RECON-001)

This page summarizes KDE runtime settings; the approved `.kde/governance/runtime/defaults.yaml` defines the actual default engine and seed. Do not change approved defaults by editing this README. Engine readiness does not authorize a CWAL CURRENT, code changes, dependency installation or governance promotion.
