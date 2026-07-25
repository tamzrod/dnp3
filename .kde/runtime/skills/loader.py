"""
Runtime Skill Loader

Implements dynamic skill loading for the KDE Runtime.

The Runtime SHALL:
1. Identify engineering task
2. Determine required skills
3. Resolve dependencies
4. Load selected skills
5. Construct engine context
6. Execute Engine

No Engine modifications are permitted.
"""

import json
import os
import time
from typing import List, Dict, Optional, Set, Any
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class SkillStatus(Enum):
    """Skill lifecycle states."""
    DRAFT = "draft"
    EXPERIMENTAL = "experimental"
    VALIDATED = "validated"
    PROMOTED = "promoted"
    DEPRECATED = "deprecated"
    ARCHIVED = "archived"


@dataclass
class SkillMetadata:
    """Metadata for a registered skill."""
    id: str
    name: str
    version: str
    status: SkillStatus
    description: str
    required_knowledge: List[str]
    required_sop: List[str]
    dependencies: List[str]
    input_types: List[str]
    output_types: List[str]
    triggers: List[str]
    context_contributions: Dict[str, Any]
    source: str
    validation_evidence: str


@dataclass
class LoadedSkill:
    """A skill that has been loaded for execution."""
    metadata: SkillMetadata
    load_time: float
    load_duration_ms: float


@dataclass
class ExecutionLog:
    """Log of a skill-based execution."""
    timestamp: str
    investigation_id: str
    loaded_skills: List[Dict[str, str]]
    skill_versions: Dict[str, str]
    load_duration_ms: float
    dependency_graph: Dict[str, List[str]]
    context_size: int
    outcome: str
    metadata: Dict[str, Any]


class SkillRegistry:
    """
    Registry of available skills.
    
    Contains skill definitions, metadata, and versioning.
    """
    
    def __init__(self, registry_path: str = None):
        """Initialize the registry."""
        if registry_path is None:
            registry_path = os.path.join(
                os.path.dirname(__file__),
                "registry.json"
            )
        
        with open(registry_path, 'r') as f:
            self.data = json.load(f)
        
        self._skills: Dict[str, SkillMetadata] = {}
        self._load_skills()
    
    def _load_skills(self):
        """Load skills from registry data."""
        for skill_data in self.data.get('skills', []):
            status_str = skill_data.get('status', 'draft')
            try:
                status = SkillStatus(status_str)
            except ValueError:
                status = SkillStatus.DRAFT
            
            skill = SkillMetadata(
                id=skill_data['id'],
                name=skill_data['name'],
                version=skill_data['version'],
                status=status,
                description=skill_data.get('description', ''),
                required_knowledge=skill_data.get('required_knowledge', []),
                required_sop=skill_data.get('required_sop', []),
                dependencies=skill_data.get('dependencies', []),
                input_types=skill_data.get('input_types', []),
                output_types=skill_data.get('output_types', []),
                triggers=skill_data.get('triggers', []),
                context_contributions=skill_data.get('context_contributions', {}),
                source=skill_data.get('source', ''),
                validation_evidence=skill_data.get('validation_evidence', '')
            )
            self._skills[skill.id] = skill
    
    def get_skill(self, skill_id: str) -> Optional[SkillMetadata]:
        """Get a skill by ID."""
        return self._skills.get(skill_id)
    
    def get_skills_by_status(self, status: SkillStatus) -> List[SkillMetadata]:
        """Get all skills with a given status."""
        return [s for s in self._skills.values() if s.status == status]
    
    def get_skills_by_trigger(self, trigger: str) -> List[SkillMetadata]:
        """Get all skills that respond to a trigger."""
        return [
            s for s in self._skills.values()
            if trigger in s.triggers and s.status == SkillStatus.PROMOTED
        ]
    
    def get_promoted_skills(self) -> List[SkillMetadata]:
        """Get all promoted skills."""
        return self.get_skills_by_status(SkillStatus.PROMOTED)
    
    def list_all(self) -> List[SkillMetadata]:
        """List all registered skills."""
        return list(self._skills.values())


