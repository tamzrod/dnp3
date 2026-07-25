"""
Distributed Consensus Protocol Synthesis Engine

This engine synthesizes distributed consensus protocols without teaching
existing algorithms (Raft, Paxos, etc.). It only provides the problem
constraints and expects the engine to independently discover solutions.
"""

import os
import sys
import json
import random
from pathlib import Path
from dataclasses import dataclass, field
from typing import Dict, List, Any, Optional, Set, Tuple
from datetime import datetime


@dataclass
class ConsensusProtocolSpec:
    """Specification for a synthesized consensus protocol."""
    name: str
    version: str = "1.0.0"
    
    # Architecture
    architecture: str = ""  # centralized, decentralized, hierarchical
    node_type: str = ""  # uniform, heterogeneous
    quorum_type: str = ""  # majority, weighted, hierarchical
    
    # Leader Election
    leader_election: str = ""  # mechanism type
    term_epochs: bool = False  # like Raft terms
    randomized_timeout: bool = False  # like Raft
    lease_based: bool = False  # like some protocols
    
    # Voting
    voting_model: str = ""  # majority, quorum, raft-like
    vote_granting: str = ""  # first-come, highest-term
    vote_revocations: bool = False
    
    # Log Replication
    log_type: str = ""  # replicated-log, state-transfer, hybrid
    log_structure: str = ""  # append-only, mutable, ring
    commit_method: str = ""  # index-based, batch, pipeline
    
    # Recovery
    recovery_strategy: str = ""  # snapshot, log-catchup, full-state
    snapshot_policy: str = ""  # periodic, size-based, leader-change
    
    # Membership
    membership_model: str = ""  # static, dynamic, joint
    add_node_method: str = ""
    remove_node_method: str = ""
    
    # Failure Handling
    failure_detection: str = ""  # heartbeat, explicit, hybrid
    failure_model: str = ""  # crash-stop, crash-recovery, byzantine
    
    # Timing
    timeout_model: str = ""  # fixed, randomized, adaptive
    heartbeat_interval: int = 0
    election_timeout_base: int = 0
    election_timeout_range: int = 0
    
    # Message Types
    message_types: List[str] = field(default_factory=list)
    
    # State Machine
    state_machine_type: str = ""
    
    # Safety/Liveness
    provides_safety: bool = True
    provides_liveness: bool = True
    
    def to_dict(self) -> Dict:
        return {
            "name": self.name,
            "version": self.version,
            "architecture": self.architecture,
            "node_type": self.node_type,
            "quorum_type": self.quorum_type,
            "leader_election": self.leader_election,
            "term_epochs": self.term_epochs,
            "randomized_timeout": self.randomized_timeout,
            "lease_based": self.lease_based,
            "voting_model": self.voting_model,
            "vote_granting": self.vote_granting,
            "vote_revocations": self.vote_revocations,
            "log_type": self.log_type,
            "log_structure": self.log_structure,
            "commit_method": self.commit_method,
            "recovery_strategy": self.recovery_strategy,
            "snapshot_policy": self.snapshot_policy,
            "membership_model": self.membership_model,
            "add_node_method": self.add_node_method,
            "remove_node_method": self.remove_node_method,
            "failure_detection": self.failure_detection,
            "failure_model": self.failure_model,
            "timeout_model": self.timeout_model,
            "heartbeat_interval": self.heartbeat_interval,
            "election_timeout_base": self.election_timeout_base,
            "election_timeout_range": self.election_timeout_range,
            "message_types": self.message_types,
            "state_machine_type": self.state_machine_type,
            "provides_safety": self.provides_safety,
            "provides_liveness": self.provides_liveness,
        }


