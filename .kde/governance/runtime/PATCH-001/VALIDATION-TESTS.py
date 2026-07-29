#!/usr/bin/env python3
"""
PATCH-001 Validation Tests

Test cases for the WorkspaceResolver component.
Run with: python3 VALIDATION-TESTS.py
"""

import sys
import os

# Add runtime to path
sys.path.insert(0, '/workspace/project/kde')

from dataclasses import dataclass
from typing import Optional, Dict, Any, List


# === Copy of proposed WorkspaceResolver for testing ===

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


class WorkspaceResolver:
    """Resolves workspace locations based on classified task type."""
    
    TASK_WORKSPACE_MAP = {
        "investigation": {
            "base": "/laboratory/investigations/",
            "pattern": "{id}/",
            "requires_id": True,
        },
        "experiment": {
            "base": "/laboratory/experiments/",
            "pattern": "{id}/",
            "requires_id": True,
        },
        "validation": {
            "base": "/laboratory/validations/",
            "pattern": "{id}/",
            "requires_id": True,
        },
        "governance_review": {
            "base": "/governance/",
            "pattern": "{filename}",
            "requires_id": False,
        },
        "knowledge_creation": {
            "base": "/knowledge/",
            "pattern": "{filename}",
            "requires_id": False,
        },
        "skill_creation": {
            "base": "/runtime/skills/",
            "pattern": "skill-{name}/",
            "requires_id": False,
            "name_field": "project_name",  # Map project_name to {name}
        },
        "runtime_development": {
            "base": "/runtime/",
            "pattern": "{component}/",
            "requires_id": False,
        },
        "architecture_design": {
            "base": "/knowledge/",
            "pattern": "ARCH-{id}.md",
            "requires_id": True,
        },
        "documentation": {
            "base": "/docs/",
            "pattern": "{filename}",
            "requires_id": False,
        },
        "frontend_development": {
            "base": "/playground/{inv_id}/frontend/",
            "pattern": "{project}/",
            "requires_id": True,
        },
        "backend_development": {
            "base": "/playground/{inv_id}/backend/",
            "pattern": "{project}/",
            "requires_id": True,
        },
        "testing": {
            "base": "/playground/{inv_id}/tests/",
            "pattern": "{project}/",
            "requires_id": True,
        },
        "general": {
            "base": None,
            "pattern": None,
            "requires_id": False,
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
        config = self.TASK_WORKSPACE_MAP.get(
            task_type,
            self.TASK_WORKSPACE_MAP["general"]
        )
        
        base_path = config["base"]
        requires_id = config.get("requires_id", False)
        
        if base_path is None:
            return WorkspaceInfo(
                base_path=None,
                resolved_path="current_directory",
                exists=True,
                task_type=task_type,
                task_id=None,
                requires_id=False
            )
        
        resolved_path = base_path
        
        if task_id:
            resolved_path = resolved_path.format(id=task_id)
        
        pattern = config.get("pattern", "")
        if pattern:
            # Handle name_field mapping (e.g., project_name → name)
            name_value = project_name or ""
            if config.get("name_field") == "project_name":
                name_value = project_name or ""
            
            resolved_path += pattern.format(
                id=task_id or "",
                filename=filename or "",
                name=name_value,
                component=project_name or ""
            )
        
        return WorkspaceInfo(
            base_path=f"{self.workspace_root}{base_path}",
            resolved_path=resolved_path,
            exists=False,
            task_type=task_type,
            task_id=task_id,
            requires_id=requires_id
        )


# === Test Cases ===

def test_resolve_investigation():
    """Test workspace resolution for INVESTIGATION task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("investigation", task_id="INV-001")
    
    assert result.resolved_path == "/laboratory/investigations/INV-001/"
    assert result.requires_id == True
    print("✓ test_resolve_investigation PASSED")


def test_resolve_experiment():
    """Test workspace resolution for EXPERIMENT task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("experiment", task_id="LAB-001")
    
    assert result.resolved_path == "/laboratory/experiments/LAB-001/"
    assert result.requires_id == True
    print("✓ test_resolve_experiment PASSED")


def test_resolve_validation():
    """Test workspace resolution for VALIDATION task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("validation", task_id="LAB-001")
    
    assert result.resolved_path == "/laboratory/validations/LAB-001/"
    assert result.requires_id == True
    print("✓ test_resolve_validation PASSED")


def test_resolve_general():
    """Test workspace resolution for GENERAL task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("general")
    
    assert result.resolved_path == "current_directory"
    assert result.base_path is None
    print("✓ test_resolve_general PASSED")


def test_resolve_governance():
    """Test workspace resolution for GOVERNANCE task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("governance_review", filename="PROPOSAL.md")
    
    assert "/governance/" in result.resolved_path
    assert result.requires_id == False
    print("✓ test_resolve_governance PASSED")


def test_resolve_knowledge():
    """Test workspace resolution for KNOWLEDGE task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("knowledge_creation", filename="NEW-KNOWLEDGE.md")
    
    assert "/knowledge/" in result.resolved_path
    assert result.requires_id == False
    print("✓ test_resolve_knowledge PASSED")


def test_resolve_skill():
    """Test workspace resolution for SKILL task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("skill_creation", project_name="my-skill")
    
    assert "/runtime/skills/" in result.resolved_path
    assert "skill-my-skill" in result.resolved_path
    print("✓ test_resolve_skill PASSED")


def test_resolve_architecture():
    """Test workspace resolution for ARCHITECTURE task type."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("architecture_design", task_id="001")
    
    assert "/knowledge/" in result.resolved_path
    assert "ARCH-001.md" in result.resolved_path
    print("✓ test_resolve_architecture PASSED")


def test_resolve_unknown_type():
    """Test workspace resolution for unknown task type defaults to GENERAL."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("unknown_type")
    
    assert result.resolved_path == "current_directory"
    print("✓ test_resolve_unknown_type PASSED")


def test_resolve_with_id():
    """Test that ID is correctly included in path."""
    resolver = WorkspaceResolver()
    result = resolver.resolve("investigation", task_id="INV-999")
    
    assert "INV-999" in result.resolved_path
    print("✓ test_resolve_with_id PASSED")


def test_workspace_info_to_dict():
    """Test WorkspaceInfo serialization."""
    info = WorkspaceInfo(
        base_path="/laboratory/investigations/",
        resolved_path="/laboratory/investigations/INV-001/",
        exists=True,
        task_type="investigation",
        task_id="INV-001",
        requires_id=True
    )
    
    result = info.to_dict()
    assert result["resolved_path"] == "/laboratory/investigations/INV-001/"
    assert result["task_type"] == "investigation"
    print("✓ test_workspace_info_to_dict PASSED")


def test_custom_workspace_root():
    """Test custom workspace root."""
    resolver = WorkspaceResolver(workspace_root="/custom/path")
    result = resolver.resolve("investigation", task_id="INV-001")
    
    assert result.base_path == "/custom/path/laboratory/investigations/"
    print("✓ test_custom_workspace_root PASSED")


# === Run Tests ===

def main():
    print("=" * 60)
    print("PATCH-001 Validation Tests")
    print("=" * 60)
    print()
    
    tests = [
        test_resolve_investigation,
        test_resolve_experiment,
        test_resolve_validation,
        test_resolve_general,
        test_resolve_governance,
        test_resolve_knowledge,
        test_resolve_skill,
        test_resolve_architecture,
        test_resolve_unknown_type,
        test_resolve_with_id,
        test_workspace_info_to_dict,
        test_custom_workspace_root,
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            test()
            passed += 1
        except AssertionError as e:
            print(f"✗ {test.__name__} FAILED: {e}")
            failed += 1
        except Exception as e:
            print(f"✗ {test.__name__} ERROR: {e}")
            failed += 1
    
    print()
    print("=" * 60)
    print(f"Results: {passed} passed, {failed} failed")
    print("=" * 60)
    
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
