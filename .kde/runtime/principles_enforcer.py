"""
Five Core Principles Enforcement Module

This module provides ACTUAL enforcement of the Five Core Principles,
not just acknowledgment. It implements runtime checks and blocks
operations that violate these principles.

The Five Core Principles (FROZEN as SEED-001):
1. No Auto-Continuation - Never begin next session without human authorization
2. No Self-Approval - Never approve your own work
3. No Self-Promotion - Never promote knowledge to production
4. Distinguish Evidence - Mark fact vs. conclusion vs. speculation
5. Evidence-Based Changes - All claims must be justified
"""

import re
import os
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum
from datetime import datetime


class PrincipleType(Enum):
    """Types of principles that can be enforced."""
    NO_AUTO_CONTINUATION = "no_auto_continuation"
    NO_SELF_APPROVAL = "no_self_approval"
    NO_SELF_PROMOTION = "no_self_promotion"
    DISTINGUISH_EVIDENCE = "distinguish_evidence"
    EVIDENCE_BASED_CHANGES = "evidence_based_changes"


class EvidenceLevel(Enum):
    """Classification of content based on evidence type."""
    EVIDENCE = "evidence"  # Documented facts with citations
    INFERENCE = "inference"  # Conclusions drawn from evidence
    HYPOTHESIS = "hypothesis"  # Speculation beyond evidence


class ContinuationStatus(Enum):
    """Status of session continuation authorization."""
    AUTHORIZED = "authorized"
    REQUIRED = "required"
    BLOCKED = "blocked"


@dataclass
class PrincipleViolation:
    """A single principle violation."""
    principle: PrincipleType
    description: str
    severity: str  # "error", "warning"
    blocked: bool
    location: Optional[str] = None
    suggestion: Optional[str] = None


@dataclass
class EnforcementResult:
    """Result of a principle enforcement check."""
    passed: bool
    violations: List[PrincipleViolation] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=datetime.now)


@dataclass
class SessionCheckpoint:
    """A checkpoint requiring human authorization to proceed."""
    checkpoint_id: str
    description: str
    status: ContinuationStatus
    created_at: datetime
    authorized_by: Optional[str] = None
    authorized_at: Optional[datetime] = None


