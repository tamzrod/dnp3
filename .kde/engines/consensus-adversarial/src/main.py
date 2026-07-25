#!/usr/bin/env python3
"""
Consensus Adversarial Evaluation - Main Entry Point
"""

import argparse
import json
import os
from pathlib import Path
from datetime import datetime
from consensus_adversarial import evaluate_consensus_protocol


def main():
    parser = argparse.ArgumentParser(description="Consensus Adversarial Evaluation")
    parser.add_argument("--protocols-dir", type=str, required=True)
    parser.add_argument("--output-dir", type=str, required=True)
    args = parser.parse_args()
    
    protocols_dir = Path(args.protocols_dir)
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    results = []
    
    print("=" * 60)
    print("Consensus Adversarial Evaluation")
    print("=" * 60)
    
    for run_dir in sorted(protocols_dir.glob("RUN-*")):
        result_file = run_dir / "results.json"
        if result_file.exists():
            with open(result_file) as f:
                protocol_result = json.load(f)
            
            eval_result = evaluate_consensus_protocol(protocol_result)
            results.append({
                "protocol_name": protocol_result.get("protocol_name"),
                "run_id": run_dir.name,
                "adversarial": eval_result,
            })
            
            print(f"{run_dir.name}: Score={eval_result['overall_score']}, "
                  f"Critical={eval_result['critical_failures']}, "
                  f"High={eval_result['high_failures']}")
    
    # Write combined results
    with open(output_dir / "adversarial_results.json", "w") as f:
        json.dump({
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "total_protocols": len(results),
            "results": results,
        }, f, indent=2)
    
    print(f"\nResults written to {output_dir / 'adversarial_results.json'}")


if __name__ == "__main__":
    main()
