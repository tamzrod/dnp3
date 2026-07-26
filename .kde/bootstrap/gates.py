"""
KDE Bootstrap Gates

This module implements the three bootstrap gates from KDE-INV-051/052:

Gate B1: Bootstrap-First Gate
    - Verify runtime state before investigation work
    - Create experiment entry before investigation
    - Acknowledge Laboratory Rules
    - Check environment requirements

Gate B2: Pre-Existence Check Gate
    - Check git log for recent fixes
    - Verify issue still exists
    - If fixed, document and stop
    - If not fixed, proceed to investigation

Gate B3: Environment Verification Gate
    - Verify toolchain availability
    - Confirm project dependencies
    - If environment incomplete, note limitation
    - Do not promise execution without verification

Usage:
    from .kde.bootstrap.gates import verify_all_gates, GateResult
    
    result = verify_all_gates(project_type="go")
    if result.passed:
        print("All gates passed")
    else:
        print(f"Failed gates: {result.failed_gates}")
"""

import os
import json
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import List, Dict, Optional, Any
from datetime import datetime


@dataclass
class GateCheck:
    """Result of a single gate check."""
    name: str
    gate: str
    passed: bool
    details: str
    can_proceed: bool = True
    
    
@dataclass
class GateResult:
    """Result of gate verification."""
    timestamp: str
    project_type: str
    checks: List[GateCheck] = field(default_factory=list)
    passed: bool = True
    failed_gates: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    can_proceed: bool = True
    summary: str = ""
    
    def add_check(self, check: GateCheck):
        self.checks.append(check)
        if not check.passed:
            self.passed = False
            if check.gate not in self.failed_gates:
                self.failed_gates.append(check.gate)
        if not check.can_proceed:
            self.can_proceed = False
        if check.details.startswith("WARNING"):
            self.warnings.append(check.details)


def get_repo_root() -> Path:
    """Get the repository root directory."""
    # Navigate from .kde/bootstrap/ to repo root
    current = Path(__file__).resolve() if '__file__' in dir() else Path.cwd()
    
    # If we're in .kde/bootstrap, go up 3 levels
    if '.kde' in current.parts:
        idx = current.parts.index('.kde')
        return Path(*current.parts[:idx])
    
    # Otherwise, assume we're at repo root or below
    return Path.cwd()


# =============================================================================
# GATE B1: Bootstrap-First Gate
# =============================================================================

def check_runtime_state() -> GateCheck:
    """
    Gate B1.1: Verify runtime state.
    
    Check that .kde/runtime/state.json exists and shows "ready" status.
    """
    repo_root = get_repo_root()
    state_file = repo_root / ".kde" / "runtime" / "state.json"
    
    if not state_file.exists():
        return GateCheck(
            name="runtime_state_file",
            gate="B1",
            passed=False,
            details="FAILED: .kde/runtime/state.json does not exist",
            can_proceed=False
        )
    
    try:
        with open(state_file) as f:
            state = json.load(f)
        
        # Check both 'status' and 'state' fields for compatibility
        status = state.get("status") or state.get("state") or "unknown"
        valid_statuses = ["ready", "initialized"]
        
        if status not in valid_statuses:
            return GateCheck(
                name="runtime_status",
                gate="B1",
                passed=False,
                details=f"FAILED: Runtime status is '{status}', expected one of {valid_statuses}",
                can_proceed=True  # Warning only, not blocking
            )
        
        modules = state.get("modules", {})
        unloaded = [m for m, s in modules.items() if s != "loaded"]
        if unloaded:
            return GateCheck(
                name="runtime_modules",
                gate="B1",
                passed=False,
                details=f"WARNING: Modules not loaded: {unloaded}",
                can_proceed=True  # Warning only
            )
        
        return GateCheck(
            name="runtime_state",
            gate="B1",
            passed=True,
            details=f"PASSED: Runtime status is '{status}', all {len(modules)} modules loaded"
        )
        
    except json.JSONDecodeError as e:
        return GateCheck(
            name="runtime_state_parse",
            gate="B1",
            passed=False,
            details=f"FAILED: Cannot parse state.json: {e}",
            can_proceed=False
        )
    except Exception as e:
        return GateCheck(
            name="runtime_state_error",
            gate="B1",
            passed=False,
            details=f"FAILED: Error checking runtime state: {e}",
            can_proceed=False
        )


