"""
Laboratory Governance Module

Implements KDE Laboratory Governance Standard (GOV-LAB-001).
Includes Subject-Based Routing (KDE-INV-049).
"""

from .id_registry import IDRegistryManager, IDRegistry
from .lifecycle import LifecycleManager, ArtifactStatus
from .validation import ValidationManager, Violation, ViolationType, ViolationResponse
from .metadata import MetadataManager, ArtifactMetadata
from .integration import GovernanceIntegration, GovernanceResult
from .subject import SubjectClassifier, SubjectRouter, Subject, SubjectClassification

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
    
    # Subject Classification (KDE-INV-049)
    'SubjectClassifier',
    'SubjectRouter',
    'Subject',
    'SubjectClassification',
    
    # Integration
    'GovernanceIntegration',
    'GovernanceResult'
]
