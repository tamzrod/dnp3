"""
Runtime Task Classification and Orchestration

Implements autonomous task classification for the KDE Runtime.

The Runtime shall:
1. Classify the engineering request
2. Select required skills
3. Retrieve relevant knowledge
4. Build execution context
5. Invoke Engine (unchanged)
"""

import json
import re
import os
from typing import List, Dict, Optional, Tuple, Any
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum

# Import workspace resolver
from .workspace import WorkspaceResolver


class TaskType(Enum):
    """Supported task types."""
    INVESTIGATION = "investigation"
    EXPERIMENT = "experiment"
    GOVERNANCE_REVIEW = "governance_review"
    KNOWLEDGE_CREATION = "knowledge_creation"
    KNOWLEDGE_REVISION = "knowledge_revision"
    SKILL_CREATION = "skill_creation"
    RUNTIME_DEVELOPMENT = "runtime_development"
    ARCHITECTURE_DESIGN = "architecture_design"
    DOCUMENTATION = "documentation"
    FRONTEND_DEVELOPMENT = "frontend_development"
    BACKEND_DEVELOPMENT = "backend_development"
    TESTING = "testing"
    GENERAL = "general"


# Task classification rules
TASK_PATTERNS = {
    TaskType.INVESTIGATION: [
        r"investigat",
        r"research",
        r"study",
        r"explor",
        r"analyz",
        r"what is",
        r"why is",
        r"how does",
        r"determine",
        r"evaluat",
    ],
    TaskType.EXPERIMENT: [
        r"experiment",
        r"test hypoth",
        r"validate",
        r"compar",
        r"a/?b",
        r"ab test",
        r"measure",
        r"trial",
    ],
    TaskType.GOVERNANCE_REVIEW: [
        r"governance",
        r"approval",
        r"review and approv",
        r"promot",
        r"policy",
        r"standard adopt",
    ],
    TaskType.KNOWLEDGE_CREATION: [
        r"knowledge",
        r"learn",
        r"understand",
        r"concepts?",
        r"principles?",
    ],
    TaskType.KNOWLEDGE_REVISION: [
        r"revis",
        r"update",
        r"amend",
        r"improv",
        r"enhanc",
    ],
    TaskType.SKILL_CREATION: [
        r"skill",
        r"capability",
        r"new ability",
    ],
    TaskType.RUNTIME_DEVELOPMENT: [
        r"runtime",
        r"orchestrat",
        r"loader",
        r"registry",
    ],
    TaskType.ARCHITECTURE_DESIGN: [
        r"architectur",
        r"design system",
        r"high.?level",
        r"component",
        r"module",
    ],
    TaskType.DOCUMENTATION: [
        r"document",
        r"readme",
        r"docs?",
        r"write up",
        r"specification",
    ],
    TaskType.FRONTEND_DEVELOPMENT: [
        r"frontend",
        r"ui",
        r"user interface",
        r"visual",
        r"dashboard",
        r"component",
        r"html",
        r"css",
        r"javascript",
    ],
    TaskType.BACKEND_DEVELOPMENT: [
        r"backend",
        r"api",
        r"server",
        r"database",
        r"service",
        r"endpoint",
    ],
    TaskType.TESTING: [
        r"test",
        r"testing",
        r"unit test",
        r"integration test",
        r"validate",
        r"verify",
    ],
}


