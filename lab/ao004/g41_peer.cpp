// AO-004 isolated OpenDNP3 3.1.2 G41 peer (outstation) — harness main.
//
// Independent test harness ONLY. Builds an OpenDNP3 outstation whose command
// handler is the memory-only G41 handler. It binds a DNP3 TCP listener on the
// isolated lab address only (10.77.0.2:20000) and uses lab link addresses
// (LocalAddr=4, RemoteAddr=3).
//
// SAFETY: this program starts a listener and speaks DNP3 ONLY when explicitly
// invoked with "--listen". With no arguments it prints a readiness line and
// exits WITHOUT creating a socket. Even with "--listen", it must be run only on
// an owner-provided isolated runner under a separate protocol-execution
// authorization. NOT_BUILT: authored as static scaffolding; the exact pinned
// OpenDNP3 3.1.2 API must be re-verified against the sha256-pinned source.
#include <iostream>
#include <memory>
#include <string>

#include <opendnp3/ConsoleLogger.h>
#include <opendnp3/DNP3Manager.h>
#include <opendnp3/channel/IChannel.h>
#include <opendnp3/channel/PrintingChannelListener.h>
#include <opendnp3/gen/ServerAcceptMode.h>
#include <opendnp3/outstation/DefaultOutstationApplication.h>
#include <opendnp3/outstation/OutstationStackConfig.h>
#include <opendnp3/outstation/IOutstation.h>
#include <opendnp3/outstation/UpdateBuilder.h>
#include <opendnp3/app/IPEndpoint.h>

#include "G41MemoryCommandHandler.h"

namespace
{
constexpr const char* kLabBindAddress = "10.77.0.2";
constexpr uint16_t kLabBindPort = 20000;
constexpr uint16_t kLinkLocalAddr = 4;   // outstation
constexpr uint16_t kLinkRemoteAddr = 3;  // Go master
} // namespace

int main(int argc, char* argv[])
{
    bool listen = false;
    for (int i = 1; i < argc; ++i)
    {
        if (std::string(argv[i]) == "--listen")
        {
            listen = true;
        }
    }

    if (!listen)
    {
        // Readiness only: no socket, no DNP3 traffic, no control.
        std::cout << "ao004-g41-peer: readiness mode (no listener). "
                     "Run with --listen only on an authorized isolated runner."
                  << std::endl;
        return 0;
    }

    auto logLevels = levels::NORMAL;
    opendnp3::DNP3Manager manager(1, opendnp3::ConsoleLogger::Create());

    auto handler = std::make_shared<ao004::G41MemoryCommandHandler>();

    // Outstation stack config. G41 static variations are not advertised in
    // Class-0; this peer serves the G41 command subset only.
    opendnp3::OutstationStackConfig config(opendnp3::DatabaseConfig(1));
    config.outstation.params.allowUnsolicited = false;
    config.link.LocalAddr = kLinkLocalAddr;
    config.link.RemoteAddr = kLinkRemoteAddr;
    config.link.KeepAliveTimeout = opendnp3::TimeDuration::Max();

    auto channel = manager.AddTCPServer(
        "ao004-g41",
        logLevels,
        opendnp3::ServerAcceptMode::CloseExisting,
        opendnp3::IPEndpoint(kLabBindAddress, kLabBindPort),
        opendnp3::PrintingChannelListener::Create());

    auto outstation = channel->AddOutstation(
        "outstation",
        handler,
        opendnp3::DefaultOutstationApplication::Create(),
        config);

    outstation->Enable();

    std::cout << "ao004-g41-peer: listening on " << kLabBindAddress << ":"
              << kLabBindPort << " (link " << kLinkLocalAddr << " -> "
              << kLinkRemoteAddr << "); memory-only G41 handler active"
              << std::endl;

    // Bounded lifetime: a later authorized campaign drives this process; here it
    // simply blocks so the caller controls termination.
    std::string line;
    while (std::getline(std::cin, line))
    {
        if (line == "quit") break;
    }
    return 0;
}
