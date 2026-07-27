"""
Alias Registry Module

Provides alias management functionality for KDE Runtime.
Supports canonical commands, categorized aliases, and resolution.
"""

from .registry import (
    AliasCategory,
    AliasEntry,
    AliasRegistry,
    get_registry,
    reset_registry
)

from .discovery import (
    DiscoveryAPI,
    DiscoveryEndpoint,
    DiscoveryResponse,
    create_api
)

from .resolver import (
    CommandType,
    ResolvedCommand,
    CommandParser,
    RuntimeCommandHandler,
    create_parser,
    create_handler
)

from .governance import (
    ApprovalStatus,
    ApprovalRequest,
    DeprecationRecord,
    AuditEntry,
    AliasGovernance,
    create_governance
)

__all__ = [
    # Registry
    'AliasCategory',
    'AliasEntry',
    'AliasRegistry',
    'get_registry',
    'reset_registry',
    # Discovery
    'DiscoveryAPI',
    'DiscoveryEndpoint',
    'DiscoveryResponse',
    'create_api',
    # Resolver
    'CommandType',
    'ResolvedCommand',
    'CommandParser',
    'RuntimeCommandHandler',
    'create_parser',
    'create_handler',
    # Governance
    'ApprovalStatus',
    'ApprovalRequest',
    'DeprecationRecord',
    'AuditEntry',
    'AliasGovernance',
    'create_governance',
]

__version__ = "1.0.0"
