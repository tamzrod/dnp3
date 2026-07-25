"""
Lifecycle Manager for Laboratory Governance

Manages artifact lifecycle states and transitions.
"""

from enum import Enum
from typing import Dict, List, Optional, Set
from dataclasses import dataclass, field
from datetime import datetime


class ArtifactStatus(Enum):
    """Artifact lifecycle statuses."""
    DRAFT = "draft"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    APPROVED = "approved"
    LOCKED = "locked"
    ARCHIVED = "archived"
    SUPERSEDED = "superseded"


class LifecycleManager:
    """Manages artifact lifecycle states and transitions."""
    
    # Valid state transitions
    TRANSITIONS: Dict[ArtifactStatus, Set[ArtifactStatus]] = {
        ArtifactStatus.DRAFT: {ArtifactStatus.IN_PROGRESS, ArtifactStatus.COMPLETED},
        ArtifactStatus.IN_PROGRESS: {ArtifactStatus.COMPLETED, ArtifactStatus.DRAFT},
        ArtifactStatus.COMPLETED: {ArtifactStatus.APPROVED, ArtifactStatus.LOCKED},
        ArtifactStatus.APPROVED: {ArtifactStatus.LOCKED},
        ArtifactStatus.LOCKED: {ArtifactStatus.ARCHIVED, ArtifactStatus.SUPERSEDED},
        ArtifactStatus.ARCHIVED: {ArtifactStatus.SUPERSEDED},
        ArtifactStatus.SUPERSEDED: set()  # Terminal state
    }
    
    # Required timestamps per status
    REQUIRED_TIMESTAMPS: Dict[ArtifactStatus, List[str]] = {
        ArtifactStatus.DRAFT: ['created'],
        ArtifactStatus.IN_PROGRESS: ['created', 'started'],
        ArtifactStatus.COMPLETED: ['created', 'started', 'completed'],
        ArtifactStatus.APPROVED: ['created', 'started', 'completed', 'reviewed'],
        ArtifactStatus.LOCKED: ['created', 'started', 'completed', 'locked'],
        ArtifactStatus.ARCHIVED: ['created', 'started', 'completed', 'locked', 'archived'],
        ArtifactStatus.SUPERSEDED: ['created', 'started', 'completed', 'superseded']
    }
    
    def __init__(self):
        self.transitions = self.TRANSITIONS
        self.required_timestamps = self.REQUIRED_TIMESTAMPS
    
    def can_transition(self, from_status: ArtifactStatus, to_status: ArtifactStatus) -> bool:
        """
        Check if transition is valid.
        
        Args:
            from_status: Current status
            to_status: Target status
        
        Returns:
            True if transition is valid
        """
        if from_status not in self.transitions:
            return False
        return to_status in self.transitions[from_status]
    
    def get_valid_transitions(self, from_status: ArtifactStatus) -> List[ArtifactStatus]:
        """
        Get valid transitions from current status.
        
        Args:
            from_status: Current status
        
        Returns:
            List of valid target statuses
        """
        return list(self.transitions.get(from_status, []))
    
    def transition(self, current_status: str, target_status: str) -> tuple:
        """
        Attempt to transition artifact status.
        
        Args:
            current_status: Current status string
            target_status: Target status string
        
        Returns:
            Tuple of (success, message, new_status)
        """
        try:
            current = ArtifactStatus(current_status.lower())
            target = ArtifactStatus(target_status.lower())
        except ValueError:
            return False, f"Invalid status: {current_status} or {target_status}", None
        
        if self.can_transition(current, target):
            return True, f"Valid transition: {current.value} -> {target.value}", target.value
        else:
            valid = self.get_valid_transitions(current)
            valid_str = ', '.join([s.value for s in valid]) if valid else 'none'
            return False, f"Invalid transition: {current.value} -> {target.value}. Valid: {valid_str}", None
    
    def get_required_timestamps(self, status: str) -> List[str]:
        """
        Get required timestamps for status.
        
        Args:
            status: Status string
        
        Returns:
            List of required timestamp fields
        """
        try:
            status_enum = ArtifactStatus(status.lower())
            return self.required_timestamps.get(status_enum, [])
        except ValueError:
            return []
    
    def is_terminal(self, status: str) -> bool:
        """
        Check if status is terminal (no further transitions).
        
        Args:
            status: Status string
        
        Returns:
            True if terminal
        """
        try:
            status_enum = ArtifactStatus(status.lower())
            return len(self.transitions.get(status_enum, set())) == 0
        except ValueError:
            return False
    
    def is_locked(self, status: str) -> bool:
        """
        Check if status represents locked artifact.
        
        Args:
            status: Status string
        
        Returns:
            True if locked
        """
        try:
            status_enum = ArtifactStatus(status.lower())
            return status_enum in {ArtifactStatus.LOCKED, ArtifactStatus.ARCHIVED, ArtifactStatus.SUPERSEDED}
        except ValueError:
            return False
    
    def auto_lock_on_complete(self, status: str) -> Optional[str]:
        """
        Auto-lock if completing an artifact.
        
        Args:
            status: Current status
        
        Returns:
            New status if auto-lock applies, None otherwise
        """
        if status.lower() == 'completed':
            return 'locked'
        return None
    
    def get_lifecycle_summary(self) -> Dict:
        """Get lifecycle configuration summary."""
        return {
            'statuses': [s.value for s in ArtifactStatus],
            'transitions': {s.value: list(t.value for t in targets) 
                          for s, targets in self.transitions.items()},
            'terminal_states': [s.value for s in ArtifactStatus 
                              if len(self.transitions.get(s, set())) == 0],
            'locked_states': ['locked', 'archived', 'superseded']
        }
