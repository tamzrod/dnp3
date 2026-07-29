"""
Synthesis Checkpoint

Enforces investigation synthesis requirement before new experiments.
Requires reviewing prior experiments before creating new ones.

REC-003: From KDE-INV-GOV-SYN
"""

import os
import re
from pathlib import Path
from typing import List, Optional, Tuple
from dataclasses import dataclass, field
from datetime import datetime


@dataclass
class SynthesisStatus:
    """Status of investigation synthesis."""
    investigation_id: str
    synthesis_complete: bool
    experiments_reviewed: List[str]
    experiments_total: int
    last_synthesis: Optional[str] = None
    blockers: List[str] = field(default_factory=list)


@dataclass
class ExperimentCreationResult:
    """Result of experiment creation check."""
    allowed: bool
    reason: str
    synthesis_status: Optional[SynthesisStatus] = None
    blocker_message: Optional[str] = None


class SynthesisCheckpoint:
    """
    Enforces synthesis requirement before new experiments.
    
    Rule: Cannot create new experiment without reviewing prior experiments.
    
    Usage:
        checkpoint = SynthesisCheckpoint()
        result = checkpoint.allow_experiment_creation("INV-066", "LAB-068")
        if not result.allowed:
            print(result.blocker_message)
    """
    
    def __init__(self, kde_root: str = "/workspace/project/dnp3"):
        self.kde_root = kde_root
        self._synthesis_cache: dict = {}
    
    def check_synthesis_required(
        self, 
        investigation_id: str
    ) -> SynthesisStatus:
        """
        Check if synthesis is required for an investigation.
        
        Args:
            investigation_id: The investigation ID
            
        Returns:
            SynthesisStatus with synthesis requirements
        """
        # Get investigation path
        inv_path = os.path.join(
            self.kde_root, 
            "laboratory", 
            "investigations", 
            investigation_id
        )
        
        if not os.path.exists(inv_path):
            return SynthesisStatus(
                investigation_id=investigation_id,
                synthesis_complete=False,
                experiments_reviewed=[],
                experiments_total=0,
                blockers=["Investigation not found"]
            )
        
        # Find all experiments
        experiments = self._find_experiments(inv_path)
        
        # Check for synthesis completion
        synthesis_file = os.path.join(inv_path, "synthesis.md")
        synthesis_complete = os.path.exists(synthesis_file)
        
        # Check for new experiments since last synthesis
        blockers = []
        if synthesis_complete and experiments:
            # Get last synthesis time
            synth_mtime = os.path.getmtime(synthesis_file)
            
            # Check for experiments after synthesis
            new_experiments = [
                exp for exp in experiments
                if os.path.getmtime(exp) > synth_mtime
            ]
            
            if new_experiments:
                blockers.append(
                    f"{len(new_experiments)} new experiment(s) since last synthesis"
                )
        
        return SynthesisStatus(
            investigation_id=investigation_id,
            synthesis_complete=synthesis_complete and len(blockers) == 0,
            experiments_reviewed=[],  # Would need tracking
            experiments_total=len(experiments),
            last_synthesis=self._get_synthesis_time(synthesis_file) if synthesis_complete else None,
            blockers=blockers
        )
    
    def allow_experiment_creation(
        self, 
        investigation_id: str,
        new_experiment_id: str
    ) -> ExperimentCreationResult:
        """
        Check if a new experiment can be created.
        
        Args:
            investigation_id: The investigation ID
            new_experiment_id: The new experiment ID being created
            
        Returns:
            ExperimentCreationResult with permission
        """
        status = self.check_synthesis_required(investigation_id)
        
        if not status.experiments_total:
            # First experiment - always allowed
            return ExperimentCreationResult(
                allowed=True,
                reason="First experiment in investigation"
            )
        
        if status.blockers:
            # Synthesis required
            blocker_msg = self._format_blocker_message(status)
            return ExperimentCreationResult(
                allowed=False,
                reason="Synthesis required before new experiments",
                synthesis_status=status,
                blocker_message=blocker_msg
            )
        
        return ExperimentCreationResult(
            allowed=True,
            reason="Synthesis complete",
            synthesis_status=status
        )
    
    def _find_experiments(self, investigation_path: str) -> List[str]:
        """Find all experiment directories in an investigation."""
        experiments = []
        
        # Check experiments/ subdirectory
        exp_dir = os.path.join(investigation_path, "experiments")
        if os.path.exists(exp_dir):
            for item in os.listdir(exp_dir):
                if item.startswith("LAB-"):
                    experiments.append(os.path.join(exp_dir, item))
        
        return sorted(experiments)
    
    def _get_synthesis_time(self, synthesis_path: str) -> Optional[str]:
        """Get the timestamp of last synthesis."""
        try:
            mtime = os.path.getmtime(synthesis_path)
            return datetime.fromtimestamp(mtime).isoformat()
        except:
            return None
    
    def _format_blocker_message(self, status: SynthesisStatus) -> str:
        """Format blocker message for display."""
        blockers_text = "\n  - ".join(status.blockers)
        return f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║  ⚠️  SYNTHESIS CHECKPOINT                                                   ║
║  ═════════════════════════════════════════════                              ║
║                                                                              ║
║  Investigation: {status.investigation_id:<50}       ║
║  Total Experiments: {status.experiments_total:<42}       ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Synthesis Required:                                                         ║
║    - {blockers_text}                                                         ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Required Action:                                                            ║
║  1. Review all {status.experiments_total} experiment(s) in this investigation          ║
║  2. Create synthesis document (synthesis.md)                                 ║
║  3. Include: Findings, Patterns, Next Steps                                 ║
║  4. Then retry experiment creation                                          ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
"""
    
    def mark_synthesis_complete(
        self, 
        investigation_id: str,
        synthesis_path: str = None
    ) -> bool:
        """
        Mark synthesis as complete for an investigation.
        
        Args:
            investigation_id: The investigation ID
            synthesis_path: Optional path to synthesis document
            
        Returns:
            True if marked successfully
        """
        if synthesis_path is None:
            synthesis_path = os.path.join(
                self.kde_root,
                "laboratory",
                "investigations",
                investigation_id,
                "synthesis.md"
            )
        
        # Create synthesis file if it doesn't exist
        if not os.path.exists(synthesis_path):
            try:
                os.makedirs(os.path.dirname(synthesis_path), exist_ok=True)
                with open(synthesis_path, 'w') as f:
                    f.write(f"""# {investigation_id} - Synthesis

**Generated**: {datetime.now().isoformat()}

## Summary

## Findings

## Patterns

## Next Steps

## Open Questions

---
*Generated by SynthesisCheckpoint*
""")
            except Exception as e:
                return False
        
        # Update cache
        self._synthesis_cache[investigation_id] = True
        return True


# Global instance
_global_checkpoint: Optional[SynthesisCheckpoint] = None


def get_synthesis_checkpoint(kde_root: str = "/workspace/project/dnp3") -> SynthesisCheckpoint:
    """Get or create the global SynthesisCheckpoint instance."""
    global _global_checkpoint
    if _global_checkpoint is None:
        _global_checkpoint = SynthesisCheckpoint(kde_root)
    return _global_checkpoint


def allow_experiment(
    investigation_id: str,
    new_experiment_id: str
) -> ExperimentCreationResult:
    """
    Convenience function to check experiment creation.
    
    Usage:
        result = allow_experiment("INV-066", "LAB-068")
        if not result.allowed:
            print(result.blocker_message)
    """
    checkpoint = get_synthesis_checkpoint()
    return checkpoint.allow_experiment_creation(investigation_id, new_experiment_id)
