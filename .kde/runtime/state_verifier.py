"""
Runtime State Verifier

Provides VERIFIED state rather than CLAIMED state.
Verifies actual system conditions, not just file contents.

This module ensures that:
1. Required imports actually work
2. Directories exist and are accessible
3. Config files are valid and readable
4. Engine/Seed discovery actually functions
5. The runtime can actually start from current state

Part of INV-RUNTIME-GAPS mitigation.
"""

import os
import sys
import json
import importlib
from typing import Dict, Any, List, Optional, Tuple
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path


@dataclass
class VerificationCheck:
    """Result of a single verification check."""
    check_name: str
    check_category: str  # 'import', 'directory', 'config', 'discovery', 'state'
    passed: bool
    message: str
    details: Dict[str, Any] = field(default_factory=dict)


@dataclass
class StateVerificationReport:
    """Complete state verification report."""
    verified_at: str
    kde_root: str
    overall_status: str  # 'VERIFIED', 'FAILED', 'DEGRADED'
    total_checks: int = 0
    passed_checks: int = 0
    failed_checks: int = 0
    checks: List[VerificationCheck] = field(default_factory=list)
    can_initialize: bool = False
    blocking_issues: List[str] = field(default_factory=list)


class RuntimeStateVerifier:
    """
    Verifies runtime state by checking ACTUAL conditions.
    
    Unlike reading state.json which may contain stale/claimed state,
    this verifier actually tests that the runtime can function.
    
    This is the SECOND gate after dependency checking.
    
    Usage:
        verifier = RuntimeStateVerifier("/path/to/kde")
        report = verifier.verify_all()
        if not report.can_initialize:
            print(verifier.format_report())
            sys.exit(1)
    """
    
    # Critical imports that must work for ECU to function
    CRITICAL_IMPORTS = [
        ('yaml', 'PyYAML - ECU registry parsing'),
        ('runtime.ecu', 'ECU Core'),
        ('runtime.preflight', 'Preflight Module'),
    ]
    
    # Required directories
    REQUIRED_DIRECTORIES = [
        'runtime',
        'engines',
        'seeds',
        'laboratory',
    ]
    
    # Config files that must exist and be valid
    REQUIRED_CONFIG_FILES = [
        'runtime/state.json',
        'runtime/catalog.json',
    ]
    
    def __init__(self, kde_root: str = "/workspace/project/kde"):
        """
        Initialize the State Verifier.
        
        Args:
            kde_root: Root path to the KDE runtime directory
        """
        self.kde_root = kde_root
        self._checks: List[VerificationCheck] = []
    
    def verify_all(self) -> StateVerificationReport:
        """
        Run all verification checks.
        
        Returns:
            Complete StateVerificationReport
        """
        self._checks = []
        blocking_issues = []
        
        # Stage 1: Dependency Check (critical imports)
        self._verify_imports()
        
        # Stage 2: Directory Structure
        self._verify_directories()
        
        # Stage 3: Config File Validity
        self._verify_config_files()
        
        # Stage 4: Actual Discovery (if imports work)
        self._verify_engine_discovery()
        self._verify_seed_discovery()
        
        # Stage 5: State File Consistency
        self._verify_state_file()
        
        # Calculate results
        passed = sum(1 for c in self._checks if c.passed)
        failed = sum(1 for c in self._checks if not c.passed)
        
        # Determine if we can initialize
        # Must have: imports work, directories exist, state file valid
        critical_checks = [
            c for c in self._checks 
            if c.check_category in ['import', 'directory', 'state']
        ]
        critical_passed = sum(1 for c in critical_checks if c.passed)
        can_init = critical_passed == len(critical_checks)
        
        # Collect blocking issues
        blocking_checks = [
            c for c in self._checks 
            if not c.passed and c.check_category in ['import', 'directory', 'state']
        ]
        blocking_issues = [c.message for c in blocking_checks]
        
        # Determine overall status
        if failed == 0:
            overall = 'VERIFIED'
        elif can_init:
            overall = 'DEGRADED'  # Can run but some features missing
        else:
            overall = 'FAILED'
        
        return StateVerificationReport(
            verified_at=datetime.now().isoformat() + "Z",
            kde_root=self.kde_root,
            overall_status=overall,
            total_checks=len(self._checks),
            passed_checks=passed,
            failed_checks=failed,
            checks=self._checks,
            can_initialize=can_init,
            blocking_issues=blocking_issues
        )
    
    def _verify_imports(self) -> None:
        """Verify all critical imports work."""
        for module_name, description in self.CRITICAL_IMPORTS:
            try:
                importlib.import_module(module_name)
                self._checks.append(VerificationCheck(
                    check_name=f"import_{module_name}",
                    check_category="import",
                    passed=True,
                    message=f"✓ {description} importable",
                    details={"module": module_name}
                ))
            except ImportError as e:
                self._checks.append(VerificationCheck(
                    check_name=f"import_{module_name}",
                    check_category="import",
                    passed=False,
                    message=f"✗ {description} import FAILED",
                    details={"module": module_name, "error": str(e)}
                ))
    
    def _verify_directories(self) -> None:
        """Verify required directories exist."""
        for dir_name in self.REQUIRED_DIRECTORIES:
            path = os.path.join(self.kde_root, dir_name)
            exists = os.path.exists(path)
            is_dir = os.path.isdir(path) if exists else False
            is_accessible = exists and os.access(path, os.R_OK | os.X_OK)
            
            if exists and is_dir and is_accessible:
                self._checks.append(VerificationCheck(
                    check_name=f"dir_{dir_name}",
                    check_category="directory",
                    passed=True,
                    message=f"✓ {dir_name}/ exists and accessible",
                    details={"path": path}
                ))
            else:
                status = "missing" if not exists else ("not a directory" if not is_dir else "not accessible")
                self._checks.append(VerificationCheck(
                    check_name=f"dir_{dir_name}",
                    check_category="directory",
                    passed=False,
                    message=f"✗ {dir_name}/ {status}",
                    details={"path": path, "status": status}
                ))
    
    def _verify_config_files(self) -> None:
        """Verify configuration files are readable and valid."""
        for config_path in self.REQUIRED_CONFIG_FILES:
            full_path = os.path.join(self.kde_root, config_path)
            exists = os.path.exists(full_path)
            readable = exists and os.access(full_path, os.R_OK)
            
            # Try to parse JSON files
            valid_json = False
            parse_error = None
            if readable and config_path.endswith('.json'):
                try:
                    with open(full_path, 'r') as f:
                        json.load(f)
                    valid_json = True
                except json.JSONDecodeError as e:
                    parse_error = str(e)
            
            passed = exists and readable and (not config_path.endswith('.json') or valid_json)
            
            if passed:
                self._checks.append(VerificationCheck(
                    check_name=f"config_{config_path}",
                    check_category="config",
                    passed=True,
                    message=f"✓ {config_path} valid",
                    details={"path": full_path}
                ))
            else:
                status = "missing" if not exists else ("unreadable" if not readable else f"invalid JSON: {parse_error}")
                self._checks.append(VerificationCheck(
                    check_name=f"config_{config_path}",
                    check_category="config",
                    passed=False,
                    message=f"✗ {config_path} {status}",
                    details={"path": full_path, "status": status}
                ))
    
    def _verify_engine_discovery(self) -> None:
        """Verify engine discovery actually works."""
        try:
            # First check if we can import the module
            from runtime.ecu.registry import EngineRegistry
            
            # Try to instantiate and discover
            registry = EngineRegistry(self.kde_root)
            engines = registry.discover()
            
            self._checks.append(VerificationCheck(
                check_name="engine_discovery",
                check_category="discovery",
                passed=True,
                message=f"✓ Engine discovery: {len(engines)} engines found",
                details={"count": len(engines)}
            ))
        except ImportError as e:
            self._checks.append(VerificationCheck(
                check_name="engine_discovery",
                check_category="discovery",
                passed=False,
                message=f"⚠ Engine discovery skipped (import failed: {e})",
                details={"error": str(e)}
            ))
        except Exception as e:
            self._checks.append(VerificationCheck(
                check_name="engine_discovery",
                check_category="discovery",
                passed=False,
                message=f"✗ Engine discovery failed: {e}",
                details={"error": str(e)}
            ))
    
    def _verify_seed_discovery(self) -> None:
        """Verify seed discovery actually works."""
        try:
            # First check if we can import the module
            from runtime.ecu.registry import SeedRegistry
            
            # Try to instantiate and discover
            registry = SeedRegistry(self.kde_root)
            seeds = registry.discover()
            
            self._checks.append(VerificationCheck(
                check_name="seed_discovery",
                check_category="discovery",
                passed=True,
                message=f"✓ Seed discovery: {len(seeds)} seeds found",
                details={"count": len(seeds)}
            ))
        except ImportError as e:
            self._checks.append(VerificationCheck(
                check_name="seed_discovery",
                check_category="discovery",
                passed=False,
                message=f"⚠ Seed discovery skipped (import failed: {e})",
                details={"error": str(e)}
            ))
        except Exception as e:
            self._checks.append(VerificationCheck(
                check_name="seed_discovery",
                check_category="discovery",
                passed=False,
                message=f"✗ Seed discovery failed: {e}",
                details={"error": str(e)}
            ))
    
    def _verify_state_file(self) -> None:
        """Verify state file matches actual conditions."""
        state_path = os.path.join(self.kde_root, 'runtime', 'state.json')
        
        if not os.path.exists(state_path):
            self._checks.append(VerificationCheck(
                check_name="state_file",
                check_category="state",
                passed=False,
                message="✗ runtime/state.json missing",
                details={"path": state_path}
            ))
            return
        
        try:
            with open(state_path, 'r') as f:
                state = json.load(f)
            
            # Check if state claims initialized
            claimed_status = state.get('status')
            claimed_initialized = state.get('initialized', False)
            
            # Check if imports actually work (required for initialization)
            imports_work = all(
                c.passed for c in self._checks if c.check_name.startswith('import_')
            )
            
            if claimed_initialized and not imports_work:
                # State claims initialized but imports fail - STALE STATE!
                self._checks.append(VerificationCheck(
                    check_name="state_file",
                    check_category="state",
                    passed=False,
                    message="⚠ STALE STATE: claims 'initialized' but imports fail",
                    details={
                        "claimed_status": claimed_status,
                        "claimed_initialized": claimed_initialized,
                        "actual_imports": False
                    }
                ))
            else:
                self._checks.append(VerificationCheck(
                    check_name="state_file",
                    check_category="state",
                    passed=True,
                    message=f"✓ State file consistent (status: {claimed_status})",
                    details={
                        "claimed_status": claimed_status,
                        "actual_imports": imports_work
                    }
                ))
        except Exception as e:
            self._checks.append(VerificationCheck(
                check_name="state_file",
                check_category="state",
                passed=False,
                message=f"✗ State file error: {e}",
                details={"error": str(e)}
            ))
    
    def is_ready_for_execution(self) -> bool:
        """
        Quick check if runtime is ready for execution.
        
        Returns:
            True if all critical checks passed
        """
        report = self.verify_all()
        return report.can_initialize
    
    def format_report(self) -> str:
        """
        Generate formatted verification report.
        
        Returns:
            Human-readable report
        """
        report = self.verify_all()
        
        lines = []
        lines.append("=" * 78)
        lines.append("RUNTIME STATE VERIFICATION REPORT")
        lines.append("=" * 78)
        lines.append(f"Verified: {report.verified_at}")
        lines.append(f"KDE Root: {report.kde_root}")
        lines.append("")
        
        # Overall status
        status_icon = {
            'VERIFIED': '✅',
            'DEGRADED': '⚠️',
            'FAILED': '❌'
        }.get(report.overall_status, '?')
        
        lines.append(f"Overall Status: {status_icon} {report.overall_status}")
        lines.append(f"Can Initialize: {'✅ YES' if report.can_initialize else '❌ NO'}")
        lines.append("")
        lines.append(f"Checks: {report.passed_checks} passed, {report.failed_checks} failed")
        
        if report.blocking_issues:
            lines.append("")
            lines.append("BLOCKING ISSUES:")
            for issue in report.blocking_issues:
                lines.append(f"  • {issue}")
        
        lines.append("")
        lines.append("-" * 78)
        lines.append("CHECK DETAILS:")
        lines.append("-" * 78)
        
        # Group by category
        categories = {}
        for check in report.checks:
            if check.check_category not in categories:
                categories[check.check_category] = []
            categories[check.check_category].append(check)
        
        for category, checks in categories.items():
            lines.append(f"\n[{category.upper()}]")
            for check in checks:
                icon = "✓" if check.passed else "✗"
                lines.append(f"  {icon} {check.message}")
        
        lines.append("")
        lines.append("=" * 78)
        
        return "\n".join(lines)
    
    def get_verification_summary(self) -> Dict[str, Any]:
        """
        Get verification summary for programmatic use.
        
        Returns:
            Dictionary with verification status
        """
        report = self.verify_all()
        return {
            "verified": report.overall_status == 'VERIFIED',
            "can_initialize": report.can_initialize,
            "status": report.overall_status,
            "passed": report.passed_checks,
            "failed": report.failed_checks,
            "blocking_issues": report.blocking_issues
        }


