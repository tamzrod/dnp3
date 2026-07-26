"""
KDE Verification System

This module implements verification checks for KDE governance compliance.

Verification Types:
- Process Verification: Verify investigation processes
- Output Verification: Verify decision outputs
- Compliance Verification: Verify governance compliance
- Quality Verification: Verify artifact quality
"""

import os
import json
from pathlib import Path
from dataclasses import dataclass
from typing import List, Dict, Optional, Any
from datetime import datetime


@dataclass
class VerificationCheck:
    """Result of a single verification check."""
    check_id: str
    check_type: str  # process, output, compliance, quality
    name: str
    passed: bool
    details: str
    severity: str = "ERROR"  # ERROR, WARNING, INFO


@dataclass
class VerificationResult:
    """Result of verification run."""
    timestamp: str
    checks: List[VerificationCheck] = None
    passed: bool = True
    errors: List[str] = None
    warnings: List[str] = None
    summary: str = ""
    
    def __post_init__(self):
        if self.checks is None:
            self.checks = []
        if self.errors is None:
            self.errors = []
        if self.warnings is None:
            self.warnings = []


def get_repo_root() -> Path:
    """Get the repository root directory."""
    current = Path(__file__).resolve() if '__file__' in dir() else Path.cwd()
    if '.kde' in current.parts:
        idx = current.parts.index('.kde')
        return Path(*current.parts[:idx])
    return Path.cwd()


# =============================================================================
# Compliance Verification
# =============================================================================

def verify_artifact_naming(artifact_path: Path) -> VerificationCheck:
    """
    Verify artifact follows naming conventions.
    
    Validates:
    - Investigation: KDE-INV-*, PROJECT-INV-*
    - Experiment: PROJECT-EXP-*
    - Decision: TDR-*
    """
    name = artifact_path.name
    
    # Check for known prefixes
    valid_prefixes = [
        'KDE-INV-', 'PROJECT-INV-', 'PROJECT-EXP-', 
        'DNP3-INV-', 'DNP3-EXP-', 'TDR-'
    ]
    
    for prefix in valid_prefixes:
        if name.startswith(prefix):
            return VerificationCheck(
                check_id="naming",
                check_type="compliance",
                name=f"Artifact naming: {name}",
                passed=True,
                details=f"Valid prefix: {prefix}"
            )
    
    return VerificationCheck(
        check_id="naming",
        check_type="compliance",
        name=f"Artifact naming: {name}",
        passed=False,
        details=f"Invalid naming - no known prefix"
    )


def verify_investigation_structure(inv_path: Path) -> List[VerificationCheck]:
    """
    Verify investigation follows standard structure.
    
    Required files:
    - README.md
    - SPEC.md
    - CONCLUSION.md
    """
    checks = []
    required_files = ['README.md', 'SPEC.md', 'CONCLUSION.md']
    
    for req_file in required_files:
        file_path = inv_path / req_file
        if file_path.exists():
            checks.append(VerificationCheck(
                check_id="structure",
                check_type="compliance",
                name=f"Investigation structure: {req_file}",
                passed=True,
                details=f"Found: {req_file}"
            ))
        else:
            checks.append(VerificationCheck(
                check_id="structure",
                check_type="compliance",
                name=f"Investigation structure: {req_file}",
                passed=False,
                details=f"Missing required file: {req_file}",
                severity="ERROR"
            ))
    
    return checks


def verify_experiment_structure(exp_path: Path) -> List[VerificationCheck]:
    """
    Verify experiment follows standard structure.
    
    Required files:
    - README.md
    - SPEC.md
    - CONCLUSION.md
    """
    checks = []
    required_files = ['README.md', 'SPEC.md', 'CONCLUSION.md']
    
    for req_file in required_files:
        file_path = exp_path / req_file
        if file_path.exists():
            checks.append(VerificationCheck(
                check_id="structure",
                check_type="compliance",
                name=f"Experiment structure: {req_file}",
                passed=True,
                details=f"Found: {req_file}"
            ))
        else:
            checks.append(VerificationCheck(
                check_id="structure",
                check_type="compliance",
                name=f"Experiment structure: {req_file}",
                passed=False,
                details=f"Missing required file: {req_file}",
                severity="ERROR"
            ))
    
    return checks


