"""
Consensus Adversarial Evaluation Engine
"""

from .consensus_adversarial import (
    ConsensusAdversarialEvaluator,
    evaluate_consensus_protocol,
    AdversarialResult,
)

__version__ = "1.0.0"
__all__ = [
    "ConsensusAdversarialEvaluator",
    "evaluate_consensus_protocol",
    "AdversarialResult",
]
