"""
Result Aggregator Module

Aggregates outputs from multiple engine executions.
"""

from typing import List, Dict, Optional, Any
from dataclasses import dataclass, field
from datetime import datetime

from ..models import (
    EngineResult, AggregatedResult, ExecutionPlan, EngineMetadata
)
from ..consensus import ConsensusResult


@dataclass
class AttributionEntry:
    """Attribution for a single output element."""
    engine_id: str
    contribution_weight: float
    reasoning: str = ""


class ResultAggregator:
    """
    Aggregates engine execution results while preserving provenance.
    
    Responsibilities:
    - Collect outputs from engine executions
    - Preserve provenance and attribution
    - Maintain transparency in aggregation
    - Support result caching
    """
    
    def __init__(self):
        """Initialize the Result Aggregator."""
        self._aggregation_cache: Dict[str, AggregatedResult] = {}
        self._attribution_cache: Dict[str, List[AttributionEntry]] = {}
    
    def aggregate(
        self,
        request_id: str,
        plan: ExecutionPlan,
        results: List[EngineResult],
        consensus_result: Optional[ConsensusResult] = None
    ) -> AggregatedResult:
        """
        Aggregate engine execution results.
        
        Args:
            request_id: Request identifier
            plan: Execution plan
            results: List of engine results
            consensus_result: Optional consensus result
        
        Returns:
            AggregatedResult
        """
        # Calculate total execution time
        total_time = sum(r.execution_time_ms for r in results)
        
        # Generate attribution
        attribution = self._generate_attribution(results, plan)
        
        # Aggregate outputs
        aggregated_outputs = self._aggregate_outputs(results, consensus_result)
        
        # Create result
        aggregated = AggregatedResult(
            request_id=request_id,
            plan_id=plan.plan_id,
            engine_results=results,
            consensus_reached=consensus_result.reached if consensus_result else False,
            consensus_value=consensus_result.consensus_value if consensus_result else None,
            aggregated_outputs=aggregated_outputs,
            attribution=attribution,
            total_execution_time_ms=total_time
        )
        
        # Cache result
        self._aggregation_cache[request_id] = aggregated
        
        return aggregated
    
    def _generate_attribution(
        self,
        results: List[EngineResult],
        plan: ExecutionPlan
    ) -> Dict[str, float]:
        """
        Generate attribution map for results.
        
        Args:
            results: Engine results
            plan: Execution plan
        
        Returns:
            Attribution dictionary (engine_id -> weight)
        """
        attribution = {}
        
        if not results:
            return attribution
        
        # Count successful results per engine
        engine_successes = {}
        for result in results:
            if result.success:
                engine_successes[result.engine_id] = (
                    engine_successes.get(result.engine_id, 0) + 1
                )
        
        # Calculate weights
        total_successes = sum(engine_successes.values())
        if total_successes > 0:
            for engine_id, count in engine_successes.items():
                attribution[engine_id] = count / total_successes
        else:
            # Equal weight if no successes
            for result in results:
                attribution[result.engine_id] = 1.0 / len(results)
        
        return attribution
    
    def _aggregate_outputs(
        self,
        results: List[EngineResult],
        consensus_result: Optional[ConsensusResult]
    ) -> Dict[str, Any]:
        """
        Aggregate outputs from engines.
        
        Args:
            results: Engine results
            consensus_result: Consensus result if applicable
        
        Returns:
            Aggregated outputs dictionary
        """
        aggregated = {
            "outputs": [],
            "successful_engines": [],
            "failed_engines": [],
            "output_count": 0,
            "metadata": {}
        }
        
        # Collect all successful outputs
        for result in results:
            if result.success:
                aggregated["successful_engines"].append(result.engine_id)
                aggregated["outputs"].append({
                    "engine_id": result.engine_id,
                    "outputs": result.outputs,
                    "provenance": result.provenance
                })
                aggregated["output_count"] += len(result.outputs)
            else:
                aggregated["failed_engines"].append(result.engine_id)
        
        # Add consensus value if available
        if consensus_result and consensus_result.reached:
            aggregated["consensus_value"] = consensus_result.consensus_value
            aggregated["consensus_strategy"] = consensus_result.strategy.value
        
        # Add metadata
        aggregated["metadata"] = {
            "total_results": len(results),
            "successful_count": len(aggregated["successful_engines"]),
            "failed_count": len(aggregated["failed_engines"]),
            "timestamp": datetime.now().isoformat()
        }
        
        return aggregated
    
    def get_attribution_report(
        self,
        request_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get detailed attribution report for a request.
        
        Args:
            request_id: Request identifier
        
        Returns:
            Attribution report or None
        """
        aggregated = self._aggregation_cache.get(request_id)
        if not aggregated:
            return None
        
        # Build detailed report
        report = {
            "request_id": request_id,
            "plan_id": aggregated.plan_id,
            "total_execution_time_ms": aggregated.total_execution_time_ms,
            "consensus_reached": aggregated.consensus_reached,
            "attribution": [],
            "provenance": []
        }
        
        for result in aggregated.engine_results:
            weight = aggregated.attribution.get(result.engine_id, 0.0)
            report["attribution"].append({
                "engine_id": result.engine_id,
                "engine_version": result.engine_version,
                "weight": weight,
                "success": result.success,
                "output_count": len(result.outputs)
            })
            
            report["provenance"].append({
                "engine_id": result.engine_id,
                "provenance": result.provenance
            })
        
        return report
    
    def get_aggregated_output(
        self,
        request_id: str,
        transparent: bool = True
    ) -> Optional[Dict[str, Any]]:
        """
        Get aggregated output with optional transparency.
        
        Args:
            request_id: Request identifier
            transparent: If True, preserve all engine attributions
        
        Returns:
            Aggregated output dictionary
        """
        aggregated = self._aggregation_cache.get(request_id)
        if not aggregated:
            return None
        
        if transparent:
            # Return full transparency output
            return {
                "request_id": request_id,
                "plan_id": aggregated.plan_id,
                "consensus_reached": aggregated.consensus_reached,
                "consensus_value": aggregated.consensus_value,
                "outputs": aggregated.aggregated_outputs,
                "attribution": aggregated.attribution,
                "total_execution_time_ms": aggregated.total_execution_time_ms,
                "timestamp": aggregated.timestamp.isoformat(),
                "transparency": "full"
            }
        else:
            # Return simplified output
            return {
                "request_id": request_id,
                "consensus_value": aggregated.consensus_value,
                "output_count": aggregated.aggregated_outputs.get("output_count", 0),
                "transparency": "simplified"
            }
    
    def get_aggregation_summary(self) -> Dict[str, Any]:
        """
        Get aggregation statistics summary.
        
        Returns:
            Aggregation summary dictionary
        """
        total = len(self._aggregation_cache)
        
        consensus_reached = sum(
            1 for a in self._aggregation_cache.values()
            if a.consensus_reached
        )
        
        total_time = sum(
            a.total_execution_time_ms
            for a in self._aggregation_cache.values()
        )
        
        return {
            "total_aggregations": total,
            "consensus_reached": consensus_reached,
            "consensus_rate": consensus_reached / total if total > 0 else 0.0,
            "total_execution_time_ms": total_time,
            "avg_execution_time_ms": total_time / total if total > 0 else 0.0
        }
