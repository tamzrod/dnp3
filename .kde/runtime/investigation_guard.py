"""
Investigation Artifact Guard

Active enforcement of investigation artifact requirements during runtime.
Blocks code changes unless investigation/experiment artifacts exist.

REC-001: From KDE-INV-GOV-SYN
"""

import os
import re
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
class InvestigationCheckResult:
    """Result of an investigation artifact check."""
    allowed: bool
    tool_name: str
    operation: str
    violation: bool
    severity: ViolationSeverity
    reason: str
    requires_approval: bool
    investigation_id: Optional[str] = None
    experiment_id: Optional[str] = None
    timestamp: str = field(default_factory=lambda: datetime.now().isoformat())


class InvestigationArtifactGuard:
    """
    Enforces investigation artifact existence before code changes.
    
    Blocks terminal/file_editor/browser_navigate unless:
    - An investigation artifact exists in the current working directory
    - The working directory is within /laboratory/
    
    Pattern follows FileBoundaryGuard for consistency.
    """
    
    # Base path for KDE
    KDE_ROOT = "/workspace/project/kde"
    LABORATORY_PATH = "/laboratory"
    
    # Tools that require investigation context
    BLOCKED_TOOLS = {
        "terminal",      # Running code
        "file_editor",   # Modifying code
        "browser_navigate",  # External access
    }
    
    # Tools that are always allowed
    EXEMPT_TOOLS = {
        "invoke_skill",
        "task_tracker",
        "file_editor",  # Can always edit .kde files
    }
    
    def __init__(self, kde_root: str = "/workspace/project/dnp3"):
        self.kde_root = kde_root
        self.checks: List[InvestigationCheckResult] = []
        self.current_investigation: Optional[str] = None
        self.current_experiment: Optional[str] = None
    
    def check_tool_use(
        self, 
        tool_name: str, 
        args: dict,
        working_dir: str = None
    ) -> InvestigationCheckResult:
        """
        Check if a tool can be used based on investigation context.
        
        Args:
            tool_name: The name of the tool being used
            args: The tool arguments
            working_dir: Current working directory
            
        Returns:
            InvestigationCheckResult with violation details
        """
        # Initialize working directory
        if working_dir is None:
            working_dir = os.getcwd()
        
        working_dir = os.path.abspath(working_dir)
        
        # Check if tool is exempt
        if tool_name in self.EXEMPT_TOOLS:
            # Special case: file_editor can edit .kde files
            if tool_name == "file_editor" and args.get("path", "").startswith(self.kde_root + "/.kde"):
                return self._allowed_result(
                    tool_name, "edit",
                    "Editing .kde files is always allowed"
                )
            return self._allowed_result(
                tool_name, args.get("command", "unknown"),
                "Tool is exempt from investigation check"
            )
        
        # Check if tool requires investigation context
        if tool_name not in self.BLOCKED_TOOLS:
            return self._allowed_result(
                tool_name, args.get("command", "unknown"),
                "Tool does not require investigation context"
            )
        
        # Detect investigation context from path
        investigation_id, experiment_id = self._detect_investigation_context(working_dir)
        
        # Check if we're in the laboratory
        lab_path = os.path.join(self.kde_root, self.LABORATORY_PATH.lstrip("/"))
        if not working_dir.startswith(lab_path):
            return self._violation_result(
                tool_name, args.get("command", "unknown"),
                f"Outside {self.LABORATORY_PATH}/: {working_dir}",
                investigation_id, experiment_id
            )
        
        # If we're in laboratory but no investigation detected
        if investigation_id is None:
            return self._violation_result(
                tool_name, args.get("command", "unknown"),
                "No investigation artifact found in current path. Create investigation first.",
                None, None
            )
        
        # Investigation found - allowed
        self.current_investigation = investigation_id
        self.current_experiment = experiment_id
        
        return self._allowed_result(
            tool_name, args.get("command", "unknown"),
            f"Investigation: {investigation_id}" + 
            (f", Experiment: {experiment_id}" if experiment_id else "")
        )
    
    def _detect_investigation_context(self, path: str) -> Tuple[Optional[str], Optional[str]]:
        """
        Detect investigation and experiment IDs from path.
        
        Expected patterns:
        - /laboratory/investigations/INV-XXX/
        - /laboratory/investigations/INV-XXX/experiments/LAB-XXX/
        - /laboratory/experiments/LAB-XXX/
        """
        investigation_id = None
        experiment_id = None
        
        # Pattern for investigations
        inv_pattern = r'/investigations/(INV-\w+)/'
        match = re.search(inv_pattern, path)
        if match:
            investigation_id = match.group(1)
        
        # Pattern for experiments in investigations
        exp_in_inv_pattern = r'/experiments/(LAB-\w+)/'
        match = re.search(exp_in_inv_pattern, path)
        if match:
            experiment_id = match.group(1)
        
        # Pattern for standalone experiments
        exp_pattern = r'/experperiments/(LAB-\w+)/'
        match = re.search(exp_pattern, path)
        if match and experiment_id is None:
            experiment_id = match.group(1)
        
        return investigation_id, experiment_id
    
    def _allowed_result(
        self, 
        tool_name: str, 
        operation: str, 
        reason: str
    ) -> InvestigationCheckResult:
        """Create an allowed result."""
        result = InvestigationCheckResult(
            allowed=True,
            tool_name=tool_name,
            operation=operation,
            violation=False,
            severity=ViolationSeverity.ALLOWED,
            reason=reason,
            requires_approval=False,
            investigation_id=self.current_investigation,
            experiment_id=self.current_experiment
        )
        self.checks.append(result)
        return result
    
    def _violation_result(
        self, 
        tool_name: str, 
        operation: str, 
        reason: str,
        investigation_id: Optional[str],
        experiment_id: Optional[str]
    ) -> InvestigationCheckResult:
        """Create a violation result."""
        result = InvestigationCheckResult(
            allowed=False,
            tool_name=tool_name,
            operation=operation,
            violation=True,
            severity=ViolationSeverity.BLOCKED,
            reason=reason,
            requires_approval=False,  # BLOCK, don't ask
            investigation_id=investigation_id,
            experiment_id=experiment_id
        )
        self.checks.append(result)
        return result
    
    def get_violations(self) -> List[InvestigationCheckResult]:
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
    
    def format_violation_message(self, result: InvestigationCheckResult) -> str:
        """Format a violation message for display."""
        return f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║  ⚠️  INVESTIGATION ARTIFACT REQUIRED                                        ║
║  ════════════════════════════════════════════════════                        ║
║                                                                              ║
║  Attempted Action:  {result.tool_name:<40}                       ║
║  Operation:        {result.operation:<40}                       ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Rule: No code changes allowed without investigation artifact.              ║
║                                                                              ║
║  Reason: {result.reason:<60}          ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Required Action:                                                            ║
║  1. Create investigation artifact first (laboratory/investigations/)        ║
║  2. Navigate to investigation directory                                      ║
║  3. Retry operation                                                         ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
"""
    
    def get_current_context(self) -> dict:
        """Get current investigation context."""
        return {
            "investigation_id": self.current_investigation,
            "experiment_id": self.current_experiment,
            "has_context": self.current_investigation is not None
        }


def create_guard(kde_root: str = "/workspace/project/dnp3") -> InvestigationArtifactGuard:
    """Factory function to create an InvestigationArtifactGuard."""
    return InvestigationArtifactGuard(kde_root)
