// AO-004 isolated OpenDNP3 3.1.2 G41 peer (master) — harness main.
//
// Independent test harness ONLY. Builds an OpenDNP3 master that can later,
// under a SEPARATE owner authorization, issue single-point G41 Direct Operate
// WITH RESPONSE commands (V1 int32, V2 int16, V3 float32, V4 float64) at a
// bounded point index (--index 0..1, default 0) to our Go DNP3 outstation over
// the isolated lab link.
//
// Link addresses (master): LocalAddr=3, RemoteAddr=4 (the Go outstation).
// Target: 10.77.0.2:20000, reachable only inside the future isolated
// vec01-master <-> vec01-outstation topology (see lab/vec01/).
//
// SAFETY: this program opens a TCP channel, starts a DNP3 master session and
// sends a command ONLY when explicitly invoked with "--run". With no arguments
// (default) it constructs no channel, opens no socket, starts no session, sends
// no command, and exits after a bounded readiness/feature print. The "--run"
// mode refuses any host that is not the isolated lab address, requires an
// explicit index/variation, and waits a bounded time for each response. It must
// only ever be invoked on an owner-provided isolated runner under a separate
// protocol-execution authorization; it is NOT authorized by the microtask that
// authored this file.
//
// The Go outstation executable is NOT started by this harness.
//
// API verified against the sha256-pinned OpenDNP3 3.1.2 source (see
// docs/verification/AO-004.md Appendix L): DNP3Manager::AddTCPClient /
// IChannel::AddMaster, MasterStackConfig{master,link}, IMasterOperations::
// DirectOperate(CommandSet&&, callback, config), CommandSet/WithIndex,
// AnalogOutputInt16/Int32/Float32/Double64, CommandStatus, PrintingSOEHandler.
#include <chrono>
#include <atomic>
#include <cstdint>
#include <iostream>
#include <memory>
#include <sstream>
#include <string>
#include <thread>
#include <vector>

#include <opendnp3/ConsoleLogger.h>
#include <opendnp3/DNP3Manager.h>
#include <opendnp3/channel/ChannelRetry.h>
#include <opendnp3/channel/IChannel.h>
#include <opendnp3/channel/IChannelListener.h>
#include <opendnp3/channel/IPEndpoint.h>
#include <opendnp3/channel/PrintingChannelListener.h>
#include <opendnp3/gen/CommandPointState.h>
#include <opendnp3/gen/CommandStatus.h>
#include <opendnp3/gen/TaskCompletion.h>
#include <opendnp3/logging/LogLevels.h>
#include <opendnp3/master/CommandSet.h>
#include <opendnp3/master/DefaultMasterApplication.h>
#include <opendnp3/master/IMaster.h>
#include <opendnp3/master/MasterStackConfig.h>
#include <opendnp3/master/PrintingSOEHandler.h>

#include "G41MasterCommandCallback.h"

namespace
{
// Records the last channel state observed by the OpenDNP3 channel listener so
// the single-shot harness can wait (bounded) for OPEN before issuing its
// command, removing the connect/operate race. It takes no action of its own.
class OpenStateListener final : public opendnp3::IChannelListener
{
public:
    void OnStateChange(opendnp3::ChannelState state) override
    {
        last_.store(state, std::memory_order_release);
    }

