#!/usr/bin/env python3
"""
Consensus Protocol Synthesis Engine - Main Entry Point

Usage:
    python3 -m consensus_synth --run-id=RUN-001 --seed=1 --output-dir=./runs
    python3 -m consensus_synth --batch=100
"""

import argparse
import os
import sys
import json
from pathlib import Path
from datetime import datetime
from typing import Dict, List

# Add src to path
src_path = Path(__file__).parent
sys.path.insert(0, str(src_path))

from consensus_engine import generate_consensus_protocol, ConsensusProtocolGenerator


def run_synthesis(run_id: str, seed: int, output_dir: Path) -> Dict:
    """Run a single consensus protocol synthesis."""
    
    print(f"\n{'='*60}")
    print(f"Consensus Protocol Synthesis: {run_id}")
    print(f"Seed: {seed}")
    print(f"{'='*60}\n")
    
    # Generate protocol
    print(f"[INFO] Generating consensus protocol...")
    spec, result = generate_consensus_protocol(seed)
    print(f"[INFO] Generated: {spec.name}")
    
    # Create output directory
    run_dir = output_dir / run_id
    run_dir.mkdir(parents=True, exist_ok=True)
    
    # Write specification
    spec_file = run_dir / "specification.md"
    with open(spec_file, "w") as f:
        f.write(f"# {spec.name}\n\n")
        f.write(f"**Version**: {spec.version}\n")
        f.write(f"**Generated**: {datetime.utcnow().isoformat()}Z\n")
        f.write(f"**Run ID**: {run_id}\n")
        f.write(f"**Seed**: {seed}\n\n")
        
        f.write("---\n\n")
        
        f.write("## Architecture\n\n")
        f.write(f"- Architecture: {spec.architecture}\n")
        f.write(f"- Node Type: {spec.node_type}\n")
        f.write(f"- Quorum Type: {spec.quorum_type}\n\n")
        
        f.write("## Leader Election\n\n")
        f.write(f"- Method: {spec.leader_election}\n")
        f.write(f"- Term/Epochs: {spec.term_epochs}\n")
        f.write(f"- Randomized Timeout: {spec.randomized_timeout}\n")
        f.write(f"- Lease-Based: {spec.lease_based}\n\n")
        
        f.write("## Voting\n\n")
        f.write(f"- Model: {spec.voting_model}\n")
        f.write(f"- Vote Granting: {spec.vote_granting}\n")
        f.write(f"- Vote Revocations: {spec.vote_revocations}\n\n")
        
        f.write("## Log Replication\n\n")
        f.write(f"- Type: {spec.log_type}\n")
        f.write(f"- Structure: {spec.log_structure}\n")
        f.write(f"- Commit Method: {spec.commit_method}\n\n")
        
        f.write("## Recovery\n\n")
        f.write(f"- Strategy: {spec.recovery_strategy}\n")
        f.write(f"- Snapshot Policy: {spec.snapshot_policy}\n\n")
        
        f.write("## Membership\n\n")
        f.write(f"- Model: {spec.membership_model}\n")
        f.write(f"- Add Node: {spec.add_node_method}\n")
        f.write(f"- Remove Node: {spec.remove_node_method}\n\n")
        
        f.write("## Failure Handling\n\n")
        f.write(f"- Detection: {spec.failure_detection}\n")
        f.write(f"- Model: {spec.failure_model}\n\n")
        
        f.write("## Timing\n\n")
        f.write(f"- Timeout Model: {spec.timeout_model}\n")
        f.write(f"- Heartbeat Interval: {spec.heartbeat_interval}ms\n")
        f.write(f"- Election Timeout Base: {spec.election_timeout_base}ms\n")
        f.write(f"- Election Timeout Range: {spec.election_timeout_range}ms\n\n")
        
        f.write("## Message Types\n\n")
        for msg in spec.message_types:
            f.write(f"- {msg}\n")
        f.write("\n")
        
        f.write("## State Machine\n\n")
        f.write(f"- Type: {spec.state_machine_type}\n")
        f.write(f"- Safety: {spec.provides_safety}\n")
        f.write(f"- Liveness: {spec.provides_liveness}\n\n")
        
        f.write("---\n\n")
        f.write("## Discovered Features\n\n")
        for feature in result["discovered_features"]:
            f.write(f"- {feature}\n")
        f.write("\n")
        
        f.write("---\n\n")
        f.write("## Simulation Results\n\n")
        for sim in result["simulations"]:
            f.write(f"### {sim['num_nodes']} Nodes\n\n")
            f.write(f"- Election Attempts: {sim['measurements']['election_attempts']}\n")
            f.write(f"- Elections Succeeded: {sim['measurements']['elections_succeeded']}\n")
            f.write(f"- Election Success Rate: {sim['metrics']['election_success_rate']:.2%}\n")
            f.write(f"- Messages Sent: {sim['measurements']['messages_sent']}\n")
            f.write(f"- Commits: {sim['measurements']['commits']}\n")
            f.write(f"- Avg Messages/Commit: {sim['metrics']['avg_messages_per_commit']:.2f}\n")
            f.write(f"- Leader Changes: {sim['measurements']['leader_changes']}\n")
            f.write(f"- Leader Stability: {sim['metrics']['leader_stability']:.2%}\n")
            f.write(f"- Safety Check: {'PASS' if sim['safety_check'] else 'FAIL'}\n")
            f.write(f"- Liveness Check: {'PASS' if sim['liveness_check'] else 'FAIL'}\n")
            f.write("\n")
    
    # Write JSON result
    result_file = run_dir / "results.json"
    with open(result_file, "w") as f:
        json.dump({
            "run_id": run_id,
            "seed": seed,
            "protocol_name": spec.name,
            "specification": spec.to_dict(),
            "discovered_features": result["discovered_features"],
            "simulations": result["simulations"],
            "timestamp": datetime.utcnow().isoformat() + "Z",
        }, f, indent=2)
    
    # Write summary
    summary_file = run_dir / "summary.md"
    with open(summary_file, "w") as f:
        f.write(f"# {run_id} Summary\n\n")
        f.write(f"**Protocol**: {spec.name}\n")
        f.write(f"**Status**: COMPLETE\n\n")
        f.write("## Key Features\n\n")
        f.write(f"- Leader Election: {spec.leader_election}\n")
        f.write(f"- Voting Model: {spec.voting_model}\n")
        f.write(f"- Log Type: {spec.log_type}\n")
        f.write(f"- Term Epochs: {spec.term_epochs}\n")
        f.write(f"- Randomized Timeout: {spec.randomized_timeout}\n\n")
        
        # Aggregate simulation results
        total_elections = sum(s["measurements"]["election_attempts"] for s in result["simulations"])
        total_succeeded = sum(s["measurements"]["elections_succeeded"] for s in result["simulations"])
        success_rate = total_succeeded / total_elections if total_elections > 0 else 0
        
        f.write("## Simulation Results\n\n")
        f.write(f"- Total Election Attempts: {total_elections}\n")
        f.write(f"- Total Elections Succeeded: {total_succeeded}\n")
        f.write(f"- Election Success Rate: {success_rate:.2%}\n")
        f.write(f"- Safety Violations: {sum(s['measurements']['safety_violations'] for s in result['simulations'])}\n")
        f.write(f"- Liveness Violations: {sum(s['measurements']['liveness_violations'] for s in result['simulations'])}\n")
    
    print(f"[INFO] Results written to {run_dir}")
    
    return result


