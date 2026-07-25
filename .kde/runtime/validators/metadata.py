"""
Metadata Validator

This module implements the Metadata Validator for the KDE Runtime.

The validator is designed to be:
- Deterministic: Same input produces same output
- Read-only: Never modifies artifacts
- Non-blocking: Never stops execution
- Append-only: Outputs to separate file

Reference: LAB-035 (Controlled Runtime Integration Trial)
"""

import re
import os
from datetime import datetime
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any


VERSION = "0.1.0"


@dataclass
class ValidationFinding:
    """A single validation finding."""
    rule: str
    severity: str
    message: str
    field: Optional[str] = None
    
    def to_dict(self) -> Dict:
        return {
            "rule": self.rule,
            "severity": self.severity,
            "message": self.message,
            "field": self.field
        }


@dataclass
class ValidationResult:
    """Result of metadata validation."""
    validator: str = "MetadataValidator"
    version: str = VERSION
    status: str = "PASS"
    artifact: str = ""
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat() + "Z")
    findings: List[ValidationFinding] = field(default_factory=list)
    fields_validated: Dict[str, Dict] = field(default_factory=dict)
    
    def to_dict(self) -> Dict:
        return {
            "validator": self.validator,
            "version": self.version,
            "status": self.status,
            "artifact": self.artifact,
            "timestamp": self.timestamp,
            "findings": [f.to_dict() for f in self.findings],
            "fields_validated": self.fields_validated
        }
    
    def to_markdown(self) -> str:
        """Generate validation report in Markdown format."""
        lines = [
            "# Metadata Validation Report",
            "",
            f"**Artifact**: {self.artifact}",
            f"**Validator**: {self.validator} v{self.version}",
            f"**Timestamp**: {self.timestamp}",
            f"**Result**: {self.status}",
            ""
        ]
        
        if self.fields_validated:
            lines.append("## Validated Fields")
            lines.append("")
            lines.append("| Field | Required | Present | Valid |")
            lines.append("|-------|----------|---------|-------|")
            
            for field_name, info in self.fields_validated.items():
                required = "Yes" if info.get("required", False) else "No"
                present = "Yes" if info.get("present", False) else "No"
                valid = "Yes" if info.get("valid", False) else "No"
                lines.append(f"| {field_name} | {required} | {present} | {valid} |")
            
            lines.append("")
        
        if self.findings:
            lines.append("## Findings")
            lines.append("")
            lines.append("| Rule | Severity | Message |")
            lines.append("|------|----------|---------|")
            
            for finding in self.findings:
                lines.append(f"| {finding.rule} | {finding.severity} | {finding.message} |")
            
            lines.append("")
        
        lines.append("---")
        
        if not self.findings:
            lines.append("")
            lines.append("*No findings.*")
        
        return "\n".join(lines)


class MetadataSchema:
    """Schema definition for experiment metadata."""
    
    REQUIRED_FIELDS = ["experiment_id", "title", "status", "created"]
    
    OPTIONAL_FIELDS = ["updated", "engine", "seed", "category", "tags", "notes"]
    
    # Note: IN_PROGRESS is kept for backward compatibility with existing experiments
    VALID_STATUS_VALUES = ["DRAFT", "ACTIVE", "IN_PROGRESS", "COMPLETE", "SUSPENDED", "ARCHIVED"]
    
    EXPERIMENT_ID_PATTERN = r"^LAB-\d{3}$"
    
    @classmethod
    def get_all_fields(cls) -> List[str]:
        return cls.REQUIRED_FIELDS + cls.OPTIONAL_FIELDS
    
    @classmethod
    def is_required(cls, field: str) -> bool:
        return field in cls.REQUIRED_FIELDS


