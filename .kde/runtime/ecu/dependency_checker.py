"""
Runtime Dependency Checker

Verifies all required Python packages are available before runtime starts.
This is the FIRST gate in initialization - dependencies must be verified
before any other runtime operations.

HUMAN EXPECTATION BEHAVIOR:
- If dependencies are missing, ATTEMPT TO INSTALL THEM automatically
- Only BLOCK if automatic installation fails
- Provide clear feedback on what was installed vs what failed

Part of INV-RUNTIME-GAPS mitigation.
"""

import importlib
import subprocess
import sys
import os
from typing import Dict, List, Tuple, Any, Optional
from dataclasses import dataclass, field
from datetime import datetime


@dataclass
class DependencyCheck:
    """Result of a single dependency check."""
    package: str
    module_path: str  # The module that requires this package
    importable: bool
    import_error: Optional[str] = None
    install_attempted: bool = False
    install_succeeded: bool = False
    install_error: Optional[str] = None


@dataclass
class DependencyCheckResult:
    """Complete result of dependency checking."""
    checked_at: str
    all_passed: bool
    total_checks: int = 0
    passed_checks: int = 0
    failed_checks: int = 0
    auto_installed: List[str] = field(default_factory=list)
    still_missing: List[str] = field(default_factory=list)
    checks: List[DependencyCheck] = field(default_factory=list)


# Package name to pip package name mapping
PACKAGE_TO_PIP: Dict[str, str] = {
    'yaml': 'pyyaml',
    'ruamel.yaml': 'ruamel.yaml',
    'json': 'json',  # Built-in, won't be in this dict
    'jsonschema': 'jsonschema',
}


class DependencyChecker:
    """
    Validates runtime dependencies before ECU initialization.
    
    This checker:
    - Identifies all packages that must be importable
    - Tests each import in isolation
    - If missing, ATTEMPTS TO INSTALL automatically
    - Only reports failure if auto-install also fails
    - Provides clear failure messages
    
    Usage:
        checker = DependencyChecker()
        result = checker.check_and_fix()
        if not result.all_passed:
            print(checker.format_failure_report())
            sys.exit(1)
    """
    
    # Core dependencies required by ECU components
    # Format: {package_name: [list of modules that require it]}
    REQUIRED_DEPENDENCIES: Dict[str, List[str]] = {
        'yaml': [
            'runtime.ecu.registry.engine_registry',
            'runtime.ecu.registry.seed_registry',
        ],
    }
    
    def __init__(self, kde_root: str = "/workspace/project/kde", auto_install: bool = True):
        """
        Initialize the Dependency Checker.
        
        Args:
            kde_root: Path to KDE runtime root (for reference)
            auto_install: If True, automatically attempt to install missing packages
        """
        self.kde_root = kde_root
        self.auto_install = auto_install
        self._checks: List[DependencyCheck] = []
    
    def check_and_fix(self) -> DependencyCheckResult:
        """
        Check all required dependencies and auto-install if missing.
        
        This is the main method that implements the human expectation:
        1. Check if dependencies are available
        2. If missing, attempt automatic installation
        3. Only report failure if both check AND install fail
        
        Returns:
            DependencyCheckResult with pass/fail status
        """
        self._checks = []
        auto_installed = []
        still_missing = []
        
        # Stage 1: Check each required package
        for package, importers in self.REQUIRED_DEPENDENCIES.items():
            for module_path in importers:
                check = self._check_package(package, module_path)
                self._checks.append(check)
                
                # Stage 2: If missing and auto_install enabled, try to install
                if not check.importable and self.auto_install:
                    install_result = self._attempt_install(package)
                    check.install_attempted = True
                    
                    if install_result["success"]:
                        # Verify the import now works
                        try:
                            importlib.import_module(package)
                            check.importable = True
                            check.install_succeeded = True
                            auto_installed.append(package)
                        except ImportError:
                            check.install_succeeded = False
                            check.install_error = "Package installed but import still fails"
                            if package not in still_missing:
                                still_missing.append(package)
                    else:
                        check.install_succeeded = False
                        check.install_error = install_result["error"]
                        if package not in still_missing:
                            still_missing.append(package)
                elif not check.importable:
                    if package not in still_missing:
                        still_missing.append(package)
        
        passed = sum(1 for c in self._checks if c.importable)
        failed = len(self._checks) - passed
        
        return DependencyCheckResult(
            checked_at=datetime.now().isoformat() + "Z",
            all_passed=failed == 0,
            total_checks=len(self._checks),
            passed_checks=passed,
            failed_checks=failed,
            auto_installed=auto_installed,
            still_missing=still_missing,
            checks=self._checks
        )
    
    def _check_package(self, package: str, module_path: str) -> DependencyCheck:
        """
        Check if a single package can be imported.
        
        Args:
            package: Package name (e.g., 'yaml')
            module_path: Path to module that requires this package
            
        Returns:
            DependencyCheck with result
        """
        try:
            importlib.import_module(package)
            return DependencyCheck(
                package=package,
                module_path=module_path,
                importable=True
            )
        except ImportError as e:
            return DependencyCheck(
                package=package,
                module_path=module_path,
                importable=False,
                import_error=str(e)
            )
    
    def _attempt_install(self, package: str) -> Dict[str, Any]:
        """
        Attempt to install a package using pip.
        
        Args:
            package: Package name to install
            
        Returns:
            Dictionary with success status and any error message
        """
        pip_name = PACKAGE_TO_PIP.get(package, package)
        
        print(f"📦 Installing missing dependency: {pip_name}...")
        
        try:
            # Run pip install
            result = subprocess.run(
                [sys.executable, '-m', 'pip', 'install', pip_name, '-q'],
                capture_output=True,
                text=True,
                timeout=120  # 2 minute timeout
            )
            
            if result.returncode == 0:
                print(f"   ✅ Successfully installed {pip_name}")
                return {"success": True}
            else:
                error_msg = result.stderr.strip() if result.stderr else "Unknown error"
                print(f"   ❌ Failed to install {pip_name}: {error_msg}")
                return {"success": False, "error": error_msg}
                
        except subprocess.TimeoutExpired:
            print(f"   ❌ Installation timed out for {pip_name}")
            return {"success": False, "error": "Installation timed out"}
        except Exception as e:
            print(f"   ❌ Error running pip: {str(e)}")
            return {"success": False, "error": str(e)}
    
    def is_ready(self) -> Tuple[bool, List[str]]:
        """
        Quick check if all dependencies are present (after auto-install attempt).
        
        Returns:
            Tuple of (all_present, missing_list)
        """
        result = self.check_and_fix()
        return result.all_passed, result.still_missing
    
    def format_report(self) -> str:
        """
        Generate human-readable report with auto-install status.
        
        Returns:
            Formatted string with status
        """
        result = self.check_and_fix()
        
        lines = []
        lines.append("=" * 78)
        
        if result.all_passed:
            lines.append("✅ RUNTIME DEPENDENCY VERIFICATION PASSED")
            lines.append("=" * 78)
            lines.append("")
            lines.append(f"Checked: {result.checked_at}")
            
            if result.auto_installed:
                lines.append(f"Auto-installed: {', '.join(result.auto_installed)}")
            
            lines.append(f"Status: ALL PASSED ({result.passed_checks}/{result.total_checks} checks)")
            lines.append("")
            lines.append("Verified Packages:")
            for check in result.checks:
                if check.importable:
                    if check.install_attempted and check.install_succeeded:
                        lines.append(f"  ✓ {check.package} (auto-installed)")
                    else:
                        lines.append(f"  ✓ {check.package}")
        else:
            lines.append("❌ RUNTIME DEPENDENCY FAILURE")
            lines.append("=" * 78)
            lines.append("")
            lines.append(f"Checked: {result.checked_at}")
            lines.append(f"Status: FAILED ({result.failed_checks}/{result.total_checks} checks failed)")
            
            if result.auto_installed:
                lines.append("")
                lines.append("Auto-installed during check:")
                for pkg in result.auto_installed:
                    lines.append(f"  📦 {pkg}")
            
            lines.append("")
            lines.append("STILL MISSING (manual intervention required):")
            for package in result.still_missing:
                lines.append(f"  • {package}")
                # Find checks for this package
                for check in self._checks:
                    if check.package == package:
                        if check.install_attempted:
                            lines.append(f"    → Auto-install attempted: {check.install_error}")
                        else:
                            lines.append(f"    → Required by: {check.module_path}")
            
            lines.append("")
            lines.append("REQUIRED ACTIONS:")
            lines.append("  pip install " + " ".join(result.still_missing))
            
        lines.append("")
        lines.append("=" * 78)
        
        return "\n".join(lines)
    
    # Alias for backward compatibility
    def check_all(self) -> DependencyCheckResult:
        """Alias for check_and_fix() for backward compatibility."""
        return self.check_and_fix()


