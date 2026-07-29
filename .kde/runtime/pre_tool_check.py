"""
Pre-Tool Check Hook

Enforces investigation artifact requirements before tool execution.
Integrates InvestigationArtifactGuard with tool execution layer.

REC-002: From KDE-INV-GOV-SYN
"""

import os
from typing import Optional, Callable, Any, Tuple
from dataclasses import dataclass

from .investigation_guard import (
    InvestigationArtifactGuard,
    InvestigationCheckResult,
    ViolationSeverity,
)


@dataclass
class ToolCheckResult:
    """Result of a pre-tool check."""
    allowed: bool
    tool_name: str
    reason: str
    violation_message: Optional[str] = None
    investigation_context: Optional[dict] = None


class PreToolCheck:
    """
    Pre-execution check for tools requiring investigation context.
    
    Usage:
        check = PreToolCheck()
        result = check.verify_tool("terminal", {"command": "go run test.go"})
        if not result.allowed:
            print(result.violation_message)
            return  # Block execution
        # Continue with tool execution
    """
    
    def __init__(self, kde_root: str = "/workspace/project/dnp3"):
        self.guard = InvestigationArtifactGuard(kde_root)
        self.enabled = True
        self._override_allowed = False  # For human override
    
    def verify_tool(
        self, 
        tool_name: str, 
        args: dict,
        working_dir: Optional[str] = None
    ) -> ToolCheckResult:
        """
        Verify if a tool can be executed.
        
        Args:
            tool_name: Name of the tool
            args: Tool arguments
            working_dir: Current working directory
            
        Returns:
            ToolCheckResult with permission and context
        """
        if not self.enabled:
            return ToolCheckResult(
                allowed=True,
                tool_name=tool_name,
                reason="Checks disabled"
            )
        
        if working_dir is None:
            working_dir = os.getcwd()
        
        check_result = self.guard.check_tool_use(tool_name, args, working_dir)
        
        if check_result.allowed:
            return ToolCheckResult(
                allowed=True,
                tool_name=tool_name,
                reason=check_result.reason,
                investigation_context=self.guard.get_current_context()
            )
        
        return ToolCheckResult(
            allowed=False,
            tool_name=tool_name,
            reason=check_result.reason,
            violation_message=self.guard.format_violation_message(check_result),
            investigation_context=None
        )
    
    def verify_or_raise(
        self, 
        tool_name: str, 
        args: dict,
        working_dir: Optional[str] = None
    ) -> dict:
        """
        Verify tool or raise exception if not allowed.
        
        Args:
            tool_name: Name of the tool
            args: Tool arguments
            working_dir: Current working directory
            
        Returns:
            Investigation context if allowed
            
        Raises:
            InvestigationViolation: If tool is not allowed
        """
        result = self.verify_tool(tool_name, args, working_dir)
        
        if result.allowed:
            return result.investigation_context
        
        raise InvestigationViolation(
            f"Tool '{tool_name}' blocked: {result.reason}\n"
            f"{result.violation_message}"
        )
    
    def get_context(self) -> dict:
        """Get current investigation context."""
        return self.guard.get_current_context()
    
    def disable(self):
        """Disable pre-tool checks."""
        self.enabled = False
    
    def enable(self):
        """Enable pre-tool checks."""
        self.enabled = True
    
    def allow_for_human(self):
        """Allow human to override checks (must be human-initiated)."""
        self._override_allowed = True
    
    def is_override_allowed(self) -> bool:
        """Check if override is allowed."""
        return self._override_allowed


class InvestigationViolation(Exception):
    """Exception raised when investigation artifact check fails."""
    pass


# Global instance for runtime use
_global_check: Optional[PreToolCheck] = None


def get_pre_tool_check(kde_root: str = "/workspace/project/dnp3") -> PreToolCheck:
    """Get or create the global PreToolCheck instance."""
    global _global_check
    if _global_check is None:
        _global_check = PreToolCheck(kde_root)
    return _global_check


def check_tool(
    tool_name: str, 
    args: dict,
    working_dir: Optional[str] = None
) -> ToolCheckResult:
    """
    Convenience function to check a tool.
    
    Usage:
        result = check_tool("terminal", {"command": "go run test.go"})
        if not result.allowed:
            print(result.violation_message)
    """
    checker = get_pre_tool_check()
    return checker.verify_tool(tool_name, args, working_dir)


def require_investigation(
    tool_name: str,
    args: dict,
    working_dir: Optional[str] = None
) -> dict:
    """
    Convenience function to require investigation context.
    
    Usage:
        ctx = require_investigation("terminal", {"command": "go run test.go"})
        # Safe to execute tool
    
    Raises:
        InvestigationViolation: If no investigation context
    """
    checker = get_pre_tool_check()
    return checker.verify_or_raise(tool_name, args, working_dir)
