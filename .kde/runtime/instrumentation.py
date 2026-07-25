"""
Runtime Instrumentation

Implements Phase 4 of Knowledge-on-Demand Runtime.

This module provides runtime logging and metrics for:
- Retrieval events
- Context construction
- Decision attribution
- Knowledge utilization statistics
"""

import json
import os
from datetime import datetime
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict
from pathlib import Path


@dataclass
class RetrievalEvent:
    """A logged retrieval event."""
    timestamp: str
    investigation_id: str
    event_type: str  # "retrieval", "context", "decision"
    trigger: str     # What triggered the event
    knowledge_retrieved: List[str]  # Knowledge IDs
    context_size: int  # Characters in context
    duration_ms: float  # Operation duration
    metadata: Dict[str, Any]


class Instrumentation:
    """
    Runtime instrumentation for Knowledge-on-Demand.
    
    Records:
    - Timestamp
    - Investigation
    - Trigger
    - Retrieval reason
    - Knowledge retrieved
    - Context size
    - Retrieval duration
    """
    
    def __init__(self, log_dir: str = None):
        """Initialize instrumentation with log directory."""
        if log_dir is None:
            log_dir = os.path.join(
                os.path.dirname(__file__),
                "logs"
            )
        
        self.log_dir = Path(log_dir)
        self.log_dir.mkdir(exist_ok=True)
        
        self.current_investigation: Optional[str] = None
        self.events: List[RetrievalEvent] = []
    
    def set_investigation(self, investigation_id: str):
        """Set the current investigation for logging."""
        self.current_investigation = investigation_id
    
    def log_retrieval(
        self,
        trigger: str,
        knowledge_ids: List[str],
        context_size: int,
        duration_ms: float,
        metadata: Dict[str, Any] = None
    ) -> RetrievalEvent:
        """Log a knowledge retrieval event."""
        event = RetrievalEvent(
            timestamp=datetime.utcnow().isoformat() + "Z",
            investigation_id=self.current_investigation or "UNKNOWN",
            event_type="retrieval",
            trigger=trigger,
            knowledge_retrieved=knowledge_ids,
            context_size=context_size,
            duration_ms=duration_ms,
            metadata=metadata or {}
        )
        
        self.events.append(event)
        self._write_event(event)
        return event
    
    def log_context_construction(
        self,
        knowledge_ids: List[str],
        context_size: int,
        duration_ms: float
    ) -> RetrievalEvent:
        """Log a context construction event."""
        event = RetrievalEvent(
            timestamp=datetime.utcnow().isoformat() + "Z",
            investigation_id=self.current_investigation or "UNKNOWN",
            event_type="context",
            trigger="context_construction",
            knowledge_retrieved=knowledge_ids,
            context_size=context_size,
            duration_ms=duration_ms,
            metadata={}
        )
        
        self.events.append(event)
        self._write_event(event)
        return event
    
    def log_sop_decision(
        self,
        decision_context: str,
        retrieval_level: str,
        rationale: str
    ) -> RetrievalEvent:
        """Log an SOP-005 decision."""
        event = RetrievalEvent(
            timestamp=datetime.utcnow().isoformat() + "Z",
            investigation_id=self.current_investigation or "UNKNOWN",
            event_type="sop_decision",
            trigger=f"sop005:{decision_context}",
            knowledge_retrieved=[],  # Decision only, no retrieval yet
            context_size=0,
            duration_ms=0,
            metadata={
                "decision_context": decision_context,
                "retrieval_level": retrieval_level,
                "rationale": rationale
            }
        )
        
        self.events.append(event)
        self._write_event(event)
        return event
    
    def _write_event(self, event: RetrievalEvent):
        """Write event to log file."""
        log_file = self.log_dir / f"{self.current_investigation or 'runtime'}.jsonl"
        
        with open(log_file, 'a') as f:
            f.write(json.dumps(asdict(event)) + '\n')
    
    def get_metrics(self) -> Dict[str, Any]:
        """Generate metrics for the current investigation."""
        if not self.current_investigation:
            return {"error": "No investigation set"}
        
        log_file = self.log_dir / f"{self.current_investigation}.jsonl"
        
        if not log_file.exists():
            return {"error": "No log file found"}
        
        events = []
        with open(log_file, 'r') as f:
            for line in f:
                events.append(json.loads(line))
        
        # Calculate metrics
        retrieval_events = [e for e in events if e['event_type'] == 'retrieval']
        sop_events = [e for e in events if e['event_type'] == 'sop_decision']
        
        all_knowledge = set()
        for e in retrieval_events:
            all_knowledge.update(e['knowledge_retrieved'])
        
        total_context_size = sum(e['context_size'] for e in retrieval_events)
        total_duration = sum(e['duration_ms'] for e in retrieval_events)
        
        return {
            "investigation_id": self.current_investigation,
            "total_events": len(events),
            "retrieval_events": len(retrieval_events),
            "sop_decisions": len(sop_events),
            "unique_knowledge_retrieved": len(all_knowledge),
            "knowledge_ids": sorted(list(all_knowledge)),
            "total_context_size": total_context_size,
            "average_context_size": (
                total_context_size / len(retrieval_events) 
                if retrieval_events else 0
            ),
            "total_duration_ms": total_duration,
            "average_duration_ms": (
                total_duration / len(retrieval_events)
                if retrieval_events else 0
            ),
            "retrieval_rate": len(all_knowledge) / 13 if events else 0  # 13 = total artifacts
        }
    
    def generate_report(self) -> str:
        """Generate a human-readable metrics report."""
        metrics = self.get_metrics()
        
        if "error" in metrics:
            return f"Error: {metrics['error']}"
        
        report = []
        report.append("=" * 60)
        report.append("KNOWLEDGE-ON-DEMAND METRICS REPORT")
        report.append("=" * 60)
        report.append("")
        report.append(f"Investigation: {metrics['investigation_id']}")
        report.append(f"Report Generated: {datetime.utcnow().isoformat()}Z")
        report.append("")
        report.append("-" * 60)
        report.append("RETRIEVAL STATISTICS")
        report.append("-" * 60)
        report.append(f"Total Events:        {metrics['total_events']}")
        report.append(f"Retrieval Events:     {metrics['retrieval_events']}")
        report.append(f"SOP Decisions:        {metrics['sop_decisions']}")
        report.append("")
        report.append("-" * 60)
        report.append("KNOWLEDGE UTILIZATION")
        report.append("-" * 60)
        report.append(f"Unique Knowledge IDs: {metrics['unique_knowledge_retrieved']}")
        report.append(f"Retrieval Rate:       {metrics['retrieval_rate']:.1%}")
        report.append("")
        report.append("Knowledge Retrieved:")
        for kid in metrics['knowledge_ids']:
            report.append(f"  - {kid}")
        report.append("")
        report.append("-" * 60)
        report.append("PERFORMANCE")
        report.append("-" * 60)
        report.append(f"Total Context Size:  {metrics['total_context_size']:,} chars")
        report.append(f"Avg Context Size:     {metrics['average_context_size']:.0f} chars")
        report.append(f"Total Duration:       {metrics['total_duration_ms']:.2f} ms")
        report.append(f"Avg Duration:         {metrics['average_duration_ms']:.2f} ms")
        report.append("")
        report.append("=" * 60)
        
        return '\n'.join(report)
    
    def export_logs(self, output_file: str = None) -> str:
        """Export all logs to a JSON file."""
        if output_file is None:
            output_file = str(
                self.log_dir / f"{self.current_investigation}_export.json"
            )
        
        log_file = self.log_dir / f"{self.current_investigation}.jsonl"
        
        events = []
        if log_file.exists():
            with open(log_file, 'r') as f:
                for line in f:
                    events.append(json.loads(line))
        
        with open(output_file, 'w') as f:
            json.dump({
                "investigation_id": self.current_investigation,
                "exported_at": datetime.utcnow().isoformat() + "Z",
                "events": events,
                "metrics": self.get_metrics()
            }, f, indent=2)
        
        return output_file


def test_instrumentation():
    """Test the instrumentation."""
    instr = Instrumentation()
    instr.set_investigation("INV-TEST")
    
    print("=== Instrumentation Test ===\n")
    
    # Log some events
    instr.log_sop_decision(
        decision_context="CONTINUATION",
        retrieval_level="FULL",
        rationale="SOP-005 §5.2: Continuation investigation"
    )
    
    instr.log_retrieval(
        trigger="domain:architecture",
        knowledge_ids=["KDE-ARCH-001", "KDE-ARCH-002", "KDE-ARCH-003"],
        context_size=4500,
        duration_ms=15.2,
        metadata={"retrieval_level": "FULL"}
    )
    
    instr.log_retrieval(
        trigger="keyword:investigation",
        knowledge_ids=["KDE-ARCH-004", "KDE-ARCH-005"],
        context_size=3200,
        duration_ms=12.1,
        metadata={"retrieval_level": "PARTIAL"}
    )
    
    # Generate report
    print(instr.generate_report())
    
    # Export
    output = instr.export_logs()
    print(f"\nExported to: {output}")


if __name__ == "__main__":
    test_instrumentation()
