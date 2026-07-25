"""
Runtime Validation Module

This module provides integration between the KDE Runtime and the validators.

Reference: LAB-035 (Controlled Runtime Integration Trial)

The validation module is designed to be:
- Non-invasive: Does not modify runtime behavior
- Append-only: Outputs to separate files
- Non-blocking: Never stops execution
- Deterministic: Same input produces same output
"""

import os
from pathlib import Path
from typing import Optional, List, Dict, Any
from datetime import datetime

from .metadata import MetadataValidator, extract_metadata_from_file


class ValidationRunner:
    """
    Runs validation on experiment artifacts.
    
    This class orchestrates validation without modifying
    runtime behavior or artifacts.
    """
    
    def __init__(self, base_path: str = None, enabled: bool = True):
        """
        Initialize the validation runner.
        
        Args:
            base_path: Base path for experiment artifacts
            enabled: Whether validation is enabled
        """
        self.base_path = base_path or os.getcwd()
        self.enabled = enabled
        self.last_result = None
    
    def validate_experiment(self, experiment_id: str) -> Optional[Dict[str, Any]]:
        """
        Validate an experiment's main artifact.
        
        Args:
            experiment_id: The experiment ID (e.g., "LAB-031")
            
        Returns:
            Validation result dictionary, or None if validation disabled
        """
        if not self.enabled:
            return None
        
        # Find experiment.md
        experiment_path = os.path.join(
            self.base_path,
            "laboratory",
            "experiments",
            experiment_id,
            "experiment.md"
        )
        
        if not os.path.exists(experiment_path):
            return None
        
        # Create validation output directory
        validation_dir = os.path.join(
            self.base_path,
            "laboratory",
            "experiments",
            experiment_id,
            "validation"
        )
        os.makedirs(validation_dir, exist_ok=True)
        
        # Run validation
        validator = MetadataValidator()
        metadata = extract_metadata_from_file(experiment_path)
        result = validator.validate(metadata, experiment_path)
        
        # Write validation report
        report_name = f"metadata-validation-experiment.md"
        report_path = os.path.join(validation_dir, report_name)
        
        with open(report_path, 'w') as f:
            f.write(result.to_markdown())
        
        self.last_result = result.to_dict()
        
        return self.last_result
    
    def validate_all_experiments(self) -> List[Dict[str, Any]]:
        """
        Validate all experiments in the laboratory.
        
        Returns:
            List of validation results
        """
        if not self.enabled:
            return []
        
        results = []
        experiments_dir = os.path.join(
            self.base_path,
            "laboratory",
            "experiments"
        )
        
        if not os.path.exists(experiments_dir):
            return []
        
        # Find all experiment directories
        for exp_dir in os.listdir(experiments_dir):
            exp_path = os.path.join(experiments_dir, exp_dir)
            
            if os.path.isdir(exp_path):
                experiment_md = os.path.join(exp_path, "experiment.md")
                
                if os.path.exists(experiment_md):
                    result = self.validate_experiment(exp_dir)
                    if result:
                        results.append(result)
        
        return results


def run_validation(experiment_id: str = None, base_path: str = None) -> Dict[str, Any]:
    """
    Run validation for an experiment or all experiments.
    
    This is the main entry point for runtime validation.
    
    Args:
        experiment_id: Specific experiment to validate, or None for all
        base_path: Base path for the laboratory
        
    Returns:
        Validation summary
    """
    runner = ValidationRunner(base_path=base_path, enabled=True)
    
    if experiment_id:
        result = runner.validate_experiment(experiment_id)
        return {
            "experiment_id": experiment_id,
            "validated": result is not None,
            "result": result
        }
    else:
        results = runner.validate_all_experiments()
        return {
            "experiments_validated": len(results),
            "results": results
        }


# Configuration for runtime integration
DEFAULT_CONFIG = {
    "validation": {
        "enabled": True,
        "validators": ["metadata"],
        "output_format": "markdown"
    }
}
