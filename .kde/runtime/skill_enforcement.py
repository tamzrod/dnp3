"""
Skill Enforcement Hook

Extends skills with enforcement sections.
Enables skills to specify rules that block non-compliant behavior.

REC-006/REC-007: From KDE-INV-GOV-SYN
"""

import os
import re
from pathlib import Path
from typing import List, Optional, Dict, Set
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class RuleType(Enum):
    """Types of enforcement rules."""
    INVESTIGATION_REQUIRED = "investigation_required"
    EXPERIMENT_REQUIRED = "experiment_required"
    RUN_DOC_REQUIRED = "run_doc_required"
    ARTIFACT_BEFORE_CODE = "artifact_before_code"
    SYNTHESIS_BEFORE_NEW = "synthesis_before_new"
    EVIDENCE_REQUIRED = "evidence_required"


@dataclass
class EnforcementRule:
    """A single enforcement rule from a skill."""
    rule_type: RuleType
    block_on_fail: bool
    tools_blocked: List[str] = field(default_factory=list)
    message: str = ""


@dataclass 
class SkillEnforcement:
    """Enforcement configuration from a skill."""
    skill_name: str
    rules: List[EnforcementRule]
    enabled: bool = True
    parsed_at: str = field(default_factory=lambda: datetime.now().isoformat())


class SkillEnforcementParser:
    """
    Parses enforcement sections from skill Markdown files.
    
    Skill Format:
    ```markdown
    <!-- ENFORCEMENT -->
    - rule: investigation_required
      block_on_fail: true
      tools_blocked: [terminal, file_editor]
      message: "Create investigation artifact first"
    ```
    
    Usage:
        parser = SkillEnforcementParser()
        enforcement = parser.parse_skill("kde-investigation-framework")
        if not enforcement.enabled:
            return "Skill has no enforcement"
        for rule in enforcement.rules:
            print(f"Rule: {rule.rule_type}")
    """
    
    ENFORCEMENT_START = "<!-- ENFORCEMENT -->"
    ENFORCEMENT_END = "-->"
    
    RULE_PATTERNS = {
        "investigation_required": RuleType.INVESTIGATION_REQUIRED,
        "experiment_required": RuleType.EXPERIMENT_REQUIRED,
        "run_doc_required": RuleType.RUN_DOC_REQUIRED,
        "artifact_before_code": RuleType.ARTIFACT_BEFORE_CODE,
        "synthesis_before_new": RuleType.SYNTHESIS_BEFORE_NEW,
        "evidence_required": RuleType.EVIDENCE_REQUIRED,
    }
    
    def __init__(self, skills_root: str = None):
        if skills_root is None:
            skills_root = os.path.join(os.path.dirname(__file__), "..", "..", ".agents", "skills")
        self.skills_root = skills_root
        self._cache: Dict[str, SkillEnforcement] = {}
    
    def parse_skill(self, skill_name: str) -> Optional[SkillEnforcement]:
        """
        Parse enforcement section from a skill.
        
        Args:
            skill_name: Name of the skill (corresponds to .md file)
            
        Returns:
            SkillEnforcement if found, None if no enforcement section
        """
        # Check cache
        if skill_name in self._cache:
            return self._cache[skill_name]
        
        # Find skill file
        skill_path = self._find_skill_file(skill_name)
        if skill_path is None:
            return None
        
        # Parse content
        try:
            with open(skill_path, 'r') as f:
                content = f.read()
        except:
            return None
        
        enforcement = self._parse_enforcement_section(skill_name, content)
        self._cache[skill_name] = enforcement
        return enforcement
    
    def _find_skill_file(self, skill_name: str) -> Optional[str]:
        """Find skill file by name."""
        if not self.skills_root or not os.path.exists(self.skills_root):
            return None
        
        # Try exact match
        path = os.path.join(self.skills_root, f"{skill_name}.md")
        if os.path.exists(path):
            return path
        
        # Try case-insensitive search
        for f in os.listdir(self.skills_root):
            if f.lower().replace("-", "_") == f"{skill_name.lower()}.md":
                return os.path.join(self.skills_root, f)
        
        return None
    
    def _parse_enforcement_section(
        self, 
        skill_name: str, 
        content: str
    ) -> Optional[SkillEnforcement]:
        """Parse enforcement section from skill content."""
        rules = []
        
        # Find enforcement section
        if self.ENFORCEMENT_START not in content:
            return SkillEnforcement(skill_name=skill_name, rules=[], enabled=False)
        
        # Extract section
        start_idx = content.index(self.ENFORCEMENT_START)
        end_idx = content.find(self.ENFORCEMENT_END, start_idx)
        if end_idx == -1:
            end_idx = len(content)
        
        section = content[start_idx:end_idx]
        
        # Parse rules
        for line in section.split("\n"):
            line = line.strip()
            if not line or line.startswith("<!--") or line.startswith("-->"):
                continue
            
            if line.startswith("- rule:"):
                rule_type_str = line.replace("- rule:", "").strip()
                rule_type = self.RULE_PATTERNS.get(rule_type_str)
                if rule_type:
                    rules.append(EnforcementRule(
                        rule_type=rule_type,
                        block_on_fail=True,
                        tools_blocked=["terminal", "file_editor"]
                    ))
            elif "block_on_fail:" in line:
                if rules:
                    rules[-1].block_on_fail = "true" in line.lower()
            elif "tools_blocked:" in line:
                if rules:
                    # Parse list
                    match = re.search(r'\[(.*?)\]', line)
                    if match:
                        tools = [t.strip() for t in match.group(1).split(",")]
                        rules[-1].tools_blocked = tools
            elif "message:" in line:
                if rules:
                    rules[-1].message = line.replace("message:", "").strip().strip('"')
        
        if not rules:
            return SkillEnforcement(skill_name=skill_name, rules=[], enabled=False)
        
        return SkillEnforcement(skill_name=skill_name, rules=rules, enabled=True)
    
    def get_blocked_tools(self, skill_name: str) -> Set[str]:
        """Get all tools blocked by a skill's enforcement."""
        enforcement = self.parse_skill(skill_name)
        if not enforcement or not enforcement.enabled:
            return set()
        
        blocked = set()
        for rule in enforcement.rules:
            blocked.update(rule.tools_blocked)
        return blocked
    
    def clear_cache(self):
        """Clear parsed skill cache."""
        self._cache.clear()


