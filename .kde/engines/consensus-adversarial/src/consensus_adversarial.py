"""
Consensus Protocol Adversarial Evaluation Engine

Evaluates consensus protocols for safety and liveness violations.
"""

import random
from dataclasses import dataclass
from typing import Dict, List, Any, Optional


@dataclass
class AdversarialResult:
    """Result of adversarial testing."""
    attack_type: str
    success: bool
    description: str
    severity: str  # Critical, High, Medium, Low


class ConsensusAdversarialEvaluator:
    """Performs adversarial testing on consensus protocols."""
    
    def __init__(self, spec: Dict, simulations: List[Dict]):
        self.spec = spec
        self.simulations = simulations
    
    def evaluate(self) -> Dict:
        """Run all adversarial tests."""
        results = {
            "split_brain": self._test_split_brain(),
            "dual_leaders": self._test_dual_leaders(),
            "log_divergence": self._test_log_divergence(),
            "infinite_election": self._test_infinite_election(),
            "deadlock": self._test_deadlock(),
            "livelock": self._test_livelock(),
            "recovery_failure": self._test_recovery_failure(),
            "state_corruption": self._test_state_corruption(),
        }
        
        # Count failures
        critical = sum(1 for r in results.values() if r.severity == "Critical" and r.success)
        high = sum(1 for r in results.values() if r.severity == "High" and r.success)
        
        return {
            "tests": {k: {
                "success": v.success,
                "description": v.description,
                "severity": v.severity,
            } for k, v in results.items()},
            "critical_failures": critical,
            "high_failures": high,
            "overall_score": max(0, 100 - critical * 30 - high * 15),
        }
    
    def _test_split_brain(self) -> AdversarialResult:
        """Test split brain resistance."""
        term_epochs = self.spec.get("term_epochs", False)
        quorum_type = self.spec.get("quorum_type", "")
        
        if term_epochs and quorum_type == "majority":
            return AdversarialResult(
                attack_type="split_brain",
                success=False,
                description="Majority quorum with term epochs prevents split brain",
                severity="Medium"
            )
        else:
            return AdversarialResult(
                attack_type="split_brain",
                success=True,
                description="Protocol may be vulnerable to split brain",
                severity="High"
            )
    
    def _test_dual_leaders(self) -> AdversarialResult:
        """Test dual leader resistance."""
        randomized_timeout = self.spec.get("randomized_timeout", False)
        term_epochs = self.spec.get("term_epochs", False)
        
        if randomized_timeout and term_epochs:
            return AdversarialResult(
                attack_type="dual_leaders",
                success=False,
                description="Randomized timeouts prevent dual leaders",
                severity="Low"
            )
        else:
            return AdversarialResult(
                attack_type="dual_leaders",
                success=True,
                description="May be vulnerable to dual leaders",
                severity="Medium"
            )
    
    def _test_log_divergence(self) -> AdversarialResult:
        """Test log divergence resistance."""
        commit_method = self.spec.get("commit_method", "")
        
        if commit_method in ["index_based", "majority_commit"]:
            return AdversarialResult(
                attack_type="log_divergence",
                success=False,
                description="Index-based commit prevents divergence",
                severity="Low"
            )
        else:
            return AdversarialResult(
                attack_type="log_divergence",
                success=True,
                description="Log divergence possible",
                severity="Medium"
            )
    
    def _test_infinite_election(self) -> AdversarialResult:
        """Test infinite election resistance."""
        timeout_model = self.spec.get("timeout_model", "")
        
        if timeout_model == "exponential":
            return AdversarialResult(
                attack_type="infinite_election",
                success=False,
                description="Exponential backoff prevents infinite elections",
                severity="Low"
            )
        else:
            return AdversarialResult(
                attack_type="infinite_election",
                success=True,
                description="May be vulnerable to infinite election loops",
                severity="Medium"
            )
    
    def _test_deadlock(self) -> AdversarialResult:
        """Test deadlock resistance."""
        membership_model = self.spec.get("membership_model", "")
        
        if membership_model != "static":
            return AdversarialResult(
                attack_type="deadlock",
                success=False,
                description="Dynamic membership provides escape paths",
                severity="Low"
            )
        else:
            return AdversarialResult(
                attack_type="deadlock",
                success=True,
                description="Static membership may lead to deadlock",
                severity="Low"
            )
    
    def _test_livelock(self) -> AdversarialResult:
        """Test livelock resistance."""
        randomized_timeout = self.spec.get("randomized_timeout", False)
        
        if randomized_timeout:
            return AdversarialResult(
                attack_type="livelock",
                success=False,
                description="Randomization prevents livelock",
                severity="Low"
            )
        else:
            return AdversarialResult(
                attack_type="livelock",
                success=True,
                description="May be vulnerable to livelock",
                severity="Low"
            )
    
    def _test_recovery_failure(self) -> AdversarialResult:
        """Test recovery failure resistance."""
        recovery_strategy = self.spec.get("recovery_strategy", "")
        
        if recovery_strategy in ["snapshot_transfer", "hybrid_recovery"]:
            return AdversarialResult(
                attack_type="recovery_failure",
                success=False,
                description="Snapshot transfer enables recovery",
                severity="Medium"
            )
        else:
            return AdversarialResult(
                attack_type="recovery_failure",
                success=True,
                description="Recovery may fail under certain conditions",
                severity="Medium"
            )
    
    def _test_state_corruption(self) -> AdversarialResult:
        """Test state corruption resistance."""
        provides_safety = self.spec.get("provides_safety", True)
        
        if provides_safety:
            return AdversarialResult(
                attack_type="state_corruption",
                success=False,
                description="Safety guarantees prevent state corruption",
                severity="Critical"
            )
        else:
            return AdversarialResult(
                attack_type="state_corruption",
                success=True,
                description="No safety guarantees - state may be corrupted",
                severity="Critical"
            )


def evaluate_consensus_protocol(result: Dict) -> Dict:
    """Evaluate a consensus protocol."""
    spec = result.get("specification", {})
    simulations = result.get("simulations", [])
    
    evaluator = ConsensusAdversarialEvaluator(spec, simulations)
    return evaluator.evaluate()
