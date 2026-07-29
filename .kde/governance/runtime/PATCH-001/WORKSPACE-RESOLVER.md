# PATCH-001: Runtime Workspace Resolution

**Patch ID**: PATCH-001  
**Title**: Add Workspace Resolution to Runtime Orchestrator  
**Type**: Runtime Enhancement  
**Status**: IMPLEMENTED (2026-07-21)  
**Date**: 2026-07-21  
**Author**: KDE Runtime Team  
**Based on**: LAB-020 Validation Results  

---

## Executive Summary

**Problem**: LAB-020 confirmed that the Runtime performs operation classification, but workspace resolution does not exist. This causes projects to be created in the current working directory instead of the appropriate workspace.

**Solution**: Add a Workspace Resolver component to the Runtime Orchestrator that maps classified task types to appropriate workspace locations.

**Impact**: Low - Non-breaking change that adds new capability without modifying existing behavior.

---

## Background

### LAB-020 Findings

| Finding | Evidence |
|---------|----------|
| Operation Classification exists | `TaskClassifier.classify()` in orchestrator |
| Skill Selection uses classification | `TASK_SKILL_MAP` in orchestrator |
| Knowledge Retrieval uses classification | `KnowledgeOnDemandRuntime` initialized |
| **Workspace Resolution missing** | **No workspace logic in orchestrator** |

### Current Orchestration Flow

```
User Request
     │
     ▼
┌─────────────────────────────────────┐
│ RuntimeOrchestrator.orchestrate()      │
└─────────────────────────────────────┘
     │
     ▼
Step 1: TaskClassifier.classify()     ← CLASSIFICATION ✓
     │
     ▼
Step 2: SkillSelector.select()         ← SKILLS ✓
     │
     ▼
Step 3: KnowledgeOnDemandRuntime        ← KNOWLEDGE ✓
     │
     ▼
Step 4: Execute
     │
     ▼
Workspace: $PWD (current directory)    ← MISSING ✗
```

---

## Architecture Impact Analysis

### Components Affected

| Component | Change | Risk |
|----------|--------|------|
| `orchestrator/__init__.py` | Add `WorkspaceResolver` class | Low |
| `RuntimeOrchestrator.orchestrate()` | Add Step 1.5 | Low |
| `OrchestrationResult` | Add `workspace` field | Low |
| Existing behavior | Unchanged for GENERAL tasks | None |

### Components Unaffected

| Component | Reason |
|-----------|--------|
| `TaskClassifier` | No changes needed |
| `SkillSelector` | No changes needed |
| `KnowledgeOnDemandRuntime` | No changes needed |
| Engine | No changes needed |

---

## Proposed Architecture

### New Component: WorkspaceResolver

```
/workspace/project/kde/runtime/orchestrator/
├── __init__.py              # (modified)
├── classifier.py             # (new file)
├── selector.py              # (existing)
├── workspace.py             # (new file)
└── types.py                # (new file)
```

### TASK_WORKSPACE_MAP

```python
# Workspace locations for each task type
TASK_WORKSPACE_MAP = {
    TaskType.INVESTIGATION: {
        "base": "/laboratory/investigations/",
        "pattern": "{id}/",
        "example": "/laboratory/investigations/INV-001/"
    },
    TaskType.EXPERIMENT: {
        "base": "/laboratory/experiments/",
        "pattern": "{id}/",
        "example": "/laboratory/experiments/LAB-001/"
    },
    TaskType.VALIDATION: {
        "base": "/laboratory/validations/",
        "pattern": "{id}/",
        "example": "/laboratory/validations/LAB-001/"
    },
    TaskType.GOVERNANCE_REVIEW: {
        "base": "/governance/",
        "pattern": "{filename}",
        "example": "/governance/PROPOSAL-001.md"
    },
    TaskType.KNOWLEDGE_CREATION: {
        "base": "/knowledge/",
        "pattern": "{filename}",
        "example": "/knowledge/NEW-DEFINITION.md"
    },
    TaskType.SKILL_CREATION: {
        "base": "/runtime/skills/",
        "pattern": "skill-{name}/",
        "example": "/runtime/skills/skill-new-skill/"
    },
    TaskType.RUNTIME_DEVELOPMENT: {
        "base": "/runtime/",
        "pattern": "{component}/",
        "example": "/runtime/orchestrator/"
    },
    TaskType.ARCHITECTURE_DESIGN: {
        "base": "/knowledge/",
        "pattern": "ARCH-{id}.md",
        "example": "/knowledge/ARCH-001.md"
    },
    TaskType.DOCUMENTATION: {
        "base": "/docs/",
        "pattern": "{filename}",
        "example": "/docs/README.md"
    },
    TaskType.FRONTEND_DEVELOPMENT: {
        "base": "/playground/{inv_id}/frontend/",
        "pattern": "{project}/",
        "example": "/playground/INV-001/frontend/"
    },
    TaskType.BACKEND_DEVELOPMENT: {
        "base": "/playground/{inv_id}/backend/",
        "pattern": "{project}/",
        "example": "/playground/INV-001/backend/"
    },
    TaskType.TESTING: {
        "base": "/playground/{inv_id}/tests/",
        "pattern": "{project}/",
        "example": "/playground/INV-001/tests/"
    },
    TaskType.GENERAL: {
        "base": None,  # Use current directory
        "pattern": None,
        "example": None
    },
}
```

