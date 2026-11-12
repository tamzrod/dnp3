// AO-004 isolated OpenDNP3 3.1.2 G41 master command-result callback (harness).
//
// Independent test harness ONLY. Observability for a single Direct Operate
// with response: it records the task summary and the per-point state/status so
// a later authorized campaign can read the outstation's per-point command
// status. It never reissues, retries or mutates anything.
//
// API verified against the sha256-pinned OpenDNP3 3.1.2 source:
//   * opendnp3::CommandResultCallbackT = std::function<void(const ICommandTaskResult&)>
//     (opendnp3/master/CommandResultCallbackT.h)
//   * ICommandTaskResult : ICollection<CommandPointResult> with a TaskCompletion
//     `summary` (opendnp3/master/ICommandTaskResult.h); the collection is visited
//     via ICollection::ForeachItem (opendnp3/app/parsing/ICollection.h)
//   * CommandPointResult { headerIndex, index, state, status } (opendnp3/master/CommandPointResult.h)
#pragma once

#include <atomic>
#include <cstdint>
#include <memory>
#include <vector>

#include <opendnp3/gen/CommandPointState.h>
#include <opendnp3/gen/CommandStatus.h>
#include <opendnp3/gen/TaskCompletion.h>
#include <opendnp3/master/CommandResultCallbackT.h>
#include <opendnp3/master/ICommandTaskResult.h>

namespace ao004
{

// One flattened per-point result (CommandPointResult) captured from a callback.
struct G41MasterPointResult
{
    uint32_t headerIndex = 0;
    uint16_t index = 0;
    opendnp3::CommandPointState state = opendnp3::CommandPointState::INIT;
    opendnp3::CommandStatus status = opendnp3::CommandStatus::SUCCESS;
};

// The outcome of one Direct Operate with response. `completed` is set (with
// release ordering, after `points` is populated) only when the OpenDNP3 callback
// fired; a bounded wait that expires leaves it false.
struct G41MasterCommandOutcome
{
    std::atomic<bool> completed{false};
    opendnp3::TaskCompletion summary = opendnp3::TaskCompletion::FAILURE_NO_COMMS;
    std::vector<G41MasterPointResult> points;
};

// Build a CommandResultCallbackT that records the task summary and every
// per-point result into `outcome`. The callback holds only a shared_ptr to the
// outcome, so it stays valid for the lifetime of the operation.
opendnp3::CommandResultCallbackT
MakeG41MasterCommandCallback(const std::shared_ptr<G41MasterCommandOutcome>& outcome);

} // namespace ao004
