#!/usr/bin/env bash
# VEC-01 lab container entrypoint.
#
# Default action (no arguments) is READINESS ONLY: it runs the objective Gate B
# preflight. It never starts DNP3 protocol traffic or issues a control by
# default. A later, separately authorized `operation cwal` may invoke specific
# protocol steps explicitly as `docker compose run --rm lab <command>`.
set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "vec01-lab: readiness mode (Gate B preflight; no DNP3 traffic)"
  exec /opt/vec01/preflight.sh
fi

exec "$@"
