"""
Violation Handler

Handles detected violations by presenting them to humans for approval.
All decisions are logged for audit.
"""

import json
import os
from pathlib import Path
from datetime import datetime
from typing import Optional, List, Dict
from dataclasses import dataclass, asdict

from .file_boundary_guard import FileBoundaryGuard, BoundaryCheckResult


@dataclass
class ViolationRecord:
    """Record of a violation and its resolution."""
    timestamp: str
    operation: str
    path: str
    reason: str
    resolved: bool
    approved: bool
    resolved_by: Optional[str]
    resolved_at: Optional[str]


class ViolationHandler:
    """
    Handles violations by presenting them to humans for approval.
    
    When a violation is detected:
    1. Log the violation
    2. Present to human
    3. Record decision
    4. Return approval status
    """
    
    def __init__(self, kde_root: str = "/workspace/project/kde"):
        self.kde_root = kde_root
        self.log_dir = Path(kde_root) / "runtime" / "logs" / "violations"
        self.log_dir.mkdir(parents=True, exist_ok=True)
        
        self.guard = FileBoundaryGuard(kde_root)
        self.pending_violations: List[BoundaryCheckResult] = []
        self.resolved_violations: Dict[str, ViolationRecord] = {}
    
    def check(self, operation: str, path: str) -> BoundaryCheckResult:
        """
        Check an operation and handle if violation.
        
        Args:
            operation: The operation type
            path: The file path
            
        Returns:
            BoundaryCheckResult
        """
        result = self.guard.check_operation(operation, path)
        
        if result.violation:
            self.pending_violations.append(result)
        
        return result
    
    def handle_violation(self, result: BoundaryCheckResult) -> bool:
        """
        Handle a violation by presenting to human.
        
        This is the interactive version that would be called
        when a user instruction triggers a violation.
        
        Args:
            result: The violation to handle
            
        Returns:
            True if approved, False if blocked
        """
        # Log the pending violation
        self.pending_violations.append(result)
        
        # Present violation message
        message = self.guard.format_violation_message(result)
        print(message)
        
        # In the current implementation, we return False (block by default)
        # The actual interactive approval would happen via user input
        return False
    
    def approve(self, path: str, approved_by: str = "human") -> bool:
        """
        Approve a pending violation.
        
        Args:
            path: The path that was violated
            approved_by: Who approved (must be human)
            
        Returns:
            True if approved
        """
        if approved_by.lower() == "ai":
            # SECURITY: AI cannot approve violations
            print("❌ SECURITY: AI cannot approve violations. Human authorization required.")
            return False
        
        # Find the pending violation
        for v in self.pending_violations:
            if v.path == path:
                # Log the approval
                record = self._log_resolution(v, approved=True, resolved_by=approved_by)
                self.resolved_violations[path] = record
                return True
        
        return False
    
    def block(self, path: str) -> bool:
        """
        Block a pending violation.
        
        Args:
            path: The path that was violated
            
        Returns:
            True if blocked
        """
        # Find the pending violation
        for v in self.pending_violations:
            if v.path == path:
                # Log the block
                record = self._log_resolution(v, approved=False, resolved_by="system")
                self.resolved_violations[path] = record
                return True
        
        return False
    
    def is_approved(self, path: str) -> bool:
        """Check if a path has been approved."""
        return path in self.resolved_violations and self.resolved_violations[path].approved
    
    def get_pending(self) -> List[BoundaryCheckResult]:
        """Get all pending violations."""
        return self.pending_violations.copy()
    
    def get_resolved(self) -> Dict[str, ViolationRecord]:
        """Get all resolved violations."""
        return self.resolved_violations.copy()
    
    def _log_resolution(
        self, 
        result: BoundaryCheckResult, 
        approved: bool, 
        resolved_by: str
    ) -> ViolationRecord:
        """Log the resolution of a violation."""
        record = ViolationRecord(
            timestamp=datetime.now().isoformat(),
            operation=result.operation,
            path=result.path,
            reason=result.reason,
            resolved=True,
            approved=approved,
            resolved_by=resolved_by,
            resolved_at=datetime.now().isoformat()
        )
        
        # Write to log file
        log_file = self._get_log_file(result.path)
        with open(log_file, 'w') as f:
            json.dump(asdict(record), f, indent=2)
        
        return record
    
    def _get_log_file(self, path: str) -> Path:
        """Get log file path for a violation."""
        # Create a safe filename from the path
        safe_name = path.replace("/", "_").replace(".", "_").replace("-", "_")
        return self.log_dir / f"violation_{safe_name}.json"
    
    def get_audit_log(self) -> str:
        """Generate an audit log of all violations."""
        lines = []
        lines.append("=" * 80)
        lines.append("FILE BOUNDARY VIOLATION AUDIT LOG")
        lines.append("=" * 80)
        lines.append("")
        
        if not self.resolved_violations:
            lines.append("No violations recorded.")
        else:
            for path, record in self.resolved_violations.items():
                status = "✅ APPROVED" if record.approved else "❌ BLOCKED"
                lines.append(f"Path: {record.path}")
                lines.append(f"Operation: {record.operation}")
                lines.append(f"Status: {status}")
                lines.append(f"Resolved by: {record.resolved_by}")
                lines.append(f"Time: {record.resolved_at}")
                lines.append("-" * 80)
        
        lines.append("")
        lines.append("=" * 80)
        
        return "\n".join(lines)
    
    def get_stats(self) -> dict:
        """Get handler statistics."""
        approved = sum(1 for r in self.resolved_violations.values() if r.approved)
        blocked = sum(1 for r in self.resolved_violations.values() if not r.approved)
        
        return {
            "pending": len(self.pending_violations),
            "resolved": len(self.resolved_violations),
            "approved": approved,
            "blocked": blocked,
            "approval_rate": approved / len(self.resolved_violations) if self.resolved_violations else 0
        }


def create_handler(kde_root: str = "/workspace/project/kde") -> ViolationHandler:
    """Factory function to create a ViolationHandler."""
    return ViolationHandler(kde_root)
