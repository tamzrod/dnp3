# LAB-ARCH-001: Repository vs Laboratory Responsibility Investigation

**Investigation ID**: LAB-ARCH-001
**Title**: Repository vs Laboratory Responsibility Investigation
**Authority**: KDE Runtime ECU
**Status**: COMPLETED
**Date**: 2026-07-25
**Execution Agent**: OpenHands Agent
**Bootstrap**: SUCCESS

---

## Executive Summary

This investigation analyzes the apparent architectural conflict between the **Repository** (Go code, implementations) and the **Laboratory** (investigations, discoveries, specifications). 

### Core Finding

**No fundamental conflict exists.** The Repository and Laboratory serve fundamentally different purposes that are intentionally separated at the runtime level:

```
.kde/                          → RUNTIME GOVERNED ARTIFACTS
├── engines/                   → Investigation capabilities
├── seeds/                     → Reasoning foundations
├── runtime/                   → Runtime orchestration
├── laboratory/                → Investigation artifacts
└── governance/                → Policy definitions

Repository Root                 → PROJECT ARTIFACTS
├── pkg/                       → Go implementation
├── cmd/                       → Command-line tools
├── internal/                  → Internal packages
└── docs/                      → Project documentation
```

### Key Observation

The apparent conflict arises from a **temporal rather than architectural** issue: the Runtime ECU was installed in `.kde/runtime/ecu/` but represents an implementation discovered/created through investigation. This is an artifact of how the ECU was installed, not an architectural violation.

---

## 1. Repository Responsibility Report

### 1.1 Defined Purpose

**Purpose**: Contains and maintains the project's **primary deliverables**—the software that users install, depend on, and consume.

### 1.2 Defined Authority

**Authority**: Owned by the project maintainers, governed by the KDE Runtime through policy but not by runtime-owned artifacts.

### 1.3 Defined Responsibilities

| Responsibility | Evidence |
|---------------|----------|
| **Software Implementation** | pkg/, cmd/, internal/ contain Go code |
| **Build & Release** | Makefile, go.mod, go.sum |
| **User Documentation** | docs/, README.md |
| **Examples** | examples/ directory |
| **Testing** | test/ directory, *_test.go files |

### 1.4 Defined Boundaries

| Boundary | Definition |
|----------|------------|
| **MUST own** | Go source code, build configuration, user documentation |
| **MUST NOT own** | Investigation artifacts, engine specifications, runtime code |

### 1.5 Evidence from Repository

```
/workspace/project/dnp3/
├── pkg/           → DNP3 protocol implementation (11 Go files)
├── cmd/           → Command-line tools
├── internal/      → Internal packages (SA, AL, etc.)
├── docs/          → User documentation
├── examples/      → Usage examples
├── test/          → Test data
├── Makefile       → Build automation
└── go.mod         → Dependency management
```

---

## 2. Laboratory Responsibility Report

### 2.1 Defined Purpose

**Purpose**: Contains artifacts of the **engineering process**—investigations, experiments, decisions, and the knowledge extracted from them.

### 2.2 Defined Authority

**Authority**: Owned by the KDE Runtime. All artifacts are under runtime governance with appropriate metadata.

### 2.3 Defined Responsibilities

| Responsibility | Evidence |
|---------------|----------|
| **Investigations** | investigations/ - 11 investigations with conclusions |
| **Experiments** | experiments/ - Controlled hypothesis testing |
| **Decisions** | decisions/ - Technology Decision Records |
| **Evidence** | evidence/ - Empirical observations |
| **Reviews** | reviews/ - Human review documents |
| **Planning** | planning/ - Future work planning |

### 2.4 Defined Boundaries

| Boundary | Definition |
|----------|------------|
| **MUST own** | Investigation artifacts, experiment results, decisions, evidence |
| **MUST NOT own** | Production code, build artifacts, user documentation |

### 2.5 Evidence from Laboratory

```
/workspace/project/dnp3/laboratory/
├── investigations/     → 11 investigations (KDE-INV-000 through KDE-INV-047)
├── experiments/       → Laboratory experiments
├── decisions/          → Technology Decision Records
├── evidence/          → Evidence artifacts
├── reviews/           → Human reviews (e.g., DNP3-REV-001)
├── planning/          → Planning documents
├── implementations/    → Implementation Specifications (IMPs)
└── testing/          → Shared testing infrastructure
```

---

## 3. Runtime Responsibility Report

### 3.1 Defined Purpose

