"""
Seed Registry Module

Automatic discovery and registration of KDE seeds.
"""

import os
import yaml
import re
from pathlib import Path
from typing import Dict, List, Optional, Set, Any

from ..models import (
    SeedMetadata, SeedStatus, Capability, CapabilityType
)


class SeedRegistry:
    """
    Runtime registry for KDE seeds with automatic discovery.
    
    Responsibilities:
    - Discover seeds from the seeds/ directory
    - Parse seed specifications and metadata
    - Maintain seed registry
    - Support capability queries
    """
    
    def __init__(self, kde_root: str):
        """
        Initialize the Seed Registry.
        
        Args:
            kde_root: Root path to the KDE runtime directory
        """
        self.kde_root = kde_root
        self.seeds_dir = os.path.join(kde_root, "seeds")
        self.seeds: Dict[str, SeedMetadata] = {}
        self._discovery_complete = False
    
    def discover(self) -> List[SeedMetadata]:
        """
        Automatically discover all seeds in the seeds/ directory.
        
        Returns:
            List of discovered seed metadata
        """
        discovered = []
        
        if not os.path.exists(self.seeds_dir):
            return discovered
        
        for entry in os.listdir(self.seeds_dir):
            seed_path = os.path.join(self.seeds_dir, entry)
            
            if not os.path.isdir(seed_path):
                continue
            
            # Skip README.md
            if entry.endswith('.md'):
                continue
            
            # Discover seed from directory
            seed = self._discover_seed(entry, seed_path)
            if seed:
                self.seeds[seed.seed_id] = seed
                discovered.append(seed)
        
        self._discovery_complete = True
        return discovered
    
    def _discover_seed(self, directory: str, seed_path: str) -> Optional[SeedMetadata]:
        """
        Discover seed metadata from a directory.
        
        Args:
            directory: Seed directory name
            seed_path: Full path to seed directory
        
        Returns:
            SeedMetadata if valid seed found, None otherwise
        """
        # Look for seed.yaml
        yaml_path = os.path.join(seed_path, "seed.yaml")
        
        if os.path.exists(yaml_path):
            return self._parse_yaml_seed(directory, seed_path, yaml_path)
        
        # Try to parse from README.md
        readme_path = os.path.join(seed_path, "README.md")
        if os.path.exists(readme_path):
            return self._parse_readme_seed(directory, seed_path, readme_path)
        
        # Create basic metadata from directory name
        return self._create_basic_seed(directory, seed_path)
    
    def _parse_yaml_seed(self, directory: str, seed_path: str, yaml_path: str) -> Optional[SeedMetadata]:
        """
        Parse a seed.yaml file (which is actually markdown with YAML-like content).
        
        Args:
            directory: Seed directory name
            seed_path: Full path to seed directory
            yaml_path: Path to seed.yaml
        
        Returns:
            SeedMetadata parsed from seed.yaml
        """
        try:
            with open(yaml_path, 'r') as f:
                content = f.read()
            
            # Extract seed ID
            seed_id_match = re.search(r'\*\*Seed ID\*\*:\s*(\S+)', content)
            seed_id = seed_id_match.group(1) if seed_id_match else directory.upper().replace('-', '_')
            
            # Extract name
            name_match = re.search(r'\*\*Name\*\*:\s*(.+)', content)
            name = name_match.group(1).strip() if name_match else seed_id
            
            # Extract codename
            codename_match = re.search(r'\*\*Codename\*\*:\s*(\S+)', content)
            codename = codename_match.group(1) if codename_match else seed_id
            
            # Extract version
            version_match = re.search(r'\*\*Version\*\*:\s*(\S+)', content)
            version = version_match.group(1) if version_match else "1.0.0"
            
            # Extract status
            status_match = re.search(r'\*\*Status\*\*:\s*(\S+)', content)
            status_str = status_match.group(1).lower() if status_match else "active"
            
            if 'frozen' in status_str:
                status = SeedStatus.FROZEN
            elif 'deprecated' in status_str:
                status = SeedStatus.DEPRECATED
            else:
                status = SeedStatus.ACTIVE
            
            # Extract compatible engines from markdown table
            compatible_engines = []
            engine_pattern = re.compile(r'\|\s*([A-Z]+-[A-Z]+-\d+)\s*\|')
            for match in engine_pattern.finditer(content):
                engine_id = match.group(1)
                if engine_id not in compatible_engines:
                    compatible_engines.append(engine_id)
            
            seed = SeedMetadata(
                seed_id=seed_id,
                directory=directory,
                name=name,
                codename=codename,
                version=version,
                status=status,
                compatible_engines=compatible_engines
            )
            
            # Extract capabilities from directory structure
            seed.capabilities = self._extract_seed_capabilities(seed_path, seed)
            
            return seed
            
        except Exception:
            return None
    
    def _parse_readme_seed(self, directory: str, seed_path: str, readme_path: str) -> Optional[SeedMetadata]:
        """
        Parse a seed from README.md.
        
        Args:
            directory: Seed directory name
            seed_path: Full path to seed directory
            readme_path: Path to README.md
        
        Returns:
            SeedMetadata parsed from README
        """
        try:
            with open(readme_path, 'r') as f:
                content = f.read()
            
            # Extract seed ID
            seed_id_match = re.search(r'(SEED-\d+)', content)
            seed_id = seed_id_match.group(1) if seed_id_match else directory
            
            # Extract name
            name_match = re.search(r'\*\*Name\*\*:\s*(.+)', content)
            name = name_match.group(1).strip() if name_match else seed_id
            
            # Extract version
            version_match = re.search(r'\*\*Version\*\*:\s*(\S+)', content)
            version = version_match.group(1) if version_match else "1.0.0"
            
            seed = SeedMetadata(
                seed_id=seed_id,
                directory=directory,
                name=name,
                codename=name.split()[0] if name else seed_id,
                version=version,
                status=SeedStatus.ACTIVE
            )
            
            seed.capabilities = self._extract_seed_capabilities(seed_path, seed)
            
            return seed
            
        except Exception:
            return None
    
    def _create_basic_seed(self, directory: str, seed_path: str) -> Optional[SeedMetadata]:
        """
        Create basic seed metadata from directory name.
        
        Args:
            directory: Seed directory name
            seed_path: Full path to seed directory
        
        Returns:
            Basic SeedMetadata
        """
        seed = SeedMetadata(
            seed_id=directory.upper().replace('-', '_'),
            directory=directory,
            name=directory,
            codename=directory,
            version="1.0.0",
            status=SeedStatus.ACTIVE
        )
        
        seed.capabilities = self._extract_seed_capabilities(seed_path, seed)
        
        return seed
    
    def _extract_seed_capabilities(self, seed_path: str, seed: SeedMetadata) -> List[Capability]:
        """
        Extract capabilities from seed directory structure.
        
        Args:
            seed_path: Path to seed directory
            seed: SeedMetadata to add capabilities to
        
        Returns:
            List of extracted capabilities
        """
        capabilities = []
        
        # Map directory names to capabilities
        capability_map = {
            'principles': CapabilityType.REASONING,
            'reasoning': CapabilityType.REASONING,
            'validation': CapabilityType.VALIDATION,
            'evidence-model': CapabilityType.VALIDATION,
            'confidence-model': CapabilityType.ANALYSIS,
            'knowledge-model': CapabilityType.ANALYSIS,
            'scientific-loop': CapabilityType.SYNTHESIS,
            'architecture': CapabilityType.SYNTHESIS,
            'philosophy': CapabilityType.REASONING,
        }
        
        for subdir in os.listdir(seed_path):
            subdir_lower = subdir.lower()
            if subdir_lower in capability_map:
                cap_type = capability_map[subdir_lower]
                capabilities.append(Capability(
                    name=f"{seed.codename}_{subdir_lower}",
                    type=cap_type,
                    description=f"{seed.codename} {subdir_lower} capability",
                    keywords=[seed.codename.lower(), subdir_lower]
                ))
        
        # Add default foundation capability
        if not capabilities:
            capabilities.append(Capability(
                name=f"{seed.codename}_foundation",
                type=CapabilityType.REASONING,
                description=f"{seed.codename} foundation reasoning capability",
                keywords=[seed.codename.lower(), "foundation"]
            ))
        
        return capabilities
    
    def get_seed(self, seed_id: str) -> Optional[SeedMetadata]:
        """
        Get a seed by ID.
        
        Args:
            seed_id: Seed ID
        
        Returns:
            SeedMetadata if found, None otherwise
        """
        return self.seeds.get(seed_id)
    
    def get_seeds_by_status(self, status: SeedStatus) -> List[SeedMetadata]:
        """
        Get all seeds with a specific status.
        
        Args:
            status: SeedStatus to filter by
        
        Returns:
            List of matching seeds
        """
        return [s for s in self.seeds.values() if s.status == status]
    
    def get_active_seeds(self) -> List[SeedMetadata]:
        """Get all active seeds."""
        return self.get_seeds_by_status(SeedStatus.ACTIVE)
    
    def get_frozen_seeds(self) -> List[SeedMetadata]:
        """Get all frozen seeds."""
        return self.get_seeds_by_status(SeedStatus.FROZEN)
    
    def get_seed_for_engine(self, engine_id: str) -> List[SeedMetadata]:
        """
        Get seeds compatible with an engine.
        
        Args:
            engine_id: Engine ID
        
        Returns:
            List of compatible seeds
        """
        return [
            s for s in self.seeds.values()
            if engine_id in s.compatible_engines or not s.compatible_engines
        ]
    
    def get_seeds_by_capability(self, capability: CapabilityType) -> List[SeedMetadata]:
        """
        Get all seeds that provide a specific capability.
        
        Args:
            capability: CapabilityType to filter by
        
        Returns:
            List of seeds with the capability
        """
        return [
            s for s in self.seeds.values()
            if any(c.type == capability for c in s.capabilities)
        ]
    
    def get_all_seeds(self) -> List[SeedMetadata]:
        """Get all registered seeds."""
        return list(self.seeds.values())
    
    def get_registry_summary(self) -> Dict[str, Any]:
        """
        Get a summary of the seed registry.
        
        Returns:
            Dictionary with registry statistics
        """
        return {
            "total_seeds": len(self.seeds),
            "active": len(self.get_seeds_by_status(SeedStatus.ACTIVE)),
            "frozen": len(self.get_seeds_by_status(SeedStatus.FROZEN)),
            "deprecated": len(self.get_seeds_by_status(SeedStatus.DEPRECATED)),
            "discovery_complete": self._discovery_complete,
            "seed_ids": list(self.seeds.keys())
        }