class SkillEnforcementRegistry:
    """
    Registry of active enforcement rules.
    
    Tracks which skills have been invoked and their rules.
    """
    
    def __init__(self):
        self._active_rules: List[SkillEnforcement] = []
        self._blocked_tools: Set[str] = set()
        self._parser = SkillEnforcementParser()
    
    def register_skill(self, skill_name: str) -> SkillEnforcement:
        """Register a skill's enforcement rules."""
        enforcement = self._parser.parse_skill(skill_name)
        if enforcement and enforcement.enabled:
            self._active_rules.append(enforcement)
            for rule in enforcement.rules:
                self._blocked_tools.update(rule.tools_blocked)
        return enforcement
    
    def get_blocked_tools(self) -> Set[str]:
        """Get all currently blocked tools."""
        return self._blocked_tools.copy()
    
    def is_tool_blocked(self, tool_name: str) -> bool:
        """Check if a tool is blocked by any active rule."""
        return tool_name in self._blocked_tools
    
    def get_blocked_message(self, tool_name: str) -> str:
        """Get message explaining why tool is blocked."""
        messages = []
        for enforcement in self._active_rules:
            for rule in enforcement.rules:
                if tool_name in rule.tools_blocked:
                    if rule.message:
                        messages.append(rule.message)
                    else:
                        messages.append(f"{enforcement.skill_name} requires {rule.rule_type.value}")
        
        if not messages:
            return f"Tool '{tool_name}' is not blocked"
        
        return "\n".join(f"  - {m}" for m in messages)


# Global registry
_global_registry: Optional[SkillEnforcementRegistry] = None


def get_skill_registry() -> SkillEnforcementRegistry:
    """Get or create the global SkillEnforcementRegistry."""
    global _global_registry
    if _global_registry is None:
        _global_registry = SkillEnforcementRegistry()
    return _global_registry


def register_skill_enforcement(skill_name: str) -> SkillEnforcement:
    """Convenience function to register a skill's enforcement."""
    registry = get_skill_registry()
    return registry.register_skill(skill_name)