    opendnp3::ChannelState Get() const { return last_.load(std::memory_order_acquire); }

private:
    std::atomic<opendnp3::ChannelState> last_{opendnp3::ChannelState::CLOSED};
};

// The only host this harness will ever connect to. Any other value is refused
// before a channel is created.
constexpr const char* kLabOutstationAddress = "10.77.0.2";
constexpr uint16_t kLabOutstationPort = 20000;
constexpr uint16_t kLinkLocalAddr = 3;   // master
constexpr uint16_t kLinkRemoteAddr = 4;  // Go outstation
// Bounded index selector: the harness supports exactly the in-range index 0 and
// the single invalid index 1 exercised by the authorized negative campaign. Any
// other value is rejected before a channel is created.
constexpr uint16_t kCommandIndexDefault = 0;
constexpr uint16_t kCommandIndexMax = 1;
constexpr unsigned long kTimeoutMaxSeconds = 600;

// Strict bounded unsigned parser for command-line numerics. Accepts only a
// non-empty run of ASCII digits whose value is <= `max`; rejects sign
// characters, empty input, trailing junk and overflow. Never throws, so a
// malformed argument cannot crash the harness or silently fall back to a
// default.
bool ParseBoundedUnsigned(const std::string& text, unsigned long max, unsigned long& out)
{
    if (text.empty())
    {
        return false;
    }
    unsigned long value = 0;
    for (const char c : text)
    {
        if (c < '0' || c > '9')
        {
            return false;
        }
        const unsigned long digit = static_cast<unsigned long>(c - '0');
        if (value > max / 10 || (value == max / 10 && digit > max % 10))
        {
            return false;
        }
        value = value * 10 + digit;
    }
    out = value;
    return true;
}

void PrintReadiness()
{
    std::cout << "ao004-g41-master: readiness mode (no channel, no session, no command).\n"
                 "  role          : DNP3 master (OpenDNP3 3.1.2)\n"
                 "  link          : LocalAddr=" << kLinkLocalAddr << " -> RemoteAddr=" << kLinkRemoteAddr
              << "\n"
                 "  target        : "
              << kLabOutstationAddress << ":" << kLabOutstationPort << " (isolated lab only)\n"
                 "  command index : "
              << kCommandIndexDefault << " (default; --index accepts 0.." << kCommandIndexMax << " only)\n"
                 "  variations    : V1 int32, V2 int16, V3 float32, V4 float64\n"
                 "  mode          : Direct Operate WITH RESPONSE only (no SBO, no NoResponse, no retry)\n"
                 "  run mode      : '--run --variation <1|2|3|4> [--index <0|1>] [--timeout <s>]' (NOT authorized here)\n"
                 "  values        : fixed typed defaults per variation (no --value option)\n"
              << std::endl;
}

// Build the single-point CommandSet for the requested variation at the
// requested index. Only Direct Operate with response is used; no
// SelectAndOperate, no NoResponse.
opendnp3::CommandSet BuildCommandSet(int variation, uint16_t index)
{
    switch (variation)
    {
    case 1:
        return opendnp3::CommandSet({opendnp3::WithIndex(opendnp3::AnalogOutputInt32(2300), index)});
    case 2:
        return opendnp3::CommandSet({opendnp3::WithIndex(opendnp3::AnalogOutputInt16(12345), index)});
    case 3:
        return opendnp3::CommandSet({opendnp3::WithIndex(opendnp3::AnalogOutputFloat32(1.5f), index)});
    case 4:
        return opendnp3::CommandSet({opendnp3::WithIndex(opendnp3::AnalogOutputDouble64(1.5), index)});
    default:
        return opendnp3::CommandSet();
    }
}

// Live mode. NOT authorized by the microtask that authored this file; it exists
// so a later, separately authorized campaign can drive it on an isolated runner.
int RunLive(const std::string& host, uint16_t port, int variation, uint16_t index, int timeoutSeconds)
{
    if (host != kLabOutstationAddress || port != kLabOutstationPort)
    {
        std::cerr << "ao004-g41-master: refusing non-lab target " << host << ":" << port
                  << " (only " << kLabOutstationAddress << ":" << kLabOutstationPort << " is permitted)"
                  << std::endl;
        return 2;
    }
    if (variation < 1 || variation > 4)
    {
        std::cerr << "ao004-g41-master: --variation must be 1..4" << std::endl;
        return 2;
    }
    if (index > kCommandIndexMax)
    {
        std::cerr << "ao004-g41-master: --index must be 0.." << kCommandIndexMax << std::endl;
        return 2;
    }
    if (timeoutSeconds < 1 || static_cast<unsigned long>(timeoutSeconds) > kTimeoutMaxSeconds)
    {
        std::cerr << "ao004-g41-master: --timeout must be 1.." << kTimeoutMaxSeconds << std::endl;
        return 2;
    }

    auto logLevels = opendnp3::levels::NORMAL;
    opendnp3::DNP3Manager manager(1, opendnp3::ConsoleLogger::Create());

    auto stateListener = std::make_shared<OpenStateListener>();
    auto channel = manager.AddTCPClient("ao004-g41-master",
                                        logLevels,
                                        opendnp3::ChannelRetry::Default(),
                                        {opendnp3::IPEndpoint(kLabOutstationAddress, kLabOutstationPort)},
                                        "0.0.0.0",
                                        stateListener);

    opendnp3::MasterStackConfig config;
    config.link.LocalAddr = kLinkLocalAddr;
    config.link.RemoteAddr = kLinkRemoteAddr;
    config.link.KeepAliveTimeout = opendnp3::TimeDuration::Max();

    auto master = channel->AddMaster("master",
                                     opendnp3::PrintingSOEHandler::Create(),
                                     opendnp3::DefaultMasterApplication::Create(),
                                     config);

    if (!master->Enable())
    {
        std::cerr << "ao004-g41-master: failed to enable master" << std::endl;
        return 1;
    }

    // Wait (bounded) for the channel to reach OPEN before issuing the command.
    // Without this the single DirectOperate can race the initial TCP connect and
    // return a spurious FAILURE_NO_COMMS; the wait makes each single-shot case
    // deterministic. It issues no command and sends no DNP3 traffic itself.
    const auto openDeadline = std::chrono::steady_clock::now() + std::chrono::seconds(timeoutSeconds);
    while (stateListener->Get() != opendnp3::ChannelState::OPEN
           && std::chrono::steady_clock::now() < openDeadline)
    {
        std::this_thread::sleep_for(std::chrono::milliseconds(20));
    }
    if (stateListener->Get() != opendnp3::ChannelState::OPEN)
    {
        std::cerr << "ao004-g41-master: channel did not open within " << timeoutSeconds << "s" << std::endl;
        return 1;
    }

    auto outcome = std::make_shared<ao004::G41MasterCommandOutcome>();
    auto commands = BuildCommandSet(variation, index);
    master->DirectOperate(std::move(commands), ao004::MakeG41MasterCommandCallback(outcome));

    const auto deadline = std::chrono::steady_clock::now() + std::chrono::seconds(timeoutSeconds);
    while (!outcome->completed.load(std::memory_order_acquire) && std::chrono::steady_clock::now() < deadline)
    {
        std::this_thread::sleep_for(std::chrono::milliseconds(20));
    }

    std::ostringstream line;
    line << "ao004-g41-master: variation=" << variation << " index=" << index
         << " completed=" << (outcome->completed ? "yes" : "no");
    if (outcome->completed)
    {
        line << " summary=" << opendnp3::TaskCompletionSpec::to_human_string(outcome->summary);
        for (const auto& point : outcome->points)
        {
            line << " point[index=" << point.index << " state="
                 << opendnp3::CommandPointStateSpec::to_human_string(point.state) << " status="
                 << opendnp3::CommandStatusSpec::to_human_string(point.status) << "]";
        }
    }
    std::cout << line.str() << std::endl;

    master->Shutdown();
    return outcome->completed ? 0 : 1;
}
} // namespace

