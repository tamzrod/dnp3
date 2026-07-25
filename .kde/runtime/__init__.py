"""
KDE Runtime: Knowledge-on-Demand

This module implements the validated Knowledge-on-Demand architecture
for the KDE Laboratory.

Components:
- Knowledge Catalog
- Retrieval Engine
- SOP-005 Execution
- Runtime Instrumentation
- Decision Attribution
- Runtime ECU (Execution Control Unit)
"""

__version__ = "1.0.0"
__author__ = "KDE Runtime Team"

# ECU exports
from .ecu import RuntimeECU, create_ecu, ECUInitializationResult, ECUExecutionResult
from .ecu.bootstrap import ECUBootstrap, bootstrap_ecu
from .ecu.models import (
    EngineStatus, EngineStability, SeedStatus, CapabilityType,
    ExecutionMode, ConsensusStrategy, PolicyViolation
)

__all__ = [
    # ECU
    'RuntimeECU',
    'create_ecu',
    'ECUInitializationResult',
    'ECUExecutionResult',
    'ECUBootstrap',
    'bootstrap_ecu',
    # Models
    'EngineStatus',
    'EngineStability',
    'SeedStatus',
    'CapabilityType',
    'ExecutionMode',
    'ConsensusStrategy',
    'PolicyViolation',
]