def verify_runtime_state(kde_root: str = "/workspace/project/kde") -> Tuple[bool, StateVerificationReport]:
    """
    Verify runtime state and return result.
    
    Args:
        kde_root: Path to KDE runtime root
        
    Returns:
        Tuple of (is_ready, report)
    """
    verifier = RuntimeStateVerifier(kde_root)
    report = verifier.verify_all()
    return report.can_initialize, report


def verify_and_report(kde_root: str = "/workspace/project/kde") -> str:
    """
    Verify state and return formatted report.
    
    Args:
        kde_root: Path to KDE runtime root
        
    Returns:
        Formatted report string
    """
    verifier = RuntimeStateVerifier(kde_root)
    return verifier.format_report()


def verify_runtime_or_exit(kde_root: str = "/workspace/project/kde") -> None:
    """
    Verify state and exit with appropriate code.
    
    Args:
        kde_root: Path to KDE runtime root
    """
    verifier = RuntimeStateVerifier(kde_root)
    print(verifier.format_report())
    
    if not verifier.is_ready_for_execution():
        print("")
        print("🔴 INITIALIZATION BLOCKED: Runtime state verification failed")
        sys.exit(1)
    
    sys.exit(0)


# ============================================================================
# BOOTSTRAP INTEGRATION
# ============================================================================

def verify_for_bootstrap(kde_root: str = "/workspace/project/kde") -> Dict[str, Any]:
    """
    Verify state specifically for bootstrap process.
    
    This is called by ECUBootstrap as part of initialization.
    
    Returns:
        Dictionary with verification result
    """
    verifier = RuntimeStateVerifier(kde_root)
    report = verifier.verify_all()
    
    return {
        "verified": report.can_initialize,
        "status": report.overall_status,
        "blocking_issues": report.blocking_issues,
        "report": verifier.format_report()
    }


# ============================================================================
# DEMONSTRATION
# ============================================================================

if __name__ == "__main__":
    print("Running Runtime State Verification...")
    print("")
    verify_runtime_or_exit()
