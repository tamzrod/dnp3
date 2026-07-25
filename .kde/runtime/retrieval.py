"""
Knowledge Retrieval Engine

Implements Phase 2 of Knowledge-on-Demand Runtime.

Supports:
- Domain lookup
- Keyword lookup
- Relationship lookup
- Investigation history lookup
"""

import json
import os
from typing import List, Dict, Optional, Set
from dataclasses import dataclass
from datetime import datetime


@dataclass
class RetrievalResult:
    """Result of a knowledge retrieval operation."""
    artifact_id: str
    title: str
    domain: str
    relevance_score: float
    match_reason: str
    summary: str


class RetrievalEngine:
    """
    Knowledge retrieval engine supporting multiple lookup strategies.
    """
    
    def __init__(self, catalog_path: str = None):
        """Initialize the retrieval engine with a catalog."""
        if catalog_path is None:
            catalog_path = os.path.join(
                os.path.dirname(__file__), 
                "catalog.json"
            )
        
        with open(catalog_path, 'r') as f:
            self.catalog = json.load(f)
    
    def retrieve_by_domain(self, domain: str) -> List[RetrievalResult]:
        """Retrieve all knowledge artifacts for a given domain."""
        results = []
        for artifact in self.catalog.get('artifacts', []):
            if artifact.get('domain') == domain:
                results.append(RetrievalResult(
                    artifact_id=artifact['id'],
                    title=artifact['title'],
                    domain=artifact['domain'],
                    relevance_score=1.0,  # Domain match = high relevance
                    match_reason=f"domain:{domain}",
                    summary=artifact.get('summary', '')
                ))
        return results
    
    def retrieve_by_keywords(
        self, 
        keywords: List[str], 
        min_score: float = 0.3
    ) -> List[RetrievalResult]:
        """Retrieve knowledge artifacts matching keywords."""
        results = []
        
        for artifact in self.catalog.get('artifacts', []):
            artifact_keywords = set(k.lower() for k in artifact.get('keywords', []))
            query_keywords = set(k.lower() for k in keywords)
            
            # Calculate Jaccard similarity
            intersection = artifact_keywords & query_keywords
            union = artifact_keywords | query_keywords
            
            if union:
                score = len(intersection) / len(union)
                
                if score >= min_score:
                    results.append(RetrievalResult(
                        artifact_id=artifact['id'],
                        title=artifact['title'],
                        domain=artifact['domain'],
                        relevance_score=score,
                        match_reason=f"keywords:{list(intersection)}",
                        summary=artifact.get('summary', '')
                    ))
        
        # Sort by relevance score descending
        results.sort(key=lambda x: x.relevance_score, reverse=True)
        return results
    
    def retrieve_by_investigation(self, investigation_id: str) -> List[RetrievalResult]:
        """Retrieve all knowledge artifacts from a specific investigation."""
        results = []
        
        for artifact in self.catalog.get('artifacts', []):
            if artifact.get('source') == investigation_id:
                results.append(RetrievalResult(
                    artifact_id=artifact['id'],
                    title=artifact['title'],
                    domain=artifact['domain'],
                    relevance_score=1.0,
                    match_reason=f"source:{investigation_id}",
                    summary=artifact.get('summary', '')
                ))
        
        return results
    
    def retrieve_by_relationships(
        self, 
        artifact_id: str
    ) -> List[RetrievalResult]:
        """Retrieve artifacts related to a given artifact."""
        # First find the artifact
        target = None
        for artifact in self.catalog.get('artifacts', []):
            if artifact['id'] == artifact_id:
                target = artifact
                break
        
        if not target:
            return []
        
        results = []
        related_ids = set(target.get('relationships', []))
        
        for artifact in self.catalog.get('artifacts', []):
            if artifact['id'] in related_ids:
                results.append(RetrievalResult(
                    artifact_id=artifact['id'],
                    title=artifact['title'],
                    domain=artifact['domain'],
                    relevance_score=0.8,
                    match_reason=f"related:{artifact_id}",
                    summary=artifact.get('summary', '')
                ))
        
        return results
    
    def retrieve_all(self) -> List[RetrievalResult]:
        """Retrieve all knowledge artifacts."""
        results = []
        for artifact in self.catalog.get('artifacts', []):
            results.append(RetrievalResult(
                artifact_id=artifact['id'],
                title=artifact['title'],
                domain=artifact['domain'],
                relevance_score=1.0,
                match_reason="all",
                summary=artifact.get('summary', '')
            ))
        return results
    
    def get_artifact(self, artifact_id: str) -> Optional[Dict]:
        """Get a specific artifact by ID."""
        for artifact in self.catalog.get('artifacts', []):
            if artifact['id'] == artifact_id:
                return artifact
        return None
    
    def get_domains(self) -> List[str]:
        """Get all available domains."""
        return self.catalog.get('domains', [])
    
    def search(self, query: str, limit: int = 10) -> List[RetrievalResult]:
        """
        General search combining multiple strategies.
        
        This is a simplified semantic-ish search that checks:
        - Keywords
        - Title
        - Summary
        """
        results = []
        query_lower = query.lower()
        query_terms = query_lower.split()
        
        for artifact in self.catalog.get('artifacts', []):
            score = 0.0
            match_reasons = []
            
            # Title match (highest weight)
            if query_lower in artifact.get('title', '').lower():
                score += 0.5
                match_reasons.append('title')
            
            # Keyword match
            artifact_keywords = set(k.lower() for k in artifact.get('keywords', []))
            query_keywords = set(query_terms)
            keyword_matches = artifact_keywords & query_keywords
            if keyword_matches:
                score += 0.3 * len(keyword_matches) / max(len(query_keywords), 1)
                match_reasons.append(f"keywords:{list(keyword_matches)}")
            
            # Summary match
            if query_lower in artifact.get('summary', '').lower():
                score += 0.2
                match_reasons.append('summary')
            
            if score > 0:
                results.append(RetrievalResult(
                    artifact_id=artifact['id'],
                    title=artifact['title'],
                    domain=artifact['domain'],
                    relevance_score=min(score, 1.0),
                    match_reason=','.join(match_reasons),
                    summary=artifact.get('summary', '')
                ))
        
        results.sort(key=lambda x: x.relevance_score, reverse=True)
        return results[:limit]


def test_retrieval():
    """Test the retrieval engine."""
    engine = RetrievalEngine()
    
    print("=== Knowledge Retrieval Engine Test ===\n")
    
    # Test domain retrieval
    print("1. Domain Retrieval (architecture):")
    results = engine.retrieve_by_domain('architecture')
    for r in results:
        print(f"   [{r.artifact_id}] {r.title} (score: {r.relevance_score})")
    
    print("\n2. Keyword Retrieval (investigation, experiment):")
    results = engine.retrieve_by_keywords(['investigation', 'experiment'])
    for r in results:
        print(f"   [{r.artifact_id}] {r.title} (score: {r.relevance_score})")
    
    print("\n3. Investigation History (INV-013):")
    results = engine.retrieve_by_investigation('INV-013')
    for r in results:
        print(f"   [{r.artifact_id}] {r.title}")
    
    print("\n4. General Search ('SCADA'):")
    results = engine.search('SCADA')
    for r in results:
        print(f"   [{r.artifact_id}] {r.title} (score: {r.relevance_score})")
    
    print("\n5. All Artifacts:")
    results = engine.retrieve_all()
    print(f"   Total: {len(results)} artifacts")


if __name__ == "__main__":
    test_retrieval()
