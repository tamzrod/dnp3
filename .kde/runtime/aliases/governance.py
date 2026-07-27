"""
Alias Governance Module

Provides governance integration for alias management including
approval workflows, deprecation tracking, and audit logging.
"""

import json
import os
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from enum import Enum

from .registry import (
    AliasCategory,
    AliasEntry,
    AliasRegistry,
    get_registry
)


class ApprovalStatus(Enum):
    """Approval workflow statuses."""
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    DEPRECATED = "deprecated"


@dataclass
class ApprovalRequest:
    """Represents an alias approval request."""
    alias: str
    canonical: str
    category: AliasCategory
    description: str
    namespace: str
    requester: str = "human"
    requested_at: str = ""
    status: ApprovalStatus = ApprovalStatus.PENDING
    approved_by: Optional[str] = None
    approved_at: Optional[str] = None
    rationale: str = ""
    rejection_reason: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            'alias': self.alias,
            'canonical': self.canonical,
            'category': self.category.name.lower(),
            'description': self.description,
            'namespace': self.namespace,
            'requester': self.requester,
            'requested_at': self.requested_at,
            'status': self.status.value,
            'approved_by': self.approved_by,
            'approved_at': self.approved_at,
            'rationale': self.rationale,
            'rejection_reason': self.rejection_reason
        }


@dataclass
class DeprecationRecord:
    """Represents a deprecation record."""
    alias: str
    canonical: str
    deprecated_at: str
    removal_date: str
    reason: str
    migration_guide: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            'alias': self.alias,
            'canonical': self.canonical,
            'deprecated_at': self.deprecated_at,
            'removal_date': self.removal_date,
            'reason': self.reason,
            'migration_guide': self.migration_guide
        }


@dataclass
class AuditEntry:
    """Represents an audit log entry."""
    timestamp: str
    action: str
    alias: str
    actor: str
    details: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            'timestamp': self.timestamp,
            'action': self.action,
            'alias': self.alias,
            'actor': self.actor,
            'details': self.details
        }


