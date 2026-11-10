#!/usr/bin/env bash
# Acquire and sha256-verify the pinned VEC-01 lab inputs at image-build time.
# Runtime is offline; all downloads happen here and are hash-gated.
set -euo pipefail

GO_VERSION="${GO_VERSION:-1.22.12}"
GO_SHA256="${GO_SHA256:-4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43}"
OPENDNP3_URL="${OPENDNP3_URL:-https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2}"
OPENDNP3_SHA256="${OPENDNP3_SHA256:-183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2}"
ASIO_URL="${ASIO_URL:-https://github.com/chriskohlhoff/asio/archive/asio-1-16-0.zip}"
ASIO_SHA256="${ASIO_SHA256:-18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5}"
EXE4CPP_URL="${EXE4CPP_URL:-https://github.com/automatak/exe4cpp/archive/fb878a4de598ba9d6e4338afebf83f96e03af1b8.zip}"
EXE4CPP_SHA256="${EXE4CPP_SHA256:-2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c}"
SER4CPP_URL="${SER4CPP_URL:-https://github.com/automatak/ser4cpp/archive/3c449734dc530a8f465eb0982de29165cc4e23d5.zip}"
SER4CPP_SHA256="${SER4CPP_SHA256:-5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e}"

fetch() { # url sha256 dest
  local url="$1" sha="$2" dest="$3"
  curl -fsSL -o "$dest" "$url"
  echo "${sha}  ${dest}" | sha256sum -c -
}

mkdir -p /opt/vec01/src /opt/vec01/deps

# --- Go toolchain (private install under /opt) ---
fetch "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" "$GO_SHA256" /tmp/go.tgz
tar -C /opt -xzf /tmp/go.tgz
rm -f /tmp/go.tgz

# --- OpenDNP3 3.1.2 source ---
fetch "$OPENDNP3_URL" "$OPENDNP3_SHA256" /tmp/opendnp3.tar.gz
mkdir -p /opt/vec01/src/opendnp3-3.1.2
tar -C /opt/vec01/src/opendnp3-3.1.2 --strip-components=1 -xzf /tmp/opendnp3.tar.gz
rm -f /tmp/opendnp3.tar.gz

# --- FetchContent dependencies ---
fetch "$ASIO_URL"   "$ASIO_SHA256"   /tmp/asio.zip
unzip -q /tmp/asio.zip -d /opt/vec01/deps
fetch "$EXE4CPP_URL" "$EXE4CPP_SHA256" /tmp/exe4cpp.zip
unzip -q /tmp/exe4cpp.zip -d /opt/vec01/deps
fetch "$SER4CPP_URL" "$SER4CPP_SHA256" /tmp/ser4cpp.zip
unzip -q /tmp/ser4cpp.zip -d /opt/vec01/deps
rm -f /tmp/*.zip

echo "acquire-pinned: OK (all inputs sha256-verified)"
