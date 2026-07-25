"""
Capability Resolver Module

Resolves runtime requests to engine capabilities.
"""

from typing import List, Dict, Optional, Any, Set
from dataclasses import dataclass, field

from ..models import (
    Capability, CapabilityType, EngineMetadata, SeedMetadata,
    EngineSelection, SeedSelection, ExecutionRequest
)


@dataclass
class CapabilityMatch:
    """A match between a capability request and an engine/seed."""
    capability: Capability
    engine: Optional[EngineMetadata] = None
    seed: Optional[SeedMetadata] = None
    match_score: float = 0.0
    match_reasons: List[str] = field(default_factory=list)


class CapabilityResolver:
    """
    Resolves execution requests to matching engines and seeds.
    
    Responsibilities:
    - Match required capabilities to available engines
    - Score and rank potential engine selections
    - Select appropriate seeds for engine execution
    - Generate capability resolution reports
    """
    
    def __init__(self):
        """Initialize the Capability Resolver."""
        self._resolution_cache: Dict[str, List[CapabilityMatch]] = {}
    
    def resolve(
        self,
        request: ExecutionRequest,
        engines: List[EngineMetadata],
        seeds: List[SeedMetadata]
    ) -> List[EngineSelection]:
        """
        Resolve an execution request to matching engines.
        
        Args:
            request: The execution request
            engines: List of available engines
            seeds: List of available seeds
        
        Returns:
            List of engine selections, ranked by match score
        """
        # Build capability requirements
        required_capabilities = set(request.required_capabilities)
        
        # Find matching engines for each required capability
        capability_matches: Dict[str, List[CapabilityMatch]] = {}
        
        for cap_type in required_capabilities:
            matches = self._find_engines_for_capability(cap_type, engines, request.keywords)
            if matches:
                capability_matches[cap_type.value] = matches
            else:
                # No engine found for this capability
                capability_matches[cap_type.value] = []
        
        # Generate engine selections
        selections = self._generate_engine_selections(
            capability_matches,
            required_capabilities,
            engines
        )
        
        # Sort by total match score
        selections.sort(key=lambda x: x.confidence, reverse=True)
        
        return selections
    
    def _find_engines_for_capability(
        self,
        capability: CapabilityType,
        engines: List[EngineMetadata],
        keywords: List[str]
    ) -> List[CapabilityMatch]:
        """
        Find engines that provide a specific capability.
        
        Args:
            capability: Required capability type
            engines: List of available engines
            keywords: Request keywords for relevance scoring
        
        Returns:
            List of capability matches
        """
        matches = []
        keywords_lower = [k.lower() for k in keywords]
        
        for engine in engines:
            for cap in engine.capabilities:
                if cap.type == capability:
                    # Calculate match score
                    score = self._calculate_match_score(cap, keywords_lower)
                    
                    # Generate match reasons
                    reasons = []
                    if engine.status.value == 'active':
                        reasons.append(f"{engine.codename} is active")
                    if score > 0.5:
                        reasons.append(f"keyword relevance: {score:.2f}")
                    if capability.value in [c.type.value for c in engine.capabilities]:
                        reasons.append(f"native {capability.value} capability")
                    
                    matches.append(CapabilityMatch(
                        engine=engine,
                        capability=cap,
                        match_score=score,
                        match_reasons=reasons
                    ))
        
        # Sort by match score
        matches.sort(key=lambda x: x.match_score, reverse=True)
        return matches
    
    def _calculate_match_score(
        self,
        capability: Capability,
        keywords: List[str]
    ) -> float:
        """
        Calculate relevance score between a capability and keywords.
        
        Args:
            capability: The capability to score
            keywords: Request keywords
        
        Returns:
            Match score between 0.0 and 1.0
        """
        if not keywords:
            return 0.5  # Default score without keywords
        
        score = 0.0
        cap_keywords_lower = [k.lower() for k in capability.keywords]
        
        # Count keyword matches
        matches = 0
        for keyword in keywords:
            if keyword.lower() in cap_keywords_lower:
                matches += 1
        
        if keywords:
            score = matches / len(keywords)
        
        return min(score + 0.3, 1.0)  # Add base score, cap at 1.0
    
    def _generate_engine_selections(
        self,
        capability_matches: Dict[str, List[CapabilityMatch]],
        required_capabilities: Set[CapabilityType],
        engines: List[EngineMetadata]
    ) -> List[EngineSelection]:
        """
        Generate engine selections from capability matches.
        
        Args:
            capability_matches: Matches per capability type
            required_capabilities: Set of required capabilities
            engines: List of all engines
        
        Returns:
            List of engine selections with scores
        """
        selections = []
        seen_engines: Set[str] = set()
        
        # First pass: collect engines that satisfy ALL required capabilities
        for engine in engines:
            engine_capabilities = set(c.type for c in engine.capabilities)
            
            # Check if engine satisfies all required capabilities
            if required_capabilities.issubset(engine_capabilities):
                # Calculate aggregate score
                total_score = 0.0
                reasons = []
                
                for cap_type in required_capabilities:
                    matches = capability_matches.get(cap_type.value, [])
                    for match in matches:
                        if match.engine and match.engine.engine_id == engine.engine_id:
                            total_score += match.match_score
                            reasons.extend(match.match_reasons)
                            break
                
                if required_capabilities:
                    avg_score = total_score / len(required_capabilities)
                else:
                    avg_score = 0.5
                
                # Apply engine priority bonus
                priority_bonus = min(engine.priority / 200, 0.2)
                final_score = min(avg_score + priority_bonus, 1.0)
                
                if engine.engine_id not in seen_engines:
                    seen_engines.add(engine.engine_id)
                    selections.append(EngineSelection(
                        engine=engine,
                        reason=f"Satisfies all {len(required_capabilities)} required capabilities",
                        confidence=final_score
                    ))
        
        # Second pass: add engines that satisfy most capabilities (partial match)
        for engine in engines:
            if engine.engine_id in seen_engines:
                continue
            
            engine_capabilities = set(c.type for c in engine.capabilities)
            satisfied = required_capabilities.intersection(engine_capabilities)
            
            if satisfied and len(satisfied) > 0:
                # Partial match
                coverage = len(satisfied) / len(required_capabilities) if required_capabilities else 0
                
                if coverage >= 0.5:  # At least 50% coverage
                    reasons = [f"Satisfies {len(satisfied)}/{len(required_capabilities)} capabilities"]
                    
                    if engine.engine_id not in seen_engines:
                        seen_engines.add(engine.engine_id)
                        selections.append(EngineSelection(
                            engine=engine,
                            reason="; ".join(reasons),
                            confidence=coverage * 0.8  # Penalty for partial coverage
                        ))
        
        return selections
    
    def select_seeds(
        self,
        selected_engines: List[EngineSelection],
        preferred_seeds: List[str],
        all_seeds: List[SeedMetadata]
    ) -> List[SeedSelection]:
        """
        Select appropriate seeds for selected engines.
        
        Args:
            selected_engines: Engines selected for execution
            preferred_seeds: User's preferred seed IDs
            all_seeds: All available seeds
        
        Returns:
            List of seed selections
        """
        selections = []
        
        for engine_sel in selected_engines:
            engine = engine_sel.engine
            
            # Find compatible seeds
            compatible_seeds = [
                s for s in all_seeds
                if not s.compatible_engines or engine.engine_id in s.compatible_engines
            ]
            
            # Prefer user's preferred seeds
            for pref_seed_id in preferred_seeds:
                for seed in compatible_seeds:
                    if seed.seed_id == pref_seed_id:
                        selections.append(SeedSelection(
                            seed=seed,
                            reason=f"Preferred seed for {engine.codename}",
                            confidence=1.0
                        ))
                        break
            
            # Add any remaining compatible seeds
            for seed in compatible_seeds:
                if not any(s.seed.seed_id == seed.seed_id for s in selections):
                    selections.append(SeedSelection(
                        seed=seed,
                        reason=f"Compatible with {engine.codename}",
                        confidence=0.8
                    ))
        
        return selections
    
    def generate_resolution_report(
        self,
        request: ExecutionRequest,
        engine_selections: List[EngineSelection],
        seed_selections: List[SeedSelection]
    ) -> Dict[str, Any]:
        """
        Generate a capability resolution report.
        
        Args:
            request: Original execution request
            engine_selections: Engine selections
            seed_selections: Seed selections
        
        Returns:
            Resolution report dictionary
        """
        return {
            "request_id": request.request_id,
            "required_capabilities": [c.value for c in request.required_capabilities],
            "engine_count": len(engine_selections),
            "seed_count": len(seed_selections),
            "top_engine": engine_selections[0].engine.engine_id if engine_selections else None,
            "top_engine_confidence": engine_selections[0].confidence if engine_selections else 0.0,
            "engines_by_id": [
                {
                    "engine_id": s.engine.engine_id,
                    "codename": s.engine.codename,
                    "confidence": s.confidence,
                    "reason": s.reason
                }
                for s in engine_selections[:5]  # Top 5
            ]
        }