**Purpose**: Provides the **governance infrastructure** that enables reproducible reasoning and evidence-based engineering.

### 3.2 Defined Authority

**Authority**: Owns all artifacts within `.kde/` including runtime code, engines, seeds, and laboratory artifacts.

### 3.3 Defined Responsibilities

| Responsibility | Evidence |
|---------------|----------|
| **Bootstrap** | Initializing deterministic runtime state |
| **Engine Management** | Discovering and coordinating engines |
| **Seed Management** | Loading and maintaining seeds |
| **Policy Enforcement** | Validating artifacts against policy |
| **Governance** | Authority hierarchy per KDE-INV-002 |

### 3.4 Defined Boundaries

| Boundary | Definition |
|----------|------------|
| **MUST own** | Runtime code, engines, seeds, governance policies, laboratory structure |
| **MUST NOT own** | Production code (Go), user documentation |

### 3.5 Evidence from Runtime

```
/workspace/project/dnp3/.kde/
├── bootstrap/          → Bootstrap procedures
├── runtime/           → Runtime ECU (orchestration)
│   └── ecu/            → Execution Control Unit
├── engines/            → 8 engines (Alpha, Beta, Gamma, Delta, etc.)
├── seeds/              → 2 seeds (Genesis, Evolution)
├── governance/          → Policy definitions
├── laboratory/         → Laboratory infrastructure
└── templates/         → Artifact templates (SPEC.md, IMP.md, etc.)
```

---

## 4. Component Responsibility Matrix

| Component | Purpose | Authority | Primary Owner | Boundaries |
|-----------|---------|-----------|--------------|-----------|
| **Repository** | Production software | Project maintainers | KDE Runtime (via policy) | Go code, build config, user docs |
| **Laboratory** | Engineering knowledge | KDE Runtime | KDE Runtime | Investigations, experiments, decisions |
| **Runtime** | Governance infrastructure | KDE Runtime | KDE Runtime | Bootstrap, engines, seeds, policies |
| **Engines** | Investigation capabilities | KDE Runtime | KDE Runtime | Reasoning methodologies |
| **Seeds** | Immutable reasoning | KDE Runtime | KDE Runtime | Foundational principles |
| **Bootstrap** | Initialization | KDE Runtime | KDE Runtime | Deterministic startup |

---

## 5. Responsibility Matrix by Activity

| Activity | Repository | Laboratory | Runtime | Notes |
|----------|------------|-----------|---------|-------|
| **Programming** | PRIMARY | ❌ | ❌ | Repository owns implementation |
| **Refactoring** | PRIMARY | ❌ | ❌ | Repository owns code structure |
| **Investigation** | ❌ | PRIMARY | SUPPORT | Laboratory owns process |
| **Experimentation** | ❌ | PRIMARY | SUPPORT | Laboratory owns experiments |
| **Discovery** | ❌ | PRIMARY | ❌ | Laboratory owns findings |
| **Validation** | SUPPORT | PRIMARY | SUPPORT | Laboratory owns validation protocol |
| **Documentation** | USER DOCS | ENGINEERING DOCS | ❌ | Split by audience |
| **Runtime Evolution** | ❌ | ❌ | PRIMARY | Runtime owns its own development |
| **Architecture** | IMPLEMENTATION | SPECIFICATIONS | PRINCIPLES | Split by abstraction level |
| **Knowledge Capture** | ❌ | PRIMARY | ❌ | Laboratory owns knowledge artifacts |
| **Specifications** | ❌ | PRIMARY | ❌ | Laboratory owns specs |
| **Governance** | ❌ | ❌ | PRIMARY | Runtime owns governance |

---

## 6. Engineering Lifecycle Diagram

### 6.1 Intended Lifecycle

```
IDEA
  │
  ▼
INVESTIGATION (Laboratory)
  │  - Systematic inquiry
  │  - Evidence collection
  │  - Analysis
  │
  ▼
DISCOVERY (Laboratory)
  │  - Conclusions
  │  - New knowledge
  │  - Gaps identified
  │
  ▼
ARCHITECTURE (Laboratory → Repository)
  │  - IMPs created in Laboratory
  │  - Decisions recorded
  │
  ▼
IMPLEMENTATION (Repository)
  │  - Go code written
  │  - Tests created
  │
  ▼
VALIDATION (Laboratory + Repository)
  │  - Testing in Repository
  │  - Verification protocols in Laboratory
  │
  ▼
KNOWLEDGE EXTRACTION (Laboratory)
  │  - Lessons learned
  │  - Process improvements
  │
  ▼
REPOSITORY EVOLUTION
  │  - Code updates
  │  - Documentation
  │
  ▼
LOOP → Next Investigation
```

