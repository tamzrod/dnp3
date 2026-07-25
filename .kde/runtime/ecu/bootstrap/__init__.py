"""
ECU Bootstrap Module

Bootstrap integration for the Runtime ECU.
"""

import os
from typing import Optional, Dict, Any
from pathlib import Path

from .. import RuntimeECU, create_ecu, ECUInitializationResult


class ECUBootstrap:
    """
    Bootstrap integration for the Runtime ECU.
    
    Responsibilities:
    - Locate KDE runtime root
    - Initialize ECU components
    - Validate runtime environment
    - Provide runtime validation report
    """
    
    def __init__(self, kde_root: Optional[str] = None):
        """
        Initialize the ECU Bootstrap.
        
        Args:
            kde_root: Optional KDE root path. If not provided,
                     will attempt to locate automatically.
        """
        self.kde_root = kde_root or self._locate_kde_root()
        self.ecu: Optional[RuntimeECU] = None
        self._validation_report: Optional[Dict[str, Any]] = None
    
    def _locate_kde_root(self) -> str:
        """
        Locate the KDE runtime root directory.
        
        Returns:
            Path to KDE root
        
        Raises:
            RuntimeError: If KDE root cannot be located
        """
        # Check environment variable
        env_root = os.environ.get('KDE_ROOT')
        if env_root and os.path.exists(env_root):
            return env_root
        
        # Check common locations
        common_locations = [
            '/workspace/project/dnp3/.kde',
            os.path.expanduser('~/.kde'),
            './.kde',
        ]
        
        for location in common_locations:
            if os.path.exists(location):
                return location
        
        # Default to current directory structure
        return '/workspace/project/dnp3/.kde'
    
    def bootstrap(self) -> ECUInitializationResult:
        """
        Bootstrap the ECU and all components.
        
        Returns:
            ECUInitializationResult
        """
        # Create ECU
        self.ecu = RuntimeECU(self.kde_root)
        
        # Initialize
        result = self.ecu.initialize()
        
        # Generate validation report
        if result.success:
            self._validation_report = self._generate_validation_report()
        
        return result
    
    def _generate_validation_report(self) -> Dict[str, Any]:
        """
        Generate runtime validation report.
        
        Returns:
            Validation report dictionary
        """
        if not self.ecu:
            return {"error": "ECU not initialized"}
        
        # Get component states
        engine_registry = self.ecu.engine_registry.get_registry_summary()
        seed_registry = self.ecu.seed_registry.get_registry_summary()
        policy_summary = self.ecu.policy_layer.get_policy_summary()
        runtime_state = self.ecu.get_runtime_state()
        
        # Check for issues
        issues = []
        warnings = []
        
        if engine_registry.get("total_engines", 0) == 0:
            issues.append("No engines registered")
        
        if seed_registry.get("total_seeds", 0) == 0:
            issues.append("No seeds registered")
        
        if policy_summary.get("total_violations", 0) > 0:
            warnings.append(
                f"{policy_summary['total_violations']} policy violations detected"
            )
        
        # Validate directories
        required_dirs = ['engines', 'seeds', 'runtime']
        for dir_name in required_dirs:
            dir_path = os.path.join(self.kde_root, dir_name)
            if not os.path.exists(dir_path):
                issues.append(f"Required directory missing: {dir_name}")
        
        return {
            "validation_timestamp": runtime_state.get("last_initialization"),
            "kde_root": self.kde_root,
            "components": {
                "engine_registry": engine_registry,
                "seed_registry": seed_registry,
                "policy_layer": policy_summary
            },
            "issues": issues,
            "warnings": warnings,
            "status": "VALID" if not issues else "INVALID",
            "ready_for_execution": (
                len(issues) == 0 and
                engine_registry.get("total_engines", 0) > 0
            )
        }
    
    def get_validation_report(self) -> Optional[Dict[str, Any]]:
        """
        Get the validation report.
        
        Returns:
            Validation report or None if not generated
        """
        return self._validation_report
    
    def validate_runtime(self) -> bool:
        """
        Validate the runtime environment.
        
        Returns:
            True if valid, False otherwise
        """
        if not self._validation_report:
            self._generate_validation_report()
        
        return self._validation_report.get("status") == "VALID"
    
    def get_runtime_info(self) -> Dict[str, Any]:
        """
        Get comprehensive runtime information.
        
        Returns:
            Runtime information dictionary
        """
        if not self.ecu:
            return {"error": "ECU not initialized"}
        
        return {
            "initialized": self.ecu.state.initialized,
            "kde_root": self.kde_root,
            "validation_report": self._validation_report,
            "runtime_state": self.ecu.get_runtime_state()
        }


def bootstrap_ecu(kde_root: Optional[str] = None) -> RuntimeECU:
    """
    Bootstrap a Runtime ECU.
    
    Args:
        kde_root: Optional KDE root path
    
    Returns:
        Initialized RuntimeECU
    
    Raises:
        RuntimeError: If bootstrap fails
    """
    bootstrap = ECUBootstrap(kde_root)
    result = bootstrap.bootstrap()
    
    if not result.success:
        raise RuntimeError(
            f"ECU bootstrap failed: {', '.join(result.errors)}"
        )
    
    if not bootstrap.validate_runtime():
        report = bootstrap.get_validation_report()
        issues = report.get("issues", []) if report else []
        raise RuntimeError(
            f"Runtime validation failed: {', '.join(issues)}"
        )
    
    return bootstrap.ecu