def check_experiment_entry_needed() -> GateCheck:
    """
    Gate B1.2: Check if experiment entry is needed.
    
    For this implementation, we check if the current session has an experiment entry.
    In practice, this would be enforced by the agent framework.
    """
    repo_root = get_repo_root()
    experiments_dir = repo_root / "laboratory" / "experiments"
    
    if not experiments_dir.exists():
        return GateCheck(
            name="experiments_directory",
            gate="B1",
            passed=False,
            details="FAILED: laboratory/experiments/ directory does not exist",
            can_proceed=True  # Warning, not blocking
        )
    
    return GateCheck(
        name="experiments_directory",
        gate="B1",
        passed=True,
        details=f"PASSED: laboratory/experiments/ exists"
    )


def check_laboratory_rules_acknowledged() -> GateCheck:
    """
    Gate B1.3: Verify Laboratory Rules documentation exists.
    """
    repo_root = get_repo_root()
    lab_readme = repo_root / "laboratory" / "README.md"
    
    if not lab_readme.exists():
        return GateCheck(
            name="laboratory_rules",
            gate="B1",
            passed=False,
            details="FAILED: laboratory/README.md does not exist",
            can_proceed=True  # Warning, not blocking
        )
    
    return GateCheck(
        name="laboratory_rules",
        gate="B1",
        passed=True,
        details="PASSED: Laboratory rules documentation exists"
    )


def verify_bootstrap_gate_b1() -> List[GateCheck]:
    """
    Verify Gate B1: Bootstrap-First Gate.
    
    All investigation work must:
    1. Verify runtime state
    2. Create experiment entry (if new investigation)
    3. Acknowledge Laboratory Rules
    4. Check environment requirements
    """
    return [
        check_runtime_state(),
        check_experiment_entry_needed(),
        check_laboratory_rules_acknowledged(),
    ]


# =============================================================================
# GATE B2: Pre-Existence Check Gate
# =============================================================================

