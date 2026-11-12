// AO-004 isolated OpenDNP3 3.1.2 G41 master command-result callback — translation unit.
//
// Independent test harness ONLY. No protocol exchange is initiated here; this TU
// only builds the result callback used by the master harness main.
#include "G41MasterCommandCallback.h"

namespace ao004
{

opendnp3::CommandResultCallbackT
MakeG41MasterCommandCallback(const std::shared_ptr<G41MasterCommandOutcome>& outcome)
{
    return [outcome](const opendnp3::ICommandTaskResult& result) {
        if (!outcome)
        {
            return;
        }

        outcome->summary = result.summary;
        outcome->points.clear();

        result.ForeachItem([&outcome](const opendnp3::CommandPointResult& point) {
            G41MasterPointResult captured;
            captured.headerIndex = point.headerIndex;
            captured.index = point.index;
            captured.state = point.state;
            captured.status = point.status;
            outcome->points.push_back(captured);
        });

        // Publish last so a waiting thread that observes `completed` also sees the
        // fully populated summary/points.
        outcome->completed.store(true, std::memory_order_release);
    };
}

} // namespace ao004
