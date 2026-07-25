"""
SOP-005 Knowledge Retrieval Policy Execution

Implements Phase 3 of Knowledge-on-Demand Runtime.

This module executes the Laboratory Knowledge Retrieval Policy
as defined in Laboratory SOP §SOP-005.

Retrieval Decision Matrix:
- Continuation investigation: REQUIRED
- Similar historical work: RECOMMENDED
- Novel research: OPTIONAL
- Routine engineering: MINIMAL
- Complex engineering: FULL
"""

from enum import Enum
from typing import Optional, Dict, List
from dataclasses import dataclass
from datetime import datetime


class RetrievalContext(Enum):
    """Retrieval context per SOP-005."""
    CONTINUATION = "REQUIRED"      # Direct follow-up
    SIMILAR = "RECOMMENDED"        # Related historical work
    NOVEL = "OPTIONAL"             # New research territory
    ROUTINE = "MINIMAL"            # Standard engineering
    COMPLEX = "FULL"               # High-complexity engineering


class RetrievalDecision:
    """Result of SOP-005 evaluation."""
    
    def __init__(
        self,
        context: RetrievalContext,
        required: bool,
        retrieval_level: str,
        rationale: str,
        domains: List[str] = None,
        keywords: List[str] = None
    ):
        self.context = context
        self.required = required
        self.retrieval_level = retrieval_level
        self.rationale = rationale
        self.domains = domains or []
        self.keywords = keywords or []
        self.timestamp = datetime.utcnow().isoformat() + "Z"
    
    def to_dict(self) -> Dict:
        return {
            "context": self.context.value,
            "required": self.required,
            "retrieval_level": self.retrieval_level,
            "rationale": self.rationale,
            "domains": self.domains,
            "keywords": self.keywords,
            "timestamp": self.timestamp
        }


class SOP005Executor:
    """
    Executes SOP-005 Knowledge Retrieval Policy.
    
    The Runtime SHALL determine retrieval requirements using SOP-005.
    The Engine SHALL receive only the resulting context.
    """
    
    # Context detection keywords
    CONTINUATION_KEYWORDS = [
        "continue", "follow-up", "completion", "extend", 
        "build upon", "next phase", "subsequent"
    ]
    
    SIMILAR_KEYWORDS = [
        "similar", "like", "comparison", "replication",
        "previous", "earlier", "related"
    ]
    
    NOVEL_KEYWORDS = [
        "new", "novel", "first", "pioneer", "explore",
        "discover", "unprecedented"
    ]
    
    COMPLEX_KEYWORDS = [
        "complex", "difficult", "challenging", "advanced",
        "enterprise", "production", "scalable"
    ]
    
    def __init__(self):
        """Initialize the SOP-005 executor."""
        self.last_decision: Optional[RetrievalDecision] = None
    
    def evaluate(
        self,
        investigation_id: str,
        title: str,
        description: str = "",
        parent_id: str = None,
        keywords: List[str] = None
    ) -> RetrievalDecision:
        """
        Evaluate investigation against SOP-005 retrieval matrix.
        
        Args:
            investigation_id: The investigation ID
            title: Investigation title
            description: Investigation description
            parent_id: Parent investigation ID (if continuation)
            keywords: Explicitly provided keywords
        
        Returns:
            RetrievalDecision with context and retrieval level
        """
        text = f"{title} {description}".lower()
        keywords = keywords or []
        
        # Determine context based on detection
        context = self._determine_context(
            text, 
            parent_id, 
            keywords
        )
        
        # Determine retrieval details based on context
        decision = self._make_decision(
            investigation_id,
            context,
            text,
            keywords
        )
        
        self.last_decision = decision
        return decision
    
    def _determine_context(
        self,
        text: str,
        parent_id: str,
        keywords: List[str]
    ) -> RetrievalContext:
        """Determine the retrieval context based on investigation characteristics."""
        
        # Check for continuation (highest priority)
        for keyword in self.CONTINUATION_KEYWORDS:
            if keyword in text:
                return RetrievalContext.CONTINUATION
        
        # Check if there's a parent (implicit continuation)
        if parent_id:
            return RetrievalContext.CONTINUATION
        
        # Check for similar work
        for keyword in self.SIMILAR_KEYWORDS:
            if keyword in text:
                return RetrievalContext.SIMILAR
        
        # Check for novel research
        for keyword in self.NOVEL_KEYWORDS:
            if keyword in text:
                return RetrievalContext.NOVEL
        
        # Check for complex engineering
        for keyword in self.COMPLEX_KEYWORDS:
            if keyword in text:
                return RetrievalContext.COMPLEX
        
        # Default to routine
        return RetrievalContext.ROUTINE
    
    def _make_decision(
        self,
        investigation_id: str,
        context: RetrievalContext,
        text: str,
        keywords: List[str]
    ) -> RetrievalDecision:
        """Make the retrieval decision based on context."""
        
        if context == RetrievalContext.CONTINUATION:
            return RetrievalDecision(
                context=context,
                required=True,
                retrieval_level="FULL",
                rationale=f"SOP-005 §5.2: Continuation investigation '{investigation_id}' requires full knowledge retrieval.",
                domains=self._extract_domains(text),
                keywords=keywords or ["architecture", "methodology", "investigation"]
            )
        
        elif context == RetrievalContext.SIMILAR:
            return RetrievalDecision(
                context=context,
                required=False,
                retrieval_level="PARTIAL",
                rationale=f"SOP-005 §5.3: Similar historical work detected; recommended retrieval.",
                domains=self._extract_domains(text),
                keywords=keywords or self._extract_keywords(text)
            )
        
        elif context == RetrievalContext.NOVEL:
            return RetrievalDecision(
                context=context,
                required=False,
                retrieval_level="MINIMAL",
                rationale=f"SOP-005 §5.4: Novel research; optional minimal retrieval.",
                domains=[],
                keywords=keywords or []
            )
        
        elif context == RetrievalContext.COMPLEX:
            return RetrievalDecision(
                context=context,
                required=False,
                retrieval_level="FULL",
                rationale=f"SOP-005 §5.5: Complex engineering; full retrieval recommended.",
                domains=self._extract_domains(text),
                keywords=keywords or ["architecture", "patterns", "tradeoffs"]
            )
        
        else:  # ROUTINE
            return RetrievalDecision(
                context=context,
                required=False,
                retrieval_level="MINIMAL",
                rationale=f"SOP-005 §5.6: Routine engineering; minimal retrieval.",
                domains=[],
                keywords=keywords or []
            )
    
    def _extract_domains(self, text: str) -> List[str]:
        """Extract domain keywords from text."""
        domains = []
        domain_keywords = {
            "architecture": ["architecture", "design", "system"],
            "scada": ["scada", "utility", "distribution", "grid"],
            "frontend": ["frontend", "ui", "design", "visualization"],
            "backend": ["backend", "api", "microservices"],
            "governance": ["governance", "policy", "procedure"],
            "methodology": ["methodology", "process", "workflow"]
        }
        
        for domain, keywords in domain_keywords.items():
            if any(kw in text for kw in keywords):
                domains.append(domain)
        
        return domains if domains else []
    
    def _extract_keywords(self, text: str) -> List[str]:
        """Extract relevant keywords from text."""
        # Simple extraction - in production would use NLP
        words = text.split()
        return [w for w in words if len(w) > 4][:10]
    
    def get_retrieval_level(self, decision: RetrievalDecision) -> int:
        """Get numeric retrieval level for logging."""
        levels = {
            "FULL": 4,
            "PARTIAL": 3,
            "OPTIONAL": 2,
            "MINIMAL": 1,
            "NONE": 0
        }
        return levels.get(decision.retrieval_level, 0)