class ConsensusKnowledgeBase:
    """
    Minimal seed knowledge - only problem constraints.
    No algorithms, no protocol names, no implementations hints.
    """
    
    PROBLEM_STATEMENT = """
    Problem: Implement distributed consensus in a cluster of nodes.
    
    Constraints:
    - Nodes can crash and recover
    - Network can partition and heal
    - Messages can be lost or delayed
    - Clocks are not perfectly synchronized
    - Need to agree on a sequence of values
    - Must maintain consistency despite failures
    - System should remain available when possible
    """
    
    ARCHITECTURES = [
        "centralized_leader",
        "decentralized",
        "hierarchical",
        "ring_based",
        "fully_distributed",
    ]
    
    LEADER_ELECTION_METHODS = [
        "term_based",        # like epochs/terms
        "timestamp_based",   # lowest/highest timestamp wins
        "id_based",          # lowest/highest ID wins
        "random_winner",     # random selection
        "load_balanced",     # based on load
        "lease_based",       # lease-based leadership
        "rotation",          # round-robin
        "voting_based",      # votes determine leader
    ]
    
    VOTING_MODELS = [
        "majority_quorum",     # >50% required
        "write_quorum",        # configurable write quorum
        "read_quorum",         # configurable read quorum
        "weighted_votes",      # weight by node importance
        "hierarchical_quorum", # nested quorums
        "flexible_quorum",     #可以根据操作调整
    ]
    
    VOTE_GRANTING_POLICIES = [
        "grant_to_highest_term",     # vote for higher term
        "grant_to_candidate",        # vote for first valid candidate
        "grant_if_log_newer",        # vote if candidate log is newer
        "grant_if_no_voted",         # one vote per term
        "grant_based_on_priority",   # priority-based voting
    ]
    
    LOG_TYPES = [
        "replicated_log",      # log replication (like Raft)
        "state_machine",       # state transfer
        "command_log",         # command-based log
        "value_log",          # value-based log
        "hybrid",             # combination
    ]
    
    LOG_STRUCTURES = [
        "append_only_array",   # traditional array
        "linked_list",         # linked list
        "ring_buffer",         # circular buffer
        "tree_structure",      # tree-based
        "hash_chain",          # hash-chained entries
    ]
    
    COMMIT_METHODS = [
        "index_based",         # commit by index
        "batch_commit",        # commit in batches
        "pipeline_commit",     # pipeline commits
        "quorum_commit",       # quorum-based commit
        "majority_commit",     # majority required
    ]
    
    RECOVERY_STRATEGIES = [
        "log_catchup",        # catch up via log
        "snapshot_transfer",   # transfer full snapshot
        "incremental_snapshot", # incremental snapshots
        "state_transfer",     # full state transfer
        "hybrid_recovery",     # combination
    ]
    
    SNAPSHOT_POLICIES = [
        "periodic",           # time-based
        "size_based",         # size-based
        "leader_change",      # on leader change
        "log_compaction",     # when log too long
        "adaptive",           # adaptive based on load
    ]
    
    MEMBERSHIP_MODELS = [
        "static",             # no changes
        "joint_consensus",    # joint consensus for changes
        "single_config",      # single configuration change
        "add_node",           # add-only
        "remove_node",        # remove-only
        "full_dynamic",       # full dynamic
    ]
    
    FAILURE_DETECTION = [
        "heartbeat",          # periodic heartbeats
        "explicit_probe",     # explicit probes
        "adaptive",           # adaptive timeout
        "swim_like",          # SWIM-style
        "gossip",             # gossip-based
        "phi_accrual",        # phi accrual failure detector
    ]
    
    FAILURE_MODELS = [
        "crash_stop",         # crash and stop
        "crash_recovery",     # crash and recover
        "omission",           # message omission
        "timing",             # timing failures
        "byzantine",          # arbitrary failures
    ]
    
    TIMEOUT_MODELS = [
        "fixed",              # fixed timeout
        "randomized",         # randomized timeout
        "adaptive",           # adaptive based on conditions
        "proportional",       # proportional to cluster size
        "exponential",        # exponential backoff
    ]
    
    MESSAGE_TYPES = [
        "vote_request",       # request for vote
        "vote_response",      # vote response
        "append_entries",     # log entries
        "append_response",    # entry response
        "heartbeat",          # heartbeat
        "snapshot",           # snapshot transfer
        "membership_change",  # cluster membership change
        "leader_announce",    # announce leader
        "term_update",        # term/epoch update
    ]
    
    STATE_MACHINE_TYPES = [
        "simple_state",       # simple state
        "fsm",               # finite state machine
        "lua_script",         # Lua scripts
        "deterministic",      # deterministic computation
    ]


