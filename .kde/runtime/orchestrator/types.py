"""
Orchestrator Data Types

Defines data structures used by the orchestrator.
"""

from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any
from datetime import datetime


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


@dataclass
class ClassificationResult:
    """Result of task classification."""
    task_type: str
    confidence: float
    matched_patterns: List[str]
    keywords: List[str]
    classification_reason: str
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize to dictionary."""
        return {
            "task_type": self.task_type,
            "confidence": self.confidence,
            "matched_patterns": self.matched_patterns,
            "keywords": self.keywords,
            "reason": self.classification_reason
        }


@dataclass
class SkillSelectionResult:
    """Result of skill selection."""
    task_type: str
    skills: List[str]
    skills_with_dependencies: List[str]
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize to dictionary."""
        return {
            "task_type": self.task_type,
            "skills": self.skills,
            "skills_with_dependencies": self.skills_with_dependencies
        }


@dataclass
class OrchestrationResult:
    """Complete orchestration result."""
    timestamp: str
    request: str
    classification: ClassificationResult
    skill_selection: SkillSelectionResult
    workspace: WorkspaceInfo
    loaded_skills_count: int
    knowledge_retrieved_count: int
    context_size: int
    execution_ready: bool
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize to dictionary."""
        return {
            "timestamp": self.timestamp,
            "request": self.request,
            "classification": self.classification.to_dict(),
            "skill_selection": self.skill_selection.to_dict(),
            "workspace": self.workspace.to_dict(),
            "loaded_skills_count": self.loaded_skills_count,
            "knowledge_retrieved_count": self.knowledge_retrieved_count,
            "context_size": self.context_size,
            "execution_ready": self.execution_ready
        }
