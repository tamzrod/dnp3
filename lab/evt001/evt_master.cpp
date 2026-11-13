// OpenDNP3 3.1.2 master harness for EVT-001.
// Binds the client side to 127.0.0.1 only. Memory observations only.
#include <opendnp3/ConsoleLogger.h>
#include <opendnp3/DNP3Manager.h>
#include <opendnp3/logging/LogLevels.h>
#include <opendnp3/master/IMasterApplication.h>
#include <opendnp3/master/ISOEHandler.h>
#include <opendnp3/master/MasterStackConfig.h>
#include <opendnp3/gen/GroupVariation.h>
#include <opendnp3/gen/QualifierCode.h>
#include <opendnp3/gen/TimeSyncMode.h>

#include <opendnp3/util/UTCTimestamp.h>

#include <atomic>
#include <chrono>
#include <cstdio>
#include <cstring>
#include <mutex>
#include <string>
#include <thread>

using namespace opendnp3;

namespace {

const uint16_t kMasterAddr = 1;
const uint16_t kOutstationAddr = 1024;

std::mutex gMu;
FILE* gLog = nullptr;

void Log(const char* line)
{
    std::lock_guard<std::mutex> lock(gMu);
    if (gLog != nullptr)
    {
        std::fputs(line, gLog);
        std::fputc('\n', gLog);
        std::fflush(gLog);
    }
    std::fputs(line, stdout);
    std::fputc('\n', stdout);
    std::fflush(stdout);
}

class LabListener : public IChannelListener
{
public:
    std::atomic<bool> open{false};

    void OnStateChange(ChannelState state) override
    {
        if (state == ChannelState::OPEN)
        {
            open = true;
            Log("STATE OPEN");
        }
        else if (state == ChannelState::CLOSED)
        {
            Log("STATE CLOSED");
        }
    }
};

class LabSOE : public ISOEHandler
{
public:
    bool unsol = false;

    void BeginFragment(const ResponseInfo& info) override { unsol = info.unsolicited; }
    void EndFragment(const ResponseInfo&) override {}

    void Process(const HeaderInfo& info, const ICollection<Indexed<Binary>>& values) override
    {
        values.ForeachItem([&](const Indexed<Binary>& item) {
            char buf[256];
            std::snprintf(buf, sizeof(buf),
                          "BIN idx=%u value=%d flags=%02x time=%llu gv=%04x qual=%02x unsol=%d",
                          item.index, item.value.value ? 1 : 0, item.value.flags.value,
                          static_cast<unsigned long long>(item.value.time.value),
                          static_cast<unsigned>(info.gv), static_cast<unsigned>(info.qualifier), unsol ? 1 : 0);
            Log(buf);
        });
    }
    void Process(const HeaderInfo&, const ICollection<Indexed<DoubleBitBinary>>&) override {}
    void Process(const HeaderInfo& info, const ICollection<Indexed<Analog>>& values) override
    {
        values.ForeachItem([&](const Indexed<Analog>& item) {
            char buf[256];
            std::snprintf(buf, sizeof(buf),
                          "ANA idx=%u value=%.0f flags=%02x time=%llu gv=%04x qual=%02x unsol=%d",
                          item.index, item.value.value, item.value.flags.value,
                          static_cast<unsigned long long>(item.value.time.value),
                          static_cast<unsigned>(info.gv), static_cast<unsigned>(info.qualifier), unsol ? 1 : 0);
            Log(buf);
        });
    }
    void Process(const HeaderInfo& info, const ICollection<Indexed<Counter>>& values) override
    {
        values.ForeachItem([&](const Indexed<Counter>& item) {
            char buf[256];
            std::snprintf(buf, sizeof(buf),
                          "CTR idx=%u value=%u flags=%02x time=%llu gv=%04x qual=%02x unsol=%d",
                          item.index, item.value.value, item.value.flags.value,
                          static_cast<unsigned long long>(item.value.time.value),
                          static_cast<unsigned>(info.gv), static_cast<unsigned>(info.qualifier), unsol ? 1 : 0);
            Log(buf);
        });
    }
    void Process(const HeaderInfo&, const ICollection<Indexed<FrozenCounter>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<BinaryOutputStatus>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<AnalogOutputStatus>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<OctetString>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<TimeAndInterval>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<BinaryCommandEvent>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<AnalogCommandEvent>>&) override {}
    void Process(const HeaderInfo&, const ICollection<DNPTime>&) override {}
};

class LabApp : public IMasterApplication
{
public:
    UTCTimestamp Now() override
    {
        auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
                      std::chrono::system_clock::now().time_since_epoch())
                      .count();
        return UTCTimestamp(static_cast<uint64_t>(ms));
    }

