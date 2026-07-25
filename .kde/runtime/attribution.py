"""
Decision Attribution

Implements Phase 5 of Knowledge-on-Demand Runtime.

This module records the origin of major investigation decisions.

Possible origins:
- Previous Knowledge
- Prompt
- Engine Reasoning
- Seed
- New Synthesis
"""

import json
from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional
from dataclasses import dataclass, asdict
from pathlib import Path


class DecisionOrigin(Enum):
    """Possible origins for investigation decisions."""
    PREVIOUS_KNOWLEDGE = "previous_knowledge"
    PROMPT = "prompt"
    ENGINE_REASONING = "engine_reasoning"
    SEED = "seed"
    NEW_SYNTHESIS = "new_synthesis"
    UNKNOWN = "unknown"


@dataclass
class Decision:
    """A recorded investigation decision."""
    timestamp: str
    investigation_id: str
    decision_type: str
    description: str
    origin: str
    evidence: List[str]  # Evidence supporting the attribution
    knowledge_ids: List[str]  # Knowledge that influenced the decision
    confidence: float  # 0.0 to 1.0


class DecisionAttributor:
    """
    Records decision origins for investigations.
    
    Every attribution SHALL be supported by observable runtime evidence.
    
    Evidence types:
    - Runtime log reference
    - Knowledge retrieval record
    - Prompt text
    - Tool call
    - File access
    """
    
    def __init__(self, log_dir: str = None):
        """Initialize the attributor."""
        if log_dir is None:
            log_dir = os.path.join(
                os.path.dirname(__file__),
                "logs",
                "attribution"
            )
        
        self.log_dir = Path(log_dir)
        self.log_dir.mkdir(parents=True, exist_ok=True)
        
        self.current_investigation: Optional[str] = None
        self.decisions: List[Decision] = []
    
    def set_investigation(self, investigation_id: str):
        """Set the current investigation."""
        self.current_investigation = investigation_id
        self.decisions = []
    
    def record_decision(
        self,
        decision_type: str,
        description: str,
        origin: DecisionOrigin,
        evidence: List[str],
        knowledge_ids: List[str] = None,
        confidence: float = 0.5
    ) -> Decision:
        """
        Record an investigation decision with attribution.
        
        Args:
            decision_type: Type of decision (e.g., "architecture", "implementation")
            description: Description of the decision
            origin: The decision's origin
            evidence: Evidence supporting the attribution
            knowledge_ids: Knowledge IDs that influenced the decision
            confidence: Confidence in the attribution (0.0-1.0)
        
        Returns:
            Decision record
        """
        decision = Decision(
            timestamp=datetime.utcnow().isoformat() + "Z",
            investigation_id=self.current_investigation or "UNKNOWN",
            decision_type=decision_type,
            description=description,
            origin=origin.value,
            evidence=evidence,
            knowledge_ids=knowledge_ids or [],
            confidence=confidence
        )
        
        self.decisions.append(decision)
        self._write_decision(decision)
        return decision
    
    def _write_decision(self, decision: Decision):
        """Write decision to log file."""
        log_file = self.log_dir / f"{self.current_investigation}.jsonl"
        
        with open(log_file, 'a') as f:
            f.write(json.dumps(asdict(decision)) + '\n')
    
    def get_attribution_summary(self) -> Dict:
        """Get a summary of decision attributions."""
        if not self.current_investigation:
            return {"error": "No investigation set"}
        
        log_file = self.log_dir / f"{self.current_investigation}.jsonl"
        
        if not log_file.exists():
            return {"error": "No decisions recorded"}
        
        decisions = []
        with open(log_file, 'r') as f:
            for line in f:
                decisions.append(json.loads(line))
        
        # Count by origin
        origin_counts = {}
        for d in decisions:
            origin = d['origin']
            origin_counts[origin] = origin_counts.get(origin, 0) + 1
        
        # Calculate percentages
        total = len(decisions)
        origin_percentages = {
            k: (v / total) * 100 
            for k, v in origin_counts.items()
        }
        
        # Calculate knowledge influence
        decisions_with_knowledge = sum(
            1 for d in decisions if d['knowledge_ids']
        )
        
        return {
            "investigation_id": self.current_investigation,
            "total_decisions": total,
            "decisions_with_knowledge": decisions_with_knowledge,
            "knowledge_influence_rate": (
                decisions_with_knowledge / total if total > 0 else 0
            ),
            "origin_counts": origin_counts,
            "origin_percentages": origin_percentages
        }
    
    def generate_report(self) -> str:
        """Generate a human-readable attribution report."""
        summary = self.get_attribution_summary()
        
        if "error" in summary:
            return f"Error: {summary['error']}"
        
        report = []
        report.append("=" * 60)
        report.append("DECISION ATTRIBUTION REPORT")
        report.append("=" * 60)
        report.append("")
        report.append(f"Investigation: {summary['investigation_id']}")
        report.append(f"Report Generated: {datetime.utcnow().isoformat()}Z")
        report.append("")
        report.append("-" * 60)
        report.append("ORIGIN DISTRIBUTION")
        report.append("-" * 60)
        
        for origin, count in summary['origin_counts'].items():
            pct = summary['origin_percentages'][origin]
            report.append(f"{origin:25} {count:3} ({pct:5.1f}%)")
        
        report.append("")
        report.append("-" * 60)
        report.append("KNOWLEDGE INFLUENCE")
        report.append("-" * 60)
        report.append(
            f"Decisions with Knowledge: "
            f"{summary['decisions_with_knowledge']} / {summary['total_decisions']}"
        )
        report.append(
            f"Knowledge Influence Rate: {summary['knowledge_influence_rate']:.1%}"
        )
        report.append("")
        report.append("=" * 60)
        
        return '\n'.join(report)


import os