class MetadataValidator:
    """
    Validates experiment artifact metadata against the defined schema.
    
    This validator is:
    - Deterministic: Same input = same output
    - Read-only: Never modifies input
    - Non-blocking: Never raises exceptions
    - Append-only: Outputs to separate file
    """
    
    def __init__(self):
        self.schema = MetadataSchema()
        self.result: Optional[ValidationResult] = None
    
    def validate(self, metadata: Dict[str, Any], artifact_path: str = "") -> ValidationResult:
        """
        Execute validation on artifact metadata.
        
        Args:
            metadata: Dictionary of metadata fields
            artifact_path: Path to the artifact file
            
        Returns:
            ValidationResult with PASS/WARNING/ERROR status
        """
        self.result = ValidationResult(
            artifact=artifact_path or "unknown"
        )
        
        # Initialize field tracking
        for field in self.schema.get_all_fields():
            present = field in metadata
            self.result.fields_validated[field] = {
                "required": self.schema.is_required(field),
                "present": present,
                "valid": False  # Will be set by validators
            }
        
        # Run validation rules
        self._check_required_fields(metadata)
        self._check_field_types(metadata)
        self._check_experiment_id_format(metadata)
        self._check_timestamp_format(metadata)
        self._check_status_enum(metadata)
        
        # Determine overall status
        has_errors = any(f.severity == "ERROR" for f in self.result.findings)
        has_warnings = any(f.severity == "WARNING" for f in self.result.findings)
        
        if has_errors:
            self.result.status = "ERROR"
        elif has_warnings:
            self.result.status = "WARNING"
        else:
            self.result.status = "PASS"
        
        # Update field validity
        for field in self.result.fields_validated:
            if field in metadata:
                # Check if field has any errors
                has_field_error = any(
                    f.severity == "ERROR" and f.field == field
                    for f in self.result.findings
                )
                self.result.fields_validated[field]["valid"] = not has_field_error
        
        return self.result
    
    def _add_finding(self, rule: str, severity: str, message: str, field: str = None):
        """Record a validation finding."""
        self.result.findings.append(ValidationFinding(
            rule=rule,
            severity=severity,
            message=message,
            field=field
        ))
    
    def _check_required_fields(self, metadata: Dict[str, Any]):
        """Rule: required_fields - Check all required fields are present."""
        for field in self.schema.REQUIRED_FIELDS:
            if field not in metadata:
                self._add_finding(
                    rule="required_fields",
                    severity="ERROR",
                    message=f"Required field '{field}' is missing",
                    field=field
                )
                self.result.fields_validated[field]["present"] = False
    
    def _check_field_types(self, metadata: Dict[str, Any]):
        """Rule: field_types - Check field values match expected types."""
        type_checks = {
            "experiment_id": str,
            "title": str,
            "status": str,
            "created": str,
            "updated": str,
            "engine": str,
            "seed": str,
            "category": str,
            "notes": str,
            "tags": list
        }
        
        for field, expected_type in type_checks.items():
            if field in metadata:
                value = metadata[field]
                if value is not None and not isinstance(value, expected_type):
                    self._add_finding(
                        rule="field_types",
                        severity="ERROR",
                        message=f"Field '{field}' has invalid type (expected {expected_type.__name__}, got {type(value).__name__})",
                        field=field
                    )
    
    def _check_experiment_id_format(self, metadata: Dict[str, Any]):
        """Rule: experiment_id_format - Check experiment ID matches expected pattern."""
        if "experiment_id" in metadata:
            exp_id = metadata["experiment_id"]
            if exp_id and not re.match(self.schema.EXPERIMENT_ID_PATTERN, exp_id):
                self._add_finding(
                    rule="experiment_id_format",
                    severity="ERROR",
                    message=f"experiment_id '{exp_id}' does not match expected pattern LAB-NNN",
                    field="experiment_id"
                )
    
    def _check_timestamp_format(self, metadata: Dict[str, Any]):
        """Rule: timestamp_format - Check timestamp fields are valid ISO8601."""
        timestamp_fields = ["created", "updated"]
        timestamp_pattern = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?$"
        
        for field in timestamp_fields:
            if field in metadata:
                value = metadata[field]
                if value and not re.match(timestamp_pattern, str(value)):
                    self._add_finding(
                        rule="timestamp_format",
                        severity="ERROR",
                        message=f"Field '{field}' is not in valid ISO8601 format",
                        field=field
                    )
    
    def _check_status_enum(self, metadata: Dict[str, Any]):
        """Rule: status_enum - Check status is valid enum value."""
        if "status" in metadata:
            status = metadata["status"]
            if status and status not in self.schema.VALID_STATUS_VALUES:
                self._add_finding(
                    rule="status_enum",
                    severity="ERROR",
                    message=f"status '{status}' is not a valid value (must be one of: {', '.join(self.schema.VALID_STATUS_VALUES)})",
                    field="status"
                )


def extract_metadata_from_file(file_path: str) -> Dict[str, Any]:
    """
    Extract metadata from a markdown file.
    
    Supports multiple formats:
    1. YAML frontmatter (--- ... ---)
    2. Markdown metadata format (**Key**: Value)
    """
    metadata = {}
    
    try:
        with open(file_path, 'r') as f:
            content = f.read()
        
        # Check for YAML frontmatter
        if content.startswith('---'):
            parts = content.split('---', 2)
            if len(parts) >= 3:
                frontmatter = parts[1]
                
                # Try yaml first
                try:
                    import yaml
                    metadata = yaml.safe_load(frontmatter) or {}
                except ImportError:
                    # Fall back to regex parsing for common fields
                    metadata = _parse_frontmatter_regex(frontmatter)
        
        # If no YAML frontmatter, try markdown format
        if not metadata:
            metadata = extract_metadata_from_markdown(file_path)
    
    except Exception:
        pass
    
    return metadata


