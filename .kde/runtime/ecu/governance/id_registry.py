"""
ID Registry for Laboratory Governance

Manages unique identifiers for all laboratory artifacts.
"""

import json
import os
from typing import Dict, Optional, List
from dataclasses import dataclass, asdict


@dataclass
class IDRegistry:
    """Registry for tracking next available ID for each artifact type."""
    investigation_kde: int = 0
    investigation_project: int = 0
    experiment: int = 0
    implementation: int = 0
    decision: int = 0
    review: int = 0
    planning: int = 0
    evidence: int = 0
    testing: int = 0
    
    def to_dict(self) -> Dict:
        return asdict(self)
    
    @classmethod
    def from_dict(cls, data: Dict) -> 'IDRegistry':
        return cls(**data)


class IDRegistryManager:
    """Manages laboratory artifact ID assignment."""
    
    TYPE_PREFIXES = {
        'investigation_kde': 'KDE-INV-',
        'investigation_project': 'PROJECT-INV-',
        'experiment': 'PROJECT-EXP-',
        'implementation': 'PROJECT-IMP-',
        'decision': 'TDR-',
        'review': 'PROJECT-REV-',
        'planning': 'PLAN-',
        'evidence': 'EVD-',
        'testing': 'TEST-'
    }
    
    TYPE_DIRECTORIES = {
        'investigation_kde': 'investigations/',
        'investigation_project': 'investigations/',
        'experiment': 'experiments/',
        'implementation': 'implementations/',
        'decision': 'decisions/',
        'review': 'reviews/',
        'planning': 'planning/',
        'evidence': 'evidence/',
        'testing': 'testing/'
    }
    
    def __init__(self, laboratory_path: str):
        self.laboratory_path = laboratory_path
        self.registry_path = os.path.join(laboratory_path, '.governance', 'id_registry.json')
        self.registry = self._load_or_create()
    
    def _load_or_create(self) -> IDRegistry:
        """Load existing registry or create new one."""
        if os.path.exists(self.registry_path):
            with open(self.registry_path, 'r') as f:
                data = json.load(f)
                return IDRegistry.from_dict(data)
        else:
            # Scan existing artifacts to find highest ID
            registry = IDRegistry()
            self._scan_existing(registry)
            self._save(registry)
            return registry
    
    def _scan_existing(self, registry: IDRegistry):
        """Scan existing artifacts to find highest IDs."""
        lab_path = self.laboratory_path
        
        # Scan investigations
        inv_path = os.path.join(lab_path, 'investigations')
        if os.path.exists(inv_path):
            for item in os.listdir(inv_path):
                if item.startswith('KDE-INV-'):
                    try:
                        num = int(item.split('-')[-1])
                        registry.investigation_kde = max(registry.investigation_kde, num)
                    except ValueError:
                        pass  # Skip non-standard names
                elif item.startswith('PROJECT-INV-') or item.startswith('DNP3-INV-'):
                    try:
                        # Extract number - handle both PROJECT-INV-* and DNP3-INV-*
                        parts = item.split('-')
                        num = int(parts[-1])
                        registry.investigation_project = max(registry.investigation_project, num)
                    except ValueError:
                        pass  # Skip non-standard names
        
        # Scan other directories
        for dirname, attr_name in [
            ('experiments', 'experiment'),
            ('implementations', 'implementation'),
            ('decisions', 'decision'),
            ('reviews', 'review'),
            ('planning', 'planning'),
            ('evidence', 'evidence'),
            ('testing', 'testing')
        ]:
            dir_path = os.path.join(lab_path, dirname)
            if os.path.exists(dir_path):
                for item in os.listdir(dir_path):
                    # Extract number from item name
                    try:
                        parts = item.replace('.md', '').split('-')
                        num = int(parts[-1])
                        current = getattr(registry, attr_name)
                        setattr(registry, attr_name, max(current, num))
                    except (ValueError, IndexError):
                        pass  # Skip non-standard names
    
    def _save(self, registry: IDRegistry):
        """Save registry to disk."""
        os.makedirs(os.path.dirname(self.registry_path), exist_ok=True)
        with open(self.registry_path, 'w') as f:
            json.dump(registry.to_dict(), f, indent=2)
    
    def get_next_id(self, artifact_type: str) -> str:
        """
        Get next available ID for artifact type.
        
        Args:
            artifact_type: Type of artifact (investigation, experiment, etc.)
        
        Returns:
            Next available ID string (e.g., 'KDE-INV-012')
        """
        # Map common types to registry fields
        type_mapping = {
            'investigation': 'investigation_project',  # Default to project investigation
            'investigation_kde': 'investigation_kde',
            'investigation_project': 'investigation_project',
            'experiment': 'experiment',
            'implementation': 'implementation',
            'decision': 'decision',
            'review': 'review',
            'planning': 'planning',
            'evidence': 'evidence',
            'testing': 'testing'
        }
        
        registry_key = type_mapping.get(artifact_type, 'investigation_project')
        prefix = self.TYPE_PREFIXES.get(registry_key, 'PROJECT-INV-')
        
        # Increment counter
        current = getattr(self.registry, registry_key)
        new_id = current + 1
        setattr(self.registry, registry_key, new_id)
        self._save(self.registry)
        
        return f"{prefix}{new_id:03d}"
    
    def classify_and_id(self, description: str, is_kde: bool = False) -> tuple:
        """
        Classify operation and assign ID.
        
        Args:
            description: Operation description
            is_kde: Whether this is a KDE framework investigation
        
        Returns:
            Tuple of (classification, id_string)
        """
        classification = self.classify(description)
        
        if classification == 'investigation':
            if is_kde:
                registry_key = 'investigation_kde'
            else:
                registry_key = 'investigation_project'
        else:
            registry_key = classification
        
        return classification, self.get_next_id(registry_key)
    
    def classify(self, description: str) -> str:
        """
        Classify operation based on keywords.
        
        Args:
            description: Operation description
        
        Returns:
            Classification string
        """
        desc_lower = description.lower()
        
        # Classification rules
        rules = [
            ('investigation', ['investigate', 'analyze', 'assess', 'determine', 'examine', 'investigation']),
            ('experiment', ['experiment', 'test hypothesis', 'validate approach']),
            ('implementation', ['implement', 'build', 'create', 'develop']),
            ('testing', ['test', 'verify', 'conformance', 'testing']),
            ('review', ['review', 'assess quality', 'audit']),
            ('decision', ['decide', 'evaluate options', 'select approach']),
            ('planning', ['plan', 'roadmap', 'schedule']),
            ('evidence', ['evidence', 'observation', 'data collection'])
        ]
        
        for classification, keywords in rules:
            for keyword in keywords:
                if keyword in desc_lower:
                    return classification
        
        # Default to investigation
        return 'investigation'
    
    def get_directory(self, artifact_type: str) -> str:
        """Get directory for artifact type."""
        type_mapping = {
            'investigation': 'investigations/',
            'investigation_kde': 'investigations/',
            'investigation_project': 'investigations/',
            'experiment': 'experiments/',
            'implementation': 'implementations/',
            'decision': 'decisions/',
            'review': 'reviews/',
            'planning': 'planning/',
            'evidence': 'evidence/',
            'testing': 'testing/'
        }
        return type_mapping.get(artifact_type, 'investigations/')
    
    def get_status(self) -> Dict:
        """Get registry status."""
        return {
            'next_ids': {
                'KDE-INV': f"{self.registry.investigation_kde + 1:03d}",
                'PROJECT-INV': f"{self.registry.investigation_project + 1:03d}",
                'PROJECT-EXP': f"{self.registry.experiment + 1:03d}",
                'PROJECT-IMP': f"{self.registry.implementation + 1:03d}",
                'TDR': f"{self.registry.decision + 1:03d}",
                'PROJECT-REV': f"{self.registry.review + 1:03d}",
                'PLAN': f"{self.registry.planning + 1:03d}",
                'EVD': f"{self.registry.evidence + 1:03d}",
                'TEST': f"{self.registry.testing + 1:03d}"
            },
            'registry_path': self.registry_path
        }
