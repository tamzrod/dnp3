"""
Runtime Workspace Resolution

Maps classified task types to appropriate workspace locations.
"""

import os
from dataclasses import dataclass
from typing import Optional, Dict, Any


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
        """Serialize to dictionary."""
        return {
            "base_path": self.base_path,
            "resolved_path": self.resolved_path,
            "exists": self.exists,
            "task_type": self.task_type,
            "task_id": self.task_id,
            "requires_id": self.requires_id
        }


class WorkspaceResolver:
    """
    Resolves workspace locations based on classified task type.
    
    Maps task types to appropriate workspace paths following KDE
    directory conventions.
    """
    
    # Workspace locations for each task type
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
            "name_field": "project_name",
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
            "base": None,
            "pattern": None,
            "requires_id": False,
            "description": "General tasks use current directory"
        },
    }
    
    def __init__(self, workspace_root: str = "/workspace/project/kde"):
        """
        Initialize WorkspaceResolver.
        
        Args:
            workspace_root: Root path of the KDE workspace
        """
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
            task_type: The classified task type (e.g., "investigation")
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
        
        # Handle {inv_id} placeholder for playground tasks
        if task_id and "{inv_id}" in resolved_path:
            resolved_path = resolved_path.replace("{inv_id}", task_id)
        elif task_id:
            resolved_path = resolved_path.format(id=task_id)
        
        # Handle pattern substitution
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
        
        # Check if path exists
        full_path = f"{self.workspace_root}{resolved_path}"
        exists = os.path.exists(full_path)
        
        return WorkspaceInfo(
            base_path=f"{self.workspace_root}{base_path}" if base_path else None,
            resolved_path=resolved_path,
            exists=exists,
            task_type=task_type,
            task_id=task_id,
            requires_id=requires_id
        )
    
    def check_exists(self, workspace_info: WorkspaceInfo) -> bool:
        """
        Check if the resolved workspace exists.
        
        Args:
            workspace_info: The workspace info to check
            
        Returns:
            True if the workspace exists, False otherwise
        """
        if workspace_info.base_path is None:
            return True  # Current directory always exists
        
        full_path = f"{self.workspace_root}{workspace_info.resolved_path}"
        return os.path.exists(full_path)