---

## Sequence Diagram

### Current Flow

```
Actor          Runtime                    Orchestrator
  │                │                           │
  │─Request───────>│                           │
  │                │                           │
  │                │─classify()────────────────>│
  │                │<────────Classification─────│
  │                │                           │
  │                │─select()─────────────────>│
  │                │<────────SkillSelection─────│
  │                │                           │
  │                │─initialize()──────────────│
  │                │                           │
  │<───────────────│                           │
  │─Result(wrong workspace)                    │
```

### Proposed Flow

```
Actor          Runtime                    Orchestrator            WorkspaceResolver
  │                │                           │                       │
  │─Request───────>│                           │                       │
  │                │                           │                       │
  │                │─classify()───────────────────────────────────────>│
  │                │<────────Classification─────────────────────────────│
  │                │                           │                       │
  │                │─resolve_workspace()──────>│                       │
  │                │<────────WorkspaceInfo─────│                       │
  │                │                           │                       │
  │                │─select()─────────────────>│                       │
  │                │<────────SkillSelection────│                       │
  │                │                           │                       │
  │                │─initialize()──────────────│                       │
  │                │                           │                       │
  │<───────────────│                           │                       │
  │─Result(correct workspace)                   │                       │
```

---

## Source Code Modifications

### 1. New File: `runtime/orchestrator/workspace.py`

