"""
Alias Command Resolver

Integrates alias resolution with KDE Runtime commands.
Provides a command parser that supports both canonical commands and aliases.
"""

import re
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple, Callable
from enum import Enum

from .registry import (
    AliasCategory,
    AliasEntry,
    AliasRegistry,
    get_registry
)


class CommandType(Enum):
    """Types of resolved commands."""
    CANONICAL = "canonical"
    ALIAS = "alias"
    UNKNOWN = "unknown"


@dataclass
class ResolvedCommand:
    """Result of resolving a command."""
    command_type: CommandType
    original_input: str
    resolved_command: str
    category: Optional[AliasCategory] = None
    entry: Optional[AliasEntry] = None
    deprecated: bool = False
    deprecation_warning: Optional[str] = None


class CommandParser:
    """
    Command parser with alias support.
    
    Parses user input and resolves aliases to canonical commands.
    """
    
    # Reserved words that should never be treated as aliases
    RESERVED_WORDS = {'help', 'exit', 'quit', 'debug', 'status'}
    
    def __init__(self, registry: Optional[AliasRegistry] = None):
        """
        Initialize command parser.
        
        Args:
            registry: AliasRegistry instance. Uses global registry if None.
        """
        self.registry = registry or get_registry()
    
    def parse(self, input_str: str) -> ResolvedCommand:
        """
        Parse input and resolve to canonical command.
        
        Args:
            input_str: User input string
            
        Returns:
            ResolvedCommand with resolution details
        """
        if not self.registry.is_loaded:
            self.registry.load()
        
        # Normalize input
        normalized = input_str.strip().lower()
        
        # Check for reserved words
        if normalized in self.RESERVED_WORDS:
            return ResolvedCommand(
                command_type=CommandType.CANONICAL,
                original_input=input_str,
                resolved_command=normalized
            )
        
        # Try to resolve
        entry = self.registry.resolve(normalized)
        
        if entry is None:
            return ResolvedCommand(
                command_type=CommandType.UNKNOWN,
                original_input=input_str,
                resolved_command=input_str
            )
        
        # Determine resolved command
        if entry.category == AliasCategory.CANONICAL:
            resolved = entry.alias
            cmd_type = CommandType.CANONICAL
        else:
            # Find the canonical for this alias
            canonical = self.registry.resolve(entry.canonical)
            resolved = canonical.alias if canonical else entry.canonical
            cmd_type = CommandType.ALIAS
        
        # Build deprecation warning
        deprecation_warning = None
        if entry.deprecated:
            deprecation_warning = (
                f"WARNING: '{input_str}' is deprecated"
                f" (use '{resolved}' instead)"
            )
        
        return ResolvedCommand(
            command_type=cmd_type,
            original_input=input_str,
            resolved_command=resolved,
            category=entry.category,
            entry=entry,
            deprecated=entry.deprecated,
            deprecation_warning=deprecation_warning
        )
    
    def suggest(self, partial: str, limit: int = 5) -> List[str]:
        """
        Suggest completions for partial input.
        
        Args:
            partial: Partial input
            limit: Maximum suggestions
            
        Returns:
            List of suggested completions
        """
        suggestions = self.registry.suggest(partial, limit)
        return [s.alias for s in suggestions]


