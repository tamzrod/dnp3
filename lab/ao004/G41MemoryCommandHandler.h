// AO-004 isolated OpenDNP3 3.1.2 G41 memory-only command handler (harness).
//
// Independent test harness ONLY. Memory-only, fixed index 0, Direct Operate with
// response only. No real output, no OS I/O, no NGCP/SCADA mapping, no dynamic
// points. SELECT/OPERATE (SBO) and DirectOperateNoResponse are rejected.
//
// NOT_BUILT: authored as static scaffolding. The exact pinned OpenDNP3 3.1.2 API
// (headers, member names, OperateType/CommandStatus enumerators) MUST be
// re-verified against the sha256-pinned source before any compile on an
// authorized runner.
#pragma once

#include <cstdint>
#include <cstring>
#include <cmath>

#include <opendnp3/gen/CommandStatus.h>
#include <opendnp3/outstation/ICommandHandler.h>
#include <opendnp3/outstation/IUpdateHandler.h>
#include <opendnp3/app/AnalogOutput.h>

namespace ao004
{

// Records the last accepted DirectOperate per G41 variation for later evidence.
// All fields are in-memory only; nothing is written outside this object.
struct G41Observed
{
    bool seen = false;
    uint16_t index = 0;
    uint8_t variation = 0;      // 1..4 (V1 int32, V2 int16, V3 float32, V4 float64)
    CommandStatus status = CommandStatus::SUCCESS;

    // Exact value/bits captured for evidence (little-endian wire is produced by
    // the OpenDNP3 serializer from these typed values).
    int32_t i32 = 0;
    int16_t i16 = 0;
    uint32_t f32_bits = 0;
    uint64_t f64_bits = 0;
};

class G41MemoryCommandHandler final : public opendnp3::ICommandHandler
{
public:
    G41MemoryCommandHandler() = default;

    void Begin() override {}
    void End() override {}

    // --- G41V2 int16 -------------------------------------------------------
    opendnp3::CommandStatus Select(const opendnp3::AnalogOutputInt16&, uint16_t) override
    {
        return opendnp3::CommandStatus::NOT_SUPPORTED;
    }

    opendnp3::CommandStatus Operate(const opendnp3::AnalogOutputInt16& command,
                                    uint16_t index,
                                    opendnp3::IUpdateHandler& /*handler*/,
                                    opendnp3::OperateType opType) override
    {
        if (!IsDirectOperate(opType)) return opendnp3::CommandStatus::NOT_SUPPORTED;
        if (index != 0) return opendnp3::CommandStatus::OUT_OF_RANGE;

        last_.seen = true;
        last_.index = index;
        last_.variation = 2;
        last_.i16 = command.value;
        last_.status = opendnp3::CommandStatus::SUCCESS;
        return last_.status;
    }

    // --- G41V1 int32 -------------------------------------------------------
    opendnp3::CommandStatus Select(const opendnp3::AnalogOutputInt32&, uint16_t) override
    {
        return opendnp3::CommandStatus::NOT_SUPPORTED;
    }

    opendnp3::CommandStatus Operate(const opendnp3::AnalogOutputInt32& command,
                                    uint16_t index,
                                    opendnp3::IUpdateHandler& /*handler*/,
                                    opendnp3::OperateType opType) override
    {
        if (!IsDirectOperate(opType)) return opendnp3::CommandStatus::NOT_SUPPORTED;
        if (index != 0) return opendnp3::CommandStatus::OUT_OF_RANGE;

        last_.seen = true;
        last_.index = index;
        last_.variation = 1;
        last_.i32 = command.value;
        last_.status = opendnp3::CommandStatus::SUCCESS;
        return last_.status;
    }

    // --- G41V3 float32 (finite-only) --------------------------------------
    opendnp3::CommandStatus Select(const opendnp3::AnalogOutputFloat32&, uint16_t) override
    {
        return opendnp3::CommandStatus::NOT_SUPPORTED;
    }

    opendnp3::CommandStatus Operate(const opendnp3::AnalogOutputFloat32& command,
                                    uint16_t index,
                                    opendnp3::IUpdateHandler& /*handler*/,
                                    opendnp3::OperateType opType) override
    {
        if (!IsDirectOperate(opType)) return opendnp3::CommandStatus::NOT_SUPPORTED;
        if (index != 0) return opendnp3::CommandStatus::OUT_OF_RANGE;
        // Reject NaN / +/-Inf before any state mutation (Go finite-only policy).
        if (!std::isfinite(command.value)) return opendnp3::CommandStatus::OUT_OF_RANGE;

        last_.seen = true;
        last_.index = index;
        last_.variation = 3;
        std::memcpy(&last_.f32_bits, &command.value, sizeof(last_.f32_bits));
        last_.status = opendnp3::CommandStatus::SUCCESS;
        return last_.status;
    }

    // --- G41V4 float64 (finite-only) --------------------------------------
    opendnp3::CommandStatus Select(const opendnp3::AnalogOutputDouble64&, uint16_t) override
    {
        return opendnp3::CommandStatus::NOT_SUPPORTED;
    }

    opendnp3::CommandStatus Operate(const opendnp3::AnalogOutputDouble64& command,
                                    uint16_t index,
                                    opendnp3::IUpdateHandler& /*handler*/,
                                    opendnp3::OperateType opType) override
    {
        if (!IsDirectOperate(opType)) return opendnp3::CommandStatus::NOT_SUPPORTED;
        if (index != 0) return opendnp3::CommandStatus::OUT_OF_RANGE;
        if (!std::isfinite(command.value)) return opendnp3::CommandStatus::OUT_OF_RANGE;

        last_.seen = true;
        last_.index = index;
        last_.variation = 4;
        std::memcpy(&last_.f64_bits, &command.value, sizeof(last_.f64_bits));
        last_.status = opendnp3::CommandStatus::SUCCESS;
        return last_.status;
    }

    const G41Observed& Last() const { return last_; }

private:
    static bool IsDirectOperate(opendnp3::OperateType opType)
    {
        // Only Direct Operate with response is in scope. SELECT/OPERATE (SBO) and
        // DirectOperateNoResponse are rejected.
        return opType == opendnp3::OperateType::DirectOperate;
    }

    G41Observed last_;
};

} // namespace ao004