def test_sop005():
    """Test the SOP-005 executor."""
    executor = SOP005Executor()
    
    print("=== SOP-005 Execution Test ===\n")
    
    # Test continuation
    print("1. Continuation Investigation:")
    decision = executor.evaluate(
        "INV-020",
        "Continuation of INV-013 SCADA Architecture",
        parent_id="INV-013"
    )
    print(f"   Context: {decision.context.value}")
    print(f"   Required: {decision.required}")
    print(f"   Level: {decision.retrieval_level}")
    print(f"   Rationale: {decision.rationale}")
    
    print("\n2. Similar Historical Work:")
    decision = executor.evaluate(
        "INV-021",
        "Replication Study Similar to INV-016",
        keywords=["architecture", "replication"]
    )
    print(f"   Context: {decision.context.value}")
    print(f"   Required: {decision.required}")
    print(f"   Level: {decision.retrieval_level}")
    
    print("\n3. Novel Research:")
    decision = executor.evaluate(
        "INV-022",
        "Novel Approach to Knowledge Synthesis",
        keywords=["knowledge", "synthesis"]
    )
    print(f"   Context: {decision.context.value}")
    print(f"   Required: {decision.required}")
    print(f"   Level: {decision.retrieval_level}")
    
    print("\n4. Complex Engineering:")
    decision = executor.evaluate(
        "INV-023",
        "Enterprise SCADA Platform Architecture",
        keywords=["enterprise", "scalable"]
    )
    print(f"   Context: {decision.context.value}")
    print(f"   Required: {decision.required}")
    print(f"   Level: {decision.retrieval_level}")
    
    print("\n5. Routine Engineering:")
    decision = executor.evaluate(
        "INV-024",
        "Update Documentation for API",
        keywords=["documentation", "api"]
    )
    print(f"   Context: {decision.context.value}")
    print(f"   Required: {decision.required}")
    print(f"   Level: {decision.retrieval_level}")


if __name__ == "__main__":
    test_sop005()
