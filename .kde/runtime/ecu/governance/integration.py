"""
Governance Integration for Runtime ECU

Integrates all governance components into Runtime ECU.
"""

import os
import json
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass

from .id_registry import IDRegistryManager
from .lifecycle import LifecycleManager, ArtifactStatus
from .validation import ValidationManager, Violation
from .metadata import MetadataManager, ArtifactMetadata


@dataclass
class GovernanceResult:
    """Result of governance operation."""
    success: bool
    message: str
    data: Dict = None
    violations: List[Violation] = None
    
    def __post_init__(self):
        if self.data is None:
            self.data = {}
        if self.violations is None:
            self.violations = []


class GovernanceIntegration:
    """
    Integrates all governance components.
    
    This is the main interface for Runtime ECU to interact with
    Laboratory Governance (GOV-LAB-001).
    """
    
    def __init__(self, laboratory_path: str, project_name: str, runtime_version: str, bootstrap_version: str):
        self.laboratory_path = laboratory_path
        self.project_name = project_name
        
        # Initialize managers
        self.id_registry = IDRegistryManager(laboratory_path)
        self.lifecycle = LifecycleManager()
        self.validation = ValidationManager(laboratory_path)
        self.metadata = MetadataManager(project_name, runtime_version, bootstrap_version)
        
        # Governance state
        self.governance_dir = os.path.join(laboratory_path, '.governance')
        self.state_file = os.path.join(self.governance_dir, 'governance_state.json')
        self._load_state()
    
    def _load_state(self):
        """Load governance state."""
        if os.path.exists(self.state_file):
            with open(self.state_file, 'r') as f:
                self.state = json.load(f)
        else:
            self.state = {
                'governance_standard': 'GOV-LAB-001',
                'version': '1.0.0',
                'adopted': '2026-07-25',
                'operations_count': 0,
                'violations_detected': 0,
                'artifacts_locked': 0
            }
    
    def _save_state(self):
        """Save governance state."""
        os.makedirs(self.governance_dir, exist_ok=True)
        with open(self.state_file, 'w') as f:
            json.dump(self.state, f, indent=2)
    
    def classify_and_prepare(self, description: str, title: str, is_kde: bool = False) -> Tuple[str, str, ArtifactMetadata]:
        """
        Classify operation and prepare artifact metadata.
        
        Args:
            description: Operation description
            title: Human-readable title
            is_kde: Whether this is KDE framework investigation
        
        Returns:
            Tuple of (classification, artifact_id, initial_metadata)
        """
        classification, artifact_id = self.id_registry.classify_and_id(description, is_kde)
        metadata = self.metadata.generate_initial(artifact_id, classification, title)
        
        self.state['operations_count'] += 1
        self._save_state()
        
        return classification, artifact_id, metadata
    
    def validate_artifact(self, artifact_path: str, metadata: ArtifactMetadata) -> List[Violation]:
        """
        Validate artifact for governance compliance.
        
        Args:
            artifact_path: Path to artifact
            metadata: Artifact metadata
        
        Returns:
            List of violations (empty if valid)
        """
        return self.validation.detect_violations(artifact_path, metadata.to_dict())
    
    def transition_status(self, current_status: str, target_status: str) -> Tuple[bool, str]:
        """
        Validate and transition artifact status.
        
        Args:
            current_status: Current status
            target_status: Target status
        
        Returns:
            Tuple of (success, message)
        """
        success, message, _ = self.lifecycle.transition(current_status, target_status)
        return success, message
    
    def complete_and_lock(self, metadata: ArtifactMetadata) -> ArtifactMetadata:
        """
        Complete artifact and auto-lock.
        
        Args:
            metadata: Current metadata
        
        Returns:
            Updated metadata with completed and locked status
        """
        metadata = self.metadata.update_completion(metadata)
        self.state['artifacts_locked'] += 1
        self._save_state()
        return metadata
    
    def get_workspace_path(self, artifact_id: str, artifact_type: str) -> str:
        """
        Get deterministic workspace path for artifact.
        
        Args:
            artifact_id: Artifact ID
            artifact_type: Artifact type
        
        Returns:
            Full path for artifact
        """
        directory = self.id_registry.get_directory(artifact_type)
        artifact_dir = os.path.join(self.laboratory_path, directory, artifact_id)
        return artifact_dir
    
    def detect_violations(self, artifacts: List[Dict]) -> List[Violation]:
        """
        Detect violations across multiple artifacts.
        
        Args:
            artifacts: List of artifact info dicts with 'path' and 'metadata'
        
        Returns:
            List of all violations found
        """
        all_violations = []
        
        for artifact in artifacts:
            path = artifact.get('path', '')
            metadata = artifact.get('metadata', {})
            violations = self.validation.detect_violations(path, metadata)
            all_violations.extend(violations)
        
        self.state['violations_detected'] += len(all_violations)
        self._save_state()
        
        return all_violations
    
    def generate_report(self) -> Dict:
        """
        Generate governance status report.
        
        Returns:
            Governance status dictionary
        """
        id_status = self.id_registry.get_status()
        lifecycle_summary = self.lifecycle.get_lifecycle_summary()
        validation_summary = self.validation.get_violation_summary()
        
        return {
            'governance_standard': self.state['governance_standard'],
            'version': self.state['version'],
            'adopted': self.state['adopted'],
            'statistics': {
                'operations_count': self.state['operations_count'],
                'violations_detected': self.state['violations_detected'],
                'artifacts_locked': self.state['artifacts_locked']
            },
            'next_ids': id_status['next_ids'],
            'lifecycle': lifecycle_summary,
            'validation': validation_summary
        }
    
    def migrate_existing_artifacts(self) -> Dict:
        """
        Scan and validate existing laboratory artifacts.
        
        Returns:
            Migration report
        """
        artifacts_checked = 0
        violations_found = 0
        artifacts_needing_review = []
        
        # Scan all artifact directories
        for artifact_type, directory in [
            ('investigation', 'investigations/'),
            ('experiment', 'experiments/'),
            ('implementation', 'implementations/'),
            ('decision', 'decisions/'),
            ('review', 'reviews/'),
            ('planning', 'planning/'),
            ('evidence', 'evidence/')
        ]:
            dir_path = os.path.join(self.laboratory_path, directory)
            if not os.path.exists(dir_path):
                continue
            
            for item in os.listdir(dir_path):
                item_path = os.path.join(dir_path, item)
                if os.path.isdir(item_path):
                    # Check for markdown files
                    for filename in os.listdir(item_path):
                        if filename.endswith('.md'):
                            full_path = os.path.join(item_path, filename)
                            metadata = self.metadata.parse_artifact(full_path)
                            if metadata:
                                artifacts_checked += 1
                                violations = self.validation.detect_violations(full_path, metadata.to_dict())
                                violations_found += len(violations)
                                if violations:
                                    artifacts_needing_review.append({
                                        'path': full_path,
                                        'violations': [(v.violation_type.value, v.message) for v in violations]
                                    })
        
        return {
            'artifacts_checked': artifacts_checked,
            'violations_found': violations_found,
            'artifacts_needing_review': artifacts_needing_review
        }
