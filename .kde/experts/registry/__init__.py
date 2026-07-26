"""
KDE Expert Registry

Provides centralized access to domain expert knowledge.
"""

import os
import yaml
from pathlib import Path
from dataclasses import dataclass
from typing import Dict, List, Optional, Any


@dataclass
class Expert:
    """Represents a domain expert."""
    id: str
    name: str
    domain: str
    version: str
    status: str
    path: Path
    rules: List[str] = None
    constraints: List[str] = None
    
    def __post_init__(self):
        if self.rules is None:
            self.rules = []
        if self.constraints is None:
            self.constraints = []


class ExpertRegistry:
    """
    Central registry for KDE domain experts.
    
    Usage:
        registry = ExpertRegistry()
        expert = registry.get("DNP3-EXPERT-001")
        print(expert.domain)
        print(expert.rules)
    """
    
    def __init__(self, repo_root: Path = None):
        """
        Initialize the expert registry.
        
        Args:
            repo_root: Path to repository root. Defaults to cwd.
        """
        if repo_root is None:
            # Navigate from .kde/experts/registry/ to repo root
            current = Path(__file__).resolve()
            if '.kde' in current.parts:
                idx = current.parts.index('.kde')
                repo_root = Path(*current.parts[:idx])
            else:
                repo_root = Path.cwd()
        
        self.repo_root = repo_root
        self.experts_dir = repo_root / ".kde" / "experts"
        self._experts: Dict[str, Expert] = {}
        self._loaded = False
    
    def _discover_experts(self) -> None:
        """Discover all experts in the experts directory."""
        if self._loaded:
            return
        
        self._experts = {}
        
        for expert_dir in self.experts_dir.iterdir():
            if not expert_dir.is_dir():
                continue
            
            readme = expert_dir / "README.md"
            if not readme.exists():
                continue
            
            try:
                expert = self._parse_expert(readme)
                self._experts[expert.id] = expert
            except Exception as e:
                print(f"Warning: Failed to parse expert at {readme}: {e}")
        
        self._loaded = True
    
    def _parse_expert(self, readme_path: Path) -> Expert:
        """
        Parse an expert README.md file.
        
        Extracts expert metadata from the file.
        """
        expert_dir = readme_path.parent
        
        with open(readme_path) as f:
            content = f.read()
        
        # Extract expert ID from first line
        lines = content.split('\n')
        expert_id = None
        expert_name = "Unknown"
        domain = "Unknown"
        version = "1.0.0"
        status = "Active"
        
        for line in lines[:20]:  # Check first 20 lines
            if '**Expert ID**:' in line or 'Expert ID' in line:
                parts = line.split(':')
                if len(parts) > 1:
                    expert_id = parts[-1].strip().replace('**', '')
            elif '**Expert ID**' in line:
                # Handle markdown bold
                idx = line.find('Expert ID')
                if idx > 0:
                    end_idx = line.find('**', idx + 10)
                    if end_idx > idx:
                        expert_id = line[idx + 11:end_idx]
        
        if expert_id is None:
            # Try to get from directory name
            expert_id = expert_dir.name.upper().replace('-', '-')
        
        # Extract name from directory
        name = expert_dir.name.replace('-', ' ').replace('_', ' ').title()
        
        return Expert(
            id=expert_id,
            name=name,
            domain=domain,
            version=version,
            status=status,
            path=readme_path.parent
        )
    
    def get(self, expert_id: str) -> Optional[Expert]:
        """
        Get an expert by ID.
        
        Args:
            expert_id: The expert ID (e.g., "DNP3-EXPERT-001")
            
        Returns:
            Expert object or None if not found
        """
        self._discover_experts()
        return self._experts.get(expert_id)
    
    def list_experts(self) -> List[Expert]:
        """
        List all available experts.
        
        Returns:
            List of Expert objects
        """
        self._discover_experts()
        return list(self._experts.values())
    
    def search_experts(self, query: str) -> List[Expert]:
        """
        Search for experts by domain or name.
        
        Args:
            query: Search query
            
        Returns:
            List of matching Expert objects
        """
        self._discover_experts()
        query_lower = query.lower()
        
        results = []
        for expert in self._experts.values():
            if (query_lower in expert.domain.lower() or
                query_lower in expert.name.lower() or
                query_lower in expert.id.lower()):
                results.append(expert)
        
        return results
    
    def get_expert_rules(self, expert_id: str) -> List[str]:
        """
        Get the rules for an expert.
        
        Args:
            expert_id: The expert ID
            
        Returns:
            List of rule strings
        """
        expert = self.get(expert_id)
        if expert is None:
            return []
        
        # Parse rules from README
        rules = []
        readme_path = expert.path / "README.md"
        
        if readme_path.exists():
            with open(readme_path) as f:
                content = f.read()
            
            # Extract rules section
            in_rules = False
            for line in content.split('\n'):
                if 'Rules' in line or 'rules' in line:
                    in_rules = True
                    continue
                if in_rules:
                    if line.startswith('##') or line.startswith('---'):
                        break
                    if line.strip():
                        rules.append(line.strip())
        
        return rules
    
    def reload(self) -> None:
        """Reload the expert registry."""
        self._loaded = False
        self._discover_experts()


# Module-level convenience functions
_registry: Optional[ExpertRegistry] = None


def get_registry() -> ExpertRegistry:
    """Get the global expert registry."""
    global _registry
    if _registry is None:
        _registry = ExpertRegistry()
    return _registry


def get_expert(expert_id: str) -> Optional[Expert]:
    """Get an expert by ID from the global registry."""
    return get_registry().get(expert_id)


def list_experts() -> List[Expert]:
    """List all experts from the global registry."""
    return get_registry().list_experts()


if __name__ == "__main__":
    # Demo usage
    registry = ExpertRegistry()
    
    print("=== KDE Expert Registry ===")
    print()
    
    experts = registry.list_experts()
    print(f"Found {len(experts)} experts:")
    for expert in experts:
        print(f"  - {expert.id}: {expert.name}")
    
    print()
    
    # Search example
    dnp3_experts = registry.search_experts("DNP3")
    if dnp3_experts:
        print(f"DNP3 experts: {[e.id for e in dnp3_experts]}")
