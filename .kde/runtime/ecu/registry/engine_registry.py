"""
Engine Registry Module

Automatic discovery and registration of KDE engines.
"""

import os
import yaml
import re
from pathlib import Path
from typing import Dict, List, Optional, Set, Any
from dataclasses import asdict

from ..models import (
    EngineMetadata, EngineStatus, EngineStability, Capability, CapabilityType
)


class EngineRegistry:
    """
    Runtime registry for KDE engines with automatic discovery.
    
    Responsibilities:
    - Discover engines from the engines/ directory
    - Parse engine specifications and metadata
    - Maintain engine registry
    - Support capability queries
    """
    
    def __init__(self, kde_root: str):
        """
        Initialize the Engine Registry.
        
        Args:
            kde_root: Root path to the KDE runtime directory
        """
        self.kde_root = kde_root
        self.engines_dir = os.path.join(kde_root, "engines")
        self.engines: Dict[str, EngineMetadata] = {}
        self._discovery_complete = False
    
    def discover(self) -> List[EngineMetadata]:
        """
        Automatically discover all engines in the engines/ directory.
        
        Returns:
            List of discovered engine metadata
        """
        discovered = []
        
        if not os.path.exists(self.engines_dir):
            return discovered
        
        for entry in os.listdir(self.engines_dir):
            engine_path = os.path.join(self.engines_dir, entry)
            
            if not os.path.isdir(engine_path):
                continue
            
            # Skip markdown files (documentation files)
            if entry.endswith('.md'):
                continue
            
            # Try to discover engine from directory
            engine = self._discover_engine(entry, engine_path)
            if engine:
                self.engines[engine.engine_id] = engine
                discovered.append(engine)
        
        self._discovery_complete = True
        return discovered
    
    def _discover_engine(self, directory: str, engine_path: str) -> Optional[EngineMetadata]:
        """
        Discover engine metadata from a directory.
        
        Args:
            directory: Engine directory name
            engine_path: Full path to engine directory
        
        Returns:
            EngineMetadata if valid engine found, None otherwise
        """
        # Look for specification files
        spec_path = os.path.join(engine_path, "specification.md")
        spec_alt_path = os.path.join(engine_path, "SPEC.md")  # Alternative naming
        manifest_path = os.path.join(engine_path, "manifest.yaml")
        
        metadata = None
        
        if os.path.exists(spec_path):
            # Parse specification
            metadata = self._parse_specification(spec_path)
            if metadata:
                metadata.specification_path = spec_path
                
                # Look for manifest.yaml for additional metadata
                if os.path.exists(manifest_path):
                    self._parse_manifest(metadata, manifest_path)
        elif os.path.exists(spec_alt_path):
            # Parse from SPEC.md
            metadata = self._parse_specification(spec_alt_path)
            if metadata:
                metadata.specification_path = spec_alt_path
                
                # Look for manifest.yaml for additional metadata
                if os.path.exists(manifest_path):
                    self._parse_manifest(metadata, manifest_path)
        elif os.path.exists(manifest_path):
            # Parse from manifest.yaml only
            metadata = self._parse_manifest_only(directory, manifest_path)
        
        if not metadata:
            return None
        
        # Set directory
        metadata.directory = directory
        
        # Look for methodology
        methodology_path = os.path.join(engine_path, "methodology.md")
        if os.path.exists(methodology_path):
            metadata.methodology_path = methodology_path
        
        # Extract capabilities from metadata
        metadata.capabilities = self._extract_capabilities(metadata)
        
        return metadata
    
    def _parse_specification(self, spec_path: str) -> Optional[EngineMetadata]:
        """
        Parse an engine specification.md file.
        
        Args:
            spec_path: Path to specification.md
        
        Returns:
            EngineMetadata parsed from specification
        """
        try:
            with open(spec_path, 'r') as f:
                content = f.read()
            
            # Extract engine ID
            engine_id_match = re.search(r'\*\*Engine ID\*\*:\s*(\S+)', content)
            if not engine_id_match:
                return None
            engine_id = engine_id_match.group(1)
            
            # Extract version
            version_match = re.search(r'\*\*Version\*\*:\s*(\S+)', content)
            version = version_match.group(1) if version_match else "0.0.0"
            
            # Extract codename
            codename_match = re.search(r'\*\*Codename\*\*:\s*(\S+)', content)
            codename = codename_match.group(1) if codename_match else engine_id
            
            # Extract status
            status_match = re.search(r'\*\*Status\*\*:\s*(\S+)', content)
            status_str = status_match.group(1).lower() if status_match else "active"
            
            if "historical" in status_str:
                status = EngineStatus.HISTORICAL
            elif "deprecated" in status_str:
                status = EngineStatus.DEPRECATED
            elif "experimental" in status_str:
                status = EngineStatus.EXPERIMENTAL
            else:
                status = EngineStatus.ACTIVE
            
            return EngineMetadata(
                engine_id=engine_id,
                directory="",
                name=codename,
                codename=codename,
                version=version,
                status=status,
                stability=EngineStability.STABLE,
                provenance=spec_path
            )
            
        except Exception:
            return None
    
    def _parse_manifest(self, metadata: EngineMetadata, manifest_path: str) -> None:
        """
        Parse a manifest.yaml file for additional metadata.
        
        Args:
            metadata: EngineMetadata to update
            manifest_path: Path to manifest.yaml
        """
        try:
            with open(manifest_path, 'r') as f:
                manifest = yaml.safe_load(f)
            
            if not manifest:
                return
            
            # Extract stability
            if 'stability' in manifest:
                stability_str = manifest['stability'].lower()
                if 'unstable' in stability_str:
                    metadata.stability = EngineStability.UNSTABLE
                elif 'testing' in stability_str:
                    metadata.stability = EngineStability.TESTING
                else:
                    metadata.stability = EngineStability.STABLE
            
            # Extract priority
            if 'priority' in manifest:
                metadata.priority = int(manifest['priority'])
            
            # Extract dependencies
            if 'dependencies' in manifest:
                metadata.dependencies = manifest['dependencies']
            
            # Extract compatible seeds
            if 'compatible_seeds' in manifest:
                metadata.compatible_seeds = manifest['compatible_seeds']
            
            # Extract capabilities
            if 'capabilities' in manifest:
                for cap in manifest['capabilities']:
                    if isinstance(cap, dict):
                        metadata.capabilities.append(Capability(
                            name=cap.get('name', ''),
                            type=CapabilityType(cap.get('type', 'reasoning')),
                            description=cap.get('description', ''),
                            keywords=cap.get('keywords', [])
                        ))
            
        except Exception:
            pass
    
    def _parse_manifest_only(self, directory: str, manifest_path: str) -> Optional[EngineMetadata]:
        """
        Parse engine from manifest.yaml (markdown format, no specification.md).
        
        Args:
            directory: Engine directory name
            manifest_path: Path to manifest.yaml
        
        Returns:
            EngineMetadata if valid, None otherwise
        """
        try:
            with open(manifest_path, 'r') as f:
                content = f.read()
            
            # Extract basic metadata using regex (markdown format)
            engine_id_match = re.search(r'\*\*Engine ID\*\*:\s*(\S+)', content)
            engine_id = engine_id_match.group(1) if engine_id_match else f"KDE-{directory.upper()}"
            
            version_match = re.search(r'\*\*Version\*\*:\s*(\S+)', content)
            version = version_match.group(1) if version_match else "1.0.0"
            
            # Extract codename from filename
            codename = directory
            
            # Extract status
            status_match = re.search(r'\*\*Status\*\*:\s*(\S+)', content)
            status_str = status_match.group(1).lower() if status_match else "active"
            
            if "historical" in status_str:
                status = EngineStatus.HISTORICAL
            elif "deprecated" in status_str:
                status = EngineStatus.DEPRECATED
            elif "experimental" in status_str:
                status = EngineStatus.EXPERIMENTAL
            else:
                status = EngineStatus.ACTIVE
            
            # Default stability
            stability = EngineStability.TESTING
            
            # Extract capabilities from markdown list
            capabilities = []
            capabilities_section = re.search(r'## Capabilities\s*\n(.*?)(?:\n##|\Z)', content, re.DOTALL)
            if capabilities_section:
                cap_lines = capabilities_section.group(1).strip().split('\n')
                for line in cap_lines:
                    if line.strip().startswith('-'):
                        cap_name = line.strip().lstrip('-').strip()
                        # Determine type from keywords
                        cap_type = CapabilityType.ANALYSIS
                        for kw in ['synthes', 'generat', 'creat']:
                            if kw in cap_name.lower():
                                cap_type = CapabilityType.SYNTHESIS
                                break
                        for kw in ['validat', 'test', 'check', 'review']:
                            if kw in cap_name.lower():
                                cap_type = CapabilityType.VALIDATION
                                break
                        for kw in ['evaluat', 'critique', 'assess']:
                            if kw in cap_name.lower():
                                cap_type = CapabilityType.EVALUATION
                                break
                        capabilities.append(Capability(
                            name=cap_name,
                            type=cap_type,
                            description=cap_name,
                            keywords=[]
                        ))
            
            metadata = EngineMetadata(
                engine_id=engine_id,
                directory=directory,
                name=codename,
                codename=codename,
                version=version,
                status=status,
                stability=stability,
                priority=100,
                capabilities=capabilities,
                provenance=f"Discovered from manifest.yaml in {directory}"
            )
            
            return metadata
            
        except Exception as e:
            print(f"Error parsing manifest {manifest_path}: {e}")
            return None
    
    def _extract_capabilities(self, metadata: EngineMetadata) -> List[Capability]:
        """
        Extract capabilities from engine metadata.
        
        Args:
            metadata: EngineMetadata to extract capabilities from
        
        Returns:
            List of extracted capabilities
        """
        capabilities = []
        
        # Map codenames to default capabilities
        capability_map = {
            'alpha': [CapabilityType.REASONING, CapabilityType.ANALYSIS],
            'beta': [CapabilityType.REASONING, CapabilityType.ANALYSIS, CapabilityType.SYNTHESIS],
            'gamma': [CapabilityType.REASONING, CapabilityType.ANALYSIS, CapabilityType.SYNTHESIS, CapabilityType.VALIDATION],
            'delta': [CapabilityType.REASONING, CapabilityType.ANALYSIS, CapabilityType.GENERATION],
            'epsilon': [CapabilityType.VALIDATION, CapabilityType.EVALUATION],
            'adversarial': [CapabilityType.EVALUATION, CapabilityType.ANALYSIS],
            'consensus': [CapabilityType.SYNTHESIS, CapabilityType.VALIDATION],
            'protocol': [CapabilityType.SYNTHESIS, CapabilityType.GENERATION],
        }
        
        codename_lower = metadata.codename.lower()
        for key, types in capability_map.items():
            if key in codename_lower:
                for cap_type in types:
                    capabilities.append(Capability(
                        name=f"{metadata.codename}_{cap_type.value}",
                        type=cap_type,
                        description=f"{metadata.codename} engine {cap_type.value} capability",
                        keywords=[metadata.codename.lower(), cap_type.value]
                    ))
                break
        
        # If no match, add basic reasoning capability
        if not capabilities:
            capabilities.append(Capability(
                name=f"{metadata.codename}_reasoning",
                type=CapabilityType.REASONING,
                description=f"{metadata.codename} engine reasoning capability",
                keywords=[metadata.codename.lower()]
            ))
        
        return capabilities
    
    def get_engine(self, engine_id: str) -> Optional[EngineMetadata]:
        """
        Get an engine by ID.
        
        Args:
            engine_id: Engine ID
        
        Returns:
            EngineMetadata if found, None otherwise
        """
        return self.engines.get(engine_id)
    
    def get_engines_by_status(self, status: EngineStatus) -> List[EngineMetadata]:
        """
        Get all engines with a specific status.
        
        Args:
            status: EngineStatus to filter by
        
        Returns:
            List of matching engines
        """
        return [e for e in self.engines.values() if e.status == status]
    
    def get_active_engines(self) -> List[EngineMetadata]:
        """Get all active engines."""
        return self.get_engines_by_status(EngineStatus.ACTIVE)
    
    def get_engines_by_capability(self, capability: CapabilityType) -> List[EngineMetadata]:
        """
        Get all engines that provide a specific capability.
        
        Args:
            capability: CapabilityType to filter by
        
        Returns:
            List of engines with the capability
        """
        return [
            e for e in self.engines.values()
            if any(c.type == capability for c in e.capabilities)
        ]
    
    def get_all_engines(self) -> List[EngineMetadata]:
        """Get all registered engines."""
        return list(self.engines.values())
    
    def get_registry_summary(self) -> Dict[str, Any]:
        """
        Get a summary of the engine registry.
        
        Returns:
            Dictionary with registry statistics
        """
        return {
            "total_engines": len(self.engines),
            "active": len(self.get_active_engines()),
            "historical": len(self.get_engines_by_status(EngineStatus.HISTORICAL)),
            "experimental": len(self.get_engines_by_status(EngineStatus.EXPERIMENTAL)),
            "deprecated": len(self.get_engines_by_status(EngineStatus.DEPRECATED)),
            "discovery_complete": self._discovery_complete,
            "engine_ids": list(self.engines.keys())
        }
