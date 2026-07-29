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
| Default Seed | SEED-001 (Genesis) |
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
| [/engines/current.md](/workspace/project/kde/engines/current.md) | Engine registry (shows Status vs Default) |
| [/laboratory/LABORATORY-RULES.md](/workspace/project/kde/laboratory/LABORATORY-RULES.md) | Laboratory Rules (includes Human Authority directive) |

---

**Authority**: Human Authority
**Immutable by KDE**: YES