```python
"""
Runtime Workspace Resolution

Maps classified task types to appropriate workspace locations.
"""

from dataclasses import dataclass
from typing import Optional, Dict, Any
from enum import Enum


@dataclass
class WorkspaceInfo:
    """Information about the resolved workspace."""
    base_path: Optional[str]
    resolved_path: str
    exists: bool
    task_type: str
    task_id: Optional[str] = None
    requires_id: bool = False


class WorkspaceResolver:
    """
    Resolves workspace locations based on classified task type.
    """
    
    TASK_WORKSPACE_MAP = {
        "investigation": {
            "base": "/laboratory/investigations/",
            "pattern": "{id}/",
            "requires_id": True,
            "description": "Investigation artifacts and evidence"
        },
        "experiment": {
            "base": "/laboratory/experiments/",
            "pattern": "{id}/",
            "requires_id": True,
            "description": "Experimental runs and validations"
        },
        "validation": {
            "base": "/laboratory/validations/",
            "pattern": "{id}/",
            "requires_id": True,
            "description": "Validation runs and test results"
        },
        "governance_review": {
            "base": "/governance/",
            "pattern": "{filename}",
            "requires_id": False,
            "description": "Governance documents and proposals"
        },
        "knowledge_creation": {
            "base": "/knowledge/",
            "pattern": "{filename}",
            "requires_id": False,
            "description": "Promoted knowledge definitions"
        },
        "knowledge_revision": {
            "base": "/knowledge/",
            "pattern": "{filename}",
            "requires_id": False,
            "description": "Knowledge revisions and updates"
        },
        "skill_creation": {
            "base": "/runtime/skills/",
            "pattern": "skill-{name}/",
            "requires_id": False,
            "description": "New skill implementations"
        },
        "runtime_development": {
            "base": "/runtime/",
            "pattern": "{component}/",
            "requires_id": False,
            "description": "Runtime component development"
        },
        "architecture_design": {
            "base": "/knowledge/",
            "pattern": "ARCH-{id}.md",
            "requires_id": True,
            "description": "Architecture design documents"
        },
        "documentation": {
            "base": "/docs/",
            "pattern": "{filename}",
            "requires_id": False,
            "description": "Documentation files"
        },
        "frontend_development": {
            "base": "/playground/{inv_id}/frontend/",
            "pattern": "{project}/",
            "requires_id": True,
            "description": "Frontend implementation"
        },
        "backend_development": {
            "base": "/playground/{inv_id}/backend/",
            "pattern": "{project}/",
            "requires_id": True,
            "description": "Backend implementation"
        },
        "testing": {
            "base": "/playground/{inv_id}/tests/",
            "pattern": "{project}/",
            "requires_id": True,
            "description": "Test implementations"
        },
        "general": {
            "base": None,  # Use current directory
            "pattern": None,
            "requires_id": False,
            "description": "General tasks use current directory"
        },
    }
    
    def __init__(self, workspace_root: str = "/workspace/project/kde"):
        self.workspace_root = workspace_root
    
    def resolve(
        self,
        task_type: str,
        task_id: Optional[str] = None,
        filename: Optional[str] = None,
        project_name: Optional[str] = None
    ) -> WorkspaceInfo:
        """
        Resolve workspace for a given task type.
        
        Args:
            task_type: The classified task type
            task_id: Optional task ID (INV-XXX, LAB-XXX, etc.)
            filename: Optional filename for file-based tasks
            project_name: Optional project name for development tasks
            
        Returns:
            WorkspaceInfo with resolved path and metadata
        """
        config = self.TASK_WORKSPACE_MAP.get(
            task_type,
            self.TASK_WORKSPACE_MAP["general"]
        )
        
        base_path = config["base"]
        requires_id = config.get("requires_id", False)
        
        # Handle GENERAL tasks - use current directory
        if base_path is None:
            return WorkspaceInfo(
                base_path=None,
                resolved_path="current_directory",
                exists=True,
                task_type=task_type,
                task_id=None,
                requires_id=False
            )
        
        # Build resolved path
        resolved_path = base_path
        
        if task_id:
            resolved_path = resolved_path.format(id=task_id)
        
        # Apply pattern substitution
        pattern = config.get("pattern", "")
        if pattern and task_id:
            resolved_path += pattern.format(
                id=task_id,
                filename=filename or "",
                name=project_name or "",
                component=project_name or ""
            )
        
        # Full path
        full_path = f"{self.workspace_root}{resolved_path}"
        
        return WorkspaceInfo(
            base_path=f"{self.workspace_root}{base_path}",
            resolved_path=resolved_path,
            exists=False,  # Checked separately if needed
            task_type=task_type,
            task_id=task_id,
            requires_id=requires_id
        )
    
    def check_exists(self, workspace_info: WorkspaceInfo) -> bool:
        """Check if the resolved workspace exists."""
        if workspace_info.base_path is None:
            return True  # Current directory always exists
        import os
        full_path = f"{self.workspace_root}{workspace_info.resolved_path}"
        return os.path.exists(full_path)
```

### 2. Modified File: `runtime/orchestrator/__init__.py`

```python
# Add to imports section
from .workspace import WorkspaceResolver, WorkspaceInfo

# Add to RuntimeOrchestrator.__init__
class RuntimeOrchestrator:
    def __init__(self):
        self.classifier = TaskClassifier()
        self.selector = SkillSelector()
        self.workspace_resolver = WorkspaceResolver()  # NEW
        self.orchestration_count = 0

    def orchestrate(
        self,
        request: str,
        investigation_id: str = None
    ) -> OrchestrationResult:
        self.orchestration_count += 1

        # Step 1: Classify
        classification = self.classifier.classify(request)

        # Step 1.5: Resolve workspace (NEW)
        workspace_info = self.workspace_resolver.resolve(
            task_type=classification.task_type.value,
            task_id=classification.keywords.get("id"),  # Extract ID from keywords
        )

        # Step 2: Select skills
        skill_selection = self.selector.select(classification.task_type)

        # ... rest unchanged ...

        return OrchestrationResult(
            timestamp=datetime.utcnow().isoformat() + "Z",
            request=request[:100] + "..." if len(request) > 100 else request,
            classification=classification,
            skill_selection=skill_selection,
            workspace=workspace_info,  # NEW field
            loaded_skills_count=loaded_count,
            knowledge_retrieved_count=knowledge_count,
            context_size=context_size,
            execution_ready=True
        )
```

