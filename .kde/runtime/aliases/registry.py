"""
Alias Registry Module

Provides alias management functionality for KDE Runtime.
Supports canonical commands, categorized aliases, and resolution.
"""

import json
import os
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional, Set, Any
from functools import lru_cache


class AliasCategory(Enum):
    """Alias categories with priority ordering."""
    CANONICAL = 1
    OPERATIONAL = 2
    PROFESSIONAL = 3
    FRIENDLY = 4
    DEPRECATED = 5


@dataclass
class AliasEntry:
    """Represents a single alias entry."""
    alias: str
    canonical: str
    category: AliasCategory
    version: str
    description: str
    namespace: str
    examples: List[str] = field(default_factory=list)
    deprecated: bool = False
    deprecation_date: Optional[str] = None
    approved_by: str = "human"
    approval_date: str = ""
    rationale: str = ""
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'AliasEntry':
        """Create AliasEntry from dictionary."""
        category_str = data.get('category', 'friendly')
        try:
            category = AliasCategory[category_str.upper()]
        except KeyError:
            category = AliasCategory.FRIENDLY
        
        return cls(
            alias=data['alias'],
            canonical=data['canonical'],
            category=category,
            version=data.get('version', '1.0.0'),
            description=data.get('description', ''),
            namespace=data.get('namespace', ''),
            examples=data.get('examples', []),
            deprecated=data.get('deprecated', False),
            deprecation_date=data.get('deprecation_date'),
            approved_by=data.get('approved_by', 'human'),
            approval_date=data.get('approval_date', ''),
            rationale=data.get('rationale', '')
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            'alias': self.alias,
            'canonical': self.canonical,
            'category': self.category.name.lower(),
            'version': self.version,
            'description': self.description,
            'namespace': self.namespace,
            'examples': self.examples,
            'deprecated': self.deprecated,
            'deprecation_date': self.deprecation_date,
            'approved_by': self.approved_by,
            'approval_date': self.approval_date,
            'rationale': self.rationale
        }


