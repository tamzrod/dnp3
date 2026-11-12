# AO-004 isolated OpenDNP3 3.1.2 G41 peer harness — CMake build.
#
# Builds the memory-only G41 outstation harness against the sha256-pinned
# OpenDNP3 3.1.2 source installed under /opt/vec01 by lab/vec01/acquire-pinned.sh.
# NOT_BUILT: static scaffolding only; not compiled in the authoring environment.
cmake_minimum_required(VERSION 3.16)
project(ao004_g41_peer CXX)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Pinned OpenDNP3 3.1.2, installed by the (unmodified) lab/vec01/acquire-pinned.sh.
set(OPENDNP3_ROOT "/opt/vec01/src/opendnp3-3.1.2" CACHE PATH "Pinned OpenDNP3 3.1.2 source root")

add_subdirectory(${OPENDNP3_ROOT} ${CMAKE_BINARY_DIR}/opendnp3 EXCLUDE_FROM_ALL)

add_executable(g41_peer
    g41_peer.cpp
    G41MemoryCommandHandler.cpp
)
target_link_libraries(g41_peer PRIVATE opendnp3)