### 3. New File: `runtime/orchestrator/types.py`

```python
"""
Orchestrator Data Types

Defines data structures used by the orchestrator.
"""

from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any
from datetime import datetime
from enum import Enum


@dataclass
class WorkspaceInfo:
    """Information about the resolved workspace."""
    base_path: Optional[str]
    resolved_path: str
    exists: bool
    task_type: str
    task_id: Optional[str] = None
    requires_id: bool = False
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "base_path": self.base_path,
            "resolved_path": self.resolved_path,
            "exists": self.exists,
            "task_type": self.task_type,
            "task_id": self.task_id,
            "requires_id": self.requires_id
        }


@dataclass
class ClassificationResult:
    """Result of task classification."""
    task_type: str
    confidence: float
    matched_patterns: List[str]
    keywords: List[str]
    classification_reason: str


@dataclass
class SkillSelectionResult:
    """Result of skill selection."""
    task_type: str
    skills: List[str]
    skills_with_dependencies: List[str]


@dataclass
class OrchestrationResult:
    """Complete orchestration result."""
    timestamp: str
    request: str
    classification: ClassificationResult
    skill_selection: SkillSelectionResult
    workspace: WorkspaceInfo  # NEW field
    loaded_skills_count: int
    knowledge_retrieved_count: int
    context_size: int
    execution_ready: bool
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "timestamp": self.timestamp,
            "request": self.request,
            "classification": {
                "task_type": self.classification.task_type,
                "confidence": self.classification.confidence,
                "matched_patterns": self.classification.matched_patterns,
                "keywords": self.classification.keywords,
                "reason": self.classification.classification_reason
            },
            "skill_selection": {
                "task_type": self.skill_selection.task_type,
                "skills": self.skill_selection.skills,
                "skills_with_dependencies": self.skill_selection.skills_with_dependencies
            },
            "workspace": self.workspace.to_dict(),  # NEW
            "loaded_skills_count": self.loaded_skills_count,
            "knowledge_retrieved_count": self.knowledge_retrieved_count,
            "context_size": self.context_size,
            "execution_ready": self.execution_ready
        }
```

---

## Migration Notes

### Deployment Steps

1. **Backup existing orchestrator**
   ```bash
   cp runtime/orchestrator/__init__.py runtime/orchestrator/__init__.py.bak
   ```

2. **Add new files**
   - `runtime/orchestrator/workspace.py` (new)
   - `runtime/orchestrator/types.py` (new)

3. **Modify existing file**
   - `runtime/orchestrator/__init__.py` (add import, modify class)

4. **Verify syntax**
   ```bash
   python3 -m py_compile runtime/orchestrator/workspace.py
   python3 -m py_compile runtime/orchestrator/types.py
   python3 -m py_compile runtime/orchestrator/__init__.py
   ```

5. **Test import**
   ```bash
   python3 -c "from runtime.orchestrator import WorkspaceResolver; print('OK')"
   ```

### Rollback Procedure

1. Remove new files
   ```bash
   rm runtime/orchestrator/workspace.py
   rm runtime/orchestrator/types.py
   ```

2. Restore backup
   ```bash
   mv runtime/orchestrator/__init__.py.bak runtime/orchestrator/__init__.py
   ```

3. Verify
   ```bash
   python3 -c "from runtime.orchestrator import RuntimeOrchestrator; print('OK')"
   ```

### Configuration Changes

| Item | Change | Default |
|------|--------|---------|
| Workspace root | Added | `/workspace/project/kde` |
| TASK_WORKSPACE_MAP | Added | See above |
| WorkspaceInfo | Added | N/A |

---

## Validation Plan

### Unit Tests