# Task to skills mapping
TASK_SKILL_MAP = {
    TaskType.INVESTIGATION: ["skill-investigation-planning", "skill-evidence-collection", "skill-knowledge-retrieval"],
    TaskType.EXPERIMENT: ["skill-experiment-design", "skill-evidence-collection", "skill-decision-attribution"],
    TaskType.GOVERNANCE_REVIEW: ["skill-governance-review"],
    TaskType.KNOWLEDGE_CREATION: ["skill-knowledge-retrieval"],
    TaskType.KNOWLEDGE_REVISION: ["skill-knowledge-retrieval", "skill-evidence-collection"],
    TaskType.SKILL_CREATION: ["skill-investigation-planning", "skill-evidence-collection"],
    TaskType.RUNTIME_DEVELOPMENT: ["skill-investigation-planning", "skill-artifact-traceability"],
    TaskType.ARCHITECTURE_DESIGN: ["skill-investigation-planning", "skill-artifact-traceability"],
    TaskType.DOCUMENTATION: ["skill-investigation-planning"],
    TaskType.FRONTEND_DEVELOPMENT: ["skill-frontend-design"],
    TaskType.BACKEND_DEVELOPMENT: ["skill-investigation-planning"],
    TaskType.TESTING: ["skill-investigation-planning", "skill-evidence-collection"],
    TaskType.GENERAL: ["skill-investigation-planning"],
}


# Task to triggers mapping
TASK_TRIGGERS_MAP = {
    TaskType.INVESTIGATION: ["investigation", "new_investigation"],
    TaskType.EXPERIMENT: ["experiment", "hypothesis_testing"],
    TaskType.GOVERNANCE_REVIEW: ["governance_review"],
    TaskType.KNOWLEDGE_CREATION: ["continuation"],
    TaskType.KNOWLEDGE_REVISION: ["similar_work"],
    TaskType.SKILL_CREATION: ["skill_creation"],
    TaskType.RUNTIME_DEVELOPMENT: ["runtime_development"],
    TaskType.ARCHITECTURE_DESIGN: ["architecture"],
    TaskType.DOCUMENTATION: ["documentation"],
    TaskType.FRONTEND_DEVELOPMENT: ["frontend_task", "ui_requirement"],
    TaskType.BACKEND_DEVELOPMENT: ["backend_task"],
    TaskType.TESTING: ["testing", "validation"],
    TaskType.GENERAL: [],
}


@dataclass
class ClassificationResult:
    """Result of task classification."""
    task_type: TaskType
    confidence: float
    matched_patterns: List[str]
    keywords: List[str]
    classification_reason: str


@dataclass
class SkillSelection:
    """Selected skills for a task."""
    task_type: TaskType
    skills: List[str]
    skills_with_dependencies: List[str]


