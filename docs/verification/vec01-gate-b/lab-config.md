# Lab-only configuration: v0 simulator patch and prospective master

This file reproduces, verbatim, the **single-file** lab-only change to the OpenDNP3
outstation example (`cpp/examples/outstation/main.cpp`). The pinned protocol library is
**not** modified. The patch combines the B2 address-only change and the B3 v0
variation/point/BO[0] change. It is a **rebuildable input**, not a compiled binary.

Recorded patch digests (external artifacts, not committed):

- `b2-address-only.patch` SHA-256 `1b6ed5402885d757e7f04a9e5ecd25e3b8d7bcb2db701c1ec4e17f475e52247d`
- `b3-lab-demo.patch` (7687 bytes) SHA-256 `1758ad1c8999f1b948a3773aa82ce823b17b8cdefc73072764c0818fa33b8241`

> **Whitespace note.** The byte-exact unified diff contains blank/context lines carrying a
> single trailing space. The copy below has trailing whitespace stripped from blank/context
> lines only so the repository's `git diff --check` gate stays clean; **no code token
> differs**. The authoritative byte-exact originals remain the external artifacts above.

## Lab-only patch (verbatim)

```diff
--- src/opendnp3-3.1.2/cpp/examples/outstation/main.cpp 2022-04-22 17:01:02.000000000 +0000
+++ src/opendnp3-3.1.2-b3/cpp/examples/outstation/main.cpp      2026-11-08 10:54:37.063022946 +0000
@@ -25,8 +25,12 @@
 #include <opendnp3/outstation/IUpdateHandler.h>
 #include <opendnp3/outstation/SimpleCommandHandler.h>
 #include <opendnp3/outstation/UpdateBuilder.h>
+#include <opendnp3/gen/StaticCounterVariation.h>
+#include <opendnp3/gen/StaticTypeBitmask.h>
+#include <opendnp3/outstation/StaticTypeBitfield.h>

 #include <iostream>
+#include <memory>
 #include <string>
 #include <thread>

@@ -37,13 +41,107 @@
 {
     DatabaseConfig config(10); // 10 of each type with default settings

-    config.analog_input[0].clazz = PointClass::Class2;
-    config.analog_input[0].svariation = StaticAnalogVariation::Group30Var5;
-    config.analog_input[0].evariation = EventAnalogVariation::Group32Var7;
-
+    // VEC-01 lab v0 subset: exact static variations for index 0.
+    // The stock Group30Var5 / Class2 analog override is removed.
+    config.binary_input[0].svariation = StaticBinaryVariation::Group1Var1;
+    config.analog_input[0].clazz = PointClass::Class0;
+    config.analog_input[0].svariation = StaticAnalogVariation::Group30Var1;
+    config.counter[0].svariation = StaticCounterVariation::Group20Var1;
+
     return config;
 }

+// VEC-01 lab: dedicated memory-only G12V1 CROB handler for simulated BO[0].
+// Supports ONLY a single DirectOperate latch on/off at index 0. It rejects
+// select-before-operate, direct-operate-no-ack, wrong indices, unsupported
+// control codes, multi-operation counts and all analog outputs. It returns
+// SUCCESS only after the in-memory BO[0] state has actually been mutated.
+class LabCrobCommandHandler final : public ICommandHandler
+{
+public:
+    LabCrobCommandHandler() : state_(false) {}
+
+    void Begin() override {}
+    void End() override {}
+
+    // Select is not part of the v0 direct-operate subset; reject SBO explicitly.
+    CommandStatus Select(const ControlRelayOutputBlock&, uint16_t) override
+    {
+        return CommandStatus::NOT_SUPPORTED;
+    }
+
+    CommandStatus Operate(const ControlRelayOutputBlock& command,
+                          uint16_t index,
+                          IUpdateHandler& /*handler*/,
+                          OperateType opType) override
+    {
+        // Only a direct operate is accepted; SBO and no-ack are rejected.
+        if (opType != OperateType::DirectOperate)
+        {
+            return CommandStatus::NOT_SUPPORTED;
+        }
+
+        const auto status = Validate(command, index);
+        if (status != CommandStatus::SUCCESS)
+        {
+            return status;
+        }
+
+        // Memory-only simulated output state. SUCCESS is returned only after the
+        // state has been written and read back equal to the requested value.
+        const bool requested = (command.opType == OperationType::LATCH_ON);
+        state_ = requested;
+        const bool observed = state_;
+        std::cout << "[LAB] BO[0] direct-operate LATCH_" << (requested ? "ON" : "OFF")
+                  << " -> state=" << (observed ? "true" : "false") << std::endl;
+        return (observed == requested) ? CommandStatus::SUCCESS : CommandStatus::HARDWARE_ERROR;
+    }
+
+    // Analog outputs (Group 41) are outside the v0 subset.
+    CommandStatus Select(const AnalogOutputInt16&, uint16_t) override { return CommandStatus::NOT_SUPPORTED; }
+    CommandStatus Operate(const AnalogOutputInt16&, uint16_t, IUpdateHandler&, OperateType) override
+    {
+        return CommandStatus::NOT_SUPPORTED;
+    }
+    CommandStatus Select(const AnalogOutputInt32&, uint16_t) override { return CommandStatus::NOT_SUPPORTED; }
+    CommandStatus Operate(const AnalogOutputInt32&, uint16_t, IUpdateHandler&, OperateType) override
+    {
+        return CommandStatus::NOT_SUPPORTED;
+    }
+    CommandStatus Select(const AnalogOutputFloat32&, uint16_t) override { return CommandStatus::NOT_SUPPORTED; }
+    CommandStatus Operate(const AnalogOutputFloat32&, uint16_t, IUpdateHandler&, OperateType) override
+    {
+        return CommandStatus::NOT_SUPPORTED;
+    }
+    CommandStatus Select(const AnalogOutputDouble64&, uint16_t) override { return CommandStatus::NOT_SUPPORTED; }
+    CommandStatus Operate(const AnalogOutputDouble64&, uint16_t, IUpdateHandler&, OperateType) override
+    {
+        return CommandStatus::NOT_SUPPORTED;
+    }
+
+    bool OutputState() const { return state_; }
+
+private:
+    CommandStatus Validate(const ControlRelayOutputBlock& command, uint16_t index) const
+    {
+        if (index != 0)
+        {
+            return CommandStatus::OUT_OF_RANGE;
+        }
+        if (command.opType != OperationType::LATCH_ON && command.opType != OperationType::LATCH_OFF)
+        {
+            return CommandStatus::NOT_SUPPORTED;
+        }
+        if (command.count != 1)
+        {
+            return CommandStatus::FORMAT_ERROR;
+        }
+        return CommandStatus::SUCCESS;
+    }
+
+    bool state_;
+};
+
 struct State
 {
     uint32_t count = 0;
@@ -73,7 +171,7 @@
     auto channel = std::shared_ptr<IChannel>(nullptr);
     try
     {
-        channel = manager.AddTCPServer("server", logLevels, ServerAcceptMode::CloseExisting, IPEndpoint("0.0.0.0", 20000),
+        channel = manager.AddTCPServer("server", logLevels, ServerAcceptMode::CloseExisting, IPEndpoint("10.77.0.2", 20000),
                                        PrintingChannelListener::Create());
     }
     catch(const std::exception& e)
@@ -92,23 +190,41 @@
     // you can override an default outstation parameters here
     // in this example, we've enabled the oustation to use unsolicted reporting
     // if the master enables it
-    config.outstation.params.allowUnsolicited = true;
+    config.outstation.params.allowUnsolicited = false;
+
+    // VEC-01 lab: restrict Class-0 (G60V1 all-objects) to the in-profile static
+    // types only, so no out-of-profile Class-0 objects are emitted.
+    config.outstation.params.typesAllowedInClass0 = StaticTypeBitField(
+        static_cast<uint16_t>(StaticTypeBitmask::BinaryInput)
+        | static_cast<uint16_t>(StaticTypeBitmask::AnalogInput)
+        | static_cast<uint16_t>(StaticTypeBitmask::Counter));

     // You can override the default link layer settings here
     // in this example we've changed the default link layer addressing
-    config.link.LocalAddr = 10;
-    config.link.RemoteAddr = 1;
+    config.link.LocalAddr = 4;
+    config.link.RemoteAddr = 3;
     config.link.KeepAliveTimeout = TimeDuration::Max();

     // Create a new outstation with a log level, command handler, and
     // config info this        returns a thread-safe interface used for
     // updating the outstation's database.
-    auto outstation = channel->AddOutstation("outstation", SuccessCommandHandler::Create(),
+    auto crob = std::make_shared<LabCrobCommandHandler>();
+    auto outstation = channel->AddOutstation("outstation", crob,
                                              app, config);

+    // VEC-01 lab: initialize the memory-backed static database to the fixed
+    // ground truth before any protocol use: BI[0]=true, AI[0]=2300, CT[0]=42.
+    {
+        UpdateBuilder init;
+        init.Update(Binary(true), 0, EventMode::Suppress);
+        init.Update(Analog(2300.0), 0, EventMode::Suppress);
+        init.Update(Counter(42), 0, EventMode::Suppress);
+        outstation->Apply(init.Build());
+    }
+
     // Enable the outstation and start communications
     outstation->Enable();

     // variables used in example loop
     string input;
     State state;
```