def check_runtime_dependencies(auto_install: bool = True) -> Tuple[bool, List[str]]:
    """
    Convenience function to check and fix runtime dependencies.
    
    Args:
        auto_install: If True, automatically install missing packages
        
    Returns:
        Tuple of (all_present, missing_list)
    """
    checker = DependencyChecker(auto_install=auto_install)
    return checker.is_ready()


def verify_and_exit(kde_root: str = "/workspace/project/kde", auto_install: bool = True) -> None:
    """
    Verify dependencies, auto-install if needed, and exit with appropriate code.
    
    This should be called BEFORE any other runtime initialization.
    
    Args:
        kde_root: Path to KDE runtime root
        auto_install: If True, automatically install missing packages
    """
    checker = DependencyChecker(kde_root, auto_install=auto_install)
    print(checker.format_report())
    print()
    
    if not checker._checks or all(c.importable for c in checker._checks):
        sys.exit(0)
    
    print("🔴 INITIALIZATION BLOCKED: Missing required dependencies")
    sys.exit(1)


# ============================================================================
# ECU INTEGRATION
# ============================================================================

def validate_dependencies_for_ecu(auto_install: bool = True) -> Dict[str, Any]:
    """
    Validate dependencies specifically for ECU initialization.
    
    This function is called by the ECU bootstrap as the first gate.
    It will auto-install missing dependencies before failing.
    
    Args:
        auto_install: If True, automatically install missing packages
        
    Returns:
        Dictionary with validation result for ECU consumption
    """
    checker = DependencyChecker(auto_install=auto_install)
    result = checker.check_and_fix()
    
    return {
        "valid": result.all_passed,
        "checked_at": result.checked_at,
        "auto_installed": result.auto_installed,
        "still_missing": result.still_missing,
        "failed_checks": result.failed_checks,
        "report": checker.format_report()
    }


# ============================================================================
# DEMONSTRATION
# ============================================================================

if __name__ == "__main__":
    print("Running Runtime Dependency Check (with auto-install)...")
    print("")
    verify_and_exit(auto_install=True)
