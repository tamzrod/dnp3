"""
Alias Discovery API

Provides REST-style endpoints for alias discovery and resolution.
"""

import json
import os
import sys
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional, Any
from enum import Enum

# Handle imports for both package and standalone use
if __name__ == 'discovery' or '.' in __name__:
    from .registry import (
        AliasCategory,
        AliasEntry,
        AliasRegistry,
        get_registry
    )
else:
    # Standalone test import
    sys.path.insert(0, os.path.dirname(__file__))
    from registry import (
        AliasCategory,
        AliasEntry,
        AliasRegistry,
        get_registry
    )


class DiscoveryEndpoint(Enum):
    """Discovery API endpoints."""
    LIST_ALL = "list_all"
    LIST_BY_CATEGORY = "list_by_category"
    RESOLVE = "resolve"
    CANONICAL = "canonical"
    SUGGEST = "suggest"
    STATS = "stats"
    VALIDATE = "validate"


@dataclass
class DiscoveryResponse:
    """Standard API response format."""
    success: bool
    endpoint: str
    data: Any = None
    error: Optional[str] = None
    warnings: List[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            'success': self.success,
            'endpoint': self.endpoint,
            'data': self.data,
            'error': self.error,
            'warnings': self.warnings or []
        }


class DiscoveryAPI:
    """
    Discovery API for alias registry.
    
    Provides structured access to alias information through
    well-defined endpoints.
    """
    
    def __init__(self, registry: Optional[AliasRegistry] = None):
        """
        Initialize Discovery API.
        
        Args:
            registry: AliasRegistry instance. Uses global registry if None.
        """
        self.registry = registry or get_registry()
    
    def list_all(
        self,
        category: Optional[str] = None,
        namespace: Optional[str] = None,
        include_deprecated: bool = True
    ) -> DiscoveryResponse:
        """
        List all aliases.
        
        Args:
            category: Filter by category
            namespace: Filter by namespace
            include_deprecated: Include deprecated aliases
            
        Returns:
            DiscoveryResponse with alias list
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            aliases = self.registry.get_all_aliases()
            
            # Filter by category
            if category:
                try:
                    cat = AliasCategory[category.upper()]
                    aliases = [a for a in aliases if a.category == cat]
                except KeyError:
                    return DiscoveryResponse(
                        success=False,
                        endpoint=DiscoveryEndpoint.LIST_ALL.value,
                        error=f"Unknown category: {category}"
                    )
            
            # Filter by namespace
            if namespace:
                aliases = [a for a in aliases if a.namespace == namespace]
            
            # Filter deprecated
            if not include_deprecated:
                aliases = [a for a in aliases if not a.deprecated]
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.LIST_ALL.value,
                data={
                    'aliases': [a.to_dict() for a in aliases],
                    'count': len(aliases)
                }
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.LIST_ALL.value,
                error=str(e)
            )
    
    def list_by_category(
        self,
        category: str
    ) -> DiscoveryResponse:
        """
        List aliases by category.
        
        Args:
            category: Category name (friendly, operational, professional, deprecated)
            
        Returns:
            DiscoveryResponse with alias list
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            try:
                cat = AliasCategory[category.upper()]
            except KeyError:
                return DiscoveryResponse(
                    success=False,
                    endpoint=DiscoveryEndpoint.LIST_BY_CATEGORY.value,
                    error=f"Unknown category: {category}"
                )
            
            aliases = self.registry.get_by_category(cat)
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.LIST_BY_CATEGORY.value,
                data={
                    'category': category.lower(),
                    'aliases': [a.to_dict() for a in aliases],
                    'count': len(aliases)
                }
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.LIST_BY_CATEGORY.value,
                error=str(e)
            )
    
    def resolve(self, alias: str) -> DiscoveryResponse:
        """
        Resolve an alias to its canonical command.
        
        Args:
            alias: Alias to resolve
            
        Returns:
            DiscoveryResponse with resolution result
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            entry = self.registry.resolve(alias)
            
            if entry is None:
                return DiscoveryResponse(
                    success=False,
                    endpoint=DiscoveryEndpoint.RESOLVE.value,
                    error=f"Alias not found: {alias}"
                )
            
            # Find the canonical for this alias
            canonical_entry = None
            for e in self.registry.get_all_aliases():
                if e.canonical == entry.canonical and e.category == AliasCategory.CANONICAL:
                    canonical_entry = e
                    break
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.RESOLVE.value,
                data={
                    'alias': entry.alias,
                    'canonical': canonical_entry.alias if canonical_entry else entry.canonical,
                    'category': entry.category.name.lower(),
                    'description': entry.description,
                    'deprecated': entry.deprecated,
                    'examples': entry.examples
                }
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.RESOLVE.value,
                error=str(e)
            )
    
    def get_canonical(self, canonical: str) -> DiscoveryResponse:
        """
        Get all aliases for a canonical command.
        
        Args:
            canonical: Canonical command name
            
        Returns:
            DiscoveryResponse with all aliases for the command
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            aliases = self.registry.get_canonical_aliases(canonical)
            
            if not aliases:
                return DiscoveryResponse(
                    success=False,
                    endpoint=DiscoveryEndpoint.CANONICAL.value,
                    error=f"Canonical command not found: {canonical}"
                )
            
            # Separate canonical from aliases
            canonical_entry = None
            alias_list = []
            
            for a in aliases:
                if a.category == AliasCategory.CANONICAL:
                    canonical_entry = a
                else:
                    alias_list.append(a)
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.CANONICAL.value,
                data={
                    'canonical': canonical_entry.to_dict() if canonical_entry else None,
                    'aliases': [a.to_dict() for a in alias_list],
                    'total_count': len(aliases)
                }
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.CANONICAL.value,
                error=str(e)
            )
    
    def suggest(
        self,
        prefix: str,
        limit: int = 5,
        category: Optional[str] = None
    ) -> DiscoveryResponse:
        """
        Suggest aliases matching a prefix.
        
        Args:
            prefix: Prefix to match
            limit: Maximum suggestions
            category: Filter by category
            
        Returns:
            DiscoveryResponse with suggestions
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            suggestions = self.registry.suggest(prefix, limit)
            
            # Filter by category if specified
            if category:
                try:
                    cat = AliasCategory[category.upper()]
                    suggestions = [s for s in suggestions if s.category == cat]
                except KeyError:
                    return DiscoveryResponse(
                        success=False,
                        endpoint=DiscoveryEndpoint.SUGGEST.value,
                        error=f"Unknown category: {category}"
                    )
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.SUGGEST.value,
                data={
                    'prefix': prefix,
                    'suggestions': [s.to_dict() for s in suggestions],
                    'count': len(suggestions)
                }
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.SUGGEST.value,
                error=str(e)
            )
    
    def get_stats(self) -> DiscoveryResponse:
        """
        Get registry statistics.
        
        Returns:
            DiscoveryResponse with statistics
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            stats = self.registry.get_stats()
            
            return DiscoveryResponse(
                success=True,
                endpoint=DiscoveryEndpoint.STATS.value,
                data=stats
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.STATS.value,
                error=str(e)
            )
    
    def validate(self) -> DiscoveryResponse:
        """
        Validate the registry.
        
        Returns:
            DiscoveryResponse with validation results
        """
        try:
            if not self.registry.is_loaded:
                self.registry.load()
            
            result = self.registry.validate()
            
            return DiscoveryResponse(
                success=result['valid'],
                endpoint=DiscoveryEndpoint.VALIDATE.value,
                data={
                    'valid': result['valid'],
                    'errors': result['errors'],
                    'warnings': result['warnings']
                },
                warnings=result['warnings']
            )
            
        except Exception as e:
            return DiscoveryResponse(
                success=False,
                endpoint=DiscoveryEndpoint.VALIDATE.value,
                error=str(e)
            )


def create_api() -> DiscoveryAPI:
    """Create a DiscoveryAPI instance."""
    return DiscoveryAPI()