class AliasRegistry:
    """
    Alias Registry for KDE Runtime.
    
    Manages canonical commands and human-friendly aliases with
    support for categorization, deprecation, and resolution.
    """
    
    VERSION = "1.0.0"
    
    def __init__(self, registry_path: Optional[str] = None):
        """
        Initialize the Alias Registry.
        
        Args:
            registry_path: Path to registry JSON file. 
                          Defaults to /runtime/aliases/registry.json
        """
        if registry_path is None:
            registry_path = os.path.join(
                os.path.dirname(__file__),
                'registry.json'
            )
        self.registry_path = registry_path
        self._entries: Dict[str, AliasEntry] = {}
        self._canonical_to_aliases: Dict[str, List[str]] = {}
        self._namespace_index: Dict[str, Set[str]] = {}
        self._category_index: Dict[AliasCategory, Set[str]] = {}
        self._loaded = False
    
    def load(self) -> bool:
        """
        Load registry from JSON file.
        
        Returns:
            True if loaded successfully, False otherwise
        """
        try:
            with open(self.registry_path, 'r') as f:
                data = json.load(f)
            
            # Initialize category index
            for cat in AliasCategory:
                self._category_index[cat] = set()
            
            # Load entries
            for entry_data in data.get('aliases', []):
                entry = AliasEntry.from_dict(entry_data)
                self._entries[entry.alias.lower()] = entry
                
                # Update canonical index
                if entry.canonical not in self._canonical_to_aliases:
                    self._canonical_to_aliases[entry.canonical] = []
                self._canonical_to_aliases[entry.canonical].append(entry.alias.lower())
                
                # Update namespace index
                if entry.namespace:
                    if entry.namespace not in self._namespace_index:
                        self._namespace_index[entry.namespace] = set()
                    self._namespace_index[entry.namespace].add(entry.alias.lower())
                
                # Update category index
                self._category_index[entry.category].add(entry.alias.lower())
            
            self._loaded = True
            return True
            
        except Exception as e:
            print(f"Failed to load registry: {e}")
            return False
    
    @property
    def is_loaded(self) -> bool:
        """Check if registry is loaded."""
        return self._loaded
    
    def resolve(self, alias: str) -> Optional[AliasEntry]:
        """
        Resolve an alias to its canonical entry.
        
        Args:
            alias: Alias or canonical command to resolve
            
        Returns:
            AliasEntry if found, None otherwise
        """
        if not self._loaded:
            self.load()
        
        alias_lower = alias.lower()
        
        # Check if it's an alias
        if alias_lower in self._entries:
            return self._entries[alias_lower]
        
        # Check if it's a canonical command
        for entry in self._entries.values():
            if entry.canonical.lower() == alias_lower and entry.category == AliasCategory.CANONICAL:
                return entry
        
        return None
    
    def resolve_to_canonical(self, alias: str) -> Optional[str]:
        """
        Resolve an alias to its canonical command name.
        
        Args:
            alias: Alias to resolve
            
        Returns:
            Canonical command name if found, None otherwise
        """
        entry = self.resolve(alias)
        if entry:
            # Find the canonical version
            if entry.category == AliasCategory.CANONICAL:
                return entry.alias
            # Find the canonical for this alias
            for e in self._entries.values():
                if e.alias == entry.alias and e.category == AliasCategory.CANONICAL:
                    return e.alias
            # Return the alias itself if no separate canonical
            return entry.canonical
        return None
    
    def get_all_aliases(self) -> List[AliasEntry]:
        """Get all alias entries."""
        if not self._loaded:
            self.load()
        return list(self._entries.values())
    
    def get_by_category(self, category: AliasCategory) -> List[AliasEntry]:
        """Get aliases by category."""
        if not self._loaded:
            self.load()
        return [
            self._entries[alias] 
            for alias in self._category_index.get(category, set())
            if alias in self._entries
        ]
    
    def get_by_namespace(self, namespace: str) -> List[AliasEntry]:
        """Get aliases by namespace."""
        if not self._loaded:
            self.load()
        return [
            self._entries[alias]
            for alias in self._namespace_index.get(namespace, set())
            if alias in self._entries
        ]
    
    def get_canonical_aliases(self, canonical: str) -> List[AliasEntry]:
        """Get all aliases for a canonical command."""
        if not self._loaded:
            self.load()
        canonical_lower = canonical.lower()
        return [
            self._entries[alias]
            for alias in self._canonical_to_aliases.get(canonical, [])
            if alias in self._entries
        ]
    
    def get_friendly_aliases(self) -> List[AliasEntry]:
        """Get all friendly aliases for new users."""
        return self.get_by_category(AliasCategory.FRIENDLY)
    
    def get_operational_aliases(self) -> List[AliasEntry]:
        """Get all operational aliases."""
        return self.get_by_category(AliasCategory.OPERATIONAL)
    
    def get_professional_aliases(self) -> List[AliasEntry]:
        """Get all professional aliases."""
        return self.get_by_category(AliasCategory.PROFESSIONAL)
    
    def get_deprecated_aliases(self) -> List[AliasEntry]:
        """Get all deprecated aliases."""
        return self.get_by_category(AliasCategory.DEPRECATED)
    
    def suggest(self, prefix: str, limit: int = 5) -> List[AliasEntry]:
        """
        Suggest aliases matching a prefix.
        
        Args:
            prefix: Prefix to match
            limit: Maximum number of suggestions
            
        Returns:
            List of matching aliases
        """
        if not self._loaded:
            self.load()
        
        prefix_lower = prefix.lower()
        matches = [
            entry for entry in self._entries.values()
            if entry.alias.lower().startswith(prefix_lower)
        ]
        
        # Sort by priority (canonical first, then by category priority)
        matches.sort(key=lambda e: (e.category.value, e.alias))
        
        return matches[:limit]
    
    def get_stats(self) -> Dict[str, Any]:
        """Get registry statistics."""
        if not self._loaded:
            self.load()
        
        return {
            'total_aliases': len(self._entries),
            'canonical_commands': len(self._canonical_to_aliases),
            'by_category': {
                cat.name.lower(): len(aliases)
                for cat, aliases in self._category_index.items()
            },
            'namespaces': list(self._namespace_index.keys()),
            'deprecated_count': len([
                e for e in self._entries.values() if e.deprecated
            ])
        }
    
    def validate(self) -> Dict[str, Any]:
        """
        Validate the registry for conflicts and issues.
        
        Returns:
            Validation report with errors and warnings
        """
        if not self._loaded:
            self.load()
        
        errors = []
        warnings = []
        
        # Check for duplicate aliases in same category
        for cat in AliasCategory:
            aliases = list(self._category_index.get(cat, set()))
            # This is expected - no action needed
        
        # Check for deprecated aliases without deprecation date
        for entry in self._entries.values():
            if entry.deprecated and not entry.deprecation_date:
                warnings.append(
                    f"Alias '{entry.alias}' is deprecated but has no deprecation date"
                )
        
        # Check for aliases with no canonical mapping
        for entry in self._entries.values():
            if entry.category != AliasCategory.CANONICAL:
                # Verify canonical exists
                canonical_found = any(
                    e.canonical == entry.canonical and e.category == AliasCategory.CANONICAL
                    for e in self._entries.values()
                )
                if not canonical_found:
                    warnings.append(
                        f"Alias '{entry.alias}' has canonical '{entry.canonical}' "
                        "but no canonical entry exists"
                    )
        
        return {
            'valid': len(errors) == 0,
            'errors': errors,
            'warnings': warnings
        }


# Global registry instance
_global_registry: Optional[AliasRegistry] = None


def get_registry() -> AliasRegistry:
    """Get the global alias registry instance."""
    global _global_registry
    if _global_registry is None:
        _global_registry = AliasRegistry()
        _global_registry.load()
    return _global_registry


def reset_registry() -> None:
    """Reset the global registry (for testing)."""
    global _global_registry
    _global_registry = None
