# AO-004 isolated OpenDNP3 3.1.2 G41 peer harness image (readiness scaffolding).
#
# PURPOSE: a single, self-contained lab container hosting the independently
# implemented OpenDNP3 3.1.2 G41 outstation harness. It starts NO protocol
# traffic by itself: compose.yaml runs it with network_mode: none and the default
# command is a readiness echo only.
#
# Build-time network is used ONLY to acquire the pinned, checksum-verified
# toolchain and peer sources via the existing (unmodified) lab/vec01/
# acquire-pinned.sh. The runtime container is offline.
#
# NOT_BUILT: this image has not been built in the authoring environment.
FROM debian:12-slim

ENV DEBIAN_FRONTEND=noninteractive

# iproute2 = namespaces/veth/routes; build-essential+cmake = peer build;
# ca-certificates+curl+unzip = pinned acquisition; coreutils = sha256sum.
# No Docker CLI and no daemon socket are installed.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      iproute2 iputils-ping build-essential cmake make pkg-config \
      ca-certificates curl unzip xz-utils coreutils procps \
 && rm -rf /var/lib/apt/lists/*

# Reuse the existing verified pinned acquisition script UNMODIFIED.
# (Build context is the repository root so this path resolves.)
COPY lab/vec01/acquire-pinned.sh /opt/vec01/acquire-pinned.sh
RUN chmod +x /opt/vec01/acquire-pinned.sh && /opt/vec01/acquire-pinned.sh

# Harness sources (allowlisted).
COPY lab/ao004/G41MemoryCommandHandler.h  /opt/ao004/
COPY lab/ao004/G41MemoryCommandHandler.cpp /opt/ao004/
COPY lab/ao004/g41_peer.cpp               /opt/ao004/
COPY lab/ao004/CMakeLists.txt             /opt/ao004/

# Compile the harness against the pinned peer (build time only).
RUN cmake -S /opt/ao004 -B /opt/ao004/build -DCMAKE_BUILD_TYPE=Release \
 && cmake --build /opt/ao004/build -j"$(nproc)" \
 && cp /opt/ao004/build/g41_peer /opt/ao004/g41_peer

ENV PATH=/opt/go/bin:$PATH
WORKDIR /opt/ao004

# Default: readiness echo only (no listener, no DNP3 traffic). The harness starts
# a listener only when explicitly invoked as `/opt/ao004/g41_peer --listen`.
ENTRYPOINT ["/opt/ao004/g41_peer"]