## What the patch establishes (static)

| Property | Value | Reference |
|----------|-------|-----------|
| Outstation bind | `IPEndpoint("10.77.0.2", 20000)` — no wildcard | patch `AddTCPServer` line |
| Outstation link local / remote | `LocalAddr = 4`, `RemoteAddr = 3` | patch link lines |
| Static BI variation index 0 | `Group1Var1` | patch |
| Static AI variation index 0 | `Group30Var1`, `Class0` (stock `Group30Var5`/`Class2` removed) | patch |
| Static Counter variation index 0 | `Group20Var1` | patch |
| Fixed memory points | BI[0]=true, AI[0]=2300, CT[0]=42 (`EventMode::Suppress`) | patch init block |
| Unsolicited | disabled (`allowUnsolicited = false`) | patch |
| Class-0 restriction | `BinaryInput | AnalogInput | Counter` | patch |
| Memory-only BO[0] | `LabCrobCommandHandler`: DirectOperate only, index 0, LATCH_ON/OFF, `count == 1`; `SUCCESS` only after in-memory write+read-back | patch handler |
| Rejections | SBO / DirectOperateNoAck / analog outputs → `NOT_SUPPORTED`; index≠0 → `OUT_OF_RANGE`; count≠1 → `FORMAT_ERROR` | patch handler |

**No physical I/O:** no GPIO, serial, Modbus, field-bus or actuator path; the handler keeps a
single in-memory `bool state_`.

## Prospective Go master configuration (static; not executed)

The Go master is the repository implementation, configured in the lab (not modified here).
Mapping from the source at the current HEAD: `MasterAddress` → link source,
`OutstationAddress` → link destination.

| Setting | Value |
|---------|-------|
| Transport | TCP |
| Target endpoint | `10.77.0.2:20000` (from `vec01-master`) |
| Master local link address | `3` (source) |
| Outstation link address | `4` (destination) |
| Control policy | single-shot G12V1 DirectOperate; no blind retry; UNKNOWN quarantine preserved |

## Not claimed

Static inspection cannot prove the actual G1V1/G30V1/G20V1 packed wire bytes or a real
CROB state transition on the wire. The peer is **never started or connected** in this
packet. Multi-header packing, response qualifiers and sequence/confirm behaviour remain for
a later, separately authorized wire gate.