class DependencyResolver:
    """
    Resolves skill dependencies.
    
    Builds a dependency graph and determines load order.
    """
    
    def __init__(self, registry: SkillRegistry):
        self.registry = registry
    
    def resolve(
        self, 
        skill_ids: List[str],
        include_dependencies: bool = True
    ) -> List[str]:
        """
        Resolve dependencies for a list of skills.
        
        Returns skills in dependency order (dependencies first).
        """
        if not include_dependencies:
            return skill_ids
        
        resolved: List[str] = []
        seen: Set[str] = set()
        
        def visit(skill_id: str):
            if skill_id in seen:
                return
            seen.add(skill_id)
            
            skill = self.registry.get_skill(skill_id)
            if skill:
                # Visit dependencies first
                for dep_id in skill.dependencies:
                    visit(dep_id)
                resolved.append(skill_id)
        
        for skill_id in skill_ids:
            visit(skill_id)
        
        return resolved
    
    def build_dependency_graph(
        self, 
        skill_ids: List[str]
    ) -> Dict[str, List[str]]:
        """Build a dependency graph for display."""
        graph: Dict[str, List[str]] = {}
        
        for skill_id in skill_ids:
            skill = self.registry.get_skill(skill_id)
            if skill:
                graph[skill_id] = skill.dependencies.copy()
        
        return graph


class SkillLoader:
    """
    Loads and manages skills at runtime.
    
    The Runtime uses this to dynamically load skills based on task context.
    """
    
    def __init__(self, registry: SkillRegistry = None):
        """Initialize the skill loader."""
        self.registry = registry or SkillRegistry()
        self.resolver = DependencyResolver(self.registry)
        
        self.current_execution: Optional[ExecutionLog] = None
        self.loaded_skills: List[LoadedSkill] = []
    
    def select_skills_for_task(
        self,
        triggers: List[str],
        keywords: List[str] = None
    ) -> List[SkillMetadata]:
        """
        Select skills required for a task based on triggers.
        
        Args:
            triggers: Task triggers (e.g., "new_investigation", "continuation")
            keywords: Optional keywords for additional matching
        
        Returns:
            List of selected skills in dependency order
        """
        selected_ids: Set[str] = set()
        
        # Find skills matching triggers
        for trigger in triggers:
            matching_skills = self.registry.get_skills_by_trigger(trigger)
            for skill in matching_skills:
                selected_ids.add(skill.id)
        
        # Resolve dependencies
        ordered_ids = self.resolver.resolve(list(selected_ids))
        
        # Return ordered skills
        return [
            self.registry.get_skill(sid) 
            for sid in ordered_ids 
            if self.registry.get_skill(sid)
        ]
    
    def load_skills(
        self,
        skills: List[SkillMetadata],
        investigation_id: str
    ) -> List[LoadedSkill]:
        """
        Load a list of skills for execution.
        
        Args:
            skills: Skills to load (in dependency order)
            investigation_id: Current investigation ID
        
        Returns:
            List of loaded skills with timing information
        """
        self.loaded_skills = []
        start_time = time.time()
        
        for skill in skills:
            load_start = time.time()
            
            # Simulate skill loading (in production, would import module)
            loaded_skill = LoadedSkill(
                metadata=skill,
                load_time=datetime.utcnow().isoformat() + "Z",
                load_duration_ms=(time.time() - load_start) * 1000
            )
            self.loaded_skills.append(loaded_skill)
        
        load_duration = (time.time() - start_time) * 1000
        
        # Build execution log
        skill_ids = [s.metadata.id for s in self.loaded_skills]
        self.current_execution = ExecutionLog(
            timestamp=datetime.utcnow().isoformat() + "Z",
            investigation_id=investigation_id,
            loaded_skills=[
                {"id": s.metadata.id, "name": s.metadata.name}
                for s in self.loaded_skills
            ],
            skill_versions={
                s.metadata.id: s.metadata.version 
                for s in self.loaded_skills
            },
            load_duration_ms=load_duration,
            dependency_graph=self.resolver.build_dependency_graph(skill_ids),
            context_size=0,  # Set after context building
            outcome="success",
            metadata={}
        )
        
        return self.loaded_skills
    
    def build_context(self) -> Dict[str, Any]:
        """
        Build the execution context from loaded skills.
        
        Combines context contributions from all loaded skills.
        """
        combined_context = {
            "instructions": [],
            "workflows": [],
            "constraints": [],
            "examples": [],
            "knowledge_references": []
        }
        
        for skill in self.loaded_skills:
            contrib = skill.metadata.context_contributions
            
            if "instructions" in contrib:
                combined_context["instructions"].append(
                    f"[{skill.metadata.name}] {contrib['instructions']}"
                )
            
            if "workflow" in contrib:
                combined_context["workflows"].append({
                    "skill": skill.metadata.name,
                    "steps": contrib["workflow"]
                })
            
            if "constraints" in contrib:
                combined_context["constraints"].extend(contrib["constraints"])
            
            if "examples" in contrib:
                combined_context["examples"].extend(contrib["examples"])
            
            combined_context["knowledge_references"].extend(
                skill.metadata.required_knowledge
            )
        
        # Update execution log
        if self.current_execution:
            self.current_execution.context_size = len(str(combined_context))
        
        return combined_context
    
    def get_execution_log(self) -> Optional[ExecutionLog]:
        """Get the current execution log."""
        return self.current_execution
    
    def generate_context_document(self) -> str:
        """Generate a formatted context document for the engine."""
        if not self.loaded_skills:
            return "# Skill Context\n\nNo skills loaded."
        
        lines = [
            "# Skill Context",
            "",
            f"**Investigation**: {self.current_execution.investigation_id}",
            f"**Skills Loaded**: {len(self.loaded_skills)}",
            f"**Load Duration**: {self.current_execution.load_duration_ms:.2f} ms",
            "",
            "## Loaded Skills",
            ""
        ]
        
        for skill in self.loaded_skills:
            lines.append(f"### {skill.metadata.name}")
            lines.append(f"**ID**: {skill.metadata.id}")
            lines.append(f"**Version**: {skill.metadata.version}")
            lines.append(f"**Status**: {skill.metadata.status.value}")
            lines.append("")
            lines.append(f"{skill.metadata.description}")
            lines.append("")
            
            contrib = skill.metadata.context_contributions
            if "instructions" in contrib:
                lines.append(f"**Instructions**: {contrib['instructions']}")
                lines.append("")
        
        # Add combined context
        context = self.build_context()
        
        lines.append("## Combined Instructions")
        lines.append("")
        for i, instruction in enumerate(context["instructions"], 1):
            lines.append(f"{i}. {instruction}")
        lines.append("")
        
        if context["constraints"]:
            lines.append("## Constraints")
            lines.append("")
            for constraint in context["constraints"]:
                lines.append(f"- {constraint}")
            lines.append("")
        
        lines.append("---")
        lines.append(f"Context size: {self.current_execution.context_size:,} characters")
        
        return '\n'.join(lines)