def verify_policy_documents() -> List[VerificationCheck]:
    """
    Verify required policy documents exist.
    """
    checks = []
    repo_root = get_repo_root()
    gov_dir = repo_root / ".kde" / "governance"
    
    required_policies = ['NAMING-CONVENTIONS.md']
    recommended_policies = ['DEP-001.md', 'ENV-001.md']
    
    for policy in required_policies:
        policy_path = gov_dir / policy
        if policy_path.exists():
            checks.append(VerificationCheck(
                check_id="policy",
                check_type="compliance",
                name=f"Required policy: {policy}",
                passed=True,
                details=f"Found: {policy}"
            ))
        else:
            checks.append(VerificationCheck(
                check_id="policy",
                check_type="compliance",
                name=f"Required policy: {policy}",
                passed=False,
                details=f"Missing required policy: {policy}",
                severity="ERROR"
            ))
    
    for policy in recommended_policies:
        policy_path = gov_dir / policy
        if policy_path.exists():
            checks.append(VerificationCheck(
                check_id="policy",
                check_type="compliance",
                name=f"Recommended policy: {policy}",
                passed=True,
                details=f"Found: {policy}"
            ))
        else:
            checks.append(VerificationCheck(
                check_id="policy",
                check_type="compliance",
                name=f"Recommended policy: {policy}",
                passed=True,  # Not required
                details=f"Recommended but not found: {policy}",
                severity="WARNING"
            ))
    
    return checks


def verify_bootstrap_gates() -> List[VerificationCheck]:
    """
    Verify bootstrap gates exist and are documented.
    """
    checks = []
    repo_root = get_repo_root()
    bootstrap_dir = repo_root / ".kde" / "bootstrap"
    
    # Check for gates.py
    gates_file = bootstrap_dir / "gates.py"
    if gates_file.exists():
        checks.append(VerificationCheck(
            check_id="gates",
            check_type="compliance",
            name="Bootstrap gates implementation",
            passed=True,
            details="Found: gates.py"
        ))
    else:
        checks.append(VerificationCheck(
            check_id="gates",
            check_type="compliance",
            name="Bootstrap gates implementation",
            passed=False,
            details="Missing: gates.py",
            severity="ERROR"
        ))
    
    # Check for requirements.json with dependencies
    req_file = bootstrap_dir / "requirements.json"
    if req_file.exists():
        try:
            with open(req_file) as f:
                req_data = json.load(f)
            if 'python_dependencies' in req_data:
                checks.append(VerificationCheck(
                    check_id="gates",
                    check_type="compliance",
                    name="Dependency documentation",
                    passed=True,
                    details="Found: python_dependencies in requirements.json"
                ))
            else:
                checks.append(VerificationCheck(
                    check_id="gates",
                    check_type="compliance",
                    name="Dependency documentation",
                    passed=False,
                    details="Missing: python_dependencies in requirements.json",
                    severity="WARNING"
                ))
        except:
            checks.append(VerificationCheck(
                check_id="gates",
                check_type="compliance",
                name="Dependency documentation",
                passed=False,
                details="Cannot parse requirements.json",
                severity="WARNING"
            ))
    
    return checks


# =============================================================================
# Quality Verification
# =============================================================================

