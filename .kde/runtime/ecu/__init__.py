"""
Runtime ECU (Execution Control Unit)

Main orchestrator for the KDE Runtime ECU.
"""

import os
import uuid
from typing import Dict, List, Optional, Any
from datetime import datetime
from dataclasses import dataclass, field

from .models import (
    ECUState, ExecutionRequest, ExecutionPlan, ExecutionMode,
    EngineSelection, SeedSelection, EngineResult, AggregatedResult,
    CapabilityType, ConsensusStrategy, PolicyViolationResult
)
from .registry import EngineRegistry, SeedRegistry
from .resolver import CapabilityResolver
from .planner import ExecutionPlanner
from .policy import PolicyLayer
from .consensus import ConsensusManager
from .aggregator import ResultAggregator


@dataclass
class ECUInitializationResult:
    """Result of ECU initialization."""
    success: bool
    engines_registered: int = 0
    seeds_registered: int = 0
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)


@dataclass
class ECUExecutionResult:
    """Result of ECU execution."""
    request_id: str
    plan_id: Optional[str] = None
    success: bool = False
    blocked: bool = False
    policy_violations: List[str] = field(default_factory=list)
    error_message: str = ""
    aggregated_result: Optional[AggregatedResult] = None
    plan_summary: Optional[Dict[str, Any]] = None