def check_git_log_recent_fixes(limit: int = 10) -> tuple[bool, str]:
    """
    Check git log for recent commits that may have fixed issues.
    
    Returns: (has_git_repo, message)
    """
    repo_root = get_repo_root()
    git_dir = repo_root / ".git"
    
    if not git_dir.exists():
        return False, "No git repository found - skipping pre-existence check"
    
    try:
        result = subprocess.run(
            ["git", "log", f"-{limit}", "--oneline", "--format=%h %s"],
            cwd=repo_root,
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode != 0:
            return False, f"Git log failed: {result.stderr}"
        
        commits = result.stdout.strip().split('\n') if result.stdout.strip() else []
        return True, f"Recent commits:\n" + "\n".join(f"  {c}" for c in commits[:5])
        
    except subprocess.TimeoutExpired:
        return False, "Git log timed out"
    except Exception as e:
        return False, f"Git log error: {e}"


def check_git_status() -> tuple[bool, str]:
    """
    Check git status for uncommitted changes.
    
    Returns: (success, message)
    """
    repo_root = get_repo_root()
    git_dir = repo_root / ".git"
    
    if not git_dir.exists():
        return False, "No git repository"
    
    try:
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            cwd=repo_root,
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode != 0:
            return False, f"Git status failed: {result.stderr}"
        
        changes = result.stdout.strip().split('\n') if result.stdout.strip() else []
        if changes:
            return True, f"Uncommitted changes: {len(changes)} file(s)"
        return True, "Working tree clean"
        
    except Exception as e:
        return False, f"Git status error: {e}"


def verify_bootstrap_gate_b2() -> List[GateCheck]:
    """
    Verify Gate B2: Pre-Existence Check Gate.
    
    Before investigating a reported issue:
    1. Check git log for recent fixes
    2. Verify the issue still exists
    3. If fixed, document and stop
    4. If not fixed, proceed with investigation
    """
    checks = []
    
    # Check for recent git commits
    has_git, log_message = check_git_log_recent_fixes()
    checks.append(GateCheck(
        name="git_log_check",
        gate="B2",
        passed=True,  # This is informational
        details=log_message,
        can_proceed=True
    ))
    
    # Check git status
    has_git, status_message = check_git_status()
    checks.append(GateCheck(
        name="git_status_check",
        gate="B2",
        passed=has_git,
        details=status_message,
        can_proceed=True
    ))
    
    return checks


# =============================================================================
# GATE B3: Environment Verification Gate
# =============================================================================

def check_go_toolchain() -> GateCheck:
    """
    Gate B3.1: Verify Go toolchain availability.
    
    Check that:
    1. `go` command is available
    2. Go version meets minimum requirement
    
    If Go is not available, provides installation instructions.
    """
    try:
        # Check if go is available
        result = subprocess.run(
            ["which", "go"],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        if result.returncode != 0 or not result.stdout.strip():
            return GateCheck(
                name="go_available",
                gate="B3",
                passed=False,
                details="WARNING: Go toolchain not available. For Go projects, install: curl -sL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz | tar -xz -C $HOME/go --strip-components=1 && echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.bashrc",
                can_proceed=True  # Warning only - can proceed with Python-only checks
            )
        
        go_path = result.stdout.strip()
        
        # Check Go version
        version_result = subprocess.run(
            ["go", "version"],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        if version_result.returncode != 0:
            return GateCheck(
                name="go_version",
                gate="B3",
                passed=False,
                details=f"WARNING: Cannot get Go version",
                can_proceed=True
            )
        
        version = version_result.stdout.strip()
        return GateCheck(
            name="go_toolchain",
            gate="B3",
            passed=True,
            details=f"PASSED: Go available at {go_path} - {version}"
        )
        
    except subprocess.TimeoutExpired:
        return GateCheck(
            name="go_check_timeout",
            gate="B3",
            passed=False,
            details="WARNING: Go check timed out",
            can_proceed=True
        )
    except Exception as e:
        return GateCheck(
            name="go_check_error",
            gate="B3",
            passed=False,
            details=f"WARNING: Go check error: {e}",
            can_proceed=True
        )


def check_go_dependencies() -> GateCheck:
    """
    Gate B3.2: Verify Go module dependencies.
    
    Check that go.mod exists and dependencies can be verified.
    """
    repo_root = get_repo_root()
    go_mod = repo_root / "go.mod"
    
    if not go_mod.exists():
        return GateCheck(
            name="go_mod_exists",
            gate="B3",
            passed=False,
            details="WARNING: go.mod not found - not a Go project or missing go.mod",
            can_proceed=True  # Not a Go project, skip this check
        )
    
    try:
        # Verify dependencies
        result = subprocess.run(
            ["go", "mod", "verify"],
            cwd=repo_root,
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode != 0:
            return GateCheck(
                name="go_mod_verify",
                gate="B3",
                passed=False,
                details=f"WARNING: go mod verify failed: {result.stderr.strip()}",
                can_proceed=True
            )
        
        return GateCheck(
            name="go_dependencies",
            gate="B3",
            passed=True,
            details="PASSED: Go dependencies verified"
        )
        
    except subprocess.TimeoutExpired:
        return GateCheck(
            name="go_deps_timeout",
            gate="B3",
            passed=False,
            details="WARNING: Dependency verification timed out",
            can_proceed=True
        )
    except Exception as e:
        return GateCheck(
            name="go_deps_error",
            gate="B3",
            passed=False,
            details=f"WARNING: Dependency check error: {e}",
            can_proceed=True
        )


def _try_user_install(package: str) -> tuple[bool, str]:
    """
    Try to install a Python package using user-local installation.
    
    Returns: (success, message)
    """
    try:
        result = subprocess.run(
            ["pip", "install", "--user", package],
            capture_output=True,
            text=True,
            timeout=60
        )
        if result.returncode == 0:
            return True, f"Successfully installed {package} to user-local directory"
        else:
            return False, result.stderr.strip() or "Installation failed"
    except Exception as e:
        return False, str(e)


def check_python_runtime(auto_install: bool = True) -> GateCheck:
    """
    Gate B3.3: Verify Python runtime for KDE Runtime.
    
    Check that Python 3.10+ is available and PyYAML is installed.
    If not installed and auto_install is True, try user-local installation.
    """
    try:
        # Check Python version
        version_result = subprocess.run(
            ["python3", "--version"],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        if version_result.returncode != 0:
            return GateCheck(
                name="python_version",
                gate="B3",
                passed=False,
                details="FAILED: Python 3 not available",
                can_proceed=False
            )
        
        python_version = version_result.stdout.strip()
        
        # Check PyYAML
        yaml_result = subprocess.run(
            ["python3", "-c", "import yaml; print(yaml.__version__)"],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        if yaml_result.returncode != 0:
            # PyYAML not found - try auto-install
            if auto_install:
                success, msg = _try_user_install("pyyaml")
                if success:
                    # Verify installation worked
                    yaml_result = subprocess.run(
                        ["python3", "-c", "import yaml; print(yaml.__version__)"],
                        capture_output=True,
                        text=True,
                        timeout=5
                    )
                    if yaml_result.returncode == 0:
                        yaml_version = yaml_result.stdout.strip()
                        return GateCheck(
                            name="python_runtime",
                            gate="B3",
                            passed=True,
                            details=f"PASSED: {python_version}, PyYAML {yaml_version} (auto-installed)"
                        )
            
            # Still not available
            install_cmd = "pip install --user pyyaml"
            return GateCheck(
                name="pyyaml_installed",
                gate="B3",
                passed=False,
                details=f"FAILED: PyYAML not installed. Run: {install_cmd}",
                can_proceed=False
            )
        
        yaml_version = yaml_result.stdout.strip()
        
        return GateCheck(
            name="python_runtime",
            gate="B3",
            passed=True,
            details=f"PASSED: {python_version}, PyYAML {yaml_version}"
        )
        
    except subprocess.TimeoutExpired:
        return GateCheck(
            name="python_check_timeout",
            gate="B3",
            passed=False,
            details="WARNING: Python check timed out",
            can_proceed=True
        )
    except Exception as e:
        return GateCheck(
            name="python_check_error",
            gate="B3",
            passed=False,
            details=f"WARNING: Python check error: {e}",
            can_proceed=True
        )


def verify_bootstrap_gate_b3(project_type: str = "go", quick: bool = False) -> List[GateCheck]:
    """
    Verify Gate B3: Environment Verification Gate.
    
    Before promising test execution:
    1. Verify toolchain availability
    2. Confirm project dependencies
    3. If environment incomplete, note limitation
    4. Do not promise execution without verification
    
    Args:
        project_type: Type of project ("go", "python", etc.)
        quick: If True, skip slow dependency checks (e.g., go mod verify)
    """
    checks = []
    
    # Check Python runtime (always needed for KDE)
    checks.append(check_python_runtime())
    
    # Check Go toolchain for Go projects
    if project_type.lower() == "go":
        checks.append(check_go_toolchain())
        # Skip slow dependency check in quick mode (go mod verify takes ~2s)
        if not quick:
            checks.append(check_go_dependencies())
        else:
            checks.append(GateCheck(
                name="go_deps",
                gate="B3",
                passed=True,
                details="SKIPPED: go mod verify skipped (quick mode). Use --full to run all checks."
            ))
    
    return checks


# =============================================================================
# Main Gate Verification
# =============================================================================

def verify_all_gates(project_type: str = "go", quick: bool = False) -> GateResult:
    """
    Verify all bootstrap gates and return comprehensive result.
    
    Args:
        project_type: Type of project ("go", "python", etc.)
        quick: If True, skip slow dependency checks
        
    Returns:
        GateResult with all check results
    """
    result = GateResult(
        timestamp=datetime.now().isoformat(),
        project_type=project_type
    )
    
    # Gate B1: Bootstrap-First
    for check in verify_bootstrap_gate_b1():
        result.add_check(check)
    
    # Gate B2: Pre-Existence Check
    for check in verify_bootstrap_gate_b2():
        result.add_check(check)
    
    # Gate B3: Environment Verification
    for check in verify_bootstrap_gate_b3(project_type, quick=quick):
        result.add_check(check)
    
    # Generate summary
    passed_count = sum(1 for c in result.checks if c.passed)
    total_count = len(result.checks)
    mode_note = " (QUICK MODE)" if quick else ""
    
    if result.can_proceed:
        result.summary = f"Bootstrap gates verified{mode_note}: {passed_count}/{total_count} checks passed. Can proceed with investigation."
    else:
        result.summary = f"Bootstrap gates FAILED: {passed_count}/{total_count} checks passed. CANNOT proceed until critical gates pass."
    
    return result


def print_gate_result(result: GateResult) -> None:
    """Print a formatted gate result."""
    print("=" * 70)
    print("KDE BOOTSTRAP GATE VERIFICATION")
    print("=" * 70)
    print(f"Timestamp: {result.timestamp}")
    print(f"Project Type: {result.project_type}")
    print()
    
    # Group by gate
    gates = {}
    for check in result.checks:
        if check.gate not in gates:
            gates[check.gate] = []
        gates[check.gate].append(check)
    
    for gate_id in sorted(gates.keys()):
        gate_checks = gates[gate_id]
        print(f"--- Gate {gate_id} ---")
        for check in gate_checks:
            status = "✓" if check.passed else "✗"
            print(f"  [{status}] {check.name}: {check.details}")
        print()
    
    print("=" * 70)
    print(f"RESULT: {'PASSED' if result.can_proceed else 'FAILED'}")
    print(f"Summary: {result.summary}")
    print("=" * 70)


def export_gate_result(result: GateResult, path: Path) -> None:
    """Export gate result to JSON file."""
    data = {
        "timestamp": result.timestamp,
        "project_type": result.project_type,
        "passed": result.passed,
        "can_proceed": result.can_proceed,
        "failed_gates": result.failed_gates,
        "warnings": result.warnings,
        "summary": result.summary,
        "checks": [
            {
                "name": c.name,
                "gate": c.gate,
                "passed": c.passed,
                "details": c.details,
                "can_proceed": c.can_proceed
            }
            for c in result.checks
        ]
    }
    
    with open(path, 'w') as f:
        json.dump(data, f, indent=2)


# =============================================================================
# CLI Interface
# =============================================================================

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="KDE Bootstrap Gate Verification")
    parser.add_argument(
        "--project-type",
        choices=["go", "python", "other"],
        default="go",
        help="Type of project being investigated"
    )
    parser.add_argument(
        "--export",
        type=Path,
        metavar="FILE",
        help="Export result to JSON file"
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output as JSON"
    )
    quick_group = parser.add_mutually_exclusive_group()
    quick_group.add_argument(
        "--quick",
        action="store_true",
        help="Skip slow checks (e.g., go mod verify). Fast but less thorough."
    )
    quick_group.add_argument(
        "--full",
        action="store_true",
        help="Run all checks including slow ones. Default behavior."
    )
    
    args = parser.parse_args()
    
    # --full is the default, --quick enables quick mode
    quick_mode = args.quick
    
    result = verify_all_gates(args.project_type, quick=quick_mode)
    
    if args.json:
        print(json.dumps({
            "timestamp": result.timestamp,
            "project_type": result.project_type,
            "passed": result.passed,
            "can_proceed": result.can_proceed,
            "failed_gates": result.failed_gates,
            "summary": result.summary,
            "checks": [
                {
                    "name": c.name,
                    "gate": c.gate,
                    "passed": c.passed,
                    "details": c.details
                }
                for c in result.checks
            ]
        }, indent=2))
    else:
        print_gate_result(result)
    
    if args.export:
        export_gate_result(result, args.export)
        print(f"\nResult exported to: {args.export}")
    
    # Exit with appropriate code
    sys.exit(0 if result.can_proceed else 1)
