"""
Runtime Skills Module

Implements dynamic skill loading for the KDE Runtime.

Components:
- Skill Registry: Contains skill definitions and metadata
- Skill Loader: Dynamically loads skills based on task context
- Dependency Resolver: Resolves skill dependencies
- Context Builder: Combines skill contributions into execution context

Usage:
    from runtime.skills import SkillLoader
    
    loader = SkillLoader()
    skills = loader.select_skills_for_task(["continuation"])
    loader.load_skills(skills, "INV-022")
    context = loader.generate_context_document()
"""

from .loader import (
    SkillRegistry,
    SkillLoader,
    SkillStatus,
    SkillMetadata,
    LoadedSkill,
    ExecutionLog,
    DependencyResolver
)

__all__ = [
    "SkillRegistry",
    "SkillLoader",
    "SkillStatus",
    "SkillMetadata",
    "LoadedSkill",
    "ExecutionLog",
    "DependencyResolver"
]