class RuntimeECU:
    """
    Runtime Execution Control Unit (ECU).
    
    The ECU is the runtime orchestration layer responsible for:
    - Capability Analysis
    - Runtime Policy Enforcement
    - Engine Registry
    - Seed Registry
    - Capability Resolution
    - Engine Selection
    - Seed Selection
    - Execution Planning
    - Consensus Coordination
    - Result Aggregation
    
    The ECU SHALL NOT execute engineering reasoning.
    Reasoning belongs exclusively to Engines.
    """
    
    def __init__(self, kde_root: str):
        """
        Initialize the Runtime ECU.
        
        Args:
            kde_root: Root path to the KDE runtime directory
        """
        self.kde_root = kde_root
        
        # State
        self.state = ECUState()
        
        # Components
        self.engine_registry = EngineRegistry(kde_root)
        self.seed_registry = SeedRegistry(kde_root)
        self.capability_resolver = CapabilityResolver()
        self.execution_planner = ExecutionPlanner()
        self.policy_layer = PolicyLayer(
            self.engine_registry,
            self.seed_registry,
            kde_root
        )
        self.consensus_manager = ConsensusManager()
        self.result_aggregator = ResultAggregator()
        
        # Execution history
        self._execution_history: List[ECUExecutionResult] = []
    
    def initialize(self) -> ECUInitializationResult:
        """
        Initialize the ECU and discover all engines and seeds.
        
        Returns:
            ECUInitializationResult
        """
        errors = []
        warnings = []
        
        try:
            # Discover engines
            engines = self.engine_registry.discover()
            self.state.engines_registered = len(engines)
            
            if len(engines) == 0:
                warnings.append("No engines discovered")
            
            # Discover seeds
            seeds = self.seed_registry.discover()
            self.state.seeds_registered = len(seeds)
            
            if len(seeds) == 0:
                warnings.append("No seeds discovered")
            
            # Validate registries
            for engine in engines:
                violation = self.policy_layer.validate_engine(engine)
                if violation.violated:
                    warnings.append(
                        f"Engine {engine.engine_id} has policy warnings"
                    )
            
            for seed in seeds:
                violation = self.policy_layer.validate_seed(seed)
                if violation.violated:
                    warnings.append(
                        f"Seed {seed.seed_id} has policy warnings"
                    )
            
            # Update state
            self.state.initialized = True
            self.state.last_initialization = datetime.now()
            
            return ECUInitializationResult(
                success=True,
                engines_registered=len(engines),
                seeds_registered=len(seeds),
                warnings=warnings
            )
            
        except Exception as e:
            errors.append(str(e))
            self.state.initialization_errors = errors
            return ECUInitializationResult(
                success=False,
                errors=errors
            )
    
    def analyze_capabilities(
        self,
        request: ExecutionRequest
    ) -> Dict[str, Any]:
        """
        Analyze capabilities required for a request.
        
        Args:
            request: Execution request
        
        Returns:
            Capability analysis report
        """
        return {
            "request_id": request.request_id,
            "required_capabilities": [c.value for c in request.required_capabilities],
            "keywords": request.keywords,
            "preferred_seeds": request.preferred_seeds,
            "consensus_mode": request.consensus_mode.value if request.consensus_mode else None,
            "analysis_timestamp": datetime.now().isoformat()
        }
    
    def resolve_capabilities(
        self,
        request: ExecutionRequest
    ) -> Dict[str, Any]:
        """
        Resolve capabilities to engines and seeds.
        
        Args:
            request: Execution request
        
        Returns:
            Resolution report
        """
        # Get all engines and seeds
        engines = self.engine_registry.get_all_engines()
        seeds = self.seed_registry.get_all_seeds()
        
        # Resolve
        engine_selections = self.capability_resolver.resolve(
            request, engines, seeds
        )
        
        seed_selections = self.capability_resolver.select_seeds(
            engine_selections,
            request.preferred_seeds,
            seeds
        )
        
        # Generate report
        report = self.capability_resolver.generate_resolution_report(
            request, engine_selections, seed_selections
        )
        
        return report
    
    def create_execution_plan(
        self,
        request: ExecutionRequest
    ) -> ECUExecutionResult:
        """
        Create an execution plan for a request.
        
        Args:
            request: Execution request
        
        Returns:
            ECUExecutionResult
        """
        result = ECUExecutionResult(request_id=request.request_id)
        
        try:
            # Resolve capabilities
            engines = self.engine_registry.get_all_engines()
            seeds = self.seed_registry.get_all_seeds()
            
            engine_selections = self.capability_resolver.resolve(
                request, engines, seeds
            )
            
            seed_selections = self.capability_resolver.select_seeds(
                engine_selections,
                request.preferred_seeds,
                seeds
            )
            
            if not engine_selections:
                result.error_message = "No engines match required capabilities"
                return result
            
            # Create plan
            plan = self.execution_planner.create_plan(
                request, engine_selections, seed_selections
            )
            
            # Validate plan
            self.execution_planner.validate_plan(plan)
            
            # Validate against policy
            policy_result = self.policy_layer.validate_execution_plan(plan)
            
            if policy_result.violated:
                result.blocked = policy_result.blocked
                result.policy_violations = [
                    f"{v.value}: {d}"
                    for v, d in zip(policy_result.violations, policy_result.details)
                ]
                result.error_message = "Policy violations detected"
                return result
            
            # Success
            result.success = True
            result.plan_id = plan.plan_id
            result.plan_summary = self.execution_planner.get_plan_summary(plan)
            
            return result
            
        except Exception as e:
            result.error_message = str(e)
            return result
    
    def execute_plan(
        self,
        plan: ExecutionPlan,
        execution_fn: Optional[callable] = None
    ) -> AggregatedResult:
        """
        Execute an established plan.
        
        Note: This is a stub. Actual execution requires Laboratory integration.
        The ECU coordinates execution but does not perform reasoning.
        
        Args:
            plan: Execution plan
            execution_fn: Optional function to execute engines
        
        Returns:
            AggregatedResult
        """
        # Generate stub results for demonstration
        results = []
        
        for step in plan.steps:
            if step.engine:
                results.append(EngineResult(
                    engine_id=step.engine.engine_id,
                    engine_version=step.engine.version,
                    step_id=step.step_id,
                    success=True,
                    outputs={"step": step.step_id, "engine": step.engine.codename},
                    execution_time_ms=100.0,
                    provenance={"engine": step.engine.engine_id}
                ))
        
        # Coordinate consensus if required
        consensus_result = None
        has_consensus = any(s.consensus_required for s in plan.steps)
        
        if has_consensus and plan.consensus_strategy:
            engine_metadata = {
                e.engine_id: e
                for e in self.engine_registry.get_all_engines()
            }
            consensus_result = self.consensus_manager.coordinate(
                results,
                plan.consensus_strategy,
                engine_metadata
            )
        
        # Aggregate results
        aggregated = self.result_aggregator.aggregate(
            plan.request_id,
            plan,
            results,
            consensus_result
        )
        
        return aggregated
    
    def get_runtime_state(self) -> Dict[str, Any]:
        """
        Get current ECU runtime state.
        
        Returns:
            Runtime state dictionary
        """
        return {
            "initialized": self.state.initialized,
            "engines_registered": self.state.engines_registered,
            "seeds_registered": self.state.seeds_registered,
            "total_requests_processed": self.state.total_requests_processed,
            "total_plans_generated": self.state.total_plans_generated,
            "total_policy_violations": self.state.total_policy_violations,
            "last_initialization": (
                self.state.last_initialization.isoformat()
                if self.state.last_initialization else None
            ),
            "initialization_errors": self.state.initialization_errors,
            "engine_registry": self.engine_registry.get_registry_summary(),
            "seed_registry": self.seed_registry.get_registry_summary(),
            "policy_summary": self.policy_layer.get_policy_summary(),
            "consensus_summary": self.consensus_manager.get_consensus_summary(),
            "aggregation_summary": self.result_aggregator.get_aggregation_summary()
        }
    
    def get_execution_history(
        self,
        limit: int = 10
    ) -> List[Dict[str, Any]]:
        """
        Get recent execution history.
        
        Args:
            limit: Maximum number of records to return
        
        Returns:
            List of execution result summaries
        """
        history = self._execution_history[-limit:]
        
        return [
            {
                "request_id": r.request_id,
                "plan_id": r.plan_id,
                "success": r.success,
                "blocked": r.blocked,
                "error_message": r.error_message
            }
            for r in history
        ]


def create_ecu(kde_root: str) -> RuntimeECU:
    """
    Create and initialize a Runtime ECU.
    
    Args:
        kde_root: Root path to the KDE runtime directory
    
    Returns:
        Initialized RuntimeECU instance
    """
    ecu = RuntimeECU(kde_root)
    result = ecu.initialize()
    
    if not result.success:
        raise RuntimeError(
            f"ECU initialization failed: {', '.join(result.errors)}"
        )
    
    return ecu
