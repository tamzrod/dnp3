#!/usr/bin/env bash
# Build the VEC-01 objective Gate B topology INSIDE this lab container:
# two private network namespaces joined by exactly one internal veth pair.
# No default route, no forwarding, no NAT, no host/public networking.
#
#   vec01-master     10.77.0.1/24   (native Go master)
#   vec01-outstation 10.77.0.2/24   (OpenDNP3 3.1.2 outstation, listener :20000)
set -euo pipefail

NS_MASTER="${NS_MASTER:-vec01-master}"
NS_OUTSTATION="${NS_OUTSTATION:-vec01-outstation}"
ADDR_MASTER="${ADDR_MASTER:-10.77.0.1/24}"
ADDR_OUTSTATION="${ADDR_OUTSTATION:-10.77.0.2/24}"

up() {
  ip netns add "$NS_MASTER"
  ip netns add "$NS_OUTSTATION"

  ip link add veth-m type veth peer name veth-o
  ip link set veth-m netns "$NS_MASTER"
  ip link set veth-o netns "$NS_OUTSTATION"

  ip -n "$NS_MASTER"    addr add "$ADDR_MASTER"    dev veth-m
  ip -n "$NS_OUTSTATION" addr add "$ADDR_OUTSTATION" dev veth-o
  ip -n "$NS_MASTER"    link set lo up
  ip -n "$NS_OUTSTATION" link set lo up
  ip -n "$NS_MASTER"    link set veth-m up
  ip -n "$NS_OUTSTATION" link set veth-o up

  # Fail closed: forwarding MUST be off in both namespaces. In a hardened
  # runtime /proc/sys may be mounted read-only; a container-local remount is
  # attempted, and the value is verified. Inability to force it off is fatal.
  mount -o remount,rw /proc/sys 2>/dev/null || true
  for ns in "$NS_MASTER" "$NS_OUTSTATION"; do
    ip netns exec "$ns" sysctl -qw net.ipv4.ip_forward=0 2>/dev/null || true
    if [[ "$(ip netns exec "$ns" cat /proc/sys/net/ipv4/ip_forward)" != "0" ]]; then
      echo "lab-netns: FATAL could not disable forwarding in $ns" >&2
      exit 1
    fi
  done
  echo "lab-netns: built ($NS_MASTER=$ADDR_MASTER, $NS_OUTSTATION=$ADDR_OUTSTATION)"
}

down() {
  ip netns del "$NS_MASTER" 2>/dev/null || true
  ip netns del "$NS_OUTSTATION" 2>/dev/null || true
  echo "lab-netns: torn down"
}

case "${1:-up}" in
  up)   up ;;
  down) down ;;
  *)    echo "usage: lab-netns.sh [up|down]" >&2; exit 2 ;;
esac
