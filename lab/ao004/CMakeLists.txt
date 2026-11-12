# AO-004 isolated OpenDNP3 3.1.2 G41 peer harness — CMake build (offline).
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
set(DEPS_ROOT "/opt/vec01/deps" CACHE PATH "Pinned FetchContent dependency root (lab/vec01/acquire-pinned.sh)")

# The pinned source imports deps/{asio,exe4cpp,ser4cpp}.cmake via FetchContent.
# acquire-pinned.sh already sha256-verified and unpacked those archives into
# DEPS_ROOT, so bind FetchContent to the verified extracted directories and run
# fully disconnected. No download may occur; missing/mismatched sources fail
# closed below. Defaults follow the GitHub archive extraction convention
# (<repo>-<ref>); override the cache vars if the actual extracted names differ.
set(FETCHCONTENT_SOURCE_DIR_ASIO
    "${DEPS_ROOT}/asio-asio-1-16-0" CACHE PATH "Verified asio-1-16-0 source dir")
set(FETCHCONTENT_SOURCE_DIR_EXE4CPP
    "${DEPS_ROOT}/exe4cpp-fb878a4de598ba9d6e4338afebf83f96e03af1b8" CACHE PATH "Verified exe4cpp source dir")
set(FETCHCONTENT_SOURCE_DIR_SER4CPP
    "${DEPS_ROOT}/ser4cpp-3c449734dc530a8f465eb0982de29165cc4e23d5" CACHE PATH "Verified ser4cpp source dir")
set(FETCHCONTENT_FULLY_DISCONNECTED ON CACHE BOOL "Forbid any FetchContent download")

# Build only the static protocol library; no TLS/tests/java/fuzzing/examples so
# no additional (e.g. catch) dependency is pulled.
set(DNP3_TLS OFF CACHE BOOL "" FORCE)
set(DNP3_TESTS OFF CACHE BOOL "" FORCE)
set(DNP3_JAVA OFF CACHE BOOL "" FORCE)
set(DNP3_FUZZING OFF CACHE BOOL "" FORCE)
set(DNP3_EXAMPLES OFF CACHE BOOL "" FORCE)
set(DNP3_STATIC_LIBS ON CACHE BOOL "" FORCE)

# Fail closed: the pinned source and every verified dependency dir must exist.
foreach(required
        "${OPENDNP3_ROOT}"
        "${FETCHCONTENT_SOURCE_DIR_ASIO}"
        "${FETCHCONTENT_SOURCE_DIR_EXE4CPP}"
        "${FETCHCONTENT_SOURCE_DIR_SER4CPP}")
    if(NOT EXISTS "${required}")
        message(FATAL_ERROR
            "AO-004 harness: required pinned source '${required}' is missing. "
            "Run lab/vec01/acquire-pinned.sh (sha256-verified) on an authorized "
            "build runner, or set the matching -DFETCHCONTENT_SOURCE_DIR_* / "
            "-DOPENDNP3_ROOT cache variable. Refusing to download.")
    endif()
endforeach()

add_subdirectory(${OPENDNP3_ROOT} ${CMAKE_BINARY_DIR}/opendnp3 EXCLUDE_FROM_ALL)

add_executable(g41_peer
    g41_peer.cpp
    G41MemoryCommandHandler.cpp
)
target_link_libraries(g41_peer PRIVATE opendnp3)

# Independent OpenDNP3 3.1.2 G41 master harness (compile-only preparation). It
# shares the same pinned source/deps as the peer. Default invocation is offline
# readiness only; the live "--run" path is NOT authorized by this microtask.
add_executable(g41_master
    g41_master.cpp
    G41MasterCommandCallback.cpp
)
target_link_libraries(g41_master PRIVATE opendnp3)
