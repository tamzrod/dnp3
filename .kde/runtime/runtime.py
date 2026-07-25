"""
Knowledge-on-Demand Runtime

Main integration module for the Knowledge-on-Demand Runtime.

Integrates:
- Knowledge Catalog
- Retrieval Engine
- SOP-005 Execution
- Runtime Instrumentation
- Decision Attribution
"""

import os
import time
import json
from typing import Dict, List, Optional, Any
from dataclasses import dataclass

from .retrieval import RetrievalEngine, RetrievalResult
from .sop005 import SOP005Executor, RetrievalDecision, RetrievalContext
from .instrumentation import Instrumentation
from .attribution import DecisionAttributor, DecisionOrigin


@dataclass
class InvestigationContext:
    """Context constructed for an investigation."""
    investigation_id: str
    retrieval_decision: Dict
    knowledge_retrieved: List[RetrievalResult]
    context_documents: List[Dict]
    context_size: int
    sop_compliance: bool


class KnowledgeOnDemandRuntime:
    """
    Main Knowledge-on-Demand Runtime.
    
    This runtime implements the validated Knowledge-on-Demand architecture
    for the KDE Laboratory.
    
    Responsibilities:
    - Execute SOP-005 retrieval policy
    - Retrieve relevant knowledge
    - Construct investigation context
    - Instrument all retrieval events
    - Support decision attribution
    """
    
    def __init__(self, investigation_id: str):
        """
        Initialize the Knowledge-on-Demand Runtime.
        
        Args:
            investigation_id: The investigation ID for this session
        """
        self.investigation_id = investigation_id
        
        # Initialize components
        self.retrieval_engine = RetrievalEngine()
        self.sop_executor = SOP005Executor()
        self.instrumentation = Instrumentation()
        self.attributor = DecisionAttributor()
        
        # Set investigation context
        self.instrumentation.set_investigation(investigation_id)
        self.attributor.set_investigation(investigation_id)
        
        # Current context
        self.current_context: Optional[InvestigationContext] = None
    
    def initialize(
        self,
        title: str,
        description: str = "",
        parent_id: str = None,
        keywords: List[str] = None
    ) -> InvestigationContext:
        """
        Initialize the runtime for an investigation.
        
        Args:
            title: Investigation title
            description: Investigation description
            parent_id: Parent investigation ID (if continuation)
            keywords: Investigation keywords
        
        Returns:
            InvestigationContext with retrieved knowledge
        """
        start_time = time.time()
        
        # Step 1: Evaluate SOP-005
        retrieval_decision = self.sop_executor.evaluate(
            investigation_id=self.investigation_id,
            title=title,
            description=description,
            parent_id=parent_id,
            keywords=keywords
        )
        
        # Log SOP decision
        self.instrumentation.log_sop_decision(
            decision_context=retrieval_decision.context.value,
            retrieval_level=retrieval_decision.retrieval_level,
            rationale=retrieval_decision.rationale
        )
        
        # Step 2: Retrieve knowledge based on SOP decision
        knowledge_retrieved = self._retrieve_knowledge(retrieval_decision)
        
        # Step 3: Construct context documents
        context_documents = self._construct_context(knowledge_retrieved)
        
        # Calculate context size
        context_size = sum(len(json.dumps(doc)) for doc in context_documents)
        
        # Log retrieval event
        retrieval_time = (time.time() - start_time) * 1000
        self.instrumentation.log_retrieval(
            trigger=f"sop005:{retrieval_decision.context.value}",
            knowledge_ids=[r.artifact_id for r in knowledge_retrieved],
            context_size=context_size,
            duration_ms=retrieval_time,
            metadata={
                "retrieval_level": retrieval_decision.retrieval_level,
                "sop_compliance": True
            }
        )
        
        # Step 4: Create context
        self.current_context = InvestigationContext(
            investigation_id=self.investigation_id,
            retrieval_decision=retrieval_decision.to_dict(),
            knowledge_retrieved=knowledge_retrieved,
            context_documents=context_documents,
            context_size=context_size,
            sop_compliance=True
        )
        
        return self.current_context
    
    def _retrieve_knowledge(
        self, 
        decision: RetrievalDecision
    ) -> List[RetrievalResult]:
        """Retrieve knowledge based on the SOP decision."""
        
        if decision.retrieval_level == "NONE":
            return []
        
        all_results = []
        
        # Retrieve by domains
        for domain in decision.domains:
            results = self.retrieval_engine.retrieve_by_domain(domain)
            all_results.extend(results)
        
        # Retrieve by keywords
        if decision.keywords:
            results = self.retrieval_engine.retrieve_by_keywords(
                decision.keywords,
                min_score=0.2
            )
            all_results.extend(results)
        
        # For FULL retrieval, also include all high-relevance results
        if decision.retrieval_level == "FULL":
            all_results.extend(self.retrieval_engine.retrieve_all())
        
        # Deduplicate by artifact_id
        seen = set()
        unique_results = []
        for r in all_results:
            if r.artifact_id not in seen:
                seen.add(r.artifact_id)
                unique_results.append(r)
        
        # Sort by relevance
        unique_results.sort(key=lambda x: x.relevance_score, reverse=True)
        
        return unique_results
    
    def _construct_context(
        self,
        knowledge: List[RetrievalResult]
    ) -> List[Dict]:
        """Construct context documents from retrieved knowledge."""
        
        documents = []
        
        for item in knowledge:
            artifact = self.retrieval_engine.get_artifact(item.artifact_id)
            
            if artifact:
                documents.append({
                    "type": "knowledge",
                    "id": artifact['id'],
                    "title": artifact['title'],
                    "domain": artifact['domain'],
                    "source": artifact['source'],
                    "status": artifact['status'],
                    "summary": artifact['summary'],
                    "keywords": artifact.get('keywords', []),
                    "relationships": artifact.get('relationships', []),
                    "relevance": {
                        "score": item.relevance_score,
                        "reason": item.match_reason
                    }
                })
        
        return documents
    
    def record_decision(
        self,
        decision_type: str,
        description: str,
        origin: DecisionOrigin,
        evidence: List[str],
        confidence: float = 0.5
    ):
        """
        Record an investigation decision with attribution.
        
        Args:
            decision_type: Type of decision
            description: Description
            origin: Decision origin
            evidence: Supporting evidence
            confidence: Attribution confidence
        """
        # Determine if knowledge influenced the decision
        knowledge_ids = []
        if origin == DecisionOrigin.PREVIOUS_KNOWLEDGE and self.current_context:
            knowledge_ids = [
                r.artifact_id for r in self.current_context.knowledge_retrieved
            ]
        
        self.attributor.record_decision(
            decision_type=decision_type,
            description=description,
            origin=origin,
            evidence=evidence,
            knowledge_ids=knowledge_ids,
            confidence=confidence
        )
    
    def get_context_document(self) -> str:
        """
        Get the investigation context as a formatted document.
        
        This document can be prepended to the investigation prompt.
        """
        if not self.current_context:
            return "# Investigation Context\n\nNo knowledge retrieved."
        
        lines = [
            "# Investigation Context",
            "",
            f"**Investigation**: {self.investigation_id}",
            f"**SOP-005 Context**: {self.current_context.retrieval_decision['context']}",
            f"**Retrieval Level**: {self.current_context.retrieval_decision['retrieval_level']}",
            "",
            "## Relevant Knowledge",
            ""
        ]
        
        for doc in self.current_context.context_documents:
            lines.append(f"### {doc['title']}")
            lines.append(f"**ID**: {doc['id']}")
            lines.append(f"**Domain**: {doc['domain']}")
            lines.append(f"**Status**: {doc['status']}")
            lines.append(f"**Relevance**: {doc['relevance']['score']:.2f} ({doc['relevance']['reason']})")
            lines.append("")
            lines.append(f"{doc['summary']}")
            lines.append("")
        
        lines.append("---")
        lines.append(f"Context size: {self.current_context.context_size:,} characters")
        lines.append(f"Knowledge retrieved: {len(self.current_context.knowledge_retrieved)} artifacts")
        
        return '\n'.join(lines)
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get metrics for this investigation."""
        instr_metrics = self.instrumentation.get_metrics()
        attr_summary = self.attributor.get_attribution_summary()
        
        return {
            "investigation_id": self.investigation_id,
            "retrieval_metrics": instr_metrics,
            "attribution_summary": attr_summary,
            "sop_compliance": self.current_context.sop_compliance if self.current_context else False
        }
    
    def generate_full_report(self) -> str:
        """Generate a complete investigation report."""
        metrics = self.get_metrics()
        
        report = []
        report.append("=" * 70)
        report.append("KNOWLEDGE-ON-DEMAND RUNTIME REPORT")
        report.append("=" * 70)
        report.append("")
        report.append(f"Investigation: {self.investigation_id}")
        report.append("")
        
        # Retrieval metrics
        if "error" not in metrics['retrieval_metrics']:
            rm = metrics['retrieval_metrics']
            report.append("-" * 70)
            report.append("RETRIEVAL METRICS")
            report.append("-" * 70)
            report.append(f"Retrieval Events:    {rm['retrieval_events']}")
            report.append(f"Knowledge Retrieved: {rm['unique_knowledge_retrieved']}")
            report.append(f"Retrieval Rate:      {rm['retrieval_rate']:.1%}")
            report.append("")
        
        # Attribution summary
        if "error" not in metrics['attribution_summary']:
            am = metrics['attribution_summary']
            report.append("-" * 70)
            report.append("DECISION ATTRIBUTION")
            report.append("-" * 70)
            report.append(f"Total Decisions:     {am['total_decisions']}")
            report.append(f"Knowledge Influence: {am['knowledge_influence_rate']:.1%}")
            report.append("")
            report.append("Origin Distribution:")
            for origin, pct in am['origin_percentages'].items():
                count = am['origin_counts'][origin]
                report.append(f"  {origin:25} {count:3} ({pct:5.1f}%)")
            report.append("")
        
        # SOP compliance
        report.append("-" * 70)
        report.append("SOP COMPLIANCE")
        report.append("-" * 70)
        report.append(f"SOP-005 Compliance: {'YES' if metrics['sop_compliance'] else 'NO'}")
        report.append("")
        
        report.append("=" * 70)
        
        return '\n'.join(report)


def demo():
    """Demonstrate the Knowledge-on-Demand Runtime."""
    print("=" * 70)
    print("KNOWLEDGE-ON-DEMAND RUNTIME DEMONSTRATION")
    print("=" * 70)
    print()
    
    # Create runtime for an investigation
    runtime = KnowledgeOnDemandRuntime("INV-DEMO")
    
    # Initialize with SOP-005 evaluation
    context = runtime.initialize(
        title="Continuation of SCADA Architecture Investigation",
        description="Building upon INV-013 findings to implement the architecture",
        parent_id="INV-013",
        keywords=["architecture", "scada", "microservices"]
    )
    
    print("STEP 1: Context Initialization")
    print("-" * 70)
    print(f"SOP-005 Context: {context.retrieval_decision['context']}")
    print(f"Retrieval Level:  {context.retrieval_decision['retrieval_level']}")
    print(f"Rationale:        {context.retrieval_decision['rationale']}")
    print()
    
    print("STEP 2: Knowledge Retrieved")
    print("-" * 70)
    for item in context.knowledge_retrieved[:5]:
        print(f"  [{item.artifact_id}] {item.title}")
        print(f"                    Relevance: {item.relevance_score:.2f}")
    print()
    
    print("STEP 3: Context Document")
    print("-" * 70)
    context_doc = runtime.get_context_document()
    print(context_doc[:500] + "...")
    print()
    
    # Record some decisions
    runtime.record_decision(
        decision_type="architecture",
        description="Use microservices architecture for backend",
        origin=DecisionOrigin.PREVIOUS_KNOWLEDGE,
        evidence=["Retrieved KDE-ARCH-009", "Pattern: API Gateway"],
        confidence=0.8
    )
    
    runtime.record_decision(
        decision_type="implementation",
        description="Follow Laboratory SOP procedures",
        origin=DecisionOrigin.ENGINE_REASONING,
        evidence=["SOP-001 Investigation Lifecycle"],
        confidence=0.9
    )
    
    print("STEP 4: Decision Attribution")
    print("-" * 70)
    print(runtime.attributor.generate_report())
    print()
    
    print("STEP 5: Full Metrics Report")
    print("-" * 70)
    print(runtime.generate_full_report())


if __name__ == "__main__":
    demo()