class FivePrinciplesEnforcer:
    """
    Enforcer for the Five Core Principles.
    
    This class provides ACTUAL enforcement mechanisms:
    - Blocks self-approval transitions
    - Blocks self-promotion transitions
    - Requires authorization checkpoints for session continuation
    - Validates evidence/marker distinctions in content
    - Requires evidence citations for claims
    
    Usage:
        enforcer = FivePrinciplesEnforcer()
        result = enforcer.check_state_transition(current_state, new_state)
        if not result.passed:
            raise PrincipleViolationError(result)
    """
    
    # States that require human authorization (from STATE-MACHINE)
    APPROVED_STATES = {"APPROVED", "approved"}
    PROMOTED_STATES = {"PROMOTED", "promoted", "production"}
    REVIEW_STATES = {"REVIEW", "review"}
    
    # Evidence patterns
    EVIDENCE_PATTERNS = [
        r'\[EVIDENCE:\s*[^\]]+\]',  # [EVIDENCE: citation]
        r'\*\*Evidence:\*\*\s*',   # **Evidence:** text
        r'Source:\s*https?://',      # Source: http(s)://
    ]
    
    INFERENCE_PATTERNS = [
        r'\[INFERENCE:\s*[^\]]+\]',  # [INFERENCE: conclusion]
        r'\*\*Inference:\*\*\s*',    # **Inference:** text
        r'This suggests that',
        r'Based on the evidence,',
        r'Therefore,',
    ]
    
    HYPOTHESIS_PATTERNS = [
        r'\[HYPOTHESIS:\s*[^\]]+\]',  # [HYPOTHESIS: speculation]
        r'\*\*Hypothesis:\*\*\s*',    # **Hypothesis:** text
        r'It may be that',
        r'Possibly,',
        r'Maybe,',
        r'Perhaps,',
    ]
    
    def __init__(self, kde_root: str = "/workspace/project/kde"):
        """
        Initialize the Five Principles Enforcer.
        
        Args:
            kde_root: Path to KDE runtime root
        """
        self.kde_root = kde_root
        self.checkpoints: Dict[str, SessionCheckpoint] = {}
        self._violation_history: List[PrincipleViolation] = []
        self._session_authorized = False
    
    # =========================================================================
    # PRINCIPLE 1: No Auto-Continuation
    # =========================================================================
    
    def require_continuation_authorization(self, session_id: str) -> EnforcementResult:
        """
        Enforce Principle 1: No Auto-Continuation.
        
        Creates a checkpoint that MUST be authorized by a human before
        the next session/operation can begin.
        
        Args:
            session_id: Unique identifier for this session
            
        Returns:
            EnforcementResult with checkpoint status
        """
        checkpoint = SessionCheckpoint(
            checkpoint_id=session_id,
            description=f"Authorization required to continue session {session_id}",
            status=ContinuationStatus.REQUIRED,
            created_at=datetime.now()
        )
        self.checkpoints[session_id] = checkpoint
        
        return EnforcementResult(
            passed=False,
            violations=[
                PrincipleViolation(
                    principle=PrincipleType.NO_AUTO_CONTINUATION,
                    description="Human authorization required to continue",
                    severity="error",
                    blocked=True,
                    location=f"session:{session_id}",
                    suggestion='Say "proceed" or "authorized" to continue'
                )
            ],
            warnings=[],
            metadata={
                "checkpoint_id": session_id,
                "status": "REQUIRED"
            }
        )
    
    def authorize_continuation(
        self, 
        session_id: str, 
        authorized_by: str = "human"
    ) -> EnforcementResult:
        """
        Authorize continuation for a session checkpoint.
        
        Args:
            session_id: Checkpoint to authorize
            authorized_by: Who authorized (must be human)
            
        Returns:
            EnforcementResult with authorization status
        """
        if session_id not in self.checkpoints:
            return EnforcementResult(
                passed=False,
                violations=[
                    PrincipleViolation(
                        principle=PrincipleType.NO_AUTO_CONTINUATION,
                        description=f"Checkpoint {session_id} not found",
                        severity="error",
                        blocked=True
                    )
                ]
            )
        
        checkpoint = self.checkpoints[session_id]
        
        if authorized_by.lower() == "ai":
            return EnforcementResult(
                passed=False,
                violations=[
                    PrincipleViolation(
                        principle=PrincipleType.NO_AUTO_CONTINUATION,
                        description="AI cannot authorize its own continuation",
                        severity="error",
                        blocked=True,
                        suggestion="Only human authorization is permitted"
                    )
                ]
            )
        
        checkpoint.status = ContinuationStatus.AUTHORIZED
        checkpoint.authorized_by = authorized_by
        checkpoint.authorized_at = datetime.now()
        self._session_authorized = True
        
        return EnforcementResult(
            passed=True,
            metadata={
                "checkpoint_id": session_id,
                "authorized_by": authorized_by,
                "authorized_at": checkpoint.authorized_at.isoformat()
            }
        )
    
    def check_continuation_authorized(self, session_id: str) -> bool:
        """
        Check if continuation is authorized for a session.
        
        Args:
            session_id: Session to check
            
        Returns:
            True if authorized, False otherwise
        """
        if session_id not in self.checkpoints:
            return False
        
        checkpoint = self.checkpoints[session_id]
        return checkpoint.status == ContinuationStatus.AUTHORIZED
    
    # =========================================================================
    # PRINCIPLE 2: No Self-Approval
    # =========================================================================
    
    def check_state_transition(
        self,
        current_state: str,
        new_state: str,
        actor: str = "unknown"
    ) -> EnforcementResult:
        """
        Enforce Principle 2: No Self-Approval.
        
        Blocks AI from approving its own work by transitioning
        from REVIEW to APPROVED state.
        
        Args:
            current_state: Current document state
            new_state: Desired new state
            actor: Who is making the transition
            
        Returns:
            EnforcementResult indicating if transition is allowed
        """
        violations = []
        
        # Check for REVIEW -> APPROVED transition
        is_review_to_approved = (
            current_state.upper() in self.REVIEW_STATES and
            new_state.upper() in self.APPROVED_STATES
        )
        
        if is_review_to_approved:
            actor_lower = actor.lower()
            
            # Check if AI is trying to approve its own work
            is_ai_actor = "ai" in actor_lower or "agent" in actor_lower or actor_lower in ["openhands", "gpt", "claude", "unknown"]
            
            if is_ai_actor or actor == "unknown":
                violations.append(PrincipleViolation(
                    principle=PrincipleType.NO_SELF_APPROVAL,
                    description=f"AI cannot approve its own work: {current_state} → {new_state}",
                    severity="error",
                    blocked=True,
                    suggestion="Human review and approval required"
                ))
                
                return EnforcementResult(
                    passed=False,
                    violations=violations,
                    metadata={
                        "transition": f"{current_state} → {new_state}",
                        "actor": actor,
                        "blocked_reason": "self_approval"
                    }
                )
        
        return EnforcementResult(passed=True, metadata={
            "transition": f"{current_state} → {new_state}",
            "actor": actor,
            "approved": True
        })
    
    # =========================================================================
    # PRINCIPLE 3: No Self-Promotion
    # =========================================================================
    
    def check_promotion_transition(
        self,
        current_state: str,
        new_state: str,
        destination: str = "knowledge",
        actor: str = "unknown"
    ) -> EnforcementResult:
        """
        Enforce Principle 3: No Self-Promotion.
        
        Blocks AI from promoting knowledge to production (VALIDATED → PROMOTED).
        
        Args:
            current_state: Current document state
            new_state: Desired new state
            destination: Where it's being promoted to (e.g., "knowledge")
            actor: Who is making the promotion
            
        Returns:
            EnforcementResult indicating if promotion is allowed
        """
        violations = []
        
        # Check for VALIDATED -> PROMOTED transition
        is_validated_to_promoted = (
            current_state.upper() == "VALIDATED" and
            new_state.upper() in self.PROMOTED_STATES
        )
        
        if is_validated_to_promoted:
            actor_lower = actor.lower()
            
            # Check if AI is trying to promote
            is_ai_actor = "ai" in actor_lower or "agent" in actor_lower or actor_lower in ["openhands", "gpt", "claude", "unknown"]
            
            if is_ai_actor or actor == "unknown":
                violations.append(PrincipleViolation(
                    principle=PrincipleType.NO_SELF_PROMOTION,
                    description=f"AI cannot promote knowledge to {destination}: {current_state} → {new_state}",
                    severity="error",
                    blocked=True,
                    suggestion="Human decision required for promotion"
                ))
                
                return EnforcementResult(
                    passed=False,
                    violations=violations,
                    metadata={
                        "transition": f"{current_state} → {new_state}",
                        "destination": destination,
                        "actor": actor,
                        "blocked_reason": "self_promotion"
                    }
                )
        
        return EnforcementResult(passed=True, metadata={
            "transition": f"{current_state} → {new_state}",
            "destination": destination,
            "actor": actor,
            "approved": True
        })
    
    # =========================================================================
    # PRINCIPLE 4: Distinguish Evidence
    # =========================================================================
    
    def classify_content(self, content: str) -> Dict[str, Any]:
        """
        Classify content by evidence level.
        
        Identifies which parts of content are:
        - Evidence: Documented facts with citations
        - Inference: Conclusions drawn from evidence
        - Hypothesis: Speculation beyond evidence
        
        Args:
            content: Content to classify
            
        Returns:
            Classification results with markers
        """
        lines = content.split('\n')
        classified = []
        stats = {"evidence": 0, "inference": 0, "hypothesis": 0, "unmarked": 0}
        
        for i, line in enumerate(lines):
            level = self._classify_line(line)
            classified.append({
                "line_number": i + 1,
                "level": level.value,
                "content": line.strip()
            })
            stats[level.value] += 1
        
        unmarked_count = stats["evidence"] + stats["inference"] + stats["hypothesis"]
        stats["unmarked"] = len(lines) - unmarked_count
        
        return {
            "classified_lines": classified,
            "stats": stats,
            "properly_distinguished": stats["evidence"] > 0 or stats["unmarked"] == 0
        }
    
    def _classify_line(self, line: str) -> EvidenceLevel:
        """Classify a single line by evidence pattern."""
        stripped = line.strip()
        
        # Check for explicit markers
        if re.search(r'\[EVIDENCE:', stripped, re.IGNORECASE):
            return EvidenceLevel.EVIDENCE
        if re.search(r'\[INFERENCE:', stripped, re.IGNORECASE):
            return EvidenceLevel.INFERENCE
        if re.search(r'\[HYPOTHESIS:', stripped, re.IGNORECASE):
            return EvidenceLevel.HYPOTHESIS
        
        # Check for evidence patterns
        for pattern in self.EVIDENCE_PATTERNS:
            if re.search(pattern, stripped, re.IGNORECASE):
                return EvidenceLevel.EVIDENCE
        
        for pattern in self.INFERENCE_PATTERNS:
            if re.search(pattern, stripped, re.IGNORECASE):
                return EvidenceLevel.INFERENCE
        
        for pattern in self.HYPOTHESIS_PATTERNS:
            if re.search(pattern, stripped, re.IGNORECASE):
                return EvidenceLevel.HYPOTHESIS
        
        # Lines that look like facts/definitions without markers are treated as unmarked
        if stripped.startswith('#') or stripped.startswith('-') or stripped.startswith('*'):
            return EvidenceLevel.EVIDENCE  # Headers and list items default to evidence
        
        return EvidenceLevel.HYPOTHESIS  # Unmarked content treated as hypothesis
    
    def check_evidence_distinction(self, content: str) -> EnforcementResult:
        """
        Enforce Principle 4: Distinguish Evidence, Inference, and Hypothesis.
        
        Args:
            content: Content to check
            
        Returns:
            EnforcementResult with classification and warnings
        """
        classification = self.classify_content(content)
        violations = []
        warnings = []
        
        stats = classification["stats"]
        total = sum(stats.values())
        
        # Check if content has properly marked sections
        has_marked_content = stats["evidence"] > 0 or stats["inference"] > 0 or stats["hypothesis"] > 0
        
        # Calculate marking ratio
        marking_ratio = (total - stats["unmarked"]) / total if total > 0 else 0
        
        if has_marked_content and marking_ratio < 0.3:
            warnings.append(
                f"Only {marking_ratio:.0%} of content is properly marked for evidence level. "
                "Consider marking evidence, inferences, and hypotheses explicitly."
            )
        
        # Check for mixed unmarked content that should be marked
        unmarked_analysis = [
            c for c in classification["classified_lines"]
            if c["level"] == "hypothesis" and len(c["content"]) > 100
        ]
        
        if len(unmarked_analysis) > 5:
            violations.append(PrincipleViolation(
                principle=PrincipleType.DISTINGUISH_EVIDENCE,
                description=f"{len(unmarked_analysis)} long unmarked sections found. Mark evidence vs inference vs hypothesis.",
                severity="warning",
                blocked=False,
                suggestion="Use [EVIDENCE:], [INFERENCE:], or [HYPOTHESIS:] markers"
            ))
        
        return EnforcementResult(
            passed=len([v for v in violations if v.severity == "error"]) == 0,
            violations=violations,
            warnings=warnings,
            metadata=classification
        )
    
    # =========================================================================
    # PRINCIPLE 5: Evidence-Based Changes
    # =========================================================================
    
    def check_claims(self, content: str) -> EnforcementResult:
        """
        Enforce Principle 5: Evidence-Based Changes.
        
        Checks that all claims are justified by evidence.
        
        Args:
            content: Content to check for claims
            
        Returns:
            EnforcementResult with claim analysis
        """
        violations = []
        
        # Find claim patterns (assertions, recommendations, statements)
        claim_patterns = [
            r'^(should|must|will|needs to|requires)',
            r'^(we recommend|it is recommended|the solution is)',
            r'(therefore|consequently|thus)',
        ]
        
        lines = content.split('\n')
        claims_without_evidence = []
        
        for i, line in enumerate(lines):
            stripped = line.strip()
            if not stripped or stripped.startswith('#'):
                continue
            
            for pattern in claim_patterns:
                if re.search(pattern, stripped, re.IGNORECASE):
                    # Check if this claim has evidence citation nearby
                    has_evidence = self._has_evidence_nearby(lines, i)
                    
                    if not has_evidence:
                        claims_without_evidence.append({
                            "line_number": i + 1,
                            "content": stripped[:100],
                            "needs_evidence": True
                        })
                    break
        
        # Flag claims without evidence
        if len(claims_without_evidence) > 3:
            violations.append(PrincipleViolation(
                principle=PrincipleType.EVIDENCE_BASED_CHANGES,
                description=f"{len(claims_without_evidence)} claims found without supporting evidence citations",
                severity="warning",
                blocked=False,
                suggestion="Cite evidence for each claim (e.g., [EVIDENCE: source citation])"
            ))
        
        return EnforcementResult(
            passed=True,  # Only warnings, not blocking
            violations=violations,
            metadata={
                "total_claims": len(claims_without_evidence),
                "unjustified_claims": [c["content"][:50] for c in claims_without_evidence[:3]]
            }
        )
    
    def _has_evidence_nearby(self, lines: List[str], line_index: int, window: int = 3) -> bool:
        """Check if there's evidence citation within window of lines."""
        start = max(0, line_index - window)
        end = min(len(lines), line_index + window + 1)
        
        for i in range(start, end):
            if re.search(r'\[EVIDENCE:', lines[i], re.IGNORECASE):
                return True
            if re.search(r'Source:\s*https?://', lines[i], re.IGNORECASE):
                return True
        
        return False
    
    # =========================================================================
    # COMPREHENSIVE ENFORCEMENT
    # =========================================================================
    
    def enforce_all(self, context: Dict[str, Any]) -> EnforcementResult:
        """
        Run all principle enforcement checks.
        
        Args:
            context: Context containing:
                - content: Content to check
                - current_state: Current document state
                - new_state: Desired state transition
                - actor: Who is performing actions
                - session_id: Current session ID
                
        Returns:
            Combined EnforcementResult
        """
        all_violations = []
        all_warnings = []
        passed = True
        
        content = context.get("content", "")
        current_state = context.get("current_state", "")
        new_state = context.get("new_state", "")
        actor = context.get("actor", "unknown")
        session_id = context.get("session_id", "default")
        
        # Check Principle 1: No Auto-Continuation
        if not self.check_continuation_authorized(session_id):
            all_violations.append(PrincipleViolation(
                principle=PrincipleType.NO_AUTO_CONTINUATION,
                description="Session continuation not authorized",
                severity="error",
                blocked=True
            ))
            passed = False
        
        # Check Principle 2: No Self-Approval
        if current_state and new_state:
            approval_result = self.check_state_transition(current_state, new_state, actor)
            all_violations.extend(approval_result.violations)
            if not approval_result.passed:
                passed = False
        
        # Check Principle 3: No Self-Promotion
        if current_state and new_state:
            promotion_result = self.check_promotion_transition(current_state, new_state, actor=actor)
            all_violations.extend(promotion_result.violations)
            if not promotion_result.passed:
                passed = False
        
        # Check Principle 4: Distinguish Evidence
        if content:
            evidence_result = self.check_evidence_distinction(content)
            all_violations.extend(evidence_result.violations)
            all_warnings.extend(evidence_result.warnings)
        
        # Check Principle 5: Evidence-Based Changes
        if content:
            claims_result = self.check_claims(content)
            all_violations.extend(claims_result.violations)
            all_warnings.extend(claims_result.warnings)
        
        return EnforcementResult(
            passed=passed,
            violations=all_violations,
            warnings=all_warnings,
            metadata={
                "session_id": session_id,
                "checks_performed": 5,
                "blocking_violations": len([v for v in all_violations if v.blocked])
            }
        )
    
    # =========================================================================
    # EXPERIMENT AUTHORIZATION (INV-RUNTIME-GAPS Mitigation)
    # =========================================================================
    
    # States that require human authorization for experiments
    EXPERIMENT_APPROVED_STATES = {"COMPLETE", "APPROVED", "PROMOTED"}
    
    def check_experiment_transition(
        self,
        experiment_id: str,
        from_state: str,
        to_state: str,
        actor: str
    ) -> EnforcementResult:
        """
        Validate experiment state transitions.
        
        Enforces that experiments can only reach COMPLETE/APPROVED/PROMOTED
        states with human authorization, not AI self-completion.
        
        This is the 6th Principle (extends SEED-001):
        "No Auto-Completion - Never mark experiments complete without human review"
        
        Args:
            experiment_id: The experiment being transitioned
            from_state: Current experiment state
            to_state: Desired new state
            actor: Who is performing the transition ('human', 'ai', 'system')
            
        Returns:
            EnforcementResult with authorization status
        """
        violations = []
        warnings = []
        blocked = False
        
        # Check if this transition requires human authorization
        requires_human = to_state in self.EXPERIMENT_APPROVED_STATES
        
        if requires_human:
            if actor.lower() != "human":
                violations.append(PrincipleViolation(
                    principle=PrincipleType.NO_AUTO_CONTINUATION,  # Reuse principle
                    description=f"Experiment '{experiment_id}' requires HUMAN authorization to transition from {from_state} to {to_state}",
                    severity="error",
                    blocked=True,
                    location=f"experiment:{experiment_id}",
                    suggestion='Say "authorized" or have human approve the transition'
                ))
                blocked = True
            else:
                warnings.append(f"Human authorization confirmed for {experiment_id} {from_state}→{to_state}")
        
        return EnforcementResult(
            passed=not blocked,
            violations=violations,
            warnings=warnings,
            metadata={
                "experiment_id": experiment_id,
                "from_state": from_state,
                "to_state": to_state,
                "actor": actor,
                "requires_human": requires_human
            }
        )
    
    def require_experiment_checkpoint(
        self,
        experiment_id: str,
        checkpoint_type: str = "completion"
    ) -> EnforcementResult:
        """
        Require a checkpoint before experiment can proceed.
        
        Creates a checkpoint that must be authorized by human before
        the experiment can be marked complete or approved.
        
        Args:
            experiment_id: The experiment requiring checkpoint
            checkpoint_type: Type of checkpoint ('completion', 'approval', 'promotion')
            
        Returns:
            EnforcementResult with checkpoint requirement
        """
        checkpoint_id = f"EXP-{experiment_id}-{checkpoint_type}"
        
        checkpoint = SessionCheckpoint(
            checkpoint_id=checkpoint_id,
            description=f"Experiment '{experiment_id}' requires human authorization for {checkpoint_type}",
            status=ContinuationStatus.REQUIRED,
            created_at=datetime.now()
        )
        self.checkpoints[checkpoint_id] = checkpoint
        
        return EnforcementResult(
            passed=False,
            violations=[
                PrincipleViolation(
                    principle=PrincipleType.NO_AUTO_CONTINUATION,
                    description=f"Checkpoint required for experiment '{experiment_id}' {checkpoint_type}",
                    severity="error",
                    blocked=True,
                    location=f"experiment:{experiment_id}",
                    suggestion=f"Human must authorize checkpoint '{checkpoint_id}' to proceed"
                )
            ],
            warnings=[],
            metadata={
                "checkpoint_id": checkpoint_id,
                "experiment_id": experiment_id,
                "checkpoint_type": checkpoint_type,
                "status": "REQUIRED"
            }
        )
    
    def authorize_experiment_checkpoint(
        self,
        checkpoint_id: str,
        authorized_by: str
    ) -> EnforcementResult:
        """
        Authorize an experiment checkpoint.
        
        Args:
            checkpoint_id: The checkpoint to authorize
            authorized_by: Who is authorizing (must be 'human')
            
        Returns:
            EnforcementResult with authorization status
        """
        if checkpoint_id not in self.checkpoints:
            return EnforcementResult(
                passed=False,
                violations=[
                    PrincipleViolation(
                        principle=PrincipleType.NO_AUTO_CONTINUATION,
                        description=f"Checkpoint '{checkpoint_id}' not found",
                        severity="error",
                        blocked=True
                    )
                ]
            )
        
        if authorized_by.lower() != "human":
            return EnforcementResult(
                passed=False,
                violations=[
                    PrincipleViolation(
                        principle=PrincipleType.NO_AUTO_CONTINUATION,
                        description="Only human authorization is permitted for experiment checkpoints",
                        severity="error",
                        blocked=True
                    )
                ]
            )
        
        checkpoint = self.checkpoints[checkpoint_id]
        checkpoint.status = ContinuationStatus.AUTHORIZED
        checkpoint.authorized_by = authorized_by
        checkpoint.authorized_at = datetime.now()
        
        return EnforcementResult(
            passed=True,
            violations=[],
            warnings=[f"Checkpoint '{checkpoint_id}' authorized by {authorized_by}"],
            metadata={
                "checkpoint_id": checkpoint_id,
                "authorized_by": authorized_by,
                "authorized_at": checkpoint.authorized_at.isoformat()
            }
        )
    
    def get_enforcement_report(self) -> str:
        """Generate a human-readable enforcement report."""
        lines = []
        lines.append("=" * 78)
        lines.append("FIVE CORE PRINCIPLES - ENFORCEMENT REPORT")
        lines.append("=" * 78)
        lines.append("")
        
        # Session status
        authorized_count = sum(
            1 for cp in self.checkpoints.values() 
            if cp.status == ContinuationStatus.AUTHORIZED
        )
        blocked_count = len(self.checkpoints) - authorized_count
        
        lines.append("■ PRINCIPLE 1: No Auto-Continuation")
        lines.append("-" * 78)
        lines.append(f"  Checkpoints Created:    {len(self.checkpoints)}")
        lines.append(f"  Authorized:            {authorized_count}")
        lines.append(f"  Blocked:               {blocked_count}")
        lines.append("")
        
        # Violation history
        principle_violations = {p: [] for p in PrincipleType}
        for v in self._violation_history:
            principle_violations[v.principle].append(v)
        
        lines.append("■ VIOLATION HISTORY")
        lines.append("-" * 78)
        for principle, violations in principle_violations.items():
            lines.append(f"  {principle.value}: {len(violations)} violations")
        lines.append("")
        
        lines.append("=" * 78)
        
        return "\n".join(lines)