@dataclass
class WorkspaceInfo:
    """Information about the resolved workspace."""
    base_path: Optional[str]
    resolved_path: str
    exists: bool
    task_type: str
    task_id: Optional[str] = None
    requires_id: bool = False
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize to dictionary."""
        return {
            "base_path": self.base_path,
            "resolved_path": self.resolved_path,
            "exists": self.exists,
            "task_type": self.task_type,
            "task_id": self.task_id,
            "requires_id": self.requires_id
        }


@dataclass
class OrchestrationResult:
    """Complete orchestration result."""
    timestamp: str
    request: str
    classification: ClassificationResult
    skill_selection: SkillSelection
    workspace: WorkspaceInfo  # NEW: Resolved workspace
    loaded_skills_count: int
    knowledge_retrieved_count: int
    context_size: int
    execution_ready: bool


class TaskClassifier:
    """
    Classifies engineering requests into task types.
    
    Uses keyword pattern matching to determine task type.
    """
    
    def __init__(self):
        self.patterns = TASK_PATTERNS
        self.classification_count = 0
    
    def classify(self, request: str) -> ClassificationResult:
        """
        Classify an engineering request.
        
        Args:
            request: The engineering request text
        
        Returns:
            ClassificationResult with task type and confidence
        """
        self.classification_count += 1
        request_lower = request.lower()
        
        scores: Dict[TaskType, float] = {}
        matched_patterns: Dict[TaskType, List[str]] = {}
        
        for task_type, patterns in self.patterns.items():
            task_score = 0
            task_matches = []
            
            for pattern in patterns:
                if re.search(pattern, request_lower):
                    task_score += 1
                    task_matches.append(pattern)
            
            if task_matches:
                scores[task_type] = task_score
                matched_patterns[task_type] = task_matches
        
        if not scores:
            # Default to general
            return ClassificationResult(
                task_type=TaskType.GENERAL,
                confidence=0.5,
                matched_patterns=[],
                keywords=self._extract_keywords(request),
                classification_reason="No specific patterns matched; defaulting to GENERAL"
            )
        
        # Find highest scoring task
        best_task = max(scores.items(), key=lambda x: x[1])
        task_type, score = best_task
        
        # Calculate confidence (normalize by number of patterns)
        max_patterns = len(self.patterns[task_type])
        confidence = min(score / max_patterns * 2, 1.0)
        
        return ClassificationResult(
            task_type=task_type,
            confidence=confidence,
            matched_patterns=matched_patterns[task_type],
            keywords=self._extract_keywords(request),
            classification_reason=f"Matched {score} patterns for {task_type.value}"
        )
    
    def _extract_keywords(self, request: str) -> List[str]:
        """Extract significant keywords from request."""
        words = re.findall(r'\b[a-z]{4,}\b', request.lower())
        
        # Common words to exclude
        stop_words = {
            'what', 'where', 'when', 'which', 'would', 'could',
            'should', 'their', 'there', 'these', 'those', 'being',
            'have', 'with', 'this', 'that', 'from', 'they', 'will',
            'shall', 'must', 'need', 'also', 'into', 'more', 'such'
        }
        
        return [w for w in words if w not in stop_words][:10]


class SkillSelector:
    """
    Selects skills based on task type.
    
    Uses the task to skills mapping.
    """
    
    def __init__(self):
        self.task_skill_map = TASK_SKILL_MAP
        self.task_triggers_map = TASK_TRIGGERS_MAP
    
    def select(
        self, 
        task_type: TaskType,
        include_dependencies: bool = True
    ) -> SkillSelection:
        """
        Select skills for a task type.
        
        Args:
            task_type: The classified task type
            include_dependencies: Whether to resolve dependencies
        
        Returns:
            SkillSelection with selected skills
        """
        skill_ids = self.task_skill_map.get(task_type, [])
        
        if not include_dependencies:
            return SkillSelection(
                task_type=task_type,
                skills=skill_ids,
                skills_with_dependencies=skill_ids
            )
        
        # Resolve dependencies (simplified)
        all_skills = set(skill_ids)
        
        # Add dependencies from registry
        try:
            from runtime.skills import SkillLoader
            loader = SkillLoader()
            
            # Load to get dependency resolution
            skills = []
            for sid in skill_ids:
                skill = loader.registry.get_skill(sid)
                if skill:
                    skills.append(skill)
            
            # Get ordered skills with dependencies
            selected = loader.select_skills_for_task(
                self.task_triggers_map.get(task_type, []),
                []
            )
            resolved_ids = [s.id for s in selected]
            
        except ImportError:
            # Fallback if skills not available
            resolved_ids = skill_ids
        
        return SkillSelection(
            task_type=task_type,
            skills=skill_ids,
            skills_with_dependencies=resolved_ids
        )


class RuntimeOrchestrator:
    """
    Orchestrates the complete Runtime workflow.
    
    1. Classify the request
    2. Select required skills
    3. Retrieve relevant knowledge
    4. Build execution context
    5. Ready for Engine invocation
    """
    
    def __init__(self):
        self.classifier = TaskClassifier()
        self.selector = SkillSelector()
        self.workspace_resolver = WorkspaceResolver()  # NEW
        self.orchestration_count = 0
    
    def orchestrate(
        self, 
        request: str,
        investigation_id: str = None
    ) -> OrchestrationResult:
        """
        Orchestrate a complete runtime workflow.
        
        Args:
            request: The engineering request
            investigation_id: Optional investigation ID
        
        Returns:
            OrchestrationResult with complete workflow info
        """
        self.orchestration_count += 1
        
        # Step 1: Classify
        classification = self.classifier.classify(request)

        # Step 1.5: Resolve workspace (NEW)
        workspace_info = self.workspace_resolver.resolve(
            task_type=classification.task_type.value,
            task_id=investigation_id,
        )

        # Step 2: Select skills
        skill_selection = self.selector.select(classification.task_type)
        
        # Step 3: Load skills (if available)
        loaded_count = len(skill_selection.skills_with_dependencies)
        
        # Step 4: Retrieve knowledge (if available)
        knowledge_count = 0
        context_size = 0
        
        try:
            from runtime.runtime import KnowledgeOnDemandRuntime
            if investigation_id is None:
                investigation_id = f"INV-ORCH-{self.orchestration_count}"
            
            runtime = KnowledgeOnDemandRuntime(investigation_id)
            context = runtime.initialize(
                title=request[:50],
                description=request,
                keywords=classification.keywords
            )
            knowledge_count = len(context.knowledge_retrieved)
            context_size = context.context_size
        except ImportError:
            pass
        
        return OrchestrationResult(
            timestamp=datetime.utcnow().isoformat() + "Z",
            request=request[:100] + "..." if len(request) > 100 else request,
            classification=classification,
            skill_selection=skill_selection,
            workspace=workspace_info,  # NEW
            loaded_skills_count=loaded_count,
            knowledge_retrieved_count=knowledge_count,
            context_size=context_size,
            execution_ready=True
        )
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get orchestration statistics."""
        return {
            "total_classifications": self.classifier.classification_count,
            "total_orchestrations": self.orchestration_count,
        }


