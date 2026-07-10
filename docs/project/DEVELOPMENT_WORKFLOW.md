---
title: "Development Workflow"
layer: 4-project
---

# Development Workflow

## Purpose

This document defines the workflow for contributing to this repository. It establishes the steps required to move from idea to implementation.

## The Four-Phase Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                   PHASE 1: KNOWLEDGE                         │
│        Understand the protocol before designing              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   PHASE 2: ARCHITECTURE                     │
│          Design the system before implementing              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   PHASE 3: IMPLEMENTATION                    │
│         Implement according to architecture                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   PHASE 4: VALIDATION                       │
│            Test according to architecture                  │
└─────────────────────────────────────────────────────────────┘
```

## Phase 1: Knowledge

### Activities

1. **Read protocol specification**
2. **Analyze protocol behavior**
3. **Document understanding**
4. **Identify requirements**
5. **Review with team**

### Outputs

- Updates to protocol knowledge base
- Understanding of feature scope
- Identification of edge cases
- Clarification of ambiguities

### Gate

Must complete before Phase 2.

### Gate Criteria

- [ ] Protocol behavior understood
- [ ] Requirements identified
- [ ] Knowledge documented
- [ ] Team alignment achieved

## Phase 2: Architecture

### Activities

1. **Design component architecture**
2. **Define interfaces**
3. **Document decisions**
4. **Create ADRs**
5. **Review architecture**

### Outputs

- Architecture documents
- ADR records
- Interface specifications
- Design diagrams

### Gate

Must complete before Phase 3.

### Gate Criteria

- [ ] Architecture documented
- [ ] ADRs created for significant decisions
- [ ] Architecture reviewed
- [ ] Team approval obtained

## Phase 3: Implementation

### Activities

1. **Implement per architecture**
2. **Write unit tests**
3. **Follow code standards**
4. **Document inline**
5. **Self-review**

### Outputs

- Source code
- Unit tests
- Documentation updates
- Traceability links

### Gate

Must complete before Phase 4.

### Gate Criteria

- [ ] Implementation complete
- [ ] Code passes review
- [ ] Tests written
- [ ] Documentation updated

## Phase 4: Validation

### Activities

1. **Run test suite**
2. **Verify coverage**
3. **Validate against architecture**
4. **Document results**
5. **Merge to main**

### Outputs

- Test results
- Coverage report
- Validation summary
- Merged code

### Gate

Must complete for feature completion.

### Gate Criteria

- [ ] All tests pass
- [ ] Coverage meets target
- [ ] Validation complete
- [ ] PR merged

## The Change Request Process

### For Protocol Knowledge Changes

```
1. Propose change to protocol knowledge
2. Review against specification
3. Update documentation
4. Validate with team
5. Merge
```

### For Architecture Changes

```
1. Propose architecture change
2. Review against protocol knowledge
3. Update architecture documents
4. Update affected ADRs
5. Implement changes
6. Update tests
7. Validate and merge
```

### For Implementation Changes

```
1. Review architecture
2. Implement changes
3. Write/update tests
4. Update documentation
5. Submit PR
6. Review and merge
```

## The Review Process

### Self-Review Checklist

Before requesting review:

- [ ] Code follows style guidelines
- [ ] Tests are written
- [ ] Architecture references are documented
- [ ] Documentation is updated
- [ ] Traceability is maintained
- [ ] No debug code committed

### Peer Review Checklist

During review:

- [ ] Implementation follows architecture
- [ ] Tests are comprehensive
- [ ] Traceability is documented
- [ ] Documentation is complete
- [ ] No layer violations
- [ ] Design rationale is clear

### Architecture Review

For significant changes:

1. **Design review** before implementation
2. **Architecture approval** before coding
3. **Implementation review** before merge
4. **Validation review** before release

## Traceability Requirements

### Feature Traceability

Every feature must have:

```markdown
## Feature: [Name]

## Protocol Knowledge
- [Link to relevant protocol document]

## Architecture
- [Link to relevant architecture document]

## ADR
- [Link to relevant ADR, if applicable]

## Implementation
- [Link to source code]

## Tests
- [Link to test code]
```

### Change Traceability

Every change must reference:

- What changed
- Why it changed
- What layer governed the change
- What other layers are affected

## The PR Process

### 1. Branch Creation

```bash
git checkout -b feature/[feature-name]
```

### 2. Development

```bash
# Implement per workflow phases
```

### 3. Commit

```bash
git add .
git commit -m "feat: add [feature]"
```

### 4. Push

```bash
git push origin feature/[feature-name]
```

### 5. PR Creation

Create PR with:

- Feature description
- Protocol references
- Architecture references
- ADR references
- Test results
- Breaking changes (if any)

### 6. Review

- Address review comments
- Update as needed
- Re-request review

### 7. Merge

- Squash commits if needed
- Delete branch after merge
- Verify CI passes

## The CI Process

### Automated Checks

| Check | Purpose |
|-------|---------|
| Lint | Code style compliance |
| Format | Go fmt compliance |
| Test | Unit test execution |
| Coverage | Test coverage threshold |
| Security | Vulnerability scanning |

### Gate Criteria

All CI checks must pass before merge.

## The AI Workflow

When an AI contributes:

### Required Steps

1. **Read protocol knowledge** (docs/protocol/*)
2. **Read architecture** (docs/architecture/*)
3. **Read relevant ADRs** (docs/adr/*)
4. **Verify understanding**
5. **Implement per architecture**
6. **Write tests per architecture**
7. **Document references**

### Prohibited Actions

- Implementing without reading architecture
- Using code as protocol documentation
- Skipping layers
- Making architecture decisions without ADRs
- Implementing differently than documented

## Exception Handling

### Protocol Ambiguity

When protocol knowledge is ambiguous:

1. Document the ambiguity
2. Propose resolution
3. Update knowledge base
4. Continue with architecture

### Architecture Gap

When architecture doesn't cover implementation need:

1. Propose architecture change
2. Document rationale
3. Update architecture
4. Continue with implementation

### Time Constraints

When time pressure suggests skipping steps:

1. Document the exception
2. Complete steps partially
3. Plan to complete later
4. Do not compromise traceability

## Summary

The development workflow ensures:

1. **Knowledge guides design**
2. **Architecture guides implementation**
3. **Traceability is maintained**
4. **Quality is enforced**
5. **Documentation is complete**

Every contribution follows this workflow. No shortcuts.

---

## See Also

- [KNOWLEDGE_FIRST_ENGINEERING.md](KNOWLEDGE_FIRST_ENGINEERING.md)
- [CHAIN_OF_AUTHORITY.md](CHAIN_OF_AUTHORITY.md)
- [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md)
