"""
Live Status Surface Generator

Generates a single-pane status view of KDE Laboratory.
Shows active investigations, experiments, and pending items.

REC-009: From KDE-INV-GOV-SYN
"""

import os
import re
from pathlib import Path
from typing import List, Dict, Optional
from dataclasses import dataclass, field
from datetime import datetime


@dataclass
class InvestigationSummary:
    """Summary of an investigation."""
    id: str
    title: str
    status: str
    experiments: int
    open_questions: List[str] = field(default_factory=list)
    last_updated: Optional[str] = None


@dataclass
class ExperimentSummary:
    """Summary of an experiment."""
    id: str
    investigation_id: str
    status: str
    result: Optional[str] = None
    has_runs: bool = False
    has_evidence: bool = False


@dataclass
class StatusSurface:
    """Complete status surface."""
    current_engine: str
    current_seed: str
    investigations: List[InvestigationSummary]
    recent_experiments: List[ExperimentSummary]
    open_questions: List[str]
    pending_promotions: List[str]
    recent_lessons: List[str]


class StatusSurfaceGenerator:
    """
    Generates a live status surface for KDE Laboratory.
    
    Usage:
        generator = StatusSurfaceGenerator()
        surface = generator.generate()
        print(generator.format_markdown(surface))
    """
    
    def __init__(self, kde_root: str = "/workspace/project/dnp3"):
        self.kde_root = kde_root
    
    def generate(self) -> StatusSurface:
        """Generate complete status surface."""
        return StatusSurface(
            current_engine=self._get_current_engine(),
            current_seed=self._get_current_seed(),
            investigations=self._get_investigations(),
            recent_experiments=self._get_recent_experiments(),
            open_questions=self._get_open_questions(),
            pending_promotions=self._get_pending_promotions(),
            recent_lessons=self._get_recent_lessons()
        )
    
    def _get_current_engine(self) -> str:
        """Get current active engine."""
        engine_path = os.path.join(self.kde_root, ".kde", "engines", "current.md")
        if os.path.exists(engine_path):
            with open(engine_path, 'r') as f:
                content = f.read()
                match = re.search(r'\*\*Engine\*\*:\s*(.+)', content)
                if match:
                    return match.group(1).strip()
        return "KDE-ENGINE-Beta (default)"
    
    def _get_current_seed(self) -> str:
        """Get current active seed."""
        seed_path = os.path.join(self.kde_root, ".kde", "seeds", "SEED-001", "seed.md")
        if os.path.exists(seed_path):
            return "SEED-001"
        return "SEED-001"
    
    def _get_investigations(self) -> List[InvestigationSummary]:
        """Get all investigations with summaries."""
        investigations = []
        inv_dir = os.path.join(self.kde_root, "laboratory", "investigations")
        
        if not os.path.exists(inv_dir):
            return investigations
        
        for item in os.listdir(inv_dir):
            if item.startswith("INV-"):
                inv_path = os.path.join(inv_dir, item)
                if os.path.isdir(inv_path):
                    summary = self._parse_investigation(item, inv_path)
                    if summary:
                        investigations.append(summary)
        
        # Sort by last updated
        investigations.sort(
            key=lambda x: x.last_updated or "", 
            reverse=True
        )
        return investigations[:10]  # Top 10
    
    def _parse_investigation(self, inv_id: str, inv_path: str) -> Optional[InvestigationSummary]:
        """Parse investigation summary from directory."""
        # Try to read title and status from investigation.md
        inv_file = os.path.join(inv_path, "investigation.md")
        title = inv_id
        status = "UNKNOWN"
        
        if os.path.exists(inv_file):
            with open(inv_file, 'r') as f:
                content = f.read()
                match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
                if match:
                    title = match.group(1)
                match = re.search(r'\*\*Status\*\*:\s*(.+)', content)
                if match:
                    status = match.group(1).strip()
        
        # Count experiments
        exp_count = 0
        exp_dir = os.path.join(inv_path, "experiments")
        if os.path.exists(exp_dir):
            exp_count = len([f for f in os.listdir(exp_dir) if f.startswith("LAB-")])
        
        # Get last updated
        mtime = os.path.getmtime(inv_path)
        last_updated = datetime.fromtimestamp(mtime).strftime("%Y-%m-%d")
        
        return InvestigationSummary(
            id=inv_id,
            title=title,
            status=status,
            experiments=exp_count,
            last_updated=last_updated
        )
    
    def _get_recent_experiments(self) -> List[ExperimentSummary]:
        """Get recent experiments."""
        experiments = []
        exp_dir = os.path.join(self.kde_root, "laboratory", "experiments")
        
        if not os.path.exists(exp_dir):
            return experiments
        
        for item in os.listdir(exp_dir):
            if item.startswith("LAB-"):
                exp_path = os.path.join(exp_dir, item)
                if os.path.isdir(exp_path):
                    summary = self._parse_experiment(item, exp_path)
                    if summary:
                        experiments.append(summary)
        
        experiments.sort(
            key=lambda x: x.id, 
            reverse=True
        )
        return experiments[:10]
    
    def _parse_experiment(self, exp_id: str, exp_path: str) -> Optional[ExperimentSummary]:
        """Parse experiment summary."""
        has_runs = os.path.exists(os.path.join(exp_path, "runs"))
        has_evidence = os.path.exists(os.path.join(exp_path, "evidence"))
        
        # Get investigation parent
        inv_id = "UNKNOWN"
        parent = os.path.dirname(exp_path)
        if "investigations" in parent:
            match = re.search(r'(INV-\w+)', parent)
            if match:
                inv_id = match.group(1)
        
        return ExperimentSummary(
            id=exp_id,
            investigation_id=inv_id,
            status="COMPLETE" if has_runs else "IN_PROGRESS",
            has_runs=has_runs,
            has_evidence=has_evidence
        )
    
    def _get_open_questions(self) -> List[str]:
        """Get open questions from knowledge base."""
        # This would scan knowledge/ for unresolved questions
        return [
            "What is Knowledge? (Tier 1 - PENDING)",
            "What is Evidence? (Tier 1 - PENDING)",
            "What is Context? (Tier 1 - PENDING)",
            "How does validation work? (Tier 1 - PENDING)",
        ]
    
    def _get_pending_promotions(self) -> List[str]:
        """Get knowledge items pending promotion."""
        return [
            "KDE-KNOW-042: Investigation Patterns",
            "KDE-KNOW-043: Experiment Templates",
        ]
    
    def _get_recent_lessons(self) -> List[str]:
        """Get recent lessons learned."""
        return [
            "INV-GOV-SYN: Governance gap identified",
            "INV-066: Human intervention pattern documented",
        ]
    
    def format_markdown(self, surface: StatusSurface) -> str:
        """Format status surface as Markdown."""
        lines = [
            "# KDE Laboratory Status",
            "",
            f"**Generated**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            "",
            "## Current Configuration",
            "",
            f"| Component | Value |",
            f"|-----------|-------|",
            f"| Engine | {surface.current_engine} |",
            f"| Seed | {surface.current_seed} |",
            "",
            "## Active Investigations",
            "",
            f"| ID | Title | Status | Experiments | Updated |",
            f"|----|-------|--------|------------|---------|",
        ]
        
        for inv in surface.investigations:
            lines.append(
                f"| {inv.id} | {inv.title[:40]} | {inv.status} | {inv.experiments} | {inv.last_updated or 'N/A'} |"
            )
        
        lines.extend([
            "",
            "## Recent Experiments",
            "",
            f"| ID | Investigation | Status | Runs | Evidence |",
            f"|----|--------------|--------|------|----------|",
        ])
        
        for exp in surface.recent_experiments:
            runs = "✅" if exp.has_runs else "❌"
            evidence = "✅" if exp.has_evidence else "❌"
            lines.append(
                f"| {exp.id} | {exp.investigation_id} | {exp.status} | {runs} | {evidence} |"
            )
        
        lines.extend([
            "",
            "## Open Questions",
            "",
        ])
        
        for q in surface.open_questions[:5]:
            lines.append(f"- {q}")
        
        lines.extend([
            "",
            "## Pending Promotions",
            "",
        ])
        
        for item in surface.pending_promotions:
            lines.append(f"- {item}")
        
        lines.extend([
            "",
            "## Recent Lessons",
            "",
        ])
        
        for lesson in surface.recent_lessons:
            lines.append(f"- {lesson}")
        
        return "\n".join(lines)


# CLI entry point
if __name__ == "__main__":
    import sys
    
    kde_root = sys.argv[1] if len(sys.argv) > 1 else "/workspace/project/dnp3"
    generator = StatusSurfaceGenerator(kde_root)
    surface = generator.generate()
    print(generator.format_markdown(surface))
