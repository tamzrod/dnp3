# KDSE Runtime Environment

**Version:** 1.0  
**Repository:** go-dnp3  
**Initialized:** 2026-07-10

---

## Overview

This directory (`.kdse/`) is the KDSE Runtime Environment for the go-dnp3 repository. It serves as the local engineering governance directory, similar to how `.git/` manages version control.

## Directory Structure

```
.kdse/
├── standards/              # Pinned KDSE normative standards
├── runtime/               # Transient execution state
├── reports/              # Generated reports (never overwritten)
├── history/              # Historical records (preserved)
├── cache/               # Optional runtime cache
├── config.yaml          # Runtime configuration
├── manifest.yaml        # Version manifest
└── README.md           # This file
```

## Key Concepts

| Concept | Description |
|---------|-------------|
| Version Pinning | Repository pins exact KDSE version for reproducibility |
| Offline Execution | Standards available locally, no network required |
| Report Preservation | Reports never overwritten, unique timestamps |
| Sync Control | Updates are manual, not automatic |

## Important

**This directory is managed by KDSE.**

Do not edit manually unless instructed by KDSE documentation. All modifications should be made through KDSE commands:

```bash
kdse run          # Run a Runtime Session
kdse sync         # Sync to new KDSE version
kdse status       # Check environment status
kdse report       # Generate report
```

## Standards

The `standards/` directory contains the normative KDSE documents:
- **foundation/** - Core principles, models, definitions
- **audit/** - Audit standards and scoring
- **execution/** - Session protocol and workflows
- **templates/** - Standardized templates

## Reports

The `reports/` directory stores all generated reports:
- **sessions/** - Runtime Session Reports
- **audits/** - Compliance and Foundation Audits
- **reviews/** - Execution Reviews

Reports are never overwritten. Each report has a unique timestamp.

## History

The `history/` directory maintains execution history:
- **audit-history/** - Audit execution records
- **session-history/** - Session records
- **sync-history/** - Standard update records

## See Also

- [KDSE Standard Repository](https://github.com/tamzrod/KDSE)
- [go-dnp3 KDSE Audit Report](../docs/project/KDSE_AUDIT_REPORT.md)
- [go-dnp3 Phase Completion](../docs/project/PHASE_COMPLETION.md)

---

*This directory was created by KDSE Runtime Session on 2026-07-10 as the first KDSE Case Study reference implementation.*