    void OnReceiveIIN(const IINField& iin) override
    {
        char buf[64];
        std::snprintf(buf, sizeof(buf), "IIN lsb=%02x msb=%02x", iin.LSB, iin.MSB);
        Log(buf);
    }
};

} // namespace

int main(int argc, char** argv)
{
    if (argc != 5)
    {
        std::fprintf(stderr, "usage: evt_master <solicit|unsol|silent|class0> <port> <seconds> <logfile>\n");
        return 2;
    }
    const std::string mode = argv[1];
    const int port = std::atoi(argv[2]);
    const int seconds = std::atoi(argv[3]);
    gLog = std::fopen(argv[4], "w");
    if (gLog == nullptr || port <= 0 || port > 65535 || seconds <= 0 || seconds > 30)
    {
        std::fprintf(stderr, "bad arguments\n");
        return 2;
    }

    auto soe = std::make_shared<LabSOE>();
    auto app = std::make_shared<LabApp>();
    auto listener = std::make_shared<LabListener>();

    DNP3Manager manager(1, ConsoleLogger::Create());
    std::shared_ptr<IChannel> channel;
    try
    {
        channel = manager.AddTCPClient("evt-master", levels::NORMAL, ChannelRetry::Default(),
                                       {IPEndpoint("127.0.0.1", static_cast<uint16_t>(port))}, "127.0.0.1", listener);
    }
    catch (const std::exception& ex)
    {
        Log("STATE FAIL");
        std::fprintf(stderr, "channel: %s\n", ex.what());
        manager.Shutdown();
        {
            std::lock_guard<std::mutex> lock(gMu);
            std::fclose(gLog);
            gLog = nullptr;
        }
        return 1;
    }

    MasterStackConfig stack;
    stack.master.responseTimeout = TimeDuration::Seconds(2);
    stack.master.timeSyncMode = TimeSyncMode::None;
    stack.master.ignoreRestartIIN = true;
    stack.master.startupIntegrityClassMask = ClassField::None();
    stack.master.eventScanOnEventsAvailableClassMask = ClassField::None();
    stack.master.integrityOnEventOverflowIIN = false;
    stack.master.disableUnsolOnStartup = mode != "unsol";
    stack.master.unsolClassMask = mode == "unsol" ? ClassField::AllEventClasses() : ClassField::None();
    stack.link.LocalAddr = kMasterAddr;
    stack.link.RemoteAddr = kOutstationAddr;
    stack.link.KeepAliveTimeout = TimeDuration::Max();

    auto master = channel->AddMaster("master", soe, app, stack);
    master->Enable();

    // ScanClasses is an ad-hoc task. OpenDNP3 3.1.2 drops it when the master
    // is not online yet (MContext::ScheduleAdhocTask). AddClassScan is bound
    // and starts on link-up, so the class read is not lost to that race. The
    // period is longer than this process, so the scan runs once.
    const auto period = TimeDuration::Seconds(30);
    if (mode == "solicit")
    {
        master->AddClassScan(ClassField::AllEventClasses(), period, soe);
    }
    else if (mode == "class0")
    {
        master->AddClassScan(ClassField(true, false, false, false), period, soe);
    }

    std::this_thread::sleep_for(std::chrono::seconds(seconds));
    const bool opened = listener->open.load();
    Log(opened ? "END open=1" : "END open=0");
    // Shutdown can emit STATE CLOSED on the pool thread. Close the log only
    // after that callback has joined, otherwise Log writes a freed FILE.
    manager.Shutdown();
    {
        std::lock_guard<std::mutex> lock(gMu);
        std::fclose(gLog);
        gLog = nullptr;
    }
    return opened ? 0 : 1;
}
