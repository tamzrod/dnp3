"""
Protocol Synthesis Engine
Synthesizes novel secure communication protocol architectures.
"""

from .protocol_generator import ProtocolGenerator, ProtocolSpec
from .protocol_implementor import implement_protocol, ProtocolImplementor
from .protocol_runner import run_protocol_synthesis, ProtocolRun
from .protocol_knowledge import PROTOCOL_PATTERNS, EXISTING_PROTOCOLS

__version__ = "1.0.0"
__all__ = [
    "ProtocolGenerator",
    "ProtocolSpec", 
    "implement_protocol",
    "ProtocolImplementor",
    "run_protocol_synthesis",
    "ProtocolRun",
    "PROTOCOL_PATTERNS",
    "EXISTING_PROTOCOLS",
]