# =============================================================================
# PRINCIPLE ENFORCEMENT ERROR
# =============================================================================

class PrincipleViolationError(Exception):
    """Raised when a principle violation is detected and blocked."""
    
    def __init__(self, result: EnforcementResult):
        self.result = result
        message = "PRINCIPLE VIOLATION DETECTED\n\n"
        
        for v in result.violations:
            if v.blocked:
                message += f"🔴 [{v.principle.value}] BLOCKED: {v.description}\n"
                if v.suggestion:
                    message += f"   → {v.suggestion}\n"
            else:
                message += f"⚠️  [{v.principle.value}] WARNING: {v.description}\n"
        
        super().__init__(message)


# =============================================================================
# RUNTIME INTEGRATION
# =============================================================================

def create_enforcer(kde_root: str = "/workspace/project/kde") -> FivePrinciplesEnforcer:
    """
    Create and initialize a Five Principles Enforcer.
    
    Args:
        kde_root: Path to KDE runtime root
        
    Returns:
        Initialized FivePrinciplesEnforcer
    """
    return FivePrinciplesEnforcer(kde_root)


# =============================================================================
# DEMONSTRATION
# =============================================================================

def demonstrate_enforcement():
    """Demonstrate the Five Principles Enforcement."""
    print("=" * 78)
    print("FIVE CORE PRINCIPLES - ENFORCEMENT DEMONSTRATION")
    print("=" * 78)
    print()
    
    enforcer = FivePrinciplesEnforcer()
    
    # Demo 1: Auto-Continuation Check
    print("■ DEMO 1: No Auto-Continuation Enforcement")
    print("-" * 78)
    session_id = "INV-DEMO-001"
    
    result = enforcer.require_continuation_authorization(session_id)
    print(f"  Session created: {session_id}")
    print(f"  Status: BLOCKED (requires human authorization)")
    print()
    
    # Try AI authorization (should fail)
    ai_result = enforcer.authorize_continuation(session_id, "ai")
    print(f"  AI attempts authorization: {'BLOCKED' if not ai_result.passed else 'ALLOWED'}")
    
    # Human authorization (should succeed)
    human_result = enforcer.authorize_continuation(session_id, "human")
    print(f"  Human authorization: {'AUTHORIZED' if human_result.passed else 'BLOCKED'}")
    print()
    
    # Demo 2: Self-Approval Check
    print("■ DEMO 2: No Self-Approval Enforcement")
    print("-" * 78)
    
    # AI trying to approve (should fail)
    ai_approval = enforcer.check_state_transition("REVIEW", "APPROVED", "ai")
    print(f"  AI attempts REVIEW→APPROVED: {'BLOCKED' if not ai_approval.passed else 'ALLOWED'}")
    if not ai_approval.passed:
        for v in ai_approval.violations:
            print(f"    → {v.description}")
    
    # Human approving (should succeed)
    human_approval = enforcer.check_state_transition("REVIEW", "APPROVED", "human")
    print(f"  Human attempts REVIEW→APPROVED: {'ALLOWED' if human_approval.passed else 'BLOCKED'}")
    print()
    
    # Demo 3: Self-Promotion Check
    print("■ DEMO 3: No Self-Promotion Enforcement")
    print("-" * 78)
    
    # AI trying to promote (should fail)
    ai_promote = enforcer.check_promotion_transition("VALIDATED", "PROMOTED", "knowledge", "ai")
    print(f"  AI attempts VALIDATED→PROMOTED: {'BLOCKED' if not ai_promote.passed else 'ALLOWED'}")
    if not ai_promote.passed:
        for v in ai_promote.violations:
            print(f"    → {v.description}")
    
    # Human promoting (should succeed)
    human_promote = enforcer.check_promotion_transition("VALIDATED", "PROMOTED", "knowledge", "human")
    print(f"  Human attempts VALIDATED→PROMOTED: {'ALLOWED' if human_promote.passed else 'BLOCKED'}")
    print()
    
    # Demo 4: Evidence Distinction
    print("■ DEMO 4: Distinguish Evidence Enforcement")
    print("-" * 78)
    
    test_content = """
# Investigation Report

## Evidence
[EVIDENCE: According to the KDE governance documents...]

## Analysis
[INFERENCE: Based on the evidence, the recommended approach is...]

## Recommendation
This implementation should proceed immediately.
    """
    
    evidence_result = enforcer.check_evidence_distinction(test_content)
    print(f"  Content analyzed: 4 lines")
    print(f"  Properly marked: {evidence_result.metadata.get('stats', {}).get('evidence', 0)} evidence, "
          f"{evidence_result.metadata.get('stats', {}).get('inference', 0)} inference")
    
    if evidence_result.warnings:
        print(f"  Warnings: {len(evidence_result.warnings)}")
        for w in evidence_result.warnings:
            print(f"    → {w[:60]}...")
    print()
    
    # Demo 5: Evidence-Based Claims
    print("■ DEMO 5: Evidence-Based Changes Enforcement")
    print("-" * 78)
    
    claim_content = """
Based on our analysis, we recommend the following changes.
The solution should use microservices architecture.
Therefore, we need to update the governance documents.
    """
    
    claims_result = enforcer.check_claims(claim_content)
    print(f"  Claims analyzed: {claims_result.metadata.get('total_claims', 0)}")
    print(f"  Status: {'PASSED' if claims_result.passed else 'ISSUES FOUND'}")
    print()
    
    # Final Report
    print(enforcer.get_enforcement_report())


if __name__ == "__main__":
    demonstrate_enforcement()
