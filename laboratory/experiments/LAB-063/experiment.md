# LAB-063: Workbench Binary Integration Test

**Engine**: KDE-ENGINE-001 (v1.0.0)  
**Seed**: SEED-001 (v1.0.0)  
**Timestamp**: 2026-07-29T08:25:00Z  
**Status**: 🔬 IN_PROGRESS

## Hypothesis

Two separate workbench binaries can successfully communicate over TCP using tmux automation:
1. Outstation listens on port 20004
2. Master connects to outstation
3. TCP communication established
4. Data flows correctly between binaries

## Evidence Collection Plan

### Method
Use tmux with `script` command to record terminal sessions:
```bash
tmux new-session -d -s dnp3-test './workbench-fixed -mode outstation -port 20004'
tmux send-keys -t dnp3-test:0 's' C-m  # Start server
tmux new-session -d -s dnp3-master './workbench-fixed -mode master -port 20004'
tmux send-keys -t dnp3-master:0 's' C-m  # Connect
sleep 2
tmux capture-pane -t dnp3-test -p > evidence/run1-outstation.txt
tmux capture-pane -t dnp3-master -p > evidence/run1-master.txt
```

## Run Records

| Run | Date | Status | Evidence |
|-----|------|--------|----------|
| 1 | 2026-07-29T08:28:00Z | IN_PROGRESS | - |