def run_batch(num_runs: int, output_dir: Path, start_seed: int = None):
    """Run multiple synthesis iterations."""
    
    if start_seed is None:
        start_seed = int(datetime.utcnow().timestamp())
    
    print(f"\n{'='*60}")
    print(f"Batch Consensus Protocol Synthesis")
    print(f"Runs: {num_runs}")
    print(f"Output: {output_dir}")
    print(f"{'='*60}\n")
    
    results = []
    for i in range(num_runs):
        run_id = f"RUN-{i+1:03d}"
        seed = start_seed + i
        try:
            result = run_synthesis(run_id, seed, output_dir)
            results.append(result)
            print(f"[SUCCESS] {run_id}: {result['specification']['name']}")
        except Exception as e:
            print(f"[ERROR] {run_id}: {e}")
    
    # Write batch summary
    batch_file = output_dir / f"batch_summary_{datetime.utcnow().strftime('%Y%m%d-%H%M%S')}.json"
    with open(batch_file, "w") as f:
        json.dump({
            "batch_id": f"BATCH-{datetime.utcnow().strftime('%Y%m%d-%H%M%S')}",
            "num_runs": num_runs,
            "successful": len(results),
            "results": [
                {
                    "run_id": f"RUN-{i+1:03d}",
                    "protocol_name": r["specification"]["name"],
                    "features": r["discovered_features"],
                }
                for i, r in enumerate(results)
            ]
        }, f, indent=2)
    
    print(f"\n{'='*60}")
    print(f"Batch Complete")
    print(f"Successful: {len(results)}/{num_runs}")
    print(f"{'='*60}\n")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Consensus Protocol Synthesis Engine",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    
    parser.add_argument("--run-id", type=str, help="Run ID (e.g., RUN-001)")
    parser.add_argument("--seed", type=int, help="Random seed")
    parser.add_argument("--output-dir", type=str, default="./runs", help="Output directory")
    parser.add_argument("--batch", type=int, help="Run batch synthesis")
    
    args = parser.parse_args()
    
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    if args.batch:
        run_batch(args.batch, output_dir)
    elif args.run_id and args.seed:
        run_synthesis(args.run_id, args.seed, output_dir)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
