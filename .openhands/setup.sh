#!/usr/bin/env bash
# GOV-RECON-001: non-mutating OpenHands environment discovery.
# This hook may run BEFORE the agent reads CURRENT and acquires task authority.
# Never install, download, sudo, mutate runtime configuration, or run tests here.
set -u

printf '%s\n' 'OpenHands preflight (observation only; NOT execution authorization)'
if command -v python3 >/dev/null 2>&1; then
  python3 --version
else
  printf '%s\n' 'MISSING: python3'
fi
if python3 -c 'import yaml' >/dev/null 2>&1; then
  printf '%s\n' 'AVAILABLE: local PyYAML'
else
  printf '%s\n' 'MISSING: local PyYAML (no auto-install)'
fi
if command -v go >/dev/null 2>&1; then
  go version
else
  printf '%s\n' 'MISSING: Go on PATH (no auto-install)'
fi
printf '%s\n' 'Dependencies are NOT fetched. KDE startup and CWAL gates are NOT run by this hook.'
printf '%s\n' 'Executor must independently read AGENTS.md, workflow authority and CURRENT before acting.'