class RuntimeCommandHandler:
    """
    Handles runtime commands with alias support.
    
    Integrates with the ECU to provide alias-aware command execution.
    """
    
    # Built-in command handlers
    COMMANDS = {
        'start engine': 'start_engine_handler',
        'pre-flight check': 'preflight_handler',
        'mission ready': 'mission_ready_handler',
        'check state': 'check_state_handler',
        'bootstrap': 'bootstrap_handler',
        'run demo': 'run_demo_handler',
    }
    
    def __init__(self, parser: Optional[CommandParser] = None):
        """
        Initialize command handler.
        
        Args:
            parser: CommandParser instance
        """
        self.parser = parser or CommandParser()
    
    def handle(self, input_str: str) -> Tuple[bool, str, Optional[str]]:
        """
        Handle a command with alias resolution.
        
        Args:
            input_str: User input
            
        Returns:
            Tuple of (success, message, deprecation_warning)
        """
        resolved = self.parser.parse(input_str)
        
        # Emit deprecation warning if needed
        deprecation_warning = resolved.deprecation_warning
        
        if resolved.command_type == CommandType.UNKNOWN:
            suggestions = self.parser.suggest(input_str)
            if suggestions:
                return (
                    False,
                    f"Unknown command: '{input_str}'. Did you mean: {', '.join(suggestions)}?",
                    deprecation_warning
                )
            return (
                False,
                f"Unknown command: '{input_str}'",
                deprecation_warning
            )
        
        # Check if command is supported
        if resolved.resolved_command not in self.COMMANDS:
            return (
                False,
                f"Command '{resolved.resolved_command}' is not yet implemented",
                deprecation_warning
            )
        
        # Execute the command
        output = self._execute_command(resolved.resolved_command)
        
        return (
            True,
            output,
            deprecation_warning
        )
    
    def _execute_command(self, command: str) -> str:
        """
        Execute a resolved command.
        
        Args:
            command: The canonical command to execute
            
        Returns:
            Command output string
        """
        if command == 'pre-flight check':
            return self._run_preflight()
        elif command == 'start engine':
            return self._run_start_engine()
        elif command == 'run demo':
            return self._run_demo()
        elif command == 'check state':
            return self._check_state()
        else:
            return f"Command '{command}' executed"
    
    def _run_preflight(self) -> str:
        """Execute pre-flight check using the preflight module."""
        try:
            from runtime.preflight import run_preflight_check, format_report
            report = run_preflight_check()
            return format_report(report)
        except Exception as e:
            return f"Pre-flight check failed: {str(e)}"
    
    def _run_start_engine(self) -> str:
        """Execute start engine initialization."""
        try:
            from runtime.ecu import create_ecu
            ecu = create_ecu('/workspace/project/kde')
            state = ecu.get_runtime_state()
            return f"Engine started: {state['engines_registered']} engines, {state['seeds_registered']} seeds"
        except Exception as e:
            return f"Start engine failed: {str(e)}"
    
    def _run_demo(self) -> str:
        """Execute demo routine."""
        try:
            from runtime.runtime import demo
            import io
            from contextlib import redirect_stdout
            f = io.StringIO()
            with redirect_stdout(f):
                demo()
            return f.getvalue()
        except Exception as e:
            return f"Demo failed: {str(e)}"
    
    def _check_state(self) -> str:
        """Check runtime state."""
        try:
            import json
            with open('/workspace/project/kde/runtime/state.json') as f:
                state = json.load(f)
            return json.dumps(state, indent=2)
        except Exception as e:
            return f"State check failed: {str(e)}"
    
    def list_commands(self, category: Optional[str] = None) -> List[Dict]:
        """
        List available commands.
        
        Args:
            category: Filter by category
            
        Returns:
            List of command information
        """
        if category:
            try:
                cat = AliasCategory[category.upper()]
                aliases = self.parser.registry.get_by_category(cat)
            except KeyError:
                return []
        else:
            aliases = self.parser.registry.get_all_aliases()
        
        # Filter to canonical commands only
        canonical_commands = {}
        for alias in aliases:
            if alias.category == AliasCategory.CANONICAL:
                canonical_commands[alias.alias] = {
                    'canonical': alias.alias,
                    'description': alias.description,
                    'examples': alias.examples
                }
        
        return list(canonical_commands.values())


def create_parser() -> CommandParser:
    """Create a CommandParser instance."""
    return CommandParser()


def create_handler() -> RuntimeCommandHandler:
    """Create a RuntimeCommandHandler instance."""
    return RuntimeCommandHandler()