| Test | Description | Expected Result |
|------|-------------|-----------------|
| test_resolve_investigation | Resolve workspace for INVESTIGATION | `/laboratory/investigations/{id}/` |
| test_resolve_experiment | Resolve workspace for EXPERIMENT | `/laboratory/experiments/{id}/` |
| test_resolve_general | Resolve workspace for GENERAL | `current_directory` |
| test_resolve_with_id | Resolve with explicit ID | Path contains ID |
| test_resolve_no_id_needed | Resolve when ID not needed | Path without ID placeholder |
| test_check_exists | Check workspace exists | Boolean result |

### Integration Tests

| Test | Description | Expected Result |
|------|-------------|-----------------|
| test_orchestrate_with_workspace | Full orchestration | Workspace in result |
| test_classification_to_workspace | Classification → workspace mapping | Correct task_type → path |
| test_backward_compatibility | GENERAL task | Same as before |

### Manual Validation

1. **Test 1**: Request "Conduct Investigation INV-999"
   - Expected: Workspace = `/laboratory/investigations/INV-999/`

2. **Test 2**: Request "Create new skill"
   - Expected: Workspace = `/runtime/skills/skill-new/`

3. **Test 3**: Request "Fix bug in UI"
   - Expected: Classification = GENERAL, Workspace = `current_directory`

### Validation Commands

```bash
# Run unit tests
cd /workspace/project/kde
python3 -m pytest runtime/orchestrator/test_workspace.py -v

# Run integration tests
python3 -m pytest runtime/orchestrator/test_integration.py -v

# Manual validation
python3 << 'EOF'
from runtime.orchestrator import RuntimeOrchestrator

orchestrator = RuntimeOrchestrator()

# Test INVESTIGATION
result = orchestrator.orchestrate("Conduct Investigation INV-001")
print(f"Investigation workspace: {result.workspace.resolved_path}")

# Test GENERAL
result = orchestrator.orchestrate("Fix the bug in main.py")
print(f"General workspace: {result.workspace.resolved_path}")
EOF
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing code | Low | High | Only adds new field, doesn't modify existing |
| Incorrect path resolution | Medium | Medium | Unit tests + validation plan |
| Import errors | Low | Medium | Syntax check before deployment |
| Workspace not existing | Low | Low | check_exists() method available |

---

## Compatibility Matrix

| Component | Before | After | Compatible |
|-----------|--------|-------|------------|
| TaskClassifier | ✓ | ✓ | Yes |
| SkillSelector | ✓ | ✓ | Yes |
| KnowledgeOnDemandRuntime | ✓ | ✓ | Yes |
| Engine | ✓ | ✓ | Yes |
| OrchestrationResult consumers | ? | ✓ | Check needed |

---

## Alternatives Considered

### Alternative 1: Hard-code in Orchestrator

**Pro**: Simple, no new files
**Con**: Violates single responsibility, hard to maintain

**Rejected**: Not architecture-aligned

### Alternative 2: External Config File

**Pro**: No code changes, config-driven
**Con**: Runtime must read config, adds dependency

**Rejected**: Increases complexity without benefit

### Alternative 3: Knowledge-based Resolution

**Pro**: Flexible, uses existing knowledge system
**Con**: Over-engineering, adds latency

**Rejected**: Unnecessary complexity for static mapping

---

## Recommendation

**APPROVE** implementation of PATCH-001.

### Rationale

1. **Low risk**: Non-breaking change, adds new capability
2. **Architecture-aligned**: Follows existing patterns (TaskClassifier, SkillSelector)
3. **Testable**: Clear validation plan provided
4. **Solves the problem**: Addresses LAB-020 findings

### Next Steps (if approved)

1. Create branch: `patch/workspace-resolver`
2. Implement code changes
3. Write tests
4. Submit for code review
5. Deploy to staging
6. Run validation
7. Merge to main

---

## Approval Required

| Role | Required | Status |
|------|----------|--------|
| Governance | Yes | ✅ APPROVED |
| Architecture | Yes | ✅ APPROVED |
| Runtime Owner | Yes | ✅ APPROVED |

## Implementation Complete

Implemented in commit `8d6d17b`:
- `runtime/orchestrator/workspace.py` (WorkspaceResolver)
- `runtime/orchestrator/types.py` (WorkspaceInfo)
- `runtime/orchestrator/__init__.py` (Modified)

---

*Generated by KDE Runtime under PATCH-001*  
*Based on LAB-020 Validation Results*