def verify_investigation_quality(inv_path: Path) -> List[VerificationCheck]:
    """
    Verify investigation meets quality standards.
    
    Checks:
    - Frontmatter has required fields
    - Evidence section exists
    - Conclusion is present
    """
    checks = []
    
    readme_path = inv_path / "README.md"
    if not readme_path.exists():
        return checks
    
    try:
        with open(readme_path) as f:
            content = f.read()
        
        # Check for frontmatter
        if content.startswith('---'):
            checks.append(VerificationCheck(
                check_id="quality",
                check_type="quality",
                name="Frontmatter present",
                passed=True,
                details="Found YAML frontmatter"
            ))
            
            # Check for required frontmatter fields
            required_fields = ['id:', 'type:', 'title:', 'status:']
            for field in required_fields:
                if field in content:
                    checks.append(VerificationCheck(
                        check_id="quality",
                        check_type="quality",
                        name=f"Field: {field.strip(':')}",
                        passed=True,
                        details=f"Found: {field}"
                    ))
                else:
                    checks.append(VerificationCheck(
                        check_id="quality",
                        check_type="quality",
                        name=f"Field: {field.strip(':')}",
                        passed=False,
                        details=f"Missing: {field}",
                        severity="WARNING"
                    ))
        else:
            checks.append(VerificationCheck(
                check_id="quality",
                check_type="quality",
                name="Frontmatter present",
                passed=False,
                details="No YAML frontmatter found",
                severity="WARNING"
            ))
        
        # Check for evidence section
        if "## Evidence" in content or "## Research Questions" in content:
            checks.append(VerificationCheck(
                check_id="quality",
                check_type="quality",
                name="Evidence section",
                passed=True,
                details="Found evidence/research section"
            ))
        else:
            checks.append(VerificationCheck(
                check_id="quality",
                check_type="quality",
                name="Evidence section",
                passed=False,
                details="No evidence section found",
                severity="WARNING"
            ))
        
    except Exception as e:
        checks.append(VerificationCheck(
            check_id="quality",
            check_type="quality",
            name="Quality check error",
            passed=False,
            details=f"Error: {str(e)}",
            severity="WARNING"
        ))
    
    return checks


# =============================================================================
# Main Verification
# =============================================================================

def verify_all() -> VerificationResult:
    """
    Run all verification checks.
    
    Returns:
        VerificationResult with all check results
    """
    result = VerificationResult(timestamp=datetime.now().isoformat())
    repo_root = get_repo_root()
    
    # Compliance checks
    result.checks.extend(verify_policy_documents())
    result.checks.extend(verify_bootstrap_gates())
    
    # Artifact structure checks
    lab_dir = repo_root / "laboratory"
    if lab_dir.exists():
        investigations_dir = lab_dir / "investigations"
        if investigations_dir.exists():
            for inv in investigations_dir.iterdir():
                if inv.is_dir():
                    result.checks.extend(verify_investigation_structure(inv))
                    result.checks.extend(verify_investigation_quality(inv))
        
        experiments_dir = lab_dir / "experiments"
        if experiments_dir.exists():
            for exp in experiments_dir.iterdir():
                if exp.is_dir():
                    result.checks.extend(verify_experiment_structure(exp))
    
    # Calculate summary
    errors = [c for c in result.checks if c.severity == "ERROR" and not c.passed]
    warnings = [c for c in result.checks if c.severity == "WARNING" and not c.passed]
    
    result.errors = [f"{c.name}: {c.details}" for c in errors]
    result.warnings = [f"{c.name}: {c.details}" for c in warnings]
    result.passed = len(errors) == 0
    
    result.summary = f"Verification complete: {len(result.checks)} checks, {len(errors)} errors, {len(warnings)} warnings"
    
    return result


def print_verification_result(result: VerificationResult) -> None:
    """Print formatted verification result."""
    print("=" * 70)
    print("KDE VERIFICATION RESULT")
    print("=" * 70)
    print(f"Timestamp: {result.timestamp}")
    print()
    
    if result.errors:
        print("ERRORS:")
        for err in result.errors:
            print(f"  ✗ {err}")
        print()
    
    if result.warnings:
        print("WARNINGS:")
        for warn in result.warnings:
            print(f"  ⚠ {warn}")
        print()
    
    if not result.errors and not result.warnings:
        print("✓ All checks passed!")
        print()
    
    print("=" * 70)
    print(f"RESULT: {'PASSED' if result.passed else 'FAILED'}")
    print(f"Summary: {result.summary}")
    print("=" * 70)


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="KDE Verification")
    parser.add_argument("--json", action="store_true", help="Output as JSON")
    
    args = parser.parse_args()
    
    result = verify_all()
    
    if args.json:
        print(json.dumps({
            "timestamp": result.timestamp,
            "passed": result.passed,
            "errors": result.errors,
            "warnings": result.warnings,
            "summary": result.summary,
            "checks": [
                {
                    "check_id": c.check_id,
                    "check_type": c.check_type,
                    "name": c.name,
                    "passed": c.passed,
                    "details": c.details,
                    "severity": c.severity
                }
                for c in result.checks
            ]
        }, indent=2))
    else:
        print_verification_result(result)
    
    import sys
    sys.exit(0 if result.passed else 1)
