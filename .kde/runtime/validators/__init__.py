"""
KDE Runtime Validators

This module contains validators for the KDE Runtime validation framework.

Reference: LAB-033 (Runtime Validation Pipeline Investigation)
         LAB-034 (Runtime Validation Shadow Prototype)
         LAB-035 (Controlled Runtime Integration Trial)

Available Validators:
- MetadataValidator: Validates experiment artifact metadata

All validators are:
- Deterministic: Same input produces same output
- Read-only: Never modifies artifacts
- Non-blocking: Never stops execution
- Append-only: Outputs to separate files
"""

from .metadata import (
    MetadataValidator,
    MetadataSchema,
    ValidationResult,
    ValidationFinding,
    extract_metadata_from_file,
    validate_artifact,
    VERSION
)

__all__ = [
    "MetadataValidator",
    "MetadataSchema", 
    "ValidationResult",
    "ValidationFinding",
    "extract_metadata_from_file",
    "validate_artifact",
    "VERSION"
]

__version__ = VERSION
