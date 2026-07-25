"""
ECU Registry Module

Automatic discovery and registration of KDE engines and seeds.
"""

from .engine_registry import EngineRegistry
from .seed_registry import SeedRegistry

__all__ = ['EngineRegistry', 'SeedRegistry']
