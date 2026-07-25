"""
Consensus Manager Module

Coordinates consensus-based engine execution.
"""

from typing import List, Dict, Optional, Any, Callable
from dataclasses import dataclass, field
from enum import Enum

from ..models import (
    ConsensusStrategy, EngineResult, AggregatedResult, EngineMetadata
)


@dataclass
class ConsensusVote:
    """A single vote in the consensus process."""
    engine_id: str
    engine_version: str
    vote: Any
    confidence: float
    reasoning: str = ""


@dataclass
class ConsensusResult:
    """Result of consensus coordination."""
    reached: bool
    strategy: ConsensusStrategy
    votes: List[ConsensusVote] = field(default_factory=list)
    consensus_value: Optional[Any] = None
    disagreement_count: int = 0
    confidence: float = 0.0
    details: List[str] = field(default_factory=list)


class ConsensusManager:
    """
    Manages consensus coordination for multi-engine execution.
    
    Responsibilities:
    - Coordinate consensus among multiple engines
    - Support different consensus strategies
    - Determine consensus agreement
    - Handle disagreement resolution
    """
    
    def __init__(self):
        """Initialize the Consensus Manager."""
        self._consensus_history: List[ConsensusResult] = []
    
    def coordinate(
        self,
        results: List[EngineResult],
        strategy: ConsensusStrategy,
        engine_metadata: Dict[str, EngineMetadata]
    ) -> ConsensusResult:
        """
        Coordinate consensus among engine results.
        
        Args:
            results: List of engine results
            strategy: Consensus strategy to use
            engine_metadata: Metadata for engines
        
        Returns:
            ConsensusResult
        """
        # Convert results to votes
        votes = []
        for result in results:
            metadata = engine_metadata.get(result.engine_id)
            votes.append(ConsensusVote(
                engine_id=result.engine_id,
                engine_version=result.engine_version,
                vote=result.outputs,
                confidence=1.0 if result.success else 0.0,
                reasoning=f"Engine {result.engine_id} execution {'successful' if result.success else 'failed'}"
            ))
        
        # Apply consensus strategy
        if strategy == ConsensusStrategy.SINGLE:
            return self._consensus_single(votes)
        elif strategy == ConsensusStrategy.MAJORITY:
            return self._consensus_majority(votes)
        elif strategy == ConsensusStrategy.UNANIMOUS:
            return self._consensus_unanimous(votes)
        elif strategy == ConsensusStrategy.WEIGHTED:
            return self._consensus_weighted(votes, engine_metadata)
        elif strategy == ConsensusStrategy.ADVERSARIAL:
            return self._consensus_adversarial(votes, results)
        
        # Default: single
        return self._consensus_single(votes)
    
    def _consensus_single(self, votes: List[ConsensusVote]) -> ConsensusResult:
        """
        Single vote consensus - first successful result wins.
        
        Args:
            votes: List of consensus votes
        
        Returns:
            ConsensusResult
        """
        successful_votes = [v for v in votes if v.confidence > 0]
        
        if successful_votes:
            # Return first successful vote
            winner = successful_votes[0]
            return ConsensusResult(
                reached=True,
                strategy=ConsensusStrategy.SINGLE,
                votes=votes,
                consensus_value=winner.vote,
                confidence=winner.confidence,
                details=[f"Single consensus: {winner.engine_id} result selected"]
            )
        
        # No successful votes
        return ConsensusResult(
            reached=False,
            strategy=ConsensusStrategy.SINGLE,
            votes=votes,
            details=["No successful votes for consensus"]
        )
    
    def _consensus_majority(self, votes: List[ConsensusVote]) -> ConsensusResult:
        """
        Majority consensus - more than 50% agreement required.
        
        Args:
            votes: List of consensus votes
        
        Returns:
            ConsensusResult
        """
        successful_votes = [v for v in votes if v.confidence > 0]
        total_votes = len(votes)
        
        if total_votes == 0:
            return ConsensusResult(
                reached=False,
                strategy=ConsensusStrategy.MAJORITY,
                votes=votes,
                details=["No votes cast"]
            )
        
        # Group votes by output hash
        vote_groups: Dict[str, List[ConsensusVote]] = {}
        for vote in successful_votes:
            key = self._hash_output(vote.vote)
            if key not in vote_groups:
                vote_groups[key] = []
            vote_groups[key].append(vote)
        
        # Find majority group
        majority_count = total_votes / 2
        for key, group in vote_groups.items():
            if len(group) > majority_count:
                # Majority found
                avg_confidence = sum(v.confidence for v in group) / len(group)
                return ConsensusResult(
                    reached=True,
                    strategy=ConsensusStrategy.MAJORITY,
                    votes=votes,
                    consensus_value=group[0].vote,
                    confidence=avg_confidence,
                    disagreement_count=total_votes - len(group),
                    details=[f"Majority consensus: {len(group)}/{total_votes} agree"]
                )
        
        # No majority
        return ConsensusResult(
            reached=False,
            strategy=ConsensusStrategy.MAJORITY,
            votes=votes,
            disagreement_count=total_votes,
            details=[f"No majority: {len(vote_groups)} different outcomes"]
        )
    
    def _consensus_unanimous(self, votes: List[ConsensusVote]) -> ConsensusResult:
        """
        Unanimous consensus - all must agree.
        
        Args:
            votes: List of consensus votes
        
        Returns:
            ConsensusResult
        """
        successful_votes = [v for v in votes if v.confidence > 0]
        
        if len(successful_votes) != len(votes):
            return ConsensusResult(
                reached=False,
                strategy=ConsensusStrategy.UNANIMOUS,
                votes=votes,
                disagreement_count=len(votes) - len(successful_votes),
                details=["Not all engines successful"]
            )
        
        # Group by output
        vote_groups: Dict[str, List[ConsensusVote]] = {}
        for vote in successful_votes:
            key = self._hash_output(vote.vote)
            if key not in vote_groups:
                vote_groups[key] = []
            vote_groups[key].append(vote)
        
        if len(vote_groups) == 1:
            # All agree
            avg_confidence = sum(v.confidence for v in successful_votes) / len(successful_votes)
            return ConsensusResult(
                reached=True,
                strategy=ConsensusStrategy.UNANIMOUS,
                votes=votes,
                consensus_value=successful_votes[0].vote,
                confidence=avg_confidence,
                details=["Unanimous agreement reached"]
            )
        
        # Disagreement
        return ConsensusResult(
            reached=False,
            strategy=ConsensusStrategy.UNANIMOUS,
            votes=votes,
            disagreement_count=len(votes),
            details=[f"Disagreement: {len(vote_groups)} different outcomes"]
        )
    
    def _consensus_weighted(
        self,
        votes: List[ConsensusVote],
        engine_metadata: Dict[str, EngineMetadata]
    ) -> ConsensusResult:
        """
        Weighted consensus - weighted by engine priority.
        
        Args:
            votes: List of consensus votes
            engine_metadata: Engine metadata for weighting
        
        Returns:
            ConsensusResult
        """
        # Calculate weighted scores
        weighted_votes: Dict[str, float] = {}
        
        for vote in votes:
            metadata = engine_metadata.get(vote.engine_id)
            weight = metadata.priority if metadata else 100
            
            key = self._hash_output(vote.vote)
            if key not in weighted_votes:
                weighted_votes[key] = 0.0
            weighted_votes[key] += weight * vote.confidence
        
        if not weighted_votes:
            return ConsensusResult(
                reached=False,
                strategy=ConsensusStrategy.WEIGHTED,
                votes=votes,
                details=["No weighted votes"]
            )
        
        # Find highest weighted outcome
        total_weight = sum(weighted_votes.values())
        best_key = max(weighted_votes, key=weighted_votes.get)
        best_score = weighted_votes[best_key]
        
        # Find the vote with that key
        best_vote = None
        for vote in votes:
            if self._hash_output(vote.vote) == best_key:
                best_vote = vote
                break
        
        return ConsensusResult(
            reached=True,
            strategy=ConsensusStrategy.WEIGHTED,
            votes=votes,
            consensus_value=best_vote.vote if best_vote else None,
            confidence=best_score / total_weight if total_weight > 0 else 0.0,
            details=[f"Weighted consensus: {best_score:.0f}/{total_weight:.0f} weight"]
        )
    
    def _consensus_adversarial(
        self,
        votes: List[ConsensusVote],
        results: List[EngineResult]
    ) -> ConsensusResult:
        """
        Adversarial consensus - evaluate results against each other.
        
        Args:
            votes: List of consensus votes
            results: Engine results for evaluation
        
        Returns:
            ConsensusResult
        """
        # Find adversarial engine results
        adversarial_votes = [
            v for v in votes
            if 'adversarial' in v.engine_id.lower()
        ]
        
        if not adversarial_votes:
            # No adversarial engine, fall back to majority
            return self._consensus_majority(votes)
        
        # Evaluate each result against adversarial findings
        valid_votes = []
        for vote in votes:
            if vote.confidence <= 0:
                continue
            
            # Check if this vote contradicts adversarial findings
            is_valid = True
            for adv_vote in adversarial_votes:
                if self._contradicts(adv_vote.vote, vote.vote):
                    is_valid = False
                    break
            
            if is_valid:
                valid_votes.append(vote)
        
        if not valid_votes:
            return ConsensusResult(
                reached=False,
                strategy=ConsensusStrategy.ADVERSARIAL,
                votes=votes,
                disagreement_count=len(votes),
                details=["Adversarial evaluation invalidated all results"]
            )
        
        # Return best valid result
        best = valid_votes[0]
        return ConsensusResult(
            reached=True,
            strategy=ConsensusStrategy.ADVERSARIAL,
            votes=votes,
            consensus_value=best.vote,
            confidence=best.confidence,
            disagreement_count=len(votes) - len(valid_votes),
            details=[f"Adversarial: {len(valid_votes)}/{len(votes)} results valid"]
        )
    
    def _hash_output(self, output: Any) -> str:
        """Generate a hash key for an output."""
        if isinstance(output, dict):
            # Sort dict for consistent hashing
            items = sorted(output.items())
            return str(items)
        elif isinstance(output, (list, tuple)):
            return str(tuple(sorted(output)))
        else:
            return str(output)
    
    def _contradicts(self, output1: Any, output2: Any) -> bool:
        """Check if two outputs contradict each other."""
        # Simple contradiction check for demonstration
        if isinstance(output1, dict) and isinstance(output2, dict):
            # Check for conflicting keys
            common_keys = set(output1.keys()) & set(output2.keys())
            for key in common_keys:
                if output1[key] != output2[key]:
                    return True
        return False
    
    def get_consensus_summary(self) -> Dict[str, Any]:
        """
        Get consensus coordination summary.
        
        Returns:
            Consensus summary dictionary
        """
        total = len(self._consensus_history)
        reached = sum(1 for c in self._consensus_history if c.reached)
        
        strategy_counts = {}
        for c in self._consensus_history:
            strategy = c.strategy.value
            strategy_counts[strategy] = strategy_counts.get(strategy, 0) + 1
        
        return {
            "total_coordinations": total,
            "consensus_reached": reached,
            "success_rate": reached / total if total > 0 else 0.0,
            "by_strategy": strategy_counts
        }
