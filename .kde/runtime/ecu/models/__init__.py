"""
ECU Models Module

Core data models for the Execution Control Unit.
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import List, Dict, Set, Optional, Any
from datetime import datetime


class EngineStatus(Enum):
    """Engine lifecycle status."""
    ACTIVE = "active"
    HISTORICAL = "historical"
    DEPRECATED = "deprecated"
    EXPERIMENTAL = "experimental"


class EngineStability(Enum):
    """Engine stability level."""
    STABLE = "stable"
    TESTING = "testing"
    UNSTABLE = "unstable"


class SeedStatus(Enum):
    """Seed lifecycle status."""
    ACTIVE = "active"
    FROZEN = "frozen"
    DEPRECATED = "deprecated"


class CapabilityType(Enum):
    """Type of capability."""
    REASONING = "reasoning"
    ANALYSIS = "analysis"
    SYNTHESIS = "synthesis"
    VALIDATION = "validation"
    GENERATION = "generation"
    EVALUATION = "evaluation"


class ExecutionMode(Enum):
    """Execution pipeline mode."""
    SINGLE = "single"
    SEQUENTIAL = "sequential"
    PARALLEL = "parallel"
    CONSENSUS = "consensus"
    SEED_ASSISTED = "seed_assisted"


class ConsensusStrategy(Enum):
    """Consensus coordination strategy."""
    SINGLE = "single"  # First valid result wins
    MAJORITY = "majority"  # >50% agreement required
    UNANIMOUS = "unanimous"  # All must agree
    WEIGHTED = "weighted"  # Weighted by engine priority
    ADVERSARIAL = "adversarial"  # Adversarial evaluation


class PolicyViolation(Enum):
    """Policy violation types."""
    UNAUTHORIZED_ENGINE = "unauthorized_engine"
    PLACEHOLDER_ENGINE = "placeholder_engine"
    UNOFFICIAL_ASSET = "unofficial_asset"
    INVALID_REGISTRATION = "invalid_registration"
    INVALID_SEED_REGISTRATION = "invalid_seed_registration"
    INVALID_EXECUTION_PLAN = "invalid_execution_plan"
    MISSING_CAPABILITIES = "missing_capabilities"
    ENGINE_NOT_FOUND = "engine_not_found"
    SEED_NOT_FOUND = "seed_not_found"


@dataclass
class Capability:
    """A capability that an engine or seed provides."""
    name: str
    type: CapabilityType
    description: str
    keywords: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class EngineMetadata:
    """Metadata for a registered engine."""
    engine_id: str
    directory: str
    name: str
    codename: str
    version: str
    status: EngineStatus
    stability: EngineStability
    capabilities: List[Capability] = field(default_factory=list)
    dependencies: List[str] = field(default_factory=list)
    priority: int = 100
    compatible_seeds: List[str] = field(default_factory=list)
    provenance: str = ""
    specification_path: str = ""
    methodology_path: str = ""


@dataclass
class SeedMetadata:
    """Metadata for a registered seed."""
    seed_id: str
    directory: str
    name: str
    codename: str
    version: str
    status: SeedStatus
    capabilities: List[Capability] = field(default_factory=list)
    promotion_state: str = "foundation"
    compatible_engines: List[str] = field(default_factory=list)


@dataclass
class ExecutionRequest:
    """A request for engine execution."""
    request_id: str
    description: str
    required_capabilities: List[CapabilityType]
    keywords: List[str] = field(default_factory=list)
    preferred_seeds: List[str] = field(default_factory=list)
    consensus_mode: Optional[ConsensusStrategy] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=datetime.now)


@dataclass
class EngineSelection:
    """Selection of an engine for execution."""
    engine: EngineMetadata
    reason: str
    confidence: float = 1.0


@dataclass
class SeedSelection:
    """Selection of a seed for execution."""
    seed: SeedMetadata
    reason: str
    confidence: float = 1.0


@dataclass
class ExecutionStep:
    """A single step in an execution pipeline."""
    step_id: str
    engine: EngineMetadata
    seed: Optional[SeedMetadata] = None
    inputs: List[str] = field(default_factory=list)
    expected_outputs: List[str] = field(default_factory=list)
    consensus_required: bool = False
    position: int = 0


@dataclass
class ExecutionPlan:
    """A complete execution plan."""
    plan_id: str
    request_id: str
    mode: ExecutionMode
    steps: List[ExecutionStep] = field(default_factory=list)
    selected_engines: List[EngineSelection] = field(default_factory=list)
    selected_seeds: List[SeedSelection] = field(default_factory=list)
    consensus_strategy: Optional[ConsensusStrategy] = None
    estimated_steps: int = 0
    validated: bool = False
    validation_errors: List[str] = field(default_factory=list)


@dataclass
class EngineResult:
    """Result from an engine execution."""
    engine_id: str
    engine_version: str
    step_id: str
    success: bool
    outputs: Dict[str, Any] = field(default_factory=dict)
    errors: List[str] = field(default_factory=list)
    execution_time_ms: float = 0.0
    provenance: Dict[str, str] = field(default_factory=dict)


@dataclass
class AggregatedResult:
    """Aggregated result from multiple engines."""
    request_id: str
    plan_id: str
    engine_results: List[EngineResult] = field(default_factory=list)
    consensus_reached: bool = False
    consensus_value: Optional[Any] = None
    aggregated_outputs: Dict[str, Any] = field(default_factory=dict)
    attribution: Dict[str, float] = field(default_factory=dict)
    total_execution_time_ms: float = 0.0
    timestamp: datetime = field(default_factory=datetime.now)


@dataclass
class PolicyViolationResult:
    """Result of a policy check."""
    violated: bool
    violations: List[PolicyViolation] = field(default_factory=list)
    details: List[str] = field(default_factory=list)
    blocked: bool = False


@dataclass
class ECUState:
    """Current state of the ECU."""
    initialized: bool = False
    engines_registered: int = 0
    seeds_registered: int = 0
    total_requests_processed: int = 0
    total_plans_generated: int = 0
    total_policy_violations: int = 0
    last_initialization: Optional[datetime] = None
    initialization_errors: List[str] = field(default_factory=list)
