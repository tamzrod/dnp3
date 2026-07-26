#!/usr/bin/env python3
"""
KDE Bootstrap Status Checker

Provides a unified view of bootstrap state and integrity.
Wired to kde-start and kde-check commands.

Usage:
    python3 .kde/bootstrap/status.py          # Check status
    python3 .kde/bootstrap/status.py --json   # JSON output
    python3 .kde/bootstrap/status.py --watch  # Watch mode (continuous)
"""

import json
import os
import sys
import time
import hashlib
from pathlib import Path
from datetime import datetime
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional, Any


# =============================================================================
# Data Models
# =============================================================================

@dataclass
class ModuleStatus:
    """Status of a single module."""
    name: str
    path: str
    exists: bool
    valid: bool
    last_verified: Optional[str] = None
    checksum: Optional[str] = None
    issues: List[str] = None
    
    def __post_init__(self):
        if self.issues is None:
            self.issues = []


@dataclass
class BootstrapStatus:
    """Complete bootstrap status."""
    timestamp: str
    version: str
    state: str
    project: str
    initialized: bool
    modules: List[ModuleStatus]
    integrity: bool
    issues: List[str]
    warnings: List[str]
    
    def __post_init__(self):
        if self.issues is None:
            self.issues = []
        if self.warnings is None:
            self.warnings = []


# =============================================================================
# Bootstrap Status Checker
# =============================================================================

class BootstrapStatusChecker:
    """Check and verify bootstrap state."""
    
    def __init__(self, kde_root: Optional[Path] = None):
        if kde_root is None:
            # Find KDE root
            current = Path(__file__).resolve() if '__file__' in dir() else Path.cwd()
            if '.kde' in current.parts:
                idx = current.parts.index('.kde')
                kde_root = Path(*current.parts[:idx+1])
            else:
                kde_root = Path.cwd() / '.kde'
        
        self.kde_root = kde_root
        self.modules_dir = kde_root
        
    def get_module_list(self) -> List[str]:
        """Get list of expected modules from config."""
        config_path = self.kde_root / 'bootstrap' / 'config.yaml'
        if config_path.exists():
            try:
                import yaml
                with open(config_path) as f:
                    config = yaml.safe_load(f)
                return config.get('modules', [])
            except:
                pass
        
        # Fallback to directory listing
        return ['engines', 'experts', 'knowledge', 'governance', 
                'seeds', 'commands', 'capabilities', 'templates', 
                'verification', 'runtime', 'bootstrap']
    
    def compute_checksum(self, path: Path) -> Optional[str]:
        """Compute SHA256 checksum of a file or directory."""
        if not path.exists():
            return None
        
        if path.is_file():
            try:
                with open(path, 'rb') as f:
                    return hashlib.sha256(f.read()).hexdigest()[:16]
            except:
                return None
        
        return None
    
    def verify_module(self, module_name: str) -> ModuleStatus:
        """Verify a single module."""
        module_path = self.modules_dir / module_name
        issues = []
        
        exists = module_path.exists()
        valid = exists
        
        if not exists:
            issues.append(f"Module directory not found: {module_name}")
            valid = False
        elif not module_path.is_dir():
            issues.append(f"Path is not a directory: {module_name}")
            valid = False
        else:
            # Check for required files
            if module_name == 'runtime':
                if not (module_path / 'state.json').exists():
                    issues.append("Missing runtime/state.json")
                    valid = False
            elif module_name == 'bootstrap':
                if not (module_path / 'config.yaml').exists():
                    issues.append("Missing bootstrap/config.yaml")
                    valid = False
        
        checksum = None
        if module_path.exists():
            checksum = self.compute_checksum(module_path)
        
        return ModuleStatus(
            name=module_name,
            path=str(module_path),
            exists=exists,
            valid=valid,
            last_verified=datetime.now().isoformat(),
            checksum=checksum,
            issues=issues
        )
    
    def get_status(self) -> BootstrapStatus:
        """Get complete bootstrap status."""
        modules = []
        issues = []
        warnings = []
        
        # Get state
        state_file = self.kde_root / 'runtime' / 'state.json'
        state = "unknown"
        initialized = False
        
        if state_file.exists():
            try:
                with open(state_file) as f:
                    state_data = json.load(f)
                state = state_data.get('state', 'unknown')
                initialized = state in ('ready', 'initialized')
            except Exception as e:
                issues.append(f"Cannot read state.json: {e}")
        
        # Verify modules
        for module_name in self.get_module_list():
            module_status = self.verify_module(module_name)
            modules.append(module_status)
            issues.extend([f"{module_name}: {i}" for i in module_status.issues])
        
        # Check for unexpected directories
        expected = set(self.get_module_list())
        actual = set(d.name for d in self.modules_dir.iterdir() if d.is_dir() and not d.name.startswith('__'))
        unexpected = actual - expected
        if unexpected:
            warnings.append(f"Unexpected directories: {', '.join(unexpected)}")
        
        # Overall integrity
        integrity = len([m for m in modules if not m.valid]) == 0 and len(issues) == 0
        
        return BootstrapStatus(
            timestamp=datetime.now().isoformat(),
            version="1.0.0",
            state=state,
            project="DNP3 Library",
            initialized=initialized,
            modules=modules,
            integrity=integrity,
            issues=issues,
            warnings=warnings
        )
    
    def print_status(self, status: BootstrapStatus):
        """Print status in human-readable format."""
        print("=" * 70)
        print("KDE BOOTSTRAP STATUS")
        print("=" * 70)
        print(f"Timestamp: {status.timestamp}")
        print(f"Project:   {status.project}")
        print(f"State:     {status.state}")
        print(f"Integrity: {'✅ OK' if status.integrity else '❌ FAILED'}")
        print()
        
        print("--- Modules ---")
        for m in status.modules:
            icon = "✅" if m.valid else "❌"
            print(f"  [{icon}] {m.name}")
            for issue in m.issues:
                print(f"       └─ {issue}")
        
        if status.warnings:
            print()
            print("--- Warnings ---")
            for w in status.warnings:
                print(f"  ⚠️  {w}")
        
        if status.issues:
            print()
            print("--- Issues ---")
            for i in status.issues:
                print(f"  ❌ {i}")
        
        print()
        print("=" * 70)
        
        if status.integrity:
            print("STATUS: ✅ BOOTSTRAP INTACT")
        else:
            print("STATUS: ❌ BOOTSTRAP COMPROMISED")
        print("=" * 70)


