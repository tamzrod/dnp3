"""
Adversarial Evaluation Engine
Systematic security testing for synthesized protocols.
"""

from .adversarial_engine import (
    evaluate_protocol,
    ProtocolEvaluation,
    PhaseResult,
    Vulnerability,
    SpecificationReviewer,
    ImplementationValidator,
    FuzzTester,
    SelfCritic,
)

__version__ = "1.0.0"
__all__ = [
    "evaluate_protocol",
    "ProtocolEvaluation",
    "PhaseResult",
    "Vulnerability",
    "SpecificationReviewer",
    "ImplementationValidator",
    "FuzzTester",
    "SelfCritic",
]
