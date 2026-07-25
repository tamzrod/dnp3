#!/usr/bin/env python3
"""
Protocol Synthesis Engine - Main Entry Point

Usage:
    python -m protocol_synth --run-id=RUN-001 --seed=12345 --output-dir=./runs
    python -m protocol_synth --batch=10 --output-dir=./runs
"""

import argparse
import sys
import os
from pathlib import Path
from datetime import datetime
import json

# Add src to path for imports
sys.path.insert(0, str(Path(__file__).parent))

from protocol_runner import run_protocol_synthesis


def run_single(run_id: str, seed: int, output_dir: str):
    """Execute a single protocol synthesis run."""
    print(f"\n{'='*60}")
    print(f"Protocol Synthesis Run: {run_id}")
    print(f"Seed: {seed}")
    print(f"Output: {output_dir}")
    print(f"{'='*60}\n")
    
    results = run_protocol_synthesis(run_id, seed, output_dir)
    
    print(f"\n{'='*60}")
    print(f"Run {run_id} Summary")
    print(f"{'='*60}")
    print(f"Protocol: {results.get('protocol_name', 'N/A')}")
    print(f"Status: {results.get('status', 'N/A')}")
    print(f"Duration: {results.get('duration', 0):.2f}s")
    
    if results.get('compilation', {}).get('status') == 'success':
        print("✓ Compilation: SUCCESS")
    else:
        print("✗ Compilation: FAILED")
    
    ft = results.get('functional_test', {})
    print(f"✓ Functional: {ft.get('passed', 0)}/{ft.get('passed', 0) + ft.get('failed', 0)} passed")
    
    st = results.get('security_test', {})
    print(f"✓ Security: {st.get('resistance_rating', 'N/A')} ({st.get('resistance_score', 0)*100:.0f}%)")
    
    if results.get('errors'):
        print(f"\nErrors:")
        for err in results['errors']:
            print(f"  - {err}")
    
    return results


def run_batch(count: int, output_dir: str, start_seed: int = None):
    """Execute multiple protocol synthesis runs."""
    print(f"\n{'='*60}")
    print(f"Batch Protocol Synthesis: {count} runs")
    print(f"Output: {output_dir}")
    print(f"{'='*60}\n")
    
    # Use time-based seed if not provided
    if start_seed is None:
        start_seed = int(datetime.utcnow().timestamp())
    
    results_summary = {
        "batch_id": f"BATCH-{datetime.utcnow().strftime('%Y%m%d-%H%M%S')}",
        "run_count": count,
        "start_time": datetime.utcnow().isoformat() + "Z",
        "runs": [],
        "statistics": {},
    }
    
    successful = 0
    failed = 0
    protocol_names = set()
    
    for i in range(count):
        run_id = f"RUN-{i+1:03d}"
        seed = start_seed + i
        
        print(f"\n--- Run {i+1}/{count}: {run_id} ---")
        
        results = run_single(run_id, seed, output_dir)
        
        results_summary["runs"].append({
            "run_id": run_id,
            "seed": seed,
            "protocol_name": results.get("protocol_name"),
            "status": results.get("status"),
            "compilation": results.get("compilation", {}).get("status"),
            "functional": results.get("functional_test", {}).get("overall"),
            "security": results.get("security_test", {}).get("resistance_rating"),
        })
        
        if results.get("status") in ["completed", "completed_with_warnings"]:
            successful += 1
            if results.get("protocol_name"):
                protocol_names.add(results["protocol_name"])
        else:
            failed += 1
    
    # Calculate statistics
    end_time = datetime.utcnow()
    results_summary["end_time"] = end_time.isoformat() + "Z"
    results_summary["duration_seconds"] = (end_time - datetime.fromisoformat(results_summary["start_time"].replace("Z", ""))).total_seconds()
    
    results_summary["statistics"] = {
        "total_runs": count,
        "successful": successful,
        "failed": failed,
        "success_rate": successful / count if count > 0 else 0,
        "unique_protocols": len(protocol_names),
        "protocols": sorted(list(protocol_names)),
    }
    
    # Write batch summary
    summary_file = Path(output_dir) / f"batch_summary_{results_summary['batch_id']}.json"
    with open(summary_file, "w") as f:
        json.dump(results_summary, f, indent=2)
    
    print(f"\n{'='*60}")
    print(f"Batch Complete")
    print(f"{'='*60}")
    print(f"Total runs: {count}")
    print(f"Successful: {successful} ({successful/count*100:.1f}%)")
    print(f"Failed: {failed}")
    print(f"Unique protocols: {len(protocol_names)}")
    print(f"Duration: {results_summary['duration_seconds']:.2f}s")
    print(f"\nSummary written to: {summary_file}")
    
    return results_summary


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Protocol Synthesis Engine - Generate novel secure protocols",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Single run
  python -m protocol_synth --run-id=RUN-001 --seed=12345 --output-dir=./runs
  
  # Batch of 10 runs
  python -m protocol_synth --batch=10 --output-dir=./runs
  
  # Batch with custom start seed
  python -m protocol_synth --batch=100 --output-dir=./runs --start-seed=1000000
        """
    )
    
    parser.add_argument("--run-id", type=str, help="Run identifier (e.g., RUN-001)")
    parser.add_argument("--seed", type=int, help="Random seed for reproducibility")
    parser.add_argument("--output-dir", type=str, default="./runs", help="Output directory")
    parser.add_argument("--batch", type=int, help="Run multiple times")
    parser.add_argument("--start-seed", type=int, help="Starting seed for batch")
    
    args = parser.parse_args()
    
    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    if args.batch:
        # Batch mode
        run_batch(args.batch, str(output_dir), args.start_seed)
    elif args.run_id:
        # Single run mode
        seed = args.seed if args.seed is not None else int(datetime.utcnow().timestamp())
        run_single(args.run_id, seed, str(output_dir))
    else:
        # Interactive mode
        run_id = input("Enter run ID (e.g., RUN-001): ").strip() or f"RUN-{datetime.utcnow().strftime('%H%M%S')}"
        seed = int(datetime.utcnow().timestamp())
        run_single(run_id, seed, str(output_dir))


if __name__ == "__main__":
    main()