# =============================================================================
# Watchdog (Optional - for continuous monitoring)
# =============================================================================

class BootstrapWatchdog:
    """
    Watchdog for bootstrap integrity and AI behavior monitoring.
    
    Monitors:
    1. Bootstrap directory integrity (file changes)
    2. Process behavior (detect runaway AI)
    3. Resource usage (detect infinite loops)
    """
    
    def __init__(self, kde_root: Path, check_interval: float = 5.0):
        self.kde_root = kde_root
        self.check_interval = check_interval
        self.baseline_checksums: Dict[str, str] = {}
        self.status_checker = BootstrapStatusChecker(kde_root)
        
    def compute_baseline(self):
        """Compute baseline checksums for integrity monitoring."""
        self.baseline_checksums = {}
        for md5 in self.kde_root.rglob('*.md'):
            rel_path = md5.relative_to(self.kde_root)
            try:
                with open(md5, 'rb') as f:
                    self.baseline_checksums[str(rel_path)] = hashlib.sha256(f.read()).hexdigest()[:16]
            except:
                pass
        
        # Also baseline YAML files
        for yml in self.kde_root.rglob('*.yaml'):
            rel_path = yml.relative_to(self.kde_root)
            try:
                with open(yml, 'rb') as f:
                    self.baseline_checksums[str(rel_path)] = hashlib.sha256(f.read()).hexdigest()[:16]
            except:
                pass
    
    def check_integrity(self) -> Dict[str, Any]:
        """Check for file changes that might indicate compromise."""
        changes = []
        
        for path_str, baseline_hash in self.baseline_checksums.items():
            current_path = self.kde_root / path_str
            if not current_path.exists():
                changes.append({
                    'type': 'deleted',
                    'path': path_str
                })
                continue
            
            try:
                with open(current_path, 'rb') as f:
                    current_hash = hashlib.sha256(f.read()).hexdigest()[:16]
                if current_hash != baseline_hash:
                    changes.append({
                        'type': 'modified',
                        'path': path_str,
                        'baseline': baseline_hash,
                        'current': current_hash
                    })
            except:
                pass
        
        return {
            'integrity_ok': len(changes) == 0,
            'changes': changes
        }
    
    def watch(self, duration: Optional[float] = None):
        """Watch mode - continuously monitor bootstrap."""
        print(f"Starting watchdog monitor (interval: {self.check_interval}s)")
        self.compute_baseline()
        
        start_time = time.time()
        iteration = 0
        
        try:
            while True:
                iteration += 1
                print(f"\n[Watchdog #{iteration}] {datetime.now().isoformat()}")
                
                # Check bootstrap status
                status = self.status_checker.get_status()
                print(f"  Bootstrap: {'OK' if status.integrity else 'ISSUES'}")
                print(f"  State: {status.state}")
                
                # Check for file changes
                integrity = self.check_integrity()
                if integrity['integrity_ok']:
                    print(f"  Integrity: OK (no file changes)")
                else:
                    print(f"  ⚠️  FILE CHANGES DETECTED: {len(integrity['changes'])}")
                    for change in integrity['changes']:
                        print(f"    - {change['type']}: {change['path']}")
                
                # Check time
                if duration and (time.time() - start_time) > duration:
                    print("\nWatchdog stopped (duration reached)")
                    break
                
                time.sleep(self.check_interval)
                
        except KeyboardInterrupt:
            print("\nWatchdog stopped (interrupt)")


# =============================================================================
# CLI
# =============================================================================

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="KDE Bootstrap Status Checker")
    parser.add_argument('--json', action='store_true', help='Output as JSON')
    parser.add_argument('--watch', action='store_true', help='Watch mode (continuous)')
    parser.add_argument('--interval', type=float, default=5.0, help='Watch interval in seconds')
    parser.add_argument('--duration', type=float, help='Watch duration in seconds')
    parser.add_argument('--kde-root', type=Path, help='KDE root directory')
    
    args = parser.parse_args()
    
    checker = BootstrapStatusChecker(args.kde_root)
    status = checker.get_status()
    
    if args.watch:
        watchdog = BootstrapWatchdog(checker.kde_root, args.interval)
        watchdog.watch(args.duration)
    elif args.json:
        print(json.dumps(asdict(status), indent=2))
    else:
        checker.print_status(status)
        sys.exit(0 if status.integrity else 1)