def test_loader():
    """Test the skill loader."""
    print("=" * 60)
    print("SKILL LOADER TEST")
    print("=" * 60)
    print()
    
    # Initialize loader
    loader = SkillLoader()
    registry = loader.registry
    
    # List promoted skills
    print("1. PROMOTED SKILLS")
    print("-" * 40)
    promoted = registry.get_promoted_skills()
    for skill in promoted:
        print(f"   [{skill.id}] {skill.name} v{skill.version}")
    print()
    
    # Select skills for continuation task
    print("2. SKILL SELECTION (continuation)")
    print("-" * 40)
    selected = loader.select_skills_for_task(
        triggers=["continuation"],
        keywords=["architecture"]
    )
    print(f"   Selected {len(selected)} skills:")
    for skill in selected:
        print(f"   - {skill.name} (deps: {skill.dependencies})")
    print()
    
    # Load selected skills
    print("3. LOADING SKILLS")
    print("-" * 40)
    loaded = loader.load_skills(selected, "INV-TEST")
    print(f"   Loaded {len(loaded)} skills in {loader.current_execution.load_duration_ms:.2f} ms")
    for skill in loaded:
        print(f"   - {skill.metadata.name} v{skill.metadata.version}")
    print()
    
    # Generate context document
    print("4. CONTEXT DOCUMENT (preview)")
    print("-" * 40)
    context_doc = loader.generate_context_document()
    print(context_doc[:800] + "...")
    print()
    
    # Execution log
    print("5. EXECUTION LOG")
    print("-" * 40)
    log = loader.current_execution
    print(f"   Investigation: {log.investigation_id}")
    print(f"   Skills loaded: {len(log.loaded_skills)}")
    print(f"   Context size: {log.context_size:,} chars")
    print(f"   Dependency graph: {list(log.dependency_graph.keys())}")


if __name__ == "__main__":
    test_loader()
