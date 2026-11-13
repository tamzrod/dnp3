// OpenDNP3 3.1.2 outstation harness for EVT-001.
// Listens on 127.0.0.1 only. Points are memory values, not hardware.
#include <opendnp3/ConsoleLogger.h>
#include <opendnp3/DNP3Manager.h>
#include <opendnp3/logging/LogLevels.h>
#include <opendnp3/app/DNPTime.h>
#include <opendnp3/app/Flags.h>
#include <opendnp3/app/MeasurementTypes.h>
#include <opendnp3/outstation/DefaultOutstationApplication.h>
#include <opendnp3/outstation/EventBufferConfig.h>
#include <opendnp3/outstation/OutstationStackConfig.h>
#include <opendnp3/outstation/SimpleCommandHandler.h>
#include <opendnp3/outstation/UpdateBuilder.h>
#include <opendnp3/outstation/NumRetries.h>
#include <opendnp3/gen/EventAnalogVariation.h>
#include <opendnp3/gen/EventBinaryVariation.h>
#include <opendnp3/gen/EventCounterVariation.h>
#include <opendnp3/gen/PointClass.h>
#include <opendnp3/gen/StaticAnalogVariation.h>
#include <opendnp3/gen/StaticBinaryVariation.h>
#include <opendnp3/gen/StaticCounterVariation.h>

#include <iostream>
#include <string>

using namespace opendnp3;

namespace {

// Milliseconds since 2000-01-01T00:00:00Z. OpenDNP3 writes DNPTime::value
// onto the wire as a raw 48-bit count, so the harness stores the IEEE count
// rather than a Unix millisecond count.
const uint64_t kStamp = 676368550000ULL;       // 2021-06-07T08:09:10Z
const uint64_t kLive = 699685567008ULL;        // 2022-03-04T05:06:07.008Z
const uint16_t kOutstationAddr = 1024;
const uint16_t kMasterAddr = 1;

DatabaseConfig Configure()
{
    DatabaseConfig config;
    config.binary_input[7].clazz = PointClass::Class1;
    config.binary_input[7].svariation = StaticBinaryVariation::Group1Var1;
    config.binary_input[7].evariation = EventBinaryVariation::Group2Var2;
    config.binary_input[9].clazz = PointClass::Class1;
    config.binary_input[9].svariation = StaticBinaryVariation::Group1Var1;
    config.binary_input[9].evariation = EventBinaryVariation::Group2Var2;

    config.analog_input[3].clazz = PointClass::Class2;
    config.analog_input[3].svariation = StaticAnalogVariation::Group30Var1;
    config.analog_input[3].evariation = EventAnalogVariation::Group32Var3;

    config.counter[4].clazz = PointClass::Class3;
    config.counter[4].svariation = StaticCounterVariation::Group20Var1;
    config.counter[4].evariation = EventCounterVariation::Group22Var5;
    return config;
}

void Apply(const std::shared_ptr<IOutstation>& outstation, const char* which)
{
    UpdateBuilder builder;
    if (std::string(which) == "load")
    {
        builder.Update(Binary(true, Flags(0x01), DNPTime(kStamp)), 7);
        builder.Update(Analog(-2.0, Flags(0x01), DNPTime(kStamp)), 3);
        builder.Update(Counter(0x01020304, Flags(0x21), DNPTime(kStamp)), 4);
    }
    else if (std::string(which) == "live")
    {
        builder.Update(Binary(false, Flags(0x01), DNPTime(kLive)), 9);
    }
    else if (std::string(which) == "live-on")
    {
        builder.Update(Binary(true, Flags(0x01), DNPTime(kLive)), 9);
    }
    else
    {
        return;
    }
    outstation->Apply(builder.Build());
}

} // namespace

int main(int argc, char** argv)
{
    if (argc != 2)
    {
        std::cerr << "usage: evt_outstation <port>\n";
        return 2;
    }
    const int port = std::atoi(argv[1]);
    if (port <= 0 || port > 65535)
    {
        return 2;
    }

    DNP3Manager manager(1, ConsoleLogger::Create());
    std::shared_ptr<IChannel> channel;
    try
    {
        channel = manager.AddTCPServer("evt-outstation", levels::NORMAL, ServerAcceptMode::CloseExisting,
                                       IPEndpoint("127.0.0.1", static_cast<uint16_t>(port)), nullptr);
    }
    catch (const std::exception& ex)
    {
        std::cerr << ex.what() << "\n";
        return 1;
    }

    OutstationStackConfig stack(Configure());
    stack.outstation.eventBufferConfig = EventBufferConfig::AllTypes(50);
    stack.outstation.params.allowUnsolicited = true;
    // Class mask starts empty so a confirmed null unsolicited does not push
    // the loaded events before the master issues its class read. ENABLE
    // UNSOLICITED sets the mask for the unsolicited scenario.
    stack.outstation.params.unsolClassMask = ClassField::None();
    stack.outstation.params.numUnsolRetries = NumRetries::Fixed(2);
    stack.link.LocalAddr = kOutstationAddr;
    stack.link.RemoteAddr = kMasterAddr;
    stack.link.KeepAliveTimeout = TimeDuration::Max();

    auto outstation = channel->AddOutstation("outstation", SuccessCommandHandler::Create(),
                                             DefaultOutstationApplication::Create(), stack);
    outstation->Enable();
    std::cout << "READY\n" << std::flush;

    std::string line;
    while (std::getline(std::cin, line))
    {
        if (line == "quit")
        {
            break;
        }
        if (line == "load" || line == "live" || line == "live-on")
        {
            Apply(outstation, line.c_str());
            std::cout << "OK " << line << "\n" << std::flush;
        }
        else if (!line.empty())
        {
            std::cout << "ERR " << line << "\n" << std::flush;
        }
    }
    manager.Shutdown();
    return 0;
}