### 6.2 Phase Ownership

| Phase | Primary Owner | Support | Artifact Location |
|-------|---------------|---------|------------------|
| Idea | N/A | - | - |
| Investigation | Laboratory | Runtime (engines) | laboratory/investigations/ |
| Discovery | Laboratory | - | laboratory/investigations/*/CONCLUSION.md |
| Architecture | Laboratory | Repository | laboratory/implementations/ |
| Implementation | Repository | - | pkg/, cmd/, internal/ |
| Validation | Laboratory | Repository | laboratory/testing/, *_test.go |
| Knowledge Extraction | Laboratory | - | laboratory/reviews/ |
| Repository Evolution | Repository | - | (standard git workflow) |

---

## 7. Boundary Analysis

### 7.1 Implementation Belongs To

**Answer**: Repository (for production code)

| Criterion | Repository | Laboratory | Both | Neither |
|-----------|-----------|-----------|------|---------|
| **Executable by users** | ✅ | ❌ | ❌ | ❌ |
| **Part of project deliverable** | ✅ | ❌ | ❌ | ❌ |
| **Subject to version control** | ✅ | ✅ | ❌ | ❌ |
| **Contains business logic** | ✅ | ❌ | ❌ | ❌ |
| **Governed by runtime policy** | ✅ (indirect) | ✅ (direct) | ❌ | ❌ |

### 7.2 Specifications Belong To

**Answer**: Laboratory (for engineering specs)

| Criterion | Repository | Laboratory | Both | Neither |
|-----------|-----------|-----------|------|---------|
| **Defines requirements** | ❌ | ✅ | ❌ | ❌ |
| **Owned by runtime** | ❌ | ✅ | ❌ | ❌ |
| **Implementation contract** | ❌ | ❌ | ✅ | ❌ |
| **Version controlled** | ✅ | ✅ | ❌ | ❌ |

### 7.3 Runtime Code Belongs To

**Answer**: Runtime (`.kde/runtime/`)

| Criterion | Repository | Laboratory | Runtime | Neither |
|-----------|-----------|-----------|---------|---------|
| **Enables runtime governance** | ❌ | ❌ | ✅ | ❌ |
| **Part of KDE framework** | ❌ | ❌ | ✅ | ❌ |
| **Version controlled with runtime** | ❌ | ❌ | ✅ | ❌ |

### 7.4 The ECU Question

The Runtime ECU presents a unique case:

| Criterion | Repository | Laboratory | Runtime | Neither |
|-----------|-----------|-----------|---------|---------|
| **Enables runtime governance** | ❌ | ❌ | ✅ | ❌ |
| **Created by investigation** | ❌ | ❌ | ❌ | ✅ (ambiguous) |
| **Part of production software** | ❌ | ❌ | ✅ | ❌ |

**Resolution**: The ECU belongs in **Runtime** (`.kde/runtime/ecu/`) because:
1. Its purpose is runtime governance, not business logic
2. It's part of the KDE framework, not the DNP3 library
3. It was installed as part of Bootstrap, not developed as investigation output

The fact that it was created during an investigation does not change its ownership—**purpose determines location**.

---

## 8. Conflict Analysis

### 8.1 Observed Conflicts

| Conflict | Description | Severity |
|----------|-------------|----------|
| **ECU Location** | ECU (Python) in `.kde/runtime/` vs typical Go repository | Low |
| **Implementation Specs** | IMPs specify implementation but live in Laboratory | Low |
| **Engine Code** | Engine implementations (Python) in `.kde/engines/` | Low |

### 8.2 Why Conflicts Occurred

| Conflict | Explanation |
|----------|-------------|
| **ECU Location** | ECU was installed as part of Bootstrap, not developed as investigation output. Its purpose (runtime governance) determines its location (runtime), not its origin. |
| **Implementation Specs** | IMPs define requirements for Repository code. They live in Laboratory because they are **specifications**, not **implementations**. |
| **Engine Code** | Engines are investigation capabilities, not production code. They belong in `.kde/engines/`, not `pkg/`. |

### 8.3 Architectural Consequences

| Consequence | Severity | Mitigation |
|------------|----------|------------|
| Python in Go repository | Low | ECU is runtime infrastructure, not Go code |
| IMPs in Laboratory | None | Correct separation of concerns |
| Engine implementations in `.kde/` | None | Correct ownership per purpose |

### 8.4 No Fundamental Violations

**Conclusion**: The current architecture does **not** violate KDE principles because:

1. **Repository** owns production code (Go) ✅
2. **Laboratory** owns engineering artifacts ✅
3. **Runtime** owns governance infrastructure ✅
4. **Engines** own reasoning capabilities ✅
5. **Seeds** own immutable principles ✅

The apparent conflicts are **temporal** (things created during investigations) rather than **architectural** (violations of responsibility).

---

## 9. Recommended Repository Architecture

### 9.1 Clear Ownership Model

```
/workspace/project/dnp3/
│
├── .kde/                          ← RUNTIME OWNED (KDE Runtime)
│   ├── bootstrap/                  ← Bootstrap procedures
│   ├── runtime/                    ← Runtime infrastructure
│   │   └── ecu/                   ← Runtime ECU (orchestration)
│   ├── engines/                   ← Investigation capabilities
│   ├── seeds/                      ← Immutable reasoning DNA
│   ├── governance/                  ← Runtime policies
│   ├── laboratory/                 ← Laboratory artifacts
│   └── templates/                   ← Artifact templates
│
├── pkg/                            ← REPOSITORY OWNED (Project)
│   └── dnp3/                       ← DNP3 protocol implementation
│
├── cmd/                            ← REPOSITORY OWNED (Project)
│
├── internal/                       ← REPOSITORY OWNED (Project)
│
├── docs/                           ← REPOSITORY OWNED (User docs)
│
├── laboratory/                     ← RUNTIME OWNED (Engineering)
│   ├── investigations/             ← Investigation artifacts
│   ├── experiments/                 ← Experiment artifacts
│   ├── decisions/                  ← Decision records
│   ├── evidence/                   ← Evidence artifacts
│   ├── reviews/                    ← Review artifacts
│   ├── planning/                   ← Planning artifacts
│   ├── implementations/            ← IMPs (specs, not code)
│   └── testing/                    ← Shared testing infrastructure
│
├── pkg/, cmd/, internal/           ← REPOSITORY OWNED (Production)
├── test/, examples/, benchmarks/    ← REPOSITORY OWNED (Support)
└── README.md, Makefile, go.mod     ← REPOSITORY OWNED (Config)
```

### 9.2 Decision Rules

| If... | Then belongs to... |
|-------|-------------------|
| Code is executed by users | Repository |
| Code enables runtime governance | Runtime |
| Artifact is investigation output | Laboratory |
| Artifact defines requirements | Laboratory |
| Artifact is immutable reasoning | Seeds |
| Artifact enables investigations | Engines |

---

## 10. Recommended Laboratory Architecture

### 10.1 Clear Scope

The Laboratory owns **all engineering artifacts** that are not production code:

```
laboratory/
│
├── investigations/         ← PRIMARY: Systematic inquiries
│   └── KDE-INV-XXX/      ← Investigation artifacts
│       ├── README.md      ← Investigation document
│       ├── SPEC.md        ← Investigation specification
│       └── CONCLUSION.md  ← Investigation conclusions
│
├── experiments/           ← PRIMARY: Hypothesis testing
│   └── PROJECT-EXP-XXX/  ← Experiment artifacts
│
├── decisions/             ← PRIMARY: Decision records
│   └── TDR-XXX.md        ← Technology Decision Records
│
├── evidence/               ← SUPPORT: Empirical data
│
├── reviews/               ← SUPPORT: Human reviews
│   └── PROJECT-REV-XXX.md
│
├── planning/              ← SUPPORT: Future work
│
├── implementations/       ← SUPPORT: Implementation specs
│   └── PROJECT-IMP-XXX/  ← Implementation Specifications
│
└── testing/               ← SUPPORT: Test infrastructure
    └── (shared mocks, fixtures, etc.)
```

### 10.2 Boundary Enforcement

| Artifact Type | Location | Owner |
|--------------|----------|-------|
| Investigation conclusions | laboratory/investigations/ | Laboratory |
| IMPs (requirements) | laboratory/implementations/ | Laboratory |
| Production code | pkg/, cmd/, internal/ | Repository |
| User documentation | docs/ | Repository |
| Engine specifications | .kde/engines/ | Runtime |
| Engine implementations | .kde/engines/*/src/ | Runtime |

