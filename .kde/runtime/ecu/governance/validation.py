"""
Validation Manager for Laboratory Governance

Detects and responds to governance violations.
"""

import re
import os
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum


class ViolationType(Enum):
    """Types of governance violations."""
    INCORRECT_FOLDER = "incorrect_folder"
    INVALID_TYPE = "invalid_type"
    DUPLICATE_ID = "duplicate_id"
    INVALID_NAMING = "invalid_naming"
    ORPHAN = "orphan"
    INCOMPLETE_METADATA = "incomplete_metadata"
    INVALID_LIFECYCLE = "invalid_lifecycle"
    UNLOCKED_COMPLETED = "unlocked_completed"
    MANUAL_FOLDER = "manual_folder"
    TIMESTAMP_INCONSISTENCY = "timestamp_inconsistency"


class ViolationResponse(Enum):
    """Response actions for violations."""
    REJECT = "reject"
    WARN = "warn"
    QUARANTINE = "quarantine"
    MOVE = "move"
    ARCHIVE = "archive"
    AUTO_LOCK = "auto_lock"


@dataclass
class Violation:
    """Represents a governance violation."""
    violation_type: ViolationType
    artifact_id: str
    message: str
    severity: str  # 'critical', 'high', 'medium', 'low'
    response: ViolationResponse
    details: Dict = field(default_factory=dict)