int main(int argc, char* argv[])
{
    std::string host = kLabOutstationAddress;
    unsigned long port = kLabOutstationPort;
    unsigned long variation = 0;
    unsigned long index = kCommandIndexDefault;
    unsigned long timeoutSeconds = 10;
    bool run = false;
    bool badArg = false;

    // Reject a missing/invalid option value with a clear message and exit 2
    // BEFORE any channel is created. No std::stoi, so no exception can escape.
    auto takeValue = [&](int& i, unsigned long max, const char* name, unsigned long& out) -> bool {
        if (i + 1 >= argc)
        {
            std::cerr << "ao004-g41-master: " << name << " requires a value" << std::endl;
            return false;
        }
        if (!ParseBoundedUnsigned(argv[++i], max, out))
        {
            std::cerr << "ao004-g41-master: invalid " << name << " '" << argv[i] << "' (must be an integer 0.."
                      << max << ")" << std::endl;
            return false;
        }
        return true;
    };

    for (int i = 1; i < argc; ++i)
    {
        const std::string arg = argv[i];
        if (arg == "--run")
        {
            run = true;
        }
        else if (arg == "--host")
        {
            if (i + 1 >= argc)
            {
                std::cerr << "ao004-g41-master: --host requires a value" << std::endl;
                badArg = true;
            }
            else
            {
                host = argv[++i];
            }
        }
        else if (arg == "--port")
        {
            badArg = !takeValue(i, 65535, "--port", port);
        }
        else if (arg == "--variation")
        {
            badArg = !takeValue(i, 4, "--variation", variation);
        }
        else if (arg == "--index")
        {
            badArg = !takeValue(i, kCommandIndexMax, "--index", index);
        }
        else if (arg == "--timeout")
        {
            badArg = !takeValue(i, kTimeoutMaxSeconds, "--timeout", timeoutSeconds);
        }
        else if (arg == "--help")
        {
            PrintReadiness();
            return 0;
        }
        else
        {
            std::cerr << "ao004-g41-master: unknown argument '" << arg << "'" << std::endl;
            badArg = true;
        }

        if (badArg)
        {
            return 2;
        }
    }

    if (!run)
    {
        // Default: offline readiness only. No channel, no session, no command.
        PrintReadiness();
        return 0;
    }

    return RunLive(host, static_cast<uint16_t>(port), static_cast<int>(variation),
                   static_cast<uint16_t>(index), static_cast<int>(timeoutSeconds));
}
