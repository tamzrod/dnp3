# VEC-01 isolated DNP3 interoperability lab image (readiness scaffolding).
#
# PURPOSE: a single, self-contained lab container that can host the already
# authorized native-Go master <-> OpenDNP3 3.1.2 outstation test inside two
# private Linux network namespaces joined by one internal veth.
#
# It runs NO protocol traffic by itself. compose.yaml starts it with
# network_mode: none; the only network that exists is the private veth pair the
# entrypoint builds inside the container.
#
# Build-time network is used ONLY to acquire the pinned, checksum-verified
# toolchain and peer sources (acquire-pinned.sh). The runtime container is offline.
FROM debian:12-slim

# --- pinned inputs (single source of truth: acquire-pinned.sh) ---------------
ARG GO_VERSION=1.22.12
ARG GO_SHA256=4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43
ARG OPENDNP3_URL=https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2
ARG OPENDNP3_SHA256=183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2
ARG ASIO_URL=https://github.com/chriskohlhoff/asio/archive/asio-1-16-0.zip
ARG ASIO_SHA256=18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5
ARG EXE4CPP_URL=https://github.com/automatak/exe4cpp/archive/fb878a4de598ba9d6e4338afebf83f96e03af1b8.zip
ARG EXE4CPP_SHA256=2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c
ARG SER4CPP_URL=https://github.com/automatak/ser4cpp/archive/3c449734dc530a8f465eb0982de29165cc4e23d5.zip
ARG SER4CPP_SHA256=5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e

ENV DEBIAN_FRONTEND=noninteractive

# iproute2 = namespaces/veth/routes; tcpdump = capture; build-essential+cmake =
# peer build; ca-certificates+curl+unzip = pinned acquisition; coreutils =
# sha256sum. No Docker CLI and no daemon socket are installed.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      iproute2 tcpdump iputils-ping build-essential cmake make pkg-config \
      ca-certificates curl unzip xz-utils coreutils procps python3-minimal \
 && rm -rf /var/lib/apt/lists/*

COPY acquire-pinned.sh /opt/vec01/acquire-pinned.sh
RUN chmod +x /opt/vec01/acquire-pinned.sh && /opt/vec01/acquire-pinned.sh

COPY lab-netns.sh preflight.sh entrypoint.sh /opt/vec01/
RUN chmod +x /opt/vec01/*.sh

ENV PATH=/opt/go/bin:$PATH
WORKDIR /opt/vec01
ENTRYPOINT ["/opt/vec01/entrypoint.sh"]
