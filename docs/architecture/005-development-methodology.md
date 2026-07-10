---
title: "005 - Development Methodology"
status: approved
---

# Development Methodology

## Overview

go-dnp3 follows an architecture-first development methodology. This 
document describes our development process and why we follow it.

## The Problem with Code-First Development

Many projects start coding immediately, which leads to:

- Accumulated technical debt
- Architectural drift
- Inconsistent design decisions
- Difficulty maintaining code
- Slow iteration

## Our Solution: Architecture First

We separate architecture from implementation:

```
Research → Architecture → Review → Implementation → Testing
    ↑___________↑___________|           ↓
         iterate           |        validate
                            ↓
                      deploy/release
```

## Development Phases

### Phase 1: Research

**Duration**: Ongoing throughout project

**Activities:**
- Read specification thoroughly
- Analyze protocol behavior
- Identify edge cases
- Research existing implementations
- Document findings

**Deliverables:**
- Research notes in `docs/research/`
- Protocol analysis documents
- Edge case identification
- Implementation constraints

**Exit Criteria:**
- Complete understanding of protocol
- All edge cases documented
- Research reviewed by team

### Phase 2: Architecture

**Duration**: Per feature/subsystem

**Activities:**
- Design solution architecture
- Define interfaces
- Document design decisions
- Create ADRs for significant choices

**Deliverables:**
- Architecture documents
- ADR (Architecture Decision Records)
- Interface definitions
- Design diagrams

**Exit Criteria:**
- Architecture documented
- ADRs written for key decisions
- Architecture reviewed and approved

### Phase 3: Review

**Duration**: As needed

**Activities:**
- Team architecture review
- External expert review
- Address feedback
- Iterate on design

**Deliverables:**
- Reviewed architecture
- Incorporated feedback
- Approved design

**Exit Criteria:**
- All reviewers approve
- No blocking concerns
- Architecture signed off

### Phase 4: Implementation

**Duration**: Per feature

**Activities:**
- Write code per architecture
- Write tests first
- Follow style guidelines
- Document as you go

**Deliverables:**
- Working code
- Unit tests
- Integration tests
- Updated documentation

**Exit Criteria:**
- All tests pass
- Code reviewed
- Documentation complete
- Conformance tests pass

### Phase 5: Conformance Testing

**Duration**: Per release

**Activities:**
- Run conformance tests
- Test interoperability
- Test edge cases
- Document test results

**Deliverables:**
- Test results
- Bug reports
- Known limitations

**Exit Criteria:**
- All critical tests pass
- Interoperability verified
- Known issues documented

### Phase 6: Optimization

**Duration**: Per release (if needed)

**Activities:**
- Profile performance
- Identify bottlenecks
- Optimize selectively
- Measure improvements

**Deliverables:**
- Performance analysis
- Optimizations applied
- Benchmark results

**Exit Criteria:**
- Performance acceptable
- No correctness regressions
- Tests still pass

### Phase 7: Release

**Duration**: Ongoing

**Activities:**
- Prepare release notes
- Tag version
- Update documentation
- Announce release

**Deliverables:**
- Release artifacts
- Release notes
- Updated CHANGELOG

## Architecture Decision Records

### What is an ADR?

An Architecture Decision Record (ADR) documents a significant 
architectural decision.

### ADR Format

```markdown
# ADR-001: Package Structure

## Status
Accepted

## Context
We need to organize the codebase into packages.

## Decision
We will use a layered architecture with internal/ and pkg/ directories.

## Consequences
- Clear separation of concerns
- Implementation details hidden
- Public API constrained
```

### When to Write an ADR

Write an ADR when:

- Making a significant architectural choice
- Choosing between alternatives
- Accepting a non-obvious constraint
- Making a technology selection

### ADR Repository

All ADRs are stored in `docs/adr/`:

```
docs/adr/
├── README.md          # ADR process
├── 001-package-structure.md
├── 002-error-handling.md
├── 003-concurrency-model.md
└── ...
```

## Code Review Process

### Before Submitting

1. Run all tests locally
2. Check code formatting
3. Review your own code
4. Write clear commit messages

### During Review

- Respond to all comments
- Make requested changes
- Re-request review when ready
- Don't merge without approval

### Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests are comprehensive
- [ ] Documentation is updated
- [ ] No obvious bugs
- [ ] Design matches architecture

## Testing Requirements

### Unit Tests

- Every function has tests
- Edge cases covered
- Error paths tested
- Code coverage > 80%

### Integration Tests

- Layer interactions tested
- End-to-end scenarios
- Connection handling
- Error recovery

### Conformance Tests

- Protocol compliance
- Interoperability
- Edge case handling

## Documentation Requirements

### Per Feature

- Update relevant docs
- Add usage examples
- Document limitations

### Per Release

- Update CHANGELOG
- Update README
- Write release notes

## Branching Strategy

### Main Branch

- `main` - production-ready code
- Protected branch
- Requires PR to merge
- Requires tests to pass

### Development Branches

- `feature/*` - new features
- `bugfix/*` - bug fixes
- `docs/*` - documentation
- `research/*` - research documents

## Commit Message Convention

We follow conventional commits:

```
feat: add new feature
fix: fix a bug
docs: update documentation
refactor: refactor code
test: add or update tests
chore: maintenance tasks
research: add research findings
arch: architecture decisions
```

## Release Process

1. Update version
2. Run full test suite
3. Update CHANGELOG
4. Create release tag
5. Build release artifacts
6. Publish documentation
7. Announce release

## Continuous Improvement

We regularly:

- Review and update methodology
- Retrospect on process
- Incorporate feedback
- Improve tooling
