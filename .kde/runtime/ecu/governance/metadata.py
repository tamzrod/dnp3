"""
Metadata Manager for Laboratory Governance

Manages artifact metadata generation and validation.
"""

import os
import re
from datetime import datetime
from typing import Dict, Optional
from dataclasses import dataclass, field, asdict


@dataclass
class ArtifactMetadata:
    """Standard metadata for laboratory artifacts."""
    id: str = ""
    type: str = ""
    title: str = ""
    status: str = "draft"
    authority: str = ""
    agent: str = ""
    bootstrap_version: str = ""
    runtime_version: str = ""
    created: str = ""
    started: str = ""
    completed: str = ""
    locked: str = ""
    modified: str = ""
    parent: Optional[str] = None
    related: list = field(default_factory=list)
    supersedes: Optional[str] = None
    superseded_by: Optional[str] = None
    reviewer: Optional[str] = None
    review_date: Optional[str] = None
    approval_status: str = "pending"
    
    def to_dict(self) -> Dict:
        return asdict(self)
    
    def to_frontmatter(self) -> str:
        """Generate YAML frontmatter."""
        lines = ["---"]
        for key, value in self.to_dict().items():
            if value is None or value == "":
                continue
            if isinstance(value, list):
                if value:
                    lines.append(f"{key}:")
                    for item in value:
                        lines.append(f"  - {item}")
            else:
                lines.append(f"{key}: {value}")
        lines.append("---")
        return '\n'.join(lines)
    
    @classmethod
    def from_dict(cls, data: Dict) -> 'ArtifactMetadata':
        """Create from dictionary."""
        return cls(**{k: v for k, v in data.items() if k in cls.__annotations__})
    
    @classmethod
    def from_frontmatter(cls, content: str) -> Optional['ArtifactMetadata']:
        """Parse metadata from document frontmatter."""
        if not content.startswith('---'):
            return None
        
        # Extract YAML between --- markers
        match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
        if not match:
            return None
        
        yaml_content = match.group(1)
        data = {}
        
        for line in yaml_content.split('\n'):
            if ':' in line:
                key, value = line.split(':', 1)
                key = key.strip()
                value = value.strip()
                if value:
                    data[key] = value
        
        return cls.from_dict(data)


class MetadataManager:
    """Manages artifact metadata generation and parsing."""
    
    def __init__(self, project_name: str, runtime_version: str, bootstrap_version: str, agent: str = "Runtime ECU"):
        self.project_name = project_name
        self.runtime_version = runtime_version
        self.bootstrap_version = bootstrap_version
        self.agent = agent
        self.authority = f"KDE Runtime ({project_name})"
    
    def generate_initial(self, artifact_id: str, artifact_type: str, title: str) -> ArtifactMetadata:
        """
        Generate initial metadata for new artifact.
        
        Args:
            artifact_id: Assigned artifact ID
            artifact_type: Classified artifact type
            title: Human-readable title
        
        Returns:
            ArtifactMetadata with initial values
        """
        now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
        
        return ArtifactMetadata(
            id=artifact_id,
            type=artifact_type,
            title=title,
            status="draft",
            authority=self.authority,
            agent=self.agent,
            bootstrap_version=self.bootstrap_version,
            runtime_version=self.runtime_version,
            created=now
        )
    
    def parse_artifact(self, artifact_path: str) -> Optional[ArtifactMetadata]:
        """
        Parse metadata from existing artifact.
        
        Args:
            artifact_path: Path to artifact file
        
        Returns:
            Parsed ArtifactMetadata or None
        """
        if not os.path.exists(artifact_path):
            return None
        
        with open(artifact_path, 'r') as f:
            content = f.read()
        
        return ArtifactMetadata.from_frontmatter(content)
    
    def update_status(self, metadata: ArtifactMetadata, new_status: str) -> ArtifactMetadata:
        """
        Update artifact status with timestamp.
        
        Args:
            metadata: Current metadata
            new_status: New status value
        
        Returns:
            Updated metadata
        """
        now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
        metadata.status = new_status
        metadata.modified = now
        
        # Set specific timestamps based on status
        if new_status == 'in_progress' and not metadata.started:
            metadata.started = now
        elif new_status == 'completed' and not metadata.completed:
            metadata.completed = now
        elif new_status == 'locked' and not metadata.locked:
            metadata.locked = now
        
        return metadata
    
    def update_completion(self, metadata: ArtifactMetadata) -> ArtifactMetadata:
        """
        Mark artifact as completed with auto-lock.
        
        Args:
            metadata: Current metadata
        
        Returns:
            Updated metadata with completed status
        """
        now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
        metadata.status = 'locked'
        metadata.completed = now
        metadata.locked = now
        metadata.modified = now
        return metadata
    
    def add_relationship(self, metadata: ArtifactMetadata, relationship: str, target_id: str) -> ArtifactMetadata:
        """
        Add relationship to metadata.
        
        Args:
            metadata: Current metadata
            relationship: Relationship type (parent, related, supersedes, superseded_by)
            target_id: Target artifact ID
        
        Returns:
            Updated metadata
        """
        if relationship == 'parent':
            metadata.parent = target_id
        elif relationship == 'supersedes':
            metadata.supersedes = target_id
        elif relationship == 'superseded_by':
            metadata.superseded_by = target_id
        elif relationship == 'related':
            if metadata.related is None:
                metadata.related = []
            if target_id not in metadata.related:
                metadata.related.append(target_id)
        
        metadata.modified = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
        return metadata
    
    def add_review(self, metadata: ArtifactMetadata, reviewer: str, approval_status: str) -> ArtifactMetadata:
        """
        Add review information.
        
        Args:
            metadata: Current metadata
            reviewer: Reviewer name
            approval_status: Approval status (approved, conditional, rejected)
        
        Returns:
            Updated metadata
        """
        now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
        metadata.reviewer = reviewer
        metadata.review_date = now
        metadata.approval_status = approval_status
        metadata.modified = now
        return metadata
