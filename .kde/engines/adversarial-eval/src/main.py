#!/usr/bin/env python3
"""
Adversarial Evaluation Engine - Main Entry Point

Usage:
    python3 -m adversarial_eval --protocols-dir=../experiments/LAB-024/runs --output-dir=./results
    python3 -m adversarial_eval --batch --count=100
"""

import argparse
import os
import sys
import json
from pathlib import Path
from datetime import datetime
from collections import Counter
from typing import Dict, List, Any
from dataclasses import asdict

# Add src to path for imports
src_path = Path(__file__).parent
sys.path.insert(0, str(src_path))

from adversarial_engine import (
    evaluate_protocol,
    ProtocolEvaluation,
    PhaseResult,
    Vulnerability
)


def evaluate_batch(protocols_dir: Path, output_dir: Path) -> Dict:
    """Evaluate all protocols in the directory."""
    
    results = {
        "evaluation_id": f"EVAL-{datetime.utcnow().strftime('%Y%m%d-%H%M%S')}",
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "protocols_evaluated": 0,
        "protocols_passed": 0,
        "protocols_failed": 0,
        "protocols_partial": 0,
        "protocols_error": 0,
        "evaluations": [],
        "statistics": {},
        "vulnerability_summary": {
            "critical": 0,
            "high": 0,
            "medium": 0,
            "low": 0,
            "informational": 0,
        },
        "phase_summary": {},
        "patterns": {
            "recurring_weaknesses": [],
            "recurring_strengths": [],
        },
    }
    
    # Find all protocol run directories
    run_dirs = sorted(protocols_dir.glob("RUN-*"), key=lambda p: p.name)
    
    print(f"\n{'='*60}")
    print(f"Adversarial Evaluation Engine")
    print(f"Protocols to evaluate: {len(run_dirs)}")
    print(f"{'='*60}\n")
    
    for i, run_dir in enumerate(run_dirs):
        run_id = run_dir.name
        print(f"[{i+1}/{len(run_dirs)}] Evaluating {run_id}...", end=" ")
        
        try:
            evaluation = evaluate_protocol(run_dir, run_id)
            results["evaluations"].append(asdict(evaluation))
            results["protocols_evaluated"] += 1
            
            if evaluation.status == "pass":
                results["protocols_passed"] += 1
                print(f"PASS (score: {evaluation.overall_score:.1f})")
            elif evaluation.status == "fail":
                results["protocols_failed"] += 1
                print(f"FAIL (score: {evaluation.overall_score:.1f}, crit: {evaluation.critical_vulnerabilities})")
            elif evaluation.status == "partial":
                results["protocols_partial"] += 1
                print(f"PARTIAL (score: {evaluation.overall_score:.1f})")
            else:
                results["protocols_error"] += 1
                print(f"ERROR")
            
            # Accumulate vulnerability counts
            results["vulnerability_summary"]["critical"] += evaluation.critical_vulnerabilities
            results["vulnerability_summary"]["high"] += evaluation.high_vulnerabilities
            results["vulnerability_summary"]["medium"] += evaluation.medium_vulnerabilities
            results["vulnerability_summary"]["low"] += evaluation.low_vulnerabilities
            results["vulnerability_summary"]["informational"] += evaluation.informational
            
            # Accumulate phase scores
            for phase in evaluation.phases:
                if phase.phase not in results["phase_summary"]:
                    results["phase_summary"][phase.phase] = {
                        "name": phase.name,
                        "total_score": 0,
                        "count": 0,
                        "passes": 0,
                    }
                results["phase_summary"][phase.phase]["total_score"] += phase.score
                results["phase_summary"][phase.phase]["count"] += 1
                if phase.status == "pass":
                    results["phase_summary"][phase.phase]["passes"] += 1
                    
        except Exception as e:
            print(f"ERROR: {e}")
            results["protocols_error"] += 1
            results["evaluations"].append({
                "run_id": run_id,
                "status": "error",
                "error": str(e),
            })
    
    # Calculate phase averages
    for phase_num, phase_data in results["phase_summary"].items():
        phase_data["average_score"] = phase_data["total_score"] / phase_data["count"]
        phase_data["pass_rate"] = phase_data["passes"] / phase_data["count"]
    
    # Calculate overall statistics
    scores = [e.get("overall_score", 0) for e in results["evaluations"] if "overall_score" in e]
    if scores:
        results["statistics"]["mean_score"] = sum(scores) / len(scores)
        results["statistics"]["median_score"] = sorted(scores)[len(scores) // 2]
        results["statistics"]["min_score"] = min(scores)
        results["statistics"]["max_score"] = max(scores)
    
    # Identify recurring patterns
    all_vulnerabilities = []
    for eval_data in results["evaluations"]:
        for phase in eval_data.get("phases", []):
            all_vulnerabilities.extend(phase.get("vulnerabilities", []))
    
    vuln_titles = Counter(v.get("title", "") for v in all_vulnerabilities)
    
    results["patterns"]["recurring_weaknesses"] = [
        {"title": title, "count": count}
        for title, count in vuln_titles.most_common(10)
    ]
    
    # Write results
    output_dir.mkdir(parents=True, exist_ok=True)
    
    results_file = output_dir / f"evaluation_results_{results['evaluation_id']}.json"
    with open(results_file, "w") as f:
        json.dump(results, f, indent=2)
    
    # Generate summary report
    summary_file = output_dir / f"evaluation_summary_{results['evaluation_id']}.md"
    with open(summary_file, "w") as f:
        f.write("# Adversarial Evaluation Summary\n\n")
        f.write(f"**Evaluation ID**: {results['evaluation_id']}\n")
        f.write(f"**Timestamp**: {results['timestamp']}\n\n")
        
        f.write("## Overall Results\n\n")
        f.write(f"| Metric | Value |\n")
        f.write(f"|--------|-------|\n")
        f.write(f"| Protocols Evaluated | {results['protocols_evaluated']} |\n")
        f.write(f"| Passed | {results['protocols_passed']} ({results['protocols_passed']/max(1,results['protocols_evaluated'])*100:.1f}%) |\n")
        f.write(f"| Failed | {results['protocols_failed']} ({results['protocols_failed']/max(1,results['protocols_evaluated'])*100:.1f}%) |\n")
        f.write(f"| Partial | {results['protocols_partial']} |\n")
        f.write(f"| Errors | {results['protocols_error']} |\n\n")
        
        f.write("## Vulnerability Summary\n\n")
        f.write(f"| Severity | Count |\n")
        f.write(f"|----------|-------|\n")
        f.write(f"| Critical | {results['vulnerability_summary']['critical']} |\n")
        f.write(f"| High | {results['vulnerability_summary']['high']} |\n")
        f.write(f"| Medium | {results['vulnerability_summary']['medium']} |\n")
        f.write(f"| Low | {results['vulnerability_summary']['low']} |\n")
        f.write(f"| Informational | {results['vulnerability_summary']['informational']} |\n\n")
        
        f.write("## Score Statistics\n\n")
        if results["statistics"]:
            f.write(f"- Mean Score: {results['statistics'].get('mean_score', 0):.1f}\n")
            f.write(f"- Median Score: {results['statistics'].get('median_score', 0):.1f}\n")
            f.write(f"- Min Score: {results['statistics'].get('min_score', 0):.1f}\n")
            f.write(f"- Max Score: {results['statistics'].get('max_score', 0):.1f}\n\n")
        
        f.write("## Phase Performance\n\n")
        f.write(f"| Phase | Name | Avg Score | Pass Rate |\n")
        f.write(f"|-------|------|-----------|----------|\n")
        for phase_num in sorted(results["phase_summary"].keys()):
            phase = results["phase_summary"][phase_num]
            f.write(f"| {phase_num} | {phase['name']} | {phase['average_score']:.1f} | {phase['pass_rate']*100:.1f}% |\n")
        
        f.write("\n## Recurring Weaknesses\n\n")
        for item in results["patterns"]["recurring_weaknesses"][:5]:
            f.write(f"- {item['title']}: {item['count']} occurrences\n")
        
        f.write("\n")
    
    return results


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Adversarial Evaluation Engine - Security testing for synthesized protocols",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Evaluate all protocols from INV-003
  python3 -m adversarial_eval --protocols-dir=../experiments/LAB-024/runs --output-dir=./results
  
  # Evaluate specific protocols
  python3 -m adversarial_eval --protocols=RUN-001,RUN-002 --protocols-dir=../experiments/LAB-024/runs
        """
    )
    
    parser.add_argument("--protocols-dir", type=str, 
                       default="../experiments/LAB-024/runs",
                       help="Directory containing protocol run directories")
    parser.add_argument("--output-dir", type=str, 
                       default="./evaluation_results",
                       help="Output directory for results")
    parser.add_argument("--batch", action="store_true",
                       help="Run in batch mode (evaluate all protocols)")
    parser.add_argument("--count", type=int, default=100,
                       help="Number of protocols to evaluate")
    
    args = parser.parse_args()
    
    # Resolve paths
    protocols_dir = Path(__file__).parent.parent.parent / args.protocols_dir
    output_dir = Path(args.output_dir)
    
    if not protocols_dir.exists():
        print(f"Error: Protocols directory not found: {protocols_dir}")
        sys.exit(1)
    
    print(f"Protocols directory: {protocols_dir}")
    print(f"Output directory: {output_dir}")
    
    # Run evaluation
    results = evaluate_batch(protocols_dir, output_dir)
    
    # Print summary
    print(f"\n{'='*60}")
    print("Evaluation Complete")
    print(f"{'='*60}")
    print(f"Protocols Evaluated: {results['protocols_evaluated']}")
    print(f"Passed: {results['protocols_passed']} ({results['protocols_passed']/max(1,results['protocols_evaluated'])*100:.1f}%)")
    print(f"Failed: {results['protocols_failed']} ({results['protocols_failed']/max(1,results['protocols_evaluated'])*100:.1f}%)")
    print(f"Mean Score: {results['statistics'].get('mean_score', 0):.1f}")
    print(f"\nResults written to: {output_dir}")
    print(f"Summary written to: {output_dir}/evaluation_summary_*.md")


if __name__ == "__main__":
    main()