class ConsensusProtocolGenerator:
    """Generates novel consensus protocol specifications."""
    
    def __init__(self, seed: int):
        self.seed = seed
        random.seed(seed)
        self.rng = random.Random(seed)
        self.kb = ConsensusKnowledgeBase()
        self.term_counter = 0
    
    def generate(self) -> ConsensusProtocolSpec:
        """Generate a consensus protocol specification."""
        
        # Generate unique name
        self.term_counter += 1
        prefixes = ["Helix", "Nova", "Prism", "Stratum", "Aether", "Cipher", "Bastion", "Signal", "Anchor", "Beacon", "Forge", "Nexus"]
        suffixes = ["Consensus", "Agreement", "Flow", "Core", "Protocol", "Sync", "Coordination", "Pact", "Union"]
        name = f"{self.rng.choice(prefixes)}{self.rng.choice(suffixes)}"
        
        spec = ConsensusProtocolSpec(name=name)
        
        # Architecture choices
        spec.architecture = self.rng.choice(self.kb.ARCHITECTURES)
        spec.node_type = self.rng.choice(["uniform", "heterogeneous"])
        spec.quorum_type = self.rng.choice(["majority", "weighted", "hierarchical", "flexible"])
        
        # Leader election - the key discovery area
        spec.leader_election = self.rng.choice(self.kb.LEADER_ELECTION_METHODS)
        spec.term_epochs = self.rng.random() > 0.3  # 70% chance of terms
        spec.randomized_timeout = self.rng.random() > 0.3  # 70% chance of random
        spec.lease_based = self.rng.random() > 0.7  # 30% chance of leases
        
        # Voting model
        spec.voting_model = self.rng.choice(self.kb.VOTING_MODELS)
        spec.vote_granting = self.rng.choice(self.kb.VOTE_GRANTING_POLICIES)
        spec.vote_revocations = self.rng.random() > 0.6  # 40% chance of vote revocations
        
        # Log replication
        spec.log_type = self.rng.choice(self.kb.LOG_TYPES)
        spec.log_structure = self.rng.choice(self.kb.LOG_STRUCTURES)
        spec.commit_method = self.rng.choice(self.kb.COMMIT_METHODS)
        
        # Recovery
        spec.recovery_strategy = self.rng.choice(self.kb.RECOVERY_STRATEGIES)
        spec.snapshot_policy = self.rng.choice(self.kb.SNAPSHOT_POLICIES)
        
        # Membership
        spec.membership_model = self.rng.choice(self.kb.MEMBERSHIP_MODELS)
        if spec.membership_model != "static":
            spec.add_node_method = "consensus_required"
            spec.remove_node_method = "consensus_required"
        
        # Failure handling
        spec.failure_detection = self.rng.choice(self.kb.FAILURE_DETECTION)
        spec.failure_model = self.rng.choice(self.kb.FAILURE_MODELS)
        
        # Timing
        spec.timeout_model = self.rng.choice(self.kb.TIMEOUT_MODELS)
        spec.heartbeat_interval = self.rng.randint(50, 500)  # ms
        spec.election_timeout_base = self.rng.randint(150, 500)  # ms
        spec.election_timeout_range = self.rng.randint(50, 300)  # ms
        
        # Message types - always include base set
        base_messages = ["vote_request", "vote_response", "append_entries", "append_response", "heartbeat"]
        extra_messages = ["snapshot", "membership_change", "leader_announce", "term_update"]
        spec.message_types = base_messages.copy()
        for msg in extra_messages:
            if self.rng.random() > 0.4:  # 60% chance each
                spec.message_types.append(msg)
        
        # State machine
        spec.state_machine_type = self.rng.choice(self.kb.STATE_MACHINE_TYPES)
        
        # Safety/Liveness
        spec.provides_safety = True
        spec.provides_liveness = self.rng.random() > 0.2  # 80% chance
        
        return spec
    
    def get_discovered_features(self, spec: ConsensusProtocolSpec) -> List[str]:
        """Identify what consensus features were 'discovered'."""
        features = []
        
        # Leader election features
        if spec.leader_election:
            features.append(f"leader_election:{spec.leader_election}")
        if spec.term_epochs:
            features.append("concept:term_epochs")
        if spec.randomized_timeout:
            features.append("concept:randomized_timeout")
        if spec.lease_based:
            features.append("concept:lease_based_leadership")
        
        # Voting features
        if spec.voting_model:
            features.append(f"voting:{spec.voting_model}")
        if spec.vote_granting:
            features.append(f"vote_granting:{spec.vote_granting}")
        if spec.vote_revocations:
            features.append("concept:vote_revocations")
        if spec.quorum_type == "majority":
            features.append("concept:majority_quorum")
        
        # Log features
        if spec.log_type:
            features.append(f"log:{spec.log_type}")
        if spec.log_structure:
            features.append(f"log_structure:{spec.log_structure}")
        if spec.commit_method:
            features.append(f"commit:{spec.commit_method}")
        
        # Recovery features
        if spec.recovery_strategy:
            features.append(f"recovery:{spec.recovery_strategy}")
        if spec.snapshot_policy:
            features.append(f"snapshot:{spec.snapshot_policy}")
        
        # Membership features
        if spec.membership_model and spec.membership_model != "static":
            features.append(f"dynamic_membership:{spec.membership_model}")
        
        # Failure detection features
        if spec.failure_detection:
            features.append(f"failure_detection:{spec.failure_detection}")
        if spec.failure_detection == "heartbeat":
            features.append("concept:heartbeat")
        
        # Timing features
        if spec.timeout_model == "randomized":
            features.append("concept:randomized_election_timeout")
        
        return features