def test_orchestrator():
    """Test the runtime orchestrator."""
    print("=" * 70)
    print("RUNTIME ORCHESTRATOR TEST")
    print("=" * 70)
    print()
    
    orchestrator = RuntimeOrchestrator()
    
    # Test requests
    test_requests = [
        "Investigate whether KDE can improve frontend quality",
        "Experiment: Compare retrieval ON vs retrieval OFF",
        "Submit architecture design for governance approval",
        "Design a SCADA platform architecture",
        "Write documentation for the runtime API",
        "Create a new frontend component",
        "Implement backend API endpoint",
        "Run tests for the retrieval module",
    ]
    
    print("TASK CLASSIFICATION VALIDATION")
    print("-" * 70)
    
    results = []
    for request in test_requests:
        result = orchestrator.orchestrate(request)
        results.append(result)
        
        print(f"Request: {result.request[:50]}...")
        print(f"  → Task: {result.classification.task_type.value}")
        print(f"  → Confidence: {result.classification.confidence:.2f}")
        print(f"  → Skills: {result.skill_selection.skills}")
        print(f"  → Loaded: {result.loaded_skills_count}")
        print(f"  → Knowledge: {result.knowledge_retrieved_count} artifacts")
        print()
    
    # Evaluation matrix
    print("=" * 70)
    print("EVALUATION MATRIX")
    print("=" * 70)
    print()
    print(f"{'Request':<40} {'Task':<25} {'Confidence':<10}")
    print("-" * 70)
    for r in results:
        task = r.classification.task_type.value
        conf = f"{r.classification.confidence:.2f}"
        req = r.request[:37] + "..." if len(r.request) > 40 else r.request
        print(f"{req:<40} {task:<25} {conf:<10}")
    
    print()
    print("=" * 70)
    print("CLASSIFICATION ACCURACY")
    print("=" * 70)
    
    # Manual verification
    expected = [
        TaskType.INVESTIGATION,
        TaskType.EXPERIMENT,
        TaskType.GOVERNANCE_REVIEW,
        TaskType.ARCHITECTURE_DESIGN,
        TaskType.DOCUMENTATION,
        TaskType.FRONTEND_DEVELOPMENT,
        TaskType.BACKEND_DEVELOPMENT,
        TaskType.TESTING,
    ]
    
    correct = sum(
        1 for r, e in zip(results, expected) 
        if r.classification.task_type == e
    )
    
    accuracy = correct / len(expected) * 100
    print(f"Correct: {correct}/{len(expected)}")
    print(f"Accuracy: {accuracy:.1f}%")
    
    print()
    print(f"Statistics: {orchestrator.get_statistics()}")


if __name__ == "__main__":
    test_orchestrator()