---

## 11. Final Architectural Recommendation

### 11.1 No Fundamental Changes Required

The current architecture is **fundamentally correct**. The apparent conflicts are:

1. **Temporal** (created during investigations) not **architectural** (violations)
2. **Peripheral** (ECU in runtime) not **core** (mission-critical)
3. **Manageable** (clear rules exist) not **unsolvable** (fundamental contradictions)

### 11.2 Clarifications Needed

| Issue | Resolution |
|-------|-----------|
| **ECU belongs in Runtime** | Purpose determines location, not origin |
| **IMPs are specs, not code** | They define requirements; Repository implements |
| **Engines in .kde/** | Engines are investigation capabilities, not production |

### 11.3 Boundary Rule

> **When in doubt, ask: "What is this artifact's primary purpose?"**
> 
> - If purpose is **runtime governance** → Runtime
> - If purpose is **investigation process** → Laboratory  
> - If purpose is **production software** → Repository

### 11.4 Anti-Pattern to Avoid

Do **NOT** move investigation outputs to Repository simply because they were created during investigations. The origin (investigation) does not determine ownership (Laboratory).

### 11.5 Summary

| Component | Owner | Clear? | Overlapping? |
|-----------|-------|--------|------------|
| Repository | Project | ✅ | No |
| Laboratory | KDE Runtime | ✅ | No |
| Runtime | KDE Runtime | ✅ | No |
| Engines | KDE Runtime | ✅ | No |
| Seeds | KDE Runtime | ✅ | No |
| Bootstrap | KDE Runtime | ✅ | No |

**Conclusion**: The Repository and Laboratory have **non-overlapping responsibilities**. The ECU's location in `.kde/runtime/ecu/` is correct because its purpose is runtime governance, regardless of when or how it was created.

---

## 12. ECU Location Analysis

### 12.1 Question

Why is the Runtime ECU in `.kde/runtime/ecu/` rather than in the Repository?

### 12.2 Analysis

| Criterion | In Repository | In Runtime | In Laboratory |
|-----------|---------------|------------|---------------|
| **Executable by DNP3 users** | No | No | No |
| **Part of DNP3 protocol** | No | No | No |
| **Enables KDE governance** | No | **YES** | No |
| **Investigation output** | No | No | Yes (temporal) |
| **Part of KDE framework** | No | **YES** | No |

### 12.3 Decision

The ECU belongs in **Runtime** (`.kde/runtime/ecu/`) because:

1. **Purpose**: Enables runtime governance, not business logic
2. **Framework**: Part of KDE Runtime, not DNP3 library
3. **Ownership**: Runtime owns its own components
4. **Origin**: Installed via Bootstrap, not developed as investigation output

The fact that the ECU was exercised during investigations (LAB-ECU-001, LAB-ECU-003, LAB-ECU-OBS-001) does not make it a Laboratory artifact. **Purpose determines location.**

---

## 13. Response to Investigation Questions

### Q1: Is there a conflict between Repository and Laboratory?

**A**: No fundamental conflict exists. They serve different purposes:
- Repository → Production software
- Laboratory → Engineering process

### Q2: Is this an architectural flaw or intentional?

**A**: Intentional. The separation of concerns is deliberate:
- Laboratory owns the process (investigations, experiments)
- Repository owns the product (Go code, documentation)

### Q3: Where should implementations belong?

**A**: Implementation **code** belongs in Repository; implementation **specifications** belong in Laboratory.

### Q4: Where should the ECU belong?

**A**: Runtime (`.kde/runtime/ecu/`). Its purpose is governance, not product.

---

## 14. Conclusion

### 14.1 Findings

1. **Repository** and **Laboratory** have clear, non-overlapping responsibilities
2. The apparent conflict is **temporal** (created during investigations) not **architectural**
3. The ECU correctly belongs in **Runtime** because its **purpose** is governance
4. No fundamental architectural changes are required

### 14.2 Recommendations

| Recommendation | Priority | Rationale |
|----------------|-----------|-----------|
| **Document ECU ownership** | Medium | Clarify that ECU belongs to Runtime |
| **Enforce boundary rules** | Medium | Purpose determines location, not origin |
| **Update .kde/README.md** | Low | Add clarification about artifact ownership |

### 14.3 Final Answer

The Repository and Laboratory responsibilities are **clearly defined and non-overlapping**. The Runtime ECU's location in `.kde/runtime/ecu/` is correct. No architectural changes are required.

**The architecture is sound.**

---

*Investigation completed by LAB-ARCH-001*
*Runtime ECU verified: Bootstrap SUCCESS, 8 engines, 2 seeds, fully operational*
