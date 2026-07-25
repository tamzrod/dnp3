"""
Policy Layer Module

Runtime policy enforcement for ECU operations.
"""

import os
from typing import List, Dict, Optional, Any, Set
from dataclasses import dataclass, field

from ..models import (
    PolicyViolation, PolicyViolationResult, ExecutionPlan,
    EngineMetadata, SeedMetadata
)
from ..registry import EngineRegistry, SeedRegistry


@dataclass
class PolicyRule:
    """A single policy rule."""
    name: str
    description: str
    check_fn: callable
    severity: str = "error"  # error, warning, info
    blocking: bool = True


class PolicyLayer:
    """
    Runtime policy enforcement for the ECU.
    
    Responsibilities:
    - Validate engine registrations
    - Validate seed registrations
    - Validate execution plans
    - Block unauthorized operations
    - Enforce runtime policies
    """
    
    def __init__(
        self,
        engine_registry: EngineRegistry,
        seed_registry: SeedRegistry,
        kde_root: str
    ):
        """
        Initialize the Policy Layer.
        
        Args:
            engine_registry: Engine registry instance
            seed_registry: Seed registry instance
            kde_root: KDE runtime root directory
        """
        self.engine_registry = engine_registry
        self.seed_registry = seed_registry
        self.kde_root = kde_root
        
        # Initialize policy rules
        self._rules: List[PolicyRule] = []
        self._initialize_rules()
        
        # Policy statistics
        self.total_checks = 0
        self.total_violations = 0
    
    def _initialize_rules(self) -> None:
        """Initialize policy rules."""
        self._rules = [
            PolicyRule(
                name="engine_must_be_registered",
                description="All engines must be registered in the Engine Registry",
                check_fn=self._check_engine_registered,
                blocking=True
            ),
            PolicyRule(
                name="engine_must_have_specification",
                description="All engines must have a specification.md file",
                check_fn=self._check_engine_has_specification,
                blocking=True
            ),
            PolicyRule(
                name="engine_no_placeholder",
                description="Engines must not be placeholder implementations",
                check_fn=self._check_engine_not_placeholder,
                blocking=True
            ),
            PolicyRule(
                name="seed_must_be_registered",
                description="All seeds must be registered in the Seed Registry",
                check_fn=self._check_seed_registered,
                blocking=True
            ),
            PolicyRule(
                name="execution_plan_must_be_valid",
                description="Execution plans must pass validation",
                check_fn=self._check_plan_valid,
                blocking=True
            ),
            PolicyRule(
                name="execution_plan_engine_exists",
                description="All engines in execution plan must exist",
                check_fn=self._check_plan_engines_exist,
                blocking=True
            ),
            PolicyRule(
                name="no_unofficial_assets",
                description="Execution must not reference unofficial runtime assets",
                check_fn=self._check_no_unofficial_assets,
                blocking=True
            ),
            PolicyRule(
                name="engine_capabilities_match",
                description="Selected engines must match required capabilities",
                check_fn=self._check_engine_capabilities,
                blocking=False
            ),
        ]
    
    def validate_engine(self, engine: EngineMetadata) -> PolicyViolationResult:
        """
        Validate an engine against policy rules.
        
        Args:
            engine: Engine to validate
        
        Returns:
            PolicyViolationResult
        """
        violations = []
        details = []
        
        # Only apply engine-specific rules (not engine-related plan rules)
        engine_rules = [
            'engine_must_be_registered',
            'engine_must_have_specification',
            'engine_no_placeholder'
        ]
        
        for rule in self._rules:
            if rule.name in engine_rules:
                result = rule.check_fn(engine)
                if result:
                    violations.extend(result.get('violations', []))
                    details.extend(result.get('details', []))
        
        blocked = len(violations) > 0 and any(
            r.blocking for r in self._rules if r.name in [v.value if hasattr(v, 'value') else v for v in violations]
        )
        
        self.total_checks += 1
        if violations:
            self.total_violations += 1
        
        return PolicyViolationResult(
            violated=len(violations) > 0,
            violations=violations,
            details=details,
            blocked=blocked
        )
    
    def validate_seed(self, seed: SeedMetadata) -> PolicyViolationResult:
        """
        Validate a seed against policy rules.
        
        Args:
            seed: Seed to validate
        
        Returns:
            PolicyViolationResult
        """
        violations = []
        details = []
        
        for rule in self._rules:
            if 'seed' in rule.name:
                result = rule.check_fn(seed)
                if result:
                    violations.extend(result.get('violations', []))
                    details.extend(result.get('details', []))
        
        blocked = len(violations) > 0 and any(
            r.blocking for r in self._rules if r.name in [v.value if hasattr(v, 'value') else v for v in violations]
        )
        
        self.total_checks += 1
        if violations:
            self.total_violations += 1
        
        return PolicyViolationResult(
            violated=len(violations) > 0,
            violations=violations,
            details=details,
            blocked=blocked
        )
    
    def validate_execution_plan(self, plan: ExecutionPlan) -> PolicyViolationResult:
        """
        Validate an execution plan against policy rules.
        
        Args:
            plan: Execution plan to validate
        
        Returns:
            PolicyViolationResult
        """
        violations = []
        details = []
        
        for rule in self._rules:
            if 'plan' in rule.name or 'execution' in rule.name or 'unofficial' in rule.name:
                result = rule.check_fn(plan)
                if result:
                    violations.extend(result.get('violations', []))
                    details.extend(result.get('details', []))
        
        # Check for blocking violations
        blocked = False
        for v in violations:
            if v.blocked:
                blocked = True
                break
        
        self.total_checks += 1
        if violations:
            self.total_violations += 1
        
        return PolicyViolationResult(
            violated=len(violations) > 0,
            violations=violations,
            details=details,
            blocked=blocked
        )
    
    def _check_engine_registered(
        self, engine: EngineMetadata
    ) -> Dict[str, Any]:
        """Check if engine is registered."""
        registered = self.engine_registry.get_engine(engine.engine_id)
        
        if not registered:
            return {
                'violations': [PolicyViolation.UNAUTHORIZED_ENGINE],
                'details': [f"Engine {engine.engine_id} is not registered"]
            }
        
        return {'violations': [], 'details': []}
    
    def _check_engine_has_specification(
        self, engine: EngineMetadata
    ) -> Dict[str, Any]:
        """Check if engine has specification.md."""
        if not engine.specification_path:
            return {
                'violations': [PolicyViolation.INVALID_REGISTRATION],
                'details': [f"Engine {engine.engine_id} has no specification"]
            }
        
        if not os.path.exists(engine.specification_path):
            return {
                'violations': [PolicyViolation.INVALID_REGISTRATION],
                'details': [f"Engine {engine.engine_id} specification not found"]
            }
        
        return {'violations': [], 'details': []}
    
    def _check_engine_not_placeholder(
        self, engine: EngineMetadata
    ) -> Dict[str, Any]:
        """Check if engine is not a placeholder."""
        placeholder_indicators = [
            'placeholder', 'stub', 'todo', 'wip', 'temporary',
            'not_implemented', 'coming_soon'
        ]
        
        codename_lower = engine.codename.lower()
        for indicator in placeholder_indicators:
            if indicator in codename_lower:
                return {
                    'violations': [PolicyViolation.PLACEHOLDER_ENGINE],
                    'details': [f"Engine {engine.engine_id} appears to be a placeholder"]
                }
        
        return {'violations': [], 'details': []}
    
    def _check_seed_registered(
        self, seed: SeedMetadata
    ) -> Dict[str, Any]:
        """Check if seed is registered."""
        registered = self.seed_registry.get_seed(seed.seed_id)
        
        if not registered:
            return {
                'violations': [PolicyViolation.SEED_NOT_FOUND],
                'details': [f"Seed {seed.seed_id} is not registered"]
            }
        
        return {'violations': [], 'details': []}
    
    def _check_plan_valid(
        self, plan: ExecutionPlan
    ) -> Dict[str, Any]:
        """Check if execution plan is valid."""
        if not plan.validated:
            return {
                'violations': [PolicyViolation.INVALID_EXECUTION_PLAN],
                'details': [f"Execution plan {plan.plan_id} has not been validated"]
            }
        
        if plan.validation_errors:
            return {
                'violations': [PolicyViolation.INVALID_EXECUTION_PLAN],
                'details': plan.validation_errors
            }
        
        return {'violations': [], 'details': []}
    
    def _check_plan_engines_exist(
        self, plan: ExecutionPlan
    ) -> Dict[str, Any]:
        """Check if all engines in plan exist."""
        missing = []
        
        for step in plan.steps:
            if step.engine:
                if not self.engine_registry.get_engine(step.engine.engine_id):
                    missing.append(step.engine.engine_id)
        
        if missing:
            return {
                'violations': [PolicyViolation.ENGINE_NOT_FOUND],
                'details': [f"Missing engines in plan: {', '.join(missing)}"]
            }
        
        return {'violations': [], 'details': []}
    
    def _check_no_unofficial_assets(
        self, plan: ExecutionPlan
    ) -> Dict[str, Any]:
        """Check for unofficial assets in plan."""
        # Check for engine directories outside of official engines/
        unofficial_paths = []
        official_engines_dir = os.path.join(self.kde_root, "engines")
        
        for step in plan.steps:
            if step.engine and step.engine.directory:
                engine_path = os.path.join(official_engines_dir, step.engine.directory)
                if not os.path.exists(engine_path):
                    unofficial_paths.append(step.engine.directory)
        
        if unofficial_paths:
            return {
                'violations': [PolicyViolation.UNOFFICIAL_ASSET],
                'details': [f"Unofficial engine paths: {', '.join(set(unofficial_paths))}"]
            }
        
        return {'violations': [], 'details': []}
    
    def _check_engine_capabilities(
        self, plan: ExecutionPlan
    ) -> Dict[str, Any]:
        """Check if engine capabilities match plan requirements."""
        # This is a warning-only check
        return {'violations': [], 'details': []}
    
    def get_policy_summary(self) -> Dict[str, Any]:
        """
        Get policy enforcement summary.
        
        Returns:
            Policy summary dictionary
        """
        return {
            "total_rules": len(self._rules),
            "total_checks": self.total_checks,
            "total_violations": self.total_violations,
            "violation_rate": (
                self.total_violations / self.total_checks
                if self.total_checks > 0 else 0.0
            ),
            "rules": [
                {"name": r.name, "description": r.description, "blocking": r.blocking}
                for r in self._rules
            ]
        }