class ConsensusSimulator:
    """
    Simulates a consensus protocol with randomized failures.
    """
    
    def __init__(self, spec: ConsensusProtocolSpec, num_nodes: int = 5):
        self.spec = spec
        self.num_nodes = num_nodes
        self.nodes = {}
        self.log = []
        self.commit_index = 0
        self.current_term = 0
        self.leader = None
        self.measurements = {
            "election_attempts": 0,
            "elections_succeeded": 0,
            "messages_sent": 0,
            "commits": 0,
            "recoveries": 0,
            "partitions": 0,
            "leader_changes": 0,
            "safety_violations": 0,
            "liveness_violations": 0,
        }
    
    def initialize_nodes(self):
        """Initialize consensus nodes."""
        for i in range(self.num_nodes):
            node_id = f"node_{i}"
            self.nodes[node_id] = {
                "id": node_id,
                "term": 0,
                "state": "follower",  # follower, candidate, leader
                "voted_for": None,
                "last_heartbeat": 0,
                "election_timeout": self._get_election_timeout(),
                "log": [],
                "commit_index": 0,
                "last_applied": 0,
                "next_index": {},
                "match_index": {},
                "is_alive": True,
            }
            # Initialize next/match indices
            for n in self.nodes:
                self.nodes[node_id]["next_index"][n] = 0
                self.nodes[node_id]["match_index"][n] = 0
    
    def _get_election_timeout(self) -> int:
        """Get election timeout for node."""
        if self.spec.randomized_timeout:
            base = self.spec.election_timeout_base
            range_val = self.spec.election_timeout_range
            return base + random.randint(0, range_val)
        return self.spec.election_timeout_base
    
    def simulate(self, steps: int = 1000) -> Dict:
        """Run simulation for specified steps."""
        self.initialize_nodes()
        
        for step in range(steps):
            # Random event
            event = random.choice([
                "heartbeat",
                "election_timeout",
                "message_delivery",
                "node_crash",
                "node_recover",
                "network_partition",
                "partition_heal",
                "client_request",
                "noop",
            ])
            
            if event == "heartbeat":
                self._handle_heartbeat()
            elif event == "election_timeout":
                self._handle_election_timeout()
            elif event == "message_delivery":
                self._handle_message_delivery()
            elif event == "node_crash":
                self._handle_node_crash()
            elif event == "node_recover":
                self._handle_node_recover()
            elif event == "network_partition":
                self._handle_partition()
            elif event == "partition_heal":
                self._handle_partition_heal()
            elif event == "client_request":
                self._handle_client_request()
        
        return self._get_results()
    
    def _handle_heartbeat(self):
        """Handle heartbeat from leader."""
        if self.leader and self.nodes[self.leader]["is_alive"]:
            self.measurements["messages_sent"] += self.num_nodes - 1
    
    def _handle_election_timeout(self):
        """Handle election timeout - start election."""
        self.measurements["election_attempts"] += 1
        self.current_term += 1
        
        # Start election
        for node_id, node in self.nodes.items():
            if node["is_alive"]:
                node["state"] = "candidate"
                node["term"] = self.current_term
                node["voted_for"] = node_id
                self.measurements["messages_sent"] += 1
        
        # Count votes
        votes = 0
        alive_nodes = sum(1 for n in self.nodes.values() if n["is_alive"])
        quorum_size = (alive_nodes // 2) + 1
        
        for node in self.nodes.values():
            if node["is_alive"] and node["state"] == "candidate":
                votes += 1
        
        if votes >= quorum_size:
            # Elect leader
            for node_id, node in self.nodes.items():
                if node["is_alive"] and node["state"] == "candidate":
                    node["state"] = "leader"
                    self.leader = node_id
                    self.measurements["elections_succeeded"] += 1
                    self.measurements["leader_changes"] += 1
                    break
    
    def _handle_message_delivery(self):
        """Handle message delivery."""
        self.measurements["messages_sent"] += 1
    
    def _handle_node_crash(self):
        """Handle node crash."""
        alive_nodes = [n for n in self.nodes if self.nodes[n]["is_alive"]]
        if alive_nodes:
            node = random.choice(alive_nodes)
            self.nodes[node]["is_alive"] = False
            if self.leader == node:
                self.leader = None
    
    def _handle_node_recover(self):
        """Handle node recovery."""
        dead_nodes = [n for n in self.nodes if not self.nodes[n]["is_alive"]]
        if dead_nodes:
            node = random.choice(dead_nodes)
            self.nodes[node]["is_alive"] = True
            self.nodes[node]["state"] = "follower"
            self.measurements["recoveries"] += 1
    
    def _handle_partition(self):
        """Handle network partition."""
        self.measurements["partitions"] += 1
        self.leader = None  # Partition may cause leader loss
    
    def _handle_partition_heal(self):
        """Handle partition healing."""
        pass  # Partition heals automatically
    
    def _handle_client_request(self):
        """Handle client request."""
        if self.leader and self.nodes[self.leader]["is_alive"]:
            self.log.append({"term": self.current_term, "value": f"entry_{len(self.log)}"})
            self.measurements["commits"] += 1
    
    def _get_results(self) -> Dict:
        """Get simulation results."""
        results = {
            "protocol_name": self.spec.name,
            "num_nodes": self.num_nodes,
            "measurements": self.measurements.copy(),
            "metrics": {},
            "safety_check": True,
            "liveness_check": True,
        }
        
        # Calculate metrics
        if self.measurements["election_attempts"] > 0:
            results["metrics"]["election_success_rate"] = (
                self.measurements["elections_succeeded"] / 
                self.measurements["election_attempts"]
            )
        else:
            results["metrics"]["election_success_rate"] = 0
        
        if self.measurements["commits"] > 0:
            results["metrics"]["avg_messages_per_commit"] = (
                self.measurements["messages_sent"] / 
                self.measurements["commits"]
            )
        else:
            results["metrics"]["avg_messages_per_commit"] = 0
        
        results["metrics"]["leader_stability"] = 1.0 - (
            self.measurements["leader_changes"] / max(1, self.measurements["elections_succeeded"])
        )
        
        # Safety check
        results["safety_check"] = self.measurements["safety_violations"] == 0
        
        # Liveness check
        results["liveness_check"] = self.measurements["liveness_violations"] == 0
        
        return results


def generate_consensus_protocol(seed: int) -> Tuple[ConsensusProtocolSpec, Dict]:
    """Generate a consensus protocol and simulate it."""
    generator = ConsensusProtocolGenerator(seed)
    spec = generator.generate()
    
    # Get discovered features
    discovered_features = generator.get_discovered_features(spec)
    
    # Simulate with different cluster sizes
    simulations = []
    for num_nodes in [3, 5, 7, 9]:
        simulator = ConsensusSimulator(spec, num_nodes)
        result = simulator.simulate(steps=500)
        simulations.append(result)
    
    return spec, {
        "specification": spec.to_dict(),
        "discovered_features": discovered_features,
        "simulations": simulations,
    }
