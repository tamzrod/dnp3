"""
File Boundary Guard

Active enforcement of file boundary rules during runtime.
Checks all file write operations against permitted boundaries.
"""

import os
from pathlib import Path
from typing import List, Optional, Tuple
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class ViolationSeverity(Enum):
    """Severity levels for violations."""
    ALLOWED = "allowed"
    BLOCKED = "blocked"
    WARNING = "warning"


@dataclass
class FileOperation:
    """Represents a file operation to check."""
    operation: str  # create, write, delete, str_replace, insert
    path: str
    details: str = ""


@dataclass
class BoundaryCheckResult:
    """Result of a boundary check."""
    allowed: bool
    path: str
    operation: str
    violation: bool
    severity: ViolationSeverity
    reason: str
    requires_approval: bool
    timestamp: str = field(default_factory=lambda: datetime.now().isoformat())


class FileBoundaryGuard:
    """
    Active enforcement of file boundary rules.
    
    Checks all file write operations against permitted boundaries:
    - /laboratory/** - Allowed (Laboratory Rules apply)
    - /runtime/logs/** - Allowed (exempt)
    - Everything else - Requires human approval
    """
    
    # Base path for KDE
    KDE_ROOT = "/workspace/project/kde"
    
    # Allowed path prefixes
    ALLOWED_PREFIXES = [
        "/workspace/project/kde/laboratory/",
        "/workspace/project/kde/runtime/logs/",
    ]
    
    # Exempt specific files
    EXEMPT_FILES = [
        "/workspace/project/kde/runtime/state.json",
        "/workspace/project/kde/runtime/catalog.json",
        "/workspace/project/kde/runtime/aliases/audit.log",
        "/workspace/project/kde/runtime/aliases/discovery.log",
    ]
    
    # Write operations that require checking
    WRITE_OPERATIONS = {
        "create", "write", "str_replace", "delete", 
        "insert", "mkdir", "touch", "move", "copy"
    }
    
    def __init__(self, kde_root: str = "/workspace/project/kde"):
        self.kde_root = kde_root
        self.violations: List[BoundaryCheckResult] = []
        self.checks: List[BoundaryCheckResult] = []
    
    def check_path(self, path: str, operation: str) -> BoundaryCheckResult:
        """
        Check if a file path is within allowed boundaries.
        
        Args:
            path: The file path to check
            operation: The operation being performed
            
        Returns:
            BoundaryCheckResult with violation details
        """
        # Normalize path
        path = os.path.abspath(path)
        
        # Check exempt files first
        if path in self.EXEMPT_FILES:
            return self._allowed_result(
                path, operation, 
                f"Exempt file: {os.path.basename(path)}"
            )
        
        # Check allowed prefixes
        for prefix in self.ALLOWED_PREFIXES:
            if path.startswith(prefix):
                return self._allowed_result(
                    path, operation,
                    f"Within allowed path: {prefix}"
                )
        
        # Check if inside /laboratory/ (dynamic check)
        lab_path = os.path.join(self.kde_root, "laboratory")
        if path.startswith(lab_path + "/") or path == lab_path:
            return self._allowed_result(
                path, operation,
                f"Inside /laboratory/: {path}"
            )
        
        # VIOLATION - Outside all allowed paths
        return self._violation_result(
            path, operation,
            f"Outside /laboratory/: {path}"
        )
    
    def check_operation(self, operation: str, path: str) -> BoundaryCheckResult:
        """
        Check a file operation.
        
        Args:
            operation: The operation type
            path: The file path
            
        Returns:
            BoundaryCheckResult
        """
        result = BoundaryCheckResult(
            allowed=False,
            path=path,
            operation=operation,
            violation=False,
            severity=ViolationSeverity.ALLOWED,
            reason="",
            requires_approval=False
        )
        
        # Only check write operations
        if operation.lower() not in self.WRITE_OPERATIONS:
            result.allowed = True
            result.reason = "Read operation - not restricted"
            return result
        
        # Check the path
        result = self.check_path(path, operation)
        
        # Log the check
        self.checks.append(result)
        
        return result
    
    def is_allowed(self, path: str, operation: str) -> Tuple[bool, str]:
        """
        Quick check if operation is allowed.
        
        Returns:
            Tuple of (allowed, reason)
        """
        result = self.check_operation(operation, path)
        return result.allowed, result.reason
    
    def _allowed_result(self, path: str, operation: str, reason: str) -> BoundaryCheckResult:
        """Create an allowed result."""
        return BoundaryCheckResult(
            allowed=True,
            path=path,
            operation=operation,
            violation=False,
            severity=ViolationSeverity.ALLOWED,
            reason=reason,
            requires_approval=False
        )
    
    def _violation_result(self, path: str, operation: str, reason: str) -> BoundaryCheckResult:
        """Create a violation result."""
        return BoundaryCheckResult(
            allowed=False,
            path=path,
            operation=operation,
            violation=True,
            severity=ViolationSeverity.BLOCKED,
            reason=reason,
            requires_approval=True
        )
    
    def get_violations(self) -> List[BoundaryCheckResult]:
        """Get all violations detected."""
        return [c for c in self.checks if c.violation]
    
    def get_stats(self) -> dict:
        """Get enforcement statistics."""
        violations = self.get_violations()
        return {
            "total_checks": len(self.checks),
            "violations": len(violations),
            "allowed": len(self.checks) - len(violations),
            "violation_rate": len(violations) / len(self.checks) if self.checks else 0
        }
    
    def format_violation_message(self, result: BoundaryCheckResult) -> str:
        """Format a violation message for display."""
        return f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║  ⚠️  FILE BOUNDARY VIOLATION                                                ║
║  ════════════════════════════════════════════                                ║
║                                                                              ║
║  Attempted Action:  {result.operation:<50}           ║
║  Target Path:       {result.path:<50}                 ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Rule: No files written outside /laboratory/ without human approval         ║
║                                                                              ║
║  Reason: {result.reason:<60}          ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Required Action: Human authorization to override                             ║
║                                                                              ║
║  [Override: Yes]  [Block: No]                                               ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
"""


def create_guard(kde_root: str = "/workspace/project/kde") -> FileBoundaryGuard:
    """Factory function to create a FileBoundaryGuard."""
    return FileBoundaryGuard(kde_root)
