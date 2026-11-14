# AO-004 isolated OpenDNP3 3.1.2 G41 peer + master harness image (compile-only preparation).
#
# PURPOSE: a single, self-contained lab container hosting the independently
# implemented OpenDNP3 3.1.2 G41 outstation harness (`g41_peer`) and master
# harness (`g41_master`). It starts NO protocol traffic by itself: compose.yaml
# runs it with network_mode: none and the default commands are readiness echoes
# only.
#
# Build-time network is used ONLY to acquire the pinned, checksum-verified
# toolchain and peer sources via the existing (unmodified) lab/vec01/
# acquire-pinned.sh. The runtime container is offline.
#
# NOT_BUILT: this image has not been built in the authoring environment.
#
# BASE_IMAGE_UNPINNED: the debian:12-slim tag is not globally immutable and the
# APT packages are not version-pinned, so full-image reproducibility is NOT
# proved by the sha256-pinned OpenDNP3 input alone. No owner-verified base-image
# digest is available, so no digest is invented here. This remains an unresolved
# limitation.
FROM debian:12-slim

ENV DEBIAN_FRONTEND=noninteractive

# iproute2 = namespaces/veth/routes; build-essential+cmake = peer build;
# ca-certificates+curl+unzip = pinned acquisition; coreutils = sha256sum.
# tcpdump = AO-005 isolated raw capture (lab-only; see README/AO-005 evidence).
# python3-minimal = the unmodified lab/vec01/preflight.sh listener-bind proof.
# No Docker CLI and no daemon socket are installed.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      iproute2 iputils-ping build-essential cmake make pkg-config \
      ca-certificates curl unzip xz-utils coreutils procps tcpdump python3-minimal \
 && rm -rf /var/lib/apt/lists/*

# Reuse the existing verified pinned acquisition script UNMODIFIED. It sha256-
# verifies OpenDNP3 3.1.2 and the three FetchContent dependency archives and
# unpacks them under /opt/vec01/{src,deps}; the harness CMake binds to those
# verified extracted directories offline.
# (Build context is the repository root so these paths resolve.)
COPY lab/vec01/acquire-pinned.sh /opt/vec01/acquire-pinned.sh
RUN chmod +x /opt/vec01/acquire-pinned.sh && /opt/vec01/acquire-pinned.sh

# Approved topology/checking references, COPIED UNCHANGED from lab/vec01. They
# are NOT executed here and NOT wired into the default command: creating the
# private two-namespace veth topology is a later, explicitly authorized,
# isolated-runner-only step (see README). network_mode: none alone gives only
# loopback; it does not create 10.77.0.2.
COPY lab/vec01/lab-netns.sh /opt/vec01/lab-netns.sh
COPY lab/vec01/preflight.sh /opt/vec01/preflight.sh
RUN chmod +x /opt/vec01/lab-netns.sh /opt/vec01/preflight.sh

# AO-005 isolated raw-PCAP evidence runner (lab-only; invoked explicitly on an
# authorized isolated runner, never by the default command).
COPY lab/ao004/ao005-pcap-runner.sh /opt/ao004/ao005-pcap-runner.sh
RUN chmod +x /opt/ao004/ao005-pcap-runner.sh

# Harness sources (allowlisted).
COPY lab/ao004/G41MemoryCommandHandler.h  /opt/ao004/
COPY lab/ao004/G41MemoryCommandHandler.cpp /opt/ao004/
COPY lab/ao004/g41_peer.cpp               /opt/ao004/
COPY lab/ao004/G41MasterCommandCallback.h  /opt/ao004/
COPY lab/ao004/G41MasterCommandCallback.cpp /opt/ao004/
COPY lab/ao004/g41_master.cpp             /opt/ao004/
COPY lab/ao004/CMakeLists.txt             /opt/ao004/

# Compile the harness against the pinned peer (build time only). The pinned
# source/deps are already present, so CMake runs fully disconnected and fails
# closed if any verified directory is missing (no implicit download).
RUN cmake -S /opt/ao004 -B /opt/ao004/build -DCMAKE_BUILD_TYPE=Release \
      -DFETCHCONTENT_FULLY_DISCONNECTED=ON \
      -DOPENDNP3_ROOT=/opt/vec01/src/opendnp3-3.1.2 \
      -DDEPS_ROOT=/opt/vec01/deps \
 && cmake --build /opt/ao004/build -j"$(nproc)" \
 && cp /opt/ao004/build/g41_peer /opt/ao004/g41_peer \
 && cp /opt/ao004/build/g41_master /opt/ao004/g41_master

ENV PATH=/opt/go/bin:$PATH
WORKDIR /opt/ao004

# Default: readiness echo only (no listener, no DNP3 traffic). The peer harness
# starts a listener only when explicitly invoked as `/opt/ao004/g41_peer
# --listen`, and the master harness opens a channel only as `/opt/ao004/g41_master
# --run`, and only after a later, separately authorized isolated-runner setup
# builds the private namespace topology.
ENTRYPOINT ["/opt/ao004/g41_peer"]
