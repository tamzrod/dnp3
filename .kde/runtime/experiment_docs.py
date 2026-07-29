"""
Experiment Documentation Gate

Enforces experiment run documentation and evidence requirements.
- Requires runs/run-NNN.md for each experiment
- Requires evidence/ directory with captured outputs

REC-004/REC-005: From KDE-INV-GOV-SYN
"""

import os
import json
import hashlib
from pathlib import Path
from typing import List, Optional, Tuple
from dataclasses import dataclass, field, asdict
from datetime import datetime


@dataclass
class RunDocument:
    """Structure of an experiment run document."""
    run_id: str
    timestamp: str
    commands: List[str]
    outputs: List[str]
    result: str  # PASS, FAIL, PARTIAL
    notes: str = ""
    evidence_files: List[str] = field(default_factory=list)


@dataclass
class ExperimentDocResult:
    """Result of experiment documentation check."""
    complete: bool
    missing: List[str]
    required: List[str]
    run_count: int
    evidence_count: int


class ExperimentDocsGate:
    """
    Enforces experiment documentation requirements.
    
    Requirements:
    1. Each experiment must have at least one runs/run-*.md
    2. Each experiment should have evidence/ with captured outputs
    
    Usage:
        gate = ExperimentDocsGate()
        result = gate.check_experiment("LAB-063")
        if not result.complete:
            print(f"Missing: {result.missing}")
    """
    
    REQUIRED_STRUCTURE = {
        "runs": True,           # Must have runs/ directory
        "runs_pattern": "run-*.md",  # Pattern for run files
        "evidence": True,       # Should have evidence/ directory
        "evidence_pattern": "*",  # Any files in evidence/
    }
    
    def __init__(self, kde_root: str = "/workspace/project/dnp3"):
        self.kde_root = kde_root
    
    def check_experiment(
        self, 
        experiment_id: str,
        experiment_path: str = None
    ) -> ExperimentDocResult:
        """
        Check if an experiment has required documentation.
        
        Args:
            experiment_id: The experiment ID (e.g., "LAB-063")
            experiment_path: Optional explicit path
            
        Returns:
            ExperimentDocResult with completeness status
        """
        if experiment_path is None:
            experiment_path = self._find_experiment(experiment_id)
        
        if experiment_path is None:
            return ExperimentDocResult(
                complete=False,
                missing=["Experiment directory not found"],
                required=["experiment_dir"],
                run_count=0,
                evidence_count=0
            )
        
        missing = []
        required = []
        
        # Check runs directory
        runs_dir = os.path.join(experiment_path, "runs")
        if not os.path.exists(runs_dir):
            missing.append("runs/")
            required.append("runs/")
        else:
            run_files = [f for f in os.listdir(runs_dir) if f.startswith("run-") and f.endswith(".md")]
            if not run_files:
                missing.append("runs/run-*.md")
                required.append("runs/run-*.md")
        
        # Check evidence directory
        evidence_dir = os.path.join(experiment_path, "evidence")
        evidence_count = 0
        if os.path.exists(evidence_dir):
            evidence_count = len([f for f in os.listdir(evidence_dir) if not f.startswith(".")])
            if evidence_count == 0:
                missing.append("evidence/* (captured outputs)")
                required.append("evidence/")
        
        return ExperimentDocResult(
            complete=len(missing) == 0,
            missing=missing,
            required=required,
            run_count=len(run_files) if os.path.exists(runs_dir) else 0,
            evidence_count=evidence_count
        )
    
    def create_run_document(
        self,
        experiment_path: str,
        run_id: int,
        commands: List[str],
        outputs: List[str],
        result: str,
        notes: str = ""
    ) -> str:
        """
        Create a run document with captured evidence.
        
        Args:
            experiment_path: Path to experiment
            run_id: Run number
            commands: Commands executed
            outputs: Outputs captured
            result: PASS/FAIL/PARTIAL
            notes: Additional notes
            
        Returns:
            Path to created run document
        """
        runs_dir = os.path.join(experiment_path, "runs")
        os.makedirs(runs_dir, exist_ok=True)
        
        run_filename = f"run-{run_id:03d}.md"
        run_path = os.path.join(runs_dir, run_filename)
        
        # Capture evidence
        evidence_files = []
        if outputs:
            evidence_dir = os.path.join(experiment_path, "evidence")
            os.makedirs(evidence_dir, exist_ok=True)
            
            # Save output to evidence
            evidence_file = f"output-run-{run_id:03d}.txt"
            evidence_path = os.path.join(evidence_dir, evidence_file)
            with open(evidence_path, 'w') as f:
                f.write("\n".join(outputs))
            evidence_files.append(evidence_file)
        
        # Create run document
        doc = RunDocument(
            run_id=run_filename,
            timestamp=datetime.now().isoformat(),
            commands=commands,
            outputs=[],  # Just reference evidence
            result=result,
            notes=notes,
            evidence_files=evidence_files
        )
        
        content = self._format_run_document(doc)
        with open(run_path, 'w') as f:
            f.write(content)
        
        return run_path
    
    def _format_run_document(self, doc: RunDocument) -> str:
        """Format a run document."""
        evidence_text = ""
        if doc.evidence_files:
            evidence_text = "\n## Evidence\n\n" + "\n".join(
                f"- `{ef}`" for ef in doc.evidence_files
            )
        
        return f"""# {doc.run_id}

**Timestamp**: {doc.timestamp}  
**Result**: {doc.result}

## Commands Executed

```
{chr(10).join(doc.commands)}
```

## Notes

{doc.notes}
{evidence_text}

---
*Generated by ExperimentDocsGate*
"""
    
    def _find_experiment(self, experiment_id: str) -> Optional[str]:
        """Find experiment directory by ID."""
        # Check common locations
        search_paths = [
            os.path.join(self.kde_root, "laboratory", "experiments", experiment_id),
            os.path.join(self.kde_root, "laboratory", "investigations"),
        ]
        
        for base in search_paths:
            if os.path.exists(base):
                # Search recursively
                for root, dirs, files in os.walk(base):
                    if experiment_id in root.split(os.sep):
                        return root
        
        return None
    
    def format_incomplete_message(self, result: ExperimentDocResult) -> str:
        """Format message for incomplete documentation."""
        missing_text = "\n  - ".join(result.missing)
        return f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║  ⚠️  EXPERIMENT DOCUMENTATION REQUIRED                                       ║
║  ═══════════════════════════════════════════════════                         ║
║                                                                              ║
║  Missing Required Documentation:                                             ║
║    - {missing_text}                                                         ║
║                                                                              ║
║  Current Status:                                                              ║
║    - Run files: {result.run_count}                                                   ║
║    - Evidence files: {result.evidence_count}                                              ║
║                                                                              ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Required:                                                                    ║
║  1. Create runs/run-001.md documenting test execution                        ║
║  2. Save command outputs to evidence/ directory                               ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
"""


# Global instance
_global_gate: Optional[ExperimentDocsGate] = None


def get_experiment_gate(kde_root: str = "/workspace/project/dnp3") -> ExperimentDocsGate:
    """Get or create the global ExperimentDocsGate instance."""
    global _global_gate
    if _global_gate is None:
        _global_gate = ExperimentDocsGate(kde_root)
    return _global_gate


def check_experiment_docs(experiment_id: str) -> ExperimentDocResult:
    """Convenience function to check experiment documentation."""
    gate = get_experiment_gate()
    return gate.check_experiment(experiment_id)
