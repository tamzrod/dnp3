cmake_minimum_required(VERSION 3.11)
project(evt001_opendnp3_harness CXX)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(OPENDNP3_ROOT "/opt/evt/src/opendnp3-3.1.2" CACHE PATH "Pinned OpenDNP3 3.1.2 source")
set(DEPS_ROOT "/opt/evt/deps" CACHE PATH "Pinned FetchContent sources")

set(FETCHCONTENT_SOURCE_DIR_ASIO
    "${DEPS_ROOT}/asio-asio-1-16-0" CACHE PATH "asio-1-16-0")
set(FETCHCONTENT_SOURCE_DIR_EXE4CPP
    "${DEPS_ROOT}/exe4cpp-fb878a4de598ba9d6e4338afebf83f96e03af1b8" CACHE PATH "exe4cpp")
set(FETCHCONTENT_SOURCE_DIR_SER4CPP
    "${DEPS_ROOT}/ser4cpp-3c449734dc530a8f465eb0982de29165cc4e23d5" CACHE PATH "ser4cpp")
set(FETCHCONTENT_FULLY_DISCONNECTED ON CACHE BOOL "Forbid FetchContent downloads")

set(DNP3_TLS OFF CACHE BOOL "" FORCE)
set(DNP3_TESTS OFF CACHE BOOL "" FORCE)
set(DNP3_JAVA OFF CACHE BOOL "" FORCE)
set(DNP3_FUZZING OFF CACHE BOOL "" FORCE)
set(DNP3_EXAMPLES OFF CACHE BOOL "" FORCE)
set(DNP3_STATIC_LIBS ON CACHE BOOL "" FORCE)

foreach(required
        "${OPENDNP3_ROOT}"
        "${FETCHCONTENT_SOURCE_DIR_ASIO}"
        "${FETCHCONTENT_SOURCE_DIR_EXE4CPP}"
        "${FETCHCONTENT_SOURCE_DIR_SER4CPP}")
    if(NOT EXISTS "${required}")
        message(FATAL_ERROR "EVT-001 harness missing pinned source ${required}")
    endif()
endforeach()

add_subdirectory(${OPENDNP3_ROOT} ${CMAKE_BINARY_DIR}/opendnp3 EXCLUDE_FROM_ALL)

add_executable(evt_master evt_master.cpp)
target_link_libraries(evt_master PRIVATE opendnp3)

add_executable(evt_outstation evt_outstation.cpp)
target_link_libraries(evt_outstation PRIVATE opendnp3)
