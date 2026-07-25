"""
Execution Planner Module

Generates execution pipelines for engine coordination.
"""

import uuid
from typing import List, Dict, Optional, Any
from dataclasses import dataclass, field

from ..models import (
    ExecutionRequest, ExecutionPlan, ExecutionStep, ExecutionMode,
    EngineSelection, SeedSelection, ConsensusStrategy, EngineMetadata, SeedMetadata
)


class ExecutionPlanner:
    """
    Plans execution pipelines for engine coordination.
    
    Responsibilities:
    - Generate execution plans from capability resolutions
    - Determine execution order and dependencies
    - Support multiple execution modes
    - Validate execution plans
    """
    
    def __init__(self):
        """Initialize the Execution Planner."""
        self._plan_templates: Dict[str, Dict] = {}
    
    def create_plan(
        self,
        request: ExecutionRequest,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> ExecutionPlan:
        """
        Create an execution plan from selections.
        
        Args:
            request: Original execution request
            engine_selections: Selected engines
            seed_selections: Selected seeds
        
        Returns:
            Generated execution plan
        """
        plan_id = f"PLAN-{uuid.uuid4().hex[:8].upper()}"
        
        # Determine execution mode
        mode = self._determine_execution_mode(
            request, engine_selections, len(engine_selections)
        )
        
        # Generate steps based on mode
        steps = self._generate_steps(
            plan_id,
            request,
            engine_selections,
            seed_selections,
            mode
        )
        
        # Create execution plan
        plan = ExecutionPlan(
            plan_id=plan_id,
            request_id=request.request_id,
            mode=mode,
            steps=steps,
            selected_engines=engine_selections,
            selected_seeds=seed_selections,
            consensus_strategy=request.consensus_mode,
            estimated_steps=len(steps),
            validated=False
        )
        
        return plan
    
    def _determine_execution_mode(
        self,
        request: ExecutionRequest,
        engine_selections: List[EngineSelection],
        selection_count: int
    ) -> ExecutionMode:
        """
        Determine the appropriate execution mode.
        
        Args:
            request: Execution request
            engine_selections: Engine selections
            selection_count: Number of selected engines
        
        Returns:
            Determined execution mode
        """
        # Check explicit consensus mode
        if request.consensus_mode:
            if request.consensus_mode == ConsensusStrategy.MAJORITY:
                return ExecutionMode.CONSENSUS
            elif request.consensus_mode == ConsensusStrategy.ADVERSARIAL:
                return ExecutionMode.CONSENSUS
        
        # Check if multiple seeds are preferred
        if len(request.preferred_seeds) > 1:
            return ExecutionMode.SEED_ASSISTED
        
        # Single engine selection
        if selection_count == 1:
            return ExecutionMode.SINGLE
        
        # Multiple engines - use sequential
        if selection_count > 1:
            # Check if engines have dependencies (could use parallel)
            has_dependencies = any(
                len(e.engine.dependencies) > 0
                for e in engine_selections
            )
            
            if not has_dependencies and selection_count <= 3:
                return ExecutionMode.PARALLEL
            
            return ExecutionMode.SEQUENTIAL
        
        return ExecutionMode.SINGLE
    
    def _generate_steps(
        self,
        plan_id: str,
        request: ExecutionRequest,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection],
        mode: ExecutionMode
    ) -> List[ExecutionStep]:
        """
        Generate execution steps based on mode.
        
        Args:
            plan_id: Plan identifier
            request: Execution request
            engine_selections: Selected engines
            seed_selections: Selected seeds
            mode: Execution mode
        
        Returns:
            List of execution steps
        """
        steps = []
        
        if mode == ExecutionMode.SINGLE:
            steps = self._generate_single_steps(
                plan_id, engine_selections, seed_selections
            )
        elif mode == ExecutionMode.SEQUENTIAL:
            steps = self._generate_sequential_steps(
                plan_id, engine_selections, seed_selections
            )
        elif mode == ExecutionMode.PARALLEL:
            steps = self._generate_parallel_steps(
                plan_id, engine_selections, seed_selections
            )
        elif mode == ExecutionMode.CONSENSUS:
            steps = self._generate_consensus_steps(
                plan_id, engine_selections, seed_selections
            )
        elif mode == ExecutionMode.SEED_ASSISTED:
            steps = self._generate_seed_assisted_steps(
                plan_id, engine_selections, seed_selections
            )
        
        return steps
    
    def _generate_single_steps(
        self,
        plan_id: str,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> List[ExecutionStep]:
        """Generate steps for single engine execution."""
        if not engine_selections:
            return []
        
        engine = engine_selections[0].engine
        
        # Find compatible seed
        seed = None
        for sel in seed_selections:
            if (not sel.seed.compatible_engines or 
                engine.engine_id in sel.seed.compatible_engines):
                seed = sel.seed
                break
        
        return [ExecutionStep(
            step_id=f"{plan_id}-STEP-1",
            engine=engine,
            seed=seed,
            position=1
        )]
    
    def _generate_sequential_steps(
        self,
        plan_id: str,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> List[ExecutionStep]:
        """Generate steps for sequential engine execution."""
        steps = []
        
        for i, engine_sel in enumerate(engine_selections, 1):
            engine = engine_sel.engine
            
            # Find compatible seed
            seed = None
            for sel in seed_selections:
                if (not sel.seed.compatible_engines or 
                    engine.engine_id in sel.seed.compatible_engines):
                    seed = sel.seed
                    break
            
            # Previous step output as input
            inputs = []
            if i > 1:
                inputs.append(f"{plan_id}-STEP-{i-1}")
            
            steps.append(ExecutionStep(
                step_id=f"{plan_id}-STEP-{i}",
                engine=engine,
                seed=seed,
                inputs=inputs,
                expected_outputs=[f"{plan_id}-STEP-{i}-output"],
                position=i
            ))
        
        return steps
    
    def _generate_parallel_steps(
        self,
        plan_id: str,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> List[ExecutionStep]:
        """Generate steps for parallel engine execution."""
        steps = []
        
        for i, engine_sel in enumerate(engine_selections, 1):
            engine = engine_sel.engine
            
            # Find compatible seed
            seed = None
            for sel in seed_selections:
                if (not sel.seed.compatible_engines or 
                    engine.engine_id in sel.seed.compatible_engines):
                    seed = sel.seed
                    break
            
            steps.append(ExecutionStep(
                step_id=f"{plan_id}-STEP-{i}",
                engine=engine,
                seed=seed,
                expected_outputs=[f"{plan_id}-STEP-{i}-output"],
                position=i
            ))
        
        return steps
    
    def _generate_consensus_steps(
        self,
        plan_id: str,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> List[ExecutionStep]:
        """Generate steps for consensus-based execution."""
        steps = []
        
        # Each engine contributes to consensus
        for i, engine_sel in enumerate(engine_selections, 1):
            engine = engine_sel.engine
            
            seed = None
            for sel in seed_selections:
                if (not sel.seed.compatible_engines or 
                    engine.engine_id in sel.seed.compatible_engines):
                    seed = sel.seed
                    break
            
            steps.append(ExecutionStep(
                step_id=f"{plan_id}-CONSENSUS-{i}",
                engine=engine,
                seed=seed,
                consensus_required=True,
                expected_outputs=[f"{plan_id}-CONSENSUS-{i}-output"],
                position=i
            ))
        
        return steps
    
    def _generate_seed_assisted_steps(
        self,
        plan_id: str,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> List[ExecutionStep]:
        """Generate steps for seed-assisted execution."""
        steps = []
        
        # First: apply each seed
        for i, seed_sel in enumerate(seed_selections[:2], 1):
            steps.append(ExecutionStep(
                step_id=f"{plan_id}-SEED-{i}",
                engine=engine_selections[0].engine if engine_selections else None,
                seed=seed_sel.seed,
                consensus_required=False,
                position=i
            ))
        
        # Then: execute with seeded context
        if engine_selections:
            steps.append(ExecutionStep(
                step_id=f"{plan_id}-STEP-EXECUTED",
                engine=engine_selections[0].engine,
                inputs=[f"{plan_id}-SEED-1", f"{plan_id}-SEED-2"],
                position=len(seed_selections) + 1
            ))
        
        return steps
    
    def validate_plan(self, plan: ExecutionPlan) -> bool:
        """
        Validate an execution plan.
        
        Args:
            plan: Execution plan to validate
        
        Returns:
            True if valid, False otherwise
        """
        errors = []
        
        # Check for steps
        if not plan.steps:
            errors.append("Plan has no execution steps")
        
        # Check engine references
        for step in plan.steps:
            if not step.engine:
                errors.append(f"Step {step.step_id} has no engine")
        
        # Check step ordering
        positions = [s.position for s in plan.steps]
        if positions != sorted(positions):
            errors.append("Step positions are not in order")
        
        # Check consensus consistency
        has_consensus = any(s.consensus_required for s in plan.steps)
        if has_consensus and not plan.consensus_strategy:
            errors.append("Consensus steps exist but no consensus strategy defined")
        
        # Store validation result
        plan.validation_errors = errors
        plan.validated = len(errors) == 0
        
        return plan.validated
    
    def get_plan_summary(self, plan: ExecutionPlan) -> Dict[str, Any]:
        """
        Get a summary of an execution plan.
        
        Args:
            plan: Execution plan
        
        Returns:
            Plan summary dictionary
        """
        return {
            "plan_id": plan.plan_id,
            "request_id": plan.request_id,
            "mode": plan.mode.value,
            "step_count": len(plan.steps),
            "engine_count": len(set(s.engine.engine_id for s in plan.steps if s.engine)),
            "seed_count": len(set(s.seed.seed_id for s in plan.steps if s.seed)),
            "consensus_required": any(s.consensus_required for s in plan.steps),
            "consensus_strategy": plan.consensus_strategy.value if plan.consensus_strategy else None,
            "validated": plan.validated,
            "validation_errors": plan.validation_errors,
            "steps": [
                {
                    "step_id": s.step_id,
                    "engine": s.engine.codename if s.engine else None,
                    "seed": s.seed.codename if s.seed else None,
                    "position": s.position,
                    "consensus_required": s.consensus_required
                }
                for s in plan.steps
            ]
        }
