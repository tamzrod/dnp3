"""
KDE Runtime Skill Installation

This module installs all promoted KDE Skills into the Runtime.

Usage:
    from runtime.install import install_skills
    
    log = install_skills()
    print(f"Installed {log.installed_skills} skills")
"""

from .manager import (
    SkillManifest,
    SkillInstallationManager,
    SkillInstallRecord,
    InstallationLog,
    InstallStatus
)

__all__ = [
    "SkillManifest",
    "SkillInstallationManager",
    "SkillInstallRecord",
    "InstallationLog",
    "InstallStatus",
    "install_skills"
]


def install_skills():
    """
    Install all promoted KDE Skills.
    
    Returns:
        InstallationLog with installation results
    """
    manager = SkillInstallationManager()
    return manager.install()