class ValidationManager:
    """Manages governance validation and violation detection."""
    
    # Naming patterns for each type
    NAMING_PATTERNS = {
        'investigation_kde': r'^KDE-INV-\d{3}$',
        'investigation_project': r'^(PROJECT|DNP3)-INV-\d{3}$',
        'experiment': r'^(PROJECT|DNP3)-EXP-\d{3}$',
        'implementation': r'^(PROJECT|DNP3)-IMP-\d{3}$',
        'decision': r'^TDR-\d{3}$',
        'review': r'^(PROJECT|DNP3)-REV-\d{3}$',
        'planning': r'^PLAN-\d{3}$',
        'evidence': r'^EVD-\d{3}$',
        'testing': r'^TEST-\d{3}$'
    }
    
    # Patterns for files/directories to EXCLUDE from validation (infrastructure)
    EXCLUSION_PATTERNS = [
        'readme.md',       # README files (infrastructure documentation)
        'catalog.md',       # Catalog files (indexes)
        '.governance',      # Governance infrastructure directory
        'testing/governance/',  # Testing governance infrastructure
        '__pycache__/',     # Python cache
        '.git/',           # Git metadata
        '.kde/'            # KDE runtime
    ]
    
    # Prefix to directory mapping
    PREFIX_TO_DIRECTORY = {
        'KDE-INV': 'investigations/',
        'PROJECT-INV': 'investigations/',
        'DNP3-INV': 'investigations/',
        'PROJECT-EXP': 'experiments/',
        'DNP3-EXP': 'experiments/',
        'PROJECT-IMP': 'implementations/',
        'DNP3-IMP': 'implementations/',
        'TDR': 'decisions/',
        'PROJECT-REV': 'reviews/',
        'DNP3-REV': 'reviews/',
        'PLAN': 'planning/',
        'EVD': 'evidence/',
        'TEST': 'testing/'
    }
    
    # Mandatory metadata fields
    MANDATORY_METADATA = ['id', 'type', 'title', 'status', 'created']
    
    # Response matrix
    RESPONSE_MATRIX = {
        ViolationType.INCORRECT_FOLDER: ViolationResponse.MOVE,
        ViolationType.INVALID_TYPE: ViolationResponse.REJECT,
        ViolationType.DUPLICATE_ID: ViolationResponse.REJECT,
        ViolationType.INVALID_NAMING: ViolationResponse.REJECT,
        ViolationType.ORPHAN: ViolationResponse.WARN,
        ViolationType.INCOMPLETE_METADATA: ViolationResponse.REJECT,
        ViolationType.INVALID_LIFECYCLE: ViolationResponse.QUARANTINE,
        ViolationType.UNLOCKED_COMPLETED: ViolationResponse.AUTO_LOCK,
        ViolationType.MANUAL_FOLDER: ViolationResponse.WARN,
        ViolationType.TIMESTAMP_INCONSISTENCY: ViolationResponse.REJECT
    }
    
    def __init__(self, laboratory_path: str):
        self.laboratory_path = laboratory_path
        self.patterns = self.NAMING_PATTERNS
        self.response_matrix = self.RESPONSE_MATRIX
        self.exclusion_patterns = self.EXCLUSION_PATTERNS
    
    def is_excluded(self, path: str) -> bool:
        """
        Check if path should be excluded from validation.
        
        Args:
            path: Path to check
        
        Returns:
            True if should be excluded
        """
        path_lower = path.lower()
        for pattern in self.exclusion_patterns:
            if pattern.lower() in path_lower:
                return True
        return False
    
    def validate_naming(self, artifact_id: str) -> Tuple[bool, Optional[str]]:
        """
        Validate artifact naming convention.
        
        Args:
            artifact_id: Artifact ID to validate
        
        Returns:
            Tuple of (valid, error_message)
        """
        for pattern_name, pattern in self.patterns.items():
            if re.match(pattern, artifact_id):
                return True, None
        
        return False, f"Invalid naming format: {artifact_id}"
    
    def validate_placement(self, artifact_id: str, current_path: str) -> Tuple[bool, str, str]:
        """
        Validate artifact is in correct directory.
        
        Args:
            artifact_id: Artifact ID
            current_path: Current artifact path
        
        Returns:
            Tuple of (valid, correct_path, error_message)
        """
        # Find correct directory based on prefix
        correct_dir = None
        for prefix, directory in self.PREFIX_TO_DIRECTORY.items():
            if artifact_id.startswith(prefix):
                correct_dir = directory
                break
        
        if not correct_dir:
            return False, "", f"Unknown prefix for: {artifact_id}"
        
        # Check if in correct directory
        expected_path = os.path.join(self.laboratory_path, correct_dir)
        actual_dir = os.path.dirname(current_path).replace(self.laboratory_path + '/', '')
        
        if not current_path.startswith(os.path.join(self.laboratory_path, correct_dir)):
            return False, expected_path, f"Artifact in wrong directory: {actual_dir} (expected: {correct_dir})"
        
        return True, expected_path, ""
    
    def validate_metadata(self, metadata: Dict) -> Tuple[bool, List[str]]:
        """
        Validate artifact metadata.
        
        Args:
            metadata: Artifact metadata dict
        
        Returns:
            Tuple of (valid, missing_fields)
        """
        missing = []
        for field in self.MANDATORY_METADATA:
            if field not in metadata or not metadata[field]:
                missing.append(field)
        
        return len(missing) == 0, missing
    
    def validate_lifecycle(self, status: str, is_completed: bool) -> Tuple[bool, Optional[str]]:
        """
        Validate lifecycle state consistency.
        
        Args:
            status: Current status
            is_completed: Whether work is completed
        
        Returns:
            Tuple of (valid, error_message)
        """
        status_lower = status.lower()
        
        # Completed artifacts should be locked
        if is_completed and status_lower not in ['locked', 'archived', 'superseded']:
            return False, f"Completed artifact not locked: {status}"
        
        return True, None
    
    def detect_violations(self, artifact_path: str, metadata: Dict) -> List[Violation]:
        """
        Detect all violations for an artifact.
        
        Args:
            artifact_path: Path to artifact
            metadata: Artifact metadata
        
        Returns:
            List of detected violations
        """
        # Skip excluded paths
        if self.is_excluded(artifact_path):
            return []
        
        violations = []
        artifact_id = metadata.get('id', os.path.basename(artifact_path).replace('.md', ''))
        
        # Check naming
        valid, error = self.validate_naming(artifact_id)
        if not valid:
            violations.append(Violation(
                violation_type=ViolationType.INVALID_NAMING,
                artifact_id=artifact_id,
                message=error,
                severity='high',
                response=self.response_matrix[ViolationType.INVALID_NAMING],
                details={'path': artifact_path}
            ))
        
        # Check placement
        valid, correct_path, error = self.validate_placement(artifact_id, artifact_path)
        if not valid:
            violations.append(Violation(
                violation_type=ViolationType.INCORRECT_FOLDER,
                artifact_id=artifact_id,
                message=error,
                severity='medium',
                response=self.response_matrix[ViolationType.INCORRECT_FOLDER],
                details={'current': artifact_path, 'correct': correct_path}
            ))
        
        # Check metadata
        valid, missing = self.validate_metadata(metadata)
        if not valid:
            violations.append(Violation(
                violation_type=ViolationType.INCOMPLETE_METADATA,
                artifact_id=artifact_id,
                message=f"Missing metadata: {', '.join(missing)}",
                severity='high',
                response=self.response_matrix[ViolationType.INCOMPLETE_METADATA],
                details={'missing': missing}
            ))
        
        # Check lifecycle
        status = metadata.get('status', '')
        is_completed = metadata.get('status', '').lower() == 'completed'
        valid, error = self.validate_lifecycle(status, is_completed)
        if not valid:
            # Downgrade to LOW for pre-governance artifacts (completed but not locked)
            # This is expected behavior for existing artifacts
            severity = 'low' if is_completed else 'medium'
            violations.append(Violation(
                violation_type=ViolationType.INVALID_LIFECYCLE,
                artifact_id=artifact_id,
                message=error + " (pre-governance artifact)",
                severity=severity,
                response=self.response_matrix[ViolationType.INVALID_LIFECYCLE],
                details={'status': status, 'note': 'Auto-lock will apply on next update'}
            ))
        
        # Check unlocked completed (only warn, don't require immediate lock)
        # This allows pre-governance artifacts to remain completed until next revision
        if status.lower() == 'completed':
            violations.append(Violation(
                violation_type=ViolationType.UNLOCKED_COMPLETED,
                artifact_id=artifact_id,
                message="Completed artifact should be locked (pre-governance)",
                severity='low',
                response=ViolationResponse.WARN,  # Changed to WARN for existing artifacts
                details={'status': status, 'note': 'Auto-lock will apply on next update'}
            ))
        
        return violations
    
    def get_violation_summary(self) -> Dict:
        """Get validation configuration summary."""
        return {
            'patterns': self.patterns,
            'mandatory_metadata': self.MANDATORY_METADATA,
            'response_matrix': {v.value: r.value for v, r in self.response_matrix.items()}
        }
