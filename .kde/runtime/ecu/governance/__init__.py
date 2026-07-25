"""
Laboratory Governance Module

Implements KDE Laboratory Governance Standard (GOV-LAB-001).
"""

from .id_registry import IDRegistryManager, IDRegistry
from .lifecycle import LifecycleManager, ArtifactStatus
from .validation import ValidationManager, Violation, ViolationType, ViolationResponse
from .metadata import MetadataManager, ArtifactMetadata
from .integration import GovernanceIntegration, GovernanceResult

__all__ = [
    # ID Registry
    'IDRegistryManager',
    'IDRegistry',
    
    # Lifecycle
    'LifecycleManager',
    'ArtifactStatus',
    
    # Validation
    'ValidationManager',
    'Violation',
    'ViolationType',
    'ViolationResponse',
    
    # Metadata
    'MetadataManager',
    'ArtifactMetadata',
    
    # Integration
    'GovernanceIntegration',
    'GovernanceResult'
]