def _parse_frontmatter_regex(frontmatter: str) -> Dict[str, Any]:
    """
    Parse YAML-like frontmatter using regex.
    
    This is a fallback when yaml module is not available.
    """
    metadata = {}
    
    # Pattern for key: value
    pattern = r'^(\w+):\s*(.+)$'
    
    for line in frontmatter.split('\n'):
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        
        match = re.match(pattern, line)
        if match:
            key = match.group(1)
            value = match.group(2).strip()
            
            # Remove quotes if present
            if value.startswith('"') and value.endswith('"'):
                value = value[1:-1]
            elif value.startswith("'") and value.endswith("'"):
                value = value[1:-1]
            
            # Handle list-like values
            if value.startswith('[') and value.endswith(']'):
                value = [v.strip() for v in value[1:-1].split(',')]
            
            metadata[key] = value
    
    return metadata


def extract_metadata_from_markdown(file_path: str) -> Dict[str, Any]:
    """
    Extract metadata from markdown file format.
    
    Handles the format:
    # Experiment: LAB-031 - Title
    
    **Experiment ID**: LAB-031
    **Created**: 2026-07-22
    **Status**: COMPLETE
    ...
    """
    metadata = {}
    
    try:
        with open(file_path, 'r') as f:
            content = f.read()
        
        lines = content.split('\n')
        
        # Pattern for **Key**: Value
        key_pattern = r'\*\*(\w+(?:\s+\w+)?)\*\*:\s*(.+)'
        
        # Pattern for first line: # Experiment: LAB-XXX - Title
        title_pattern = r'^#\s+Experiment:\s+(LAB-\d+)\s*-\s*(.+)$'
        
        for line in lines[:15]:  # Only check first 15 lines
            # Check for title pattern
            title_match = re.match(title_pattern, line)
            if title_match:
                metadata['experiment_id'] = title_match.group(1)
                metadata['title'] = title_match.group(2).strip()
            
            # Check for key-value pattern
            key_match = re.match(key_pattern, line)
            if key_match:
                key = key_match.group(1).lower().replace(' ', '_')
                value = key_match.group(2).strip()
                
                # Handle status mapping
                if key == 'status':
                    # Status is already in correct format
                    metadata[key] = value
                elif key == 'created':
                    # Convert date format
                    metadata['created'] = value + 'T00:00:00Z'
                else:
                    metadata[key] = value
        
        # Handle engine and seed fields
        for line in lines[:15]:
            if '**Engine**' in line:
                match = re.search(r'\*\*Engine\*\*:\s*(.+)', line)
                if match:
                    metadata['engine'] = match.group(1).strip()
            if '**Seed**' in line:
                match = re.search(r'\*\*Seed\*\*:\s*(.+)', line)
                if match:
                    metadata['seed'] = match.group(1).strip()
            if '**Category**' in line:
                match = re.search(r'\*\*Category\*\*:\s*(.+)', line)
                if match:
                    metadata['category'] = match.group(1).strip()
    
    except Exception:
        pass
    
    return metadata


def validate_artifact(artifact_path: str, output_dir: str = None) -> ValidationResult:
    """
    Validate an artifact and optionally write the report.
    
    Args:
        artifact_path: Path to the artifact file
        output_dir: Optional directory for validation report
        
    Returns:
        ValidationResult
    """
    # Extract metadata
    metadata = extract_metadata_from_file(artifact_path)
    
    # Validate
    validator = MetadataValidator()
    result = validator.validate(metadata, artifact_path)
    
    # Write report if output directory specified
    if output_dir:
        os.makedirs(output_dir, exist_ok=True)
        
        # Generate report filename
        artifact_name = os.path.basename(artifact_path)
        report_name = f"metadata-validation-{artifact_name}"
        report_path = os.path.join(output_dir, report_name)
        
        with open(report_path, 'w') as f:
            f.write(result.to_markdown())
    
    return result


# Demonstration function
if __name__ == "__main__":
    # Test with sample metadata
    sample_metadata = {
        "experiment_id": "LAB-001",
        "title": "Test Experiment",
        "status": "COMPLETE",
        "created": "2026-07-22T12:00:00Z",
        "engine": "KDE-ENGINE-002",
        "category": "Test"
    }
    
    validator = MetadataValidator()
    result = validator.validate(sample_metadata, "experiment.md")
    
    print("Validation Result:")
    print(f"  Status: {result.status}")
    print(f"  Findings: {len(result.findings)}")
    print()
    print(result.to_markdown())
