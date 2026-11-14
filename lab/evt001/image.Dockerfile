# Build context is not this Git directory. Lay out pinned sources beside a
# copy of this harness, then build from that context:
#   src/opendnp3-3.1.2/   deps/   harness/   derive_pcap.py   run-isolated.sh
# FetchContent downloads are disabled. Runtime is a separate --network none run.
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential cmake make pkg-config \
      tcpdump iproute2 iputils-ping libpcap0.8 procps python3 ca-certificates \
 && rm -rf /var/lib/apt/lists/*
COPY src/opendnp3-3.1.2 /opt/evt/src/opendnp3-3.1.2
COPY deps /opt/evt/deps
COPY harness /opt/evt/harness
COPY derive_pcap.py /opt/evt/derive_pcap.py
COPY run-isolated.sh /opt/evt/run-isolated.sh
RUN cmake -S /opt/evt/harness -B /opt/evt/harness/build -DCMAKE_BUILD_TYPE=Release \
 && cmake --build /opt/evt/harness/build -j"$(nproc)" \
 && cp /opt/evt/harness/build/evt_master /opt/evt/evt_master \
 && cp /opt/evt/harness/build/evt_outstation /opt/evt/evt_outstation \
 && chmod +x /opt/evt/run-isolated.sh /opt/evt/derive_pcap.py
ENTRYPOINT ["/opt/evt/run-isolated.sh"]