class AliasGovernance:
    """
    Alias Governance Manager.
    
    Handles approval workflows, deprecation tracking, and audit logging
    for the alias registry.
    """
    
    DEPRECATION_PERIOD_DAYS = 365  # 12 months
    
    def __init__(
        self,
        registry: Optional[AliasRegistry] = None,
        audit_log_path: Optional[str] = None
    ):
        """
        Initialize Alias Governance.
        
        Args:
            registry: AliasRegistry instance
            audit_log_path: Path to audit log file
        """
        self.registry = registry or get_registry()
        
        if audit_log_path is None:
            audit_log_path = os.path.join(
                os.path.dirname(__file__),
                'audit.log'
            )
        self.audit_log_path = audit_log_path
        self._audit_entries: List[AuditEntry] = []
        self._deprecation_records: Dict[str, DeprecationRecord] = {}
    
    def submit_approval_request(
        self,
        alias: str,
        canonical: str,
        category: AliasCategory,
        description: str,
        namespace: str,
        requester: str = "human",
        rationale: str = ""
    ) -> ApprovalRequest:
        """
        Submit an alias for approval.
        
        Args:
            alias: Alias name
            canonical: Canonical command
            category: Alias category
            description: Alias description
            namespace: Namespace
            requester: Who submitted the request
            rationale: Justification for the alias
            
        Returns:
            ApprovalRequest
        """
        request = ApprovalRequest(
            alias=alias,
            canonical=canonical,
            category=category,
            description=description,
            namespace=namespace,
            requester=requester,
            requested_at=datetime.now().isoformat(),
            status=ApprovalStatus.PENDING,
            rationale=rationale
        )
        
        # Log the submission
        self._log_audit(
            action="approval_submitted",
            alias=alias,
            actor=requester,
            details=request.to_dict()
        )
        
        return request
    
    def approve_alias(
        self,
        request: ApprovalRequest,
        approver: str = "human"
    ) -> bool:
        """
        Approve an alias.
        
        Args:
            request: ApprovalRequest to approve
            approver: Who approved
            
        Returns:
            True if approved
        """
        request.status = ApprovalStatus.APPROVED
        request.approved_by = approver
        request.approved_at = datetime.now().isoformat()
        
        self._log_audit(
            action="alias_approved",
            alias=request.alias,
            actor=approver,
            details={'canonical': request.canonical}
        )
        
        return True
    
    def reject_alias(
        self,
        request: ApprovalRequest,
        rejector: str = "human",
        reason: str = ""
    ) -> bool:
        """
        Reject an alias.
        
        Args:
            request: ApprovalRequest to reject
            rejector: Who rejected
            reason: Rejection reason
            
        Returns:
            True if rejected
        """
        request.status = ApprovalStatus.REJECTED
        request.rejection_reason = reason
        
        self._log_audit(
            action="alias_rejected",
            alias=request.alias,
            actor=rejector,
            details={'reason': reason}
        )
        
        return True
    
    def deprecate_alias(
        self,
        alias: str,
        reason: str = "",
        migration_guide: str = ""
    ) -> Optional[DeprecationRecord]:
        """
        Deprecate an alias.
        
        Args:
            alias: Alias to deprecate
            reason: Deprecation reason
            migration_guide: How to migrate
            
        Returns:
            DeprecationRecord if successful
        """
        # Find the alias entry
        entry = self.registry.resolve(alias)
        if not entry:
            return None
        
        now = datetime.now()
        removal_date = now + timedelta(days=self.DEPRECATION_PERIOD_DAYS)
        
        record = DeprecationRecord(
            alias=alias,
            canonical=entry.canonical,
            deprecated_at=now.isoformat(),
            removal_date=removal_date.isoformat(),
            reason=reason,
            migration_guide=migration_guide
        )
        
        self._deprecation_records[alias.lower()] = record
        
        self._log_audit(
            action="alias_deprecated",
            alias=alias,
            actor="governance",
            details={
                'removal_date': record.removal_date,
                'reason': reason
            }
        )
        
        return record
    
    def get_deprecated_aliases(self) -> List[DeprecationRecord]:
        """Get all deprecated aliases."""
        return list(self._deprecation_records.values())
    
    def check_removal_due(self, alias: str) -> bool:
        """
        Check if an alias is due for removal.
        
        Args:
            alias: Alias to check
            
        Returns:
            True if removal is due
        """
        record = self._deprecation_records.get(alias.lower())
        if not record:
            return False
        
        removal_date = datetime.fromisoformat(record.removal_date)
        return datetime.now() >= removal_date
    
    def get_audit_log(
        self,
        limit: Optional[int] = None
    ) -> List[AuditEntry]:
        """
        Get audit log entries.
        
        Args:
            limit: Maximum entries to return
            
        Returns:
            List of AuditEntry
        """
        entries = self._audit_entries
        if limit:
            entries = entries[-limit:]
        return entries
    
    def _log_audit(
        self,
        action: str,
        alias: str,
        actor: str,
        details: Dict[str, Any] = None
    ) -> None:
        """Log an audit entry."""
        entry = AuditEntry(
            timestamp=datetime.now().isoformat(),
            action=action,
            alias=alias,
            actor=actor,
            details=details or {}
        )
        self._audit_entries.append(entry)
        
        # Also write to file
        try:
            with open(self.audit_log_path, 'a') as f:
                f.write(json.dumps(entry.to_dict()) + '\n')
        except Exception:
            pass  # Ignore file write errors
    
    def generate_compliance_report(self) -> Dict[str, Any]:
        """
        Generate a governance compliance report.
        
        Returns:
            Report dictionary
        """
        if not self.registry.is_loaded:
            self.registry.load()
        
        all_aliases = self.registry.get_all_aliases()
        
        # Count by category
        category_counts = {}
        for cat in AliasCategory:
            category_counts[cat.name.lower()] = len(
                [a for a in all_aliases if a.category == cat]
            )
        
        # Count deprecated
        deprecated = [a for a in all_aliases if a.deprecated]
        
        # Get pending approvals (from audit log)
        pending = [
            e for e in self._audit_entries
            if e.action == "approval_submitted"
        ]
        
        return {
            'generated_at': datetime.now().isoformat(),
            'total_aliases': len(all_aliases),
            'by_category': category_counts,
            'deprecated_count': len(deprecated),
            'pending_approvals': len(pending),
            'audit_entries': len(self._audit_entries),
            'compliance_status': 'COMPLIANT' if len(deprecated) == 0 else 'REQUIRES_ATTENTION'
        }


def create_governance() -> AliasGovernance:
    """Create an AliasGovernance instance."""
    return AliasGovernance()
