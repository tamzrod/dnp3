"""
Protocol Generator
Generates novel secure communication protocol architectures.
"""

import hashlib
import random
import json
from typing import Dict, List, Any, Optional
from dataclasses import dataclass, field, asdict
from protocol_knowledge import PROTOCOL_PATTERNS, EXISTING_PROTOCOLS, ATTACK_VECTORS


@dataclass
class ProtocolSpec:
    """Complete protocol specification."""
    name: str
    version: str = "1.0"
    
    # Architecture
    architecture_type: str = ""
    handshake_model: str = ""
    key_exchange_type: str = ""
    authentication_model: str = ""
    encryption_mode: str = ""
    forward_secrecy_level: str = ""
    message_format: str = ""
    
    # Cryptographic primitives
    key_exchange_algorithm: str = ""
    authentication_algorithm: str = ""
    encryption_algorithm: str = ""
    hash_algorithm: str = ""
    kdf_algorithm: str = ""
    
    # Features
    supports_0rtt: bool = False
    supports_resumption: bool = False
    supports_post_quantum: bool = False
    connection_migration: bool = False
    
    # Security properties
    confidentiality: bool = True
    integrity: bool = True
    authenticity: bool = True
    forward_secrecy: bool = True
    replay_protection: bool = True
    
    # Design decisions
    distinct_features: List[str] = field(default_factory=list)
    handshake_rounds: int = 0
    message_overhead: int = 0


class ProtocolGenerator:
    """Generates novel protocol architectures."""
    
    def __init__(self, seed: int):
        self.seed = seed
        random.seed(seed)
        self.rng = random.Random(seed)
        
    def _hash_seed(self, data: str) -> int:
        """Create deterministic hash from seed and data."""
        h = hashlib.sha256(f"{self.seed}_{data}".encode()).hexdigest()
        return int(h, 16)
    
    def _choose(self, options: List[str], bias_towards: List[str] = None) -> str:
        """Choose from options, optionally with bias."""
        if bias_towards:
            if self.rng.random() < 0.3:
                return self.rng.choice(bias_towards)
        return self.rng.choice(options)
    
    def _generate_protocol_name(self) -> str:
        """Generate a unique protocol name."""
        prefixes = ["Nova", "Aether", "Stratum", "Nexus", "Helix", "Prism", 
                    "Cipher", "Anchor", "Bastion", "Forge", "Beacon", "Signal"]
        suffixes = ["Handshake", "Flow", "Link", "Channel", "Protocol", 
                    "Session", "Vault", "Shield", "Core", "Mesh"]
        modifiers = ["", "Secure", "Fast", "Hybrid", "Quantum", "Zero", "Stateless"]
        
        prefix = self.rng.choice(prefixes)
        modifier = self.rng.choice(modifiers)
        suffix = self.rng.choice(suffixes)
        
        if modifier:
            return f"{prefix}{modifier}{suffix}"
        return f"{prefix}{suffix}"
    
    def _select_key_exchange(self) -> tuple:
        """Select key exchange algorithm."""
        options = [
            ("X25519", "X25519", "ECDH on Curve25519"),
            ("X448", "X448", "ECDH on Curve448"),
            ("P256_ECDHE", "P-256", "ECDH on NIST P-256"),
            ("P384_ECDHE", "P-384", "ECDH on NIST P-384"),
        ]
        
        if self.rng.random() < 0.3:  # 30% post-quantum
            options.extend([
                ("Kyber512", "Kyber768", "ML-KEM-768 (formerly Kyber)"),
                ("Kyber768", "Kyber1024", "ML-KEM-1024"),
                ("X25519_Kyber768", "X25519+Kyber768", "Hybrid post-quantum"),
            ])
        
        idx = self.rng.randint(0, len(options) - 1)
        return options[idx]
    
    def _select_authentication(self) -> tuple:
        """Select authentication algorithm."""
        options = [
            ("Ed25519", "Ed25519", "Edwards-curve Digital Signature"),
            ("Ed448", "Ed448", "Edwards-curve Digital Signature 448"),
            ("RSA_PSS_2048", "RSA-PSS 2048", "RSA Signature with PSS padding"),
            ("ECDSA_P256", "ECDSA P-256", "Elliptic Curve Digital Signature"),
        ]
        
        if self.rng.random() < 0.2:  # 20% post-quantum
            options.extend([
                ("Dilithium2", "Dilithium2", "CRYSTALS-Dilithium (L2)"),
                ("Dilithium3", "Dilithium3", "CRYSTALS-Dilithium (L3)"),
            ])
        
        idx = self.rng.randint(0, len(options) - 1)
        return options[idx]
    
    def _select_encryption(self) -> tuple:
        """Select encryption algorithm."""
        options = [
            ("AES_128_GCM", "AES-128-GCM", "AES in GCM mode with 128-bit key"),
            ("AES_256_GCM", "AES-256-GCM", "AES in GCM mode with 256-bit key"),
            ("ChaCha20", "ChaCha20-Poly1305", "ChaCha20 stream cipher with Poly1305 MAC"),
        ]
        
        idx = self.rng.randint(0, len(options) - 1)
        return options[idx]
    
    def _select_hash(self) -> tuple:
        """Select hash algorithm."""
        options = [
            ("SHA256", "SHA-256", "Secure Hash Algorithm 256-bit"),
            ("SHA384", "SHA-384", "Secure Hash Algorithm 384-bit"),
            ("SHA512", "SHA-512", "Secure Hash Algorithm 512-bit"),
            ("BLAKE3", "BLAKE3", "BLAKE3 cryptographic hash"),
        ]
        idx = self.rng.randint(0, len(options) - 1)
        return options[idx]
    
    def _select_kdf(self) -> tuple:
        """Select key derivation function."""
        options = [
            ("HKDF", "HKDF-SHA256", "HMAC-based KDF"),
            ("HKDF_SHA384", "HKDF-SHA384", "HMAC-based KDF with SHA-384"),
            ("HKDF_SHA512", "HKDF-SHA512", "HMAC-based KDF with SHA-512"),
        ]
        
        if self.rng.random() < 0.1:  # 10% memory-hard
            options.extend([
                ("Argon2", "Argon2id", "Memory-hard KDF"),
                ("Scrypt", "Scrypt", "Memory-hard KDF"),
            ])
        
        idx = self.rng.randint(0, len(options) - 1)
        return options[idx]
    
    def _generate_handshake_sequence(self, spec: ProtocolSpec) -> List[Dict]:
        """Generate handshake message sequence."""
        sequences = []
        
        if spec.handshake_model == "full_duplex":
            sequences = [
                {"from": "client", "to": "server", "type": "ClientHello", 
                 "content": "supported_versions, key_share, cipher_suites",
                 "encrypted": False},
                {"from": "server", "to": "client", "type": "ServerHello",
                 "content": "selected_version, key_share, selected_cipher",
                 "encrypted": False},
                {"from": "server", "to": "client", "type": "EncryptedExtensions",
                 "content": "server_extensions", "encrypted": True},
                {"from": "server", "to": "client", "type": "Certificate",
                 "content": "server_certificates", "encrypted": True},
                {"from": "server", "to": "client", "type": "CertificateVerify",
                 "content": "signature", "encrypted": True},
                {"from": "server", "to": "client", "type": "Finished",
                 "content": "handshake_hash", "encrypted": True},
                {"from": "client", "to": "server", "type": "Finished",
                 "content": "handshake_hash", "encrypted": True},
            ]
            spec.handshake_rounds = 1
            spec.message_overhead = 96  # typical TLS record overhead
            
        elif spec.handshake_model == "simplex":
            sequences = [
                {"from": "initiator", "to": "responder", "type": "Init",
                 "content": "version, key_share, nonce",
                 "encrypted": False},
                {"from": "initiator", "to": "responder", "type": "Auth",
                 "content": "certificate, signature", "encrypted": True},
                {"from": "initiator", "to": "responder", "type": "Data",
                 "content": "application_data", "encrypted": True},
            ]
            spec.handshake_rounds = 0  # 0-RTT from initiator
            spec.message_overhead = 32
            
        elif spec.handshake_model == "cookie":
            sequences = [
                {"from": "initiator", "to": "server", "type": "HandshakeInit",
                 "content": "ephemeral_key, cookies", "encrypted": False},
                {"from": "server", "to": "initiator", "type": "CookieResponse",
                 "content": "ack, anti_replay_nonce", "encrypted": False},
                {"from": "initiator", "to": "server", "type": "ConfirmedInit",
                 "content": "verified_cookie", "encrypted": True},
                {"from": "server", "to": "initiator", "type": "SessionTicket",
                 "content": "session_keys, ticket", "encrypted": True},
            ]
            spec.handshake_rounds = 2
            spec.message_overhead = 64
            
        elif spec.handshake_model == "resumption":
            sequences = [
                {"from": "client", "to": "server", "type": "ClientHelloPSK",
                 "content": "ticket, psk_binder", "encrypted": False},
                {"from": "server", "to": "client", "type": "ServerHello",
                 "content": "selected_identity", "encrypted": False},
                {"from": "server", "to": "client", "type": "Finished",
                 "content": "finished", "encrypted": True},
            ]
            spec.handshake_rounds = 0
            spec.message_overhead = 48
            
        elif spec.handshake_model == "dh_ratchet":
            sequences = [
                {"from": "alice", "to": "bob", "type": "RatchetInit",
                 "content": "dh_public, chain_key",
                 "encrypted": False},
                {"from": "bob", "to": "alice", "type": "RatchetResponse",
                 "content": "dh_public, message_key_1",
                 "encrypted": False},
                {"from": "alice", "to": "bob", "type": "RatchetStep",
                 "content": "dh_public, message_key_2",
                 "encrypted": False},
            ]
            spec.handshake_rounds = 2
            spec.message_overhead = 80
            
        else:  # default/novel
            sequences = [
                {"from": "client", "to": "server", "type": "Hello",
                 "content": "versions, capabilities",
                 "encrypted": False},
                {"from": "server", "to": "client", "type": "AckHello",
                 "content": "selection",
                 "encrypted": False},
                {"from": "client", "to": "server", "type": "KeyExchange",
                 "content": "public_key_material",
                 "encrypted": False},
                {"from": "server", "to": "client", "type": "Auth",
                 "content": "certificate, verify",
                 "encrypted": True},
                {"from": "client", "to": "server", "type": "AuthAck",
                 "content": "verify",
                 "encrypted": True},
            ]
            spec.handshake_rounds = 2
            spec.message_overhead = 56
        
        return sequences
    
    def _generate_threat_model(self, spec: ProtocolSpec) -> Dict:
        """Generate threat model for the protocol."""
        threats = []
        
        # Always mitigate these
        always_mitigate = [
            "Eavesdropping (confidentiality)",
            "Man-in-the-middle attack",
            "Message tampering (integrity)",
            "Session hijacking",
            "Replay attacks",
        ]
        threats.extend(always_mitigate)
        
        # Conditional based on features
        if spec.forward_secrecy:
            threats.append("Forward secrecy against key compromise")
        
        if spec.supports_post_quantum:
            threats.append("Harvest now, decrypt later (HNDDL)")
        
        if spec.connection_migration:
            threats.append("Connection ID hijacking")
        
        if spec.supports_0rtt:
            threats.extend([
                "Replay of 0-RTT data",
                "Anti-replay attacks",
            ])
        
        return {
            "threats_addressed": threats,
            "assumptions": [
                "Cryptographic primitives are secure",
                "Random number generator is secure",
                "Endpoints are not compromised",
                "Public key infrastructure is trustworthy (if used)",
            ],
            "security_goals": [
                f"Confidentiality: {spec.confidentiality}",
                f"Integrity: {spec.integrity}",
                f"Authenticity: {spec.authenticity}",
                f"Forward Secrecy: {spec.forward_secrecy}",
                f"Replay Protection: {spec.replay_protection}",
            ],
        }
    
    def _generate_state_machine(self, spec: ProtocolSpec) -> Dict:
        """Generate protocol state machine."""
        states = {
            "states": [
                {"name": "INIT", "description": "Initial state, no connection"},
                {"name": "HANDSHAKING", "description": "Key exchange in progress"},
                {"name": "AUTHENTICATING", "description": "Authentication in progress"},
                {"name": "ESTABLISHED", "description": "Secure session established"},
                {"name": "CLOSING", "description": "Graceful shutdown"},
                {"name": "CLOSED", "description": "Connection terminated"},
            ],
            "transitions": [
                {"src": "INIT", "dst": "HANDSHAKING", "event": "connect/start"},
                {"src": "HANDSHAKING", "dst": "AUTHENTICATING", "event": "key_exchange_complete"},
                {"src": "HANDSHAKING", "dst": "CLOSED", "event": "key_exchange_fail"},
                {"src": "AUTHENTICATING", "dst": "ESTABLISHED", "event": "auth_success"},
                {"src": "AUTHENTICATING", "dst": "CLOSED", "event": "auth_fail"},
                {"src": "ESTABLISHED", "dst": "ESTABLISHED", "event": "rekey"},
                {"src": "ESTABLISHED", "dst": "CLOSING", "event": "close"},
                {"src": "CLOSING", "dst": "CLOSED", "event": "close_complete"},
                {"src": "CLOSED", "dst": "INIT", "event": "reset"},
            ],
        }
        
        # Add 0-RTT state if supported
        if spec.supports_0rtt:
            states["states"].append(
                {"name": "EARLY_DATA", "description": "0-RTT data being sent/received"}
            )
            states["transitions"].append(
                {"src": "INIT", "dst": "EARLY_DATA", "event": "send_0rtt"}
            )
            states["transitions"].append(
                {"src": "EARLY_DATA", "dst": "HANDSHAKING", "event": "0rtt_complete"}
            )
        
        # Add migration state if supported
        if spec.connection_migration:
            states["states"].append(
                {"name": "MIGRATING", "description": "Connection migration in progress"}
            )
            states["transitions"].append(
                {"src": "ESTABLISHED", "dst": "MIGRATING", "event": "migrate"}
            )
            states["transitions"].append(
                {"src": "MIGRATING", "dst": "ESTABLISHED", "event": "migration_complete"}
            )
        
        return states
    
    def _generate_packet_structure(self, spec: ProtocolSpec) -> Dict:
        """Generate packet structure specification."""
        overhead = spec.message_overhead
        
        packet_types = {
            "HANDSHAKE": {
                "header_size": overhead,
                "fields": [
                    {"name": "type", "size": 1, "description": "Packet type"},
                    {"name": "version", "size": 2, "description": "Protocol version"},
                    {"name": "length", "size": 2, "description": "Payload length"},
                    {"name": "payload", "size": "variable", "description": "Encrypted payload"},
                    {"name": "tag", "size": 16, "description": "Authentication tag (for AEAD)"},
                ],
            },
            "DATA": {
                "header_size": overhead - 16,
                "fields": [
                    {"name": "type", "size": 1, "description": "Packet type"},
                    {"name": "connection_id", "size": 8, "description": "Connection identifier"},
                    {"name": "packet_number", "size": 4, "description": "Packet sequence number"},
                    {"name": "length", "size": 2, "description": "Payload length"},
                    {"name": "payload", "size": "variable", "description": "Encrypted payload"},
                ],
            },
        }
        
        return {
            "packet_types": packet_types,
            "encoding": "TLV" if spec.message_format == "tlv" else "binary",
            "byte_order": "big_endian",
        }
    
    def _generate_security_analysis(self, spec: ProtocolSpec) -> Dict:
        """Generate self-criticism and security analysis."""
        advantages = []
        weaknesses = []
        self_criticism = []
        
        # Analyze based on design choices
        if spec.handshake_rounds == 0:
            advantages.append("Zero round-trip time for session establishment")
            if spec.supports_0rtt:
                weaknesses.append("0-RTT data vulnerable to replay attacks")
                self_criticism.append(
                    "Replay attack is possible: attacker can capture and replay "
                    "0-RTT messages. Must implement anti-replay mechanism."
                )
        
        if spec.forward_secrecy:
            advantages.append("Forward secrecy protects past sessions if keys compromised")
        
        if spec.supports_post_quantum:
            advantages.append("Post-quantum key exchange resists quantum computer attacks")
            weaknesses.append("Larger key sizes may increase latency")
            self_criticism.append(
                "Post-quantum algorithms are relatively new; trust in their security "
                "is based on current cryptanalysis, not decades of scrutiny."
            )
        
        if spec.connection_migration:
            advantages.append("Connection survives IP/port changes")
            weaknesses.append("Migration could be exploited for tracking")
            self_criticism.append(
                "Connection IDs must be unlinkable to prevent tracking across migrations."
            )
        
        # Generic analysis
        advantages.extend([
            f"Uses {spec.encryption_algorithm} for authenticated encryption",
            f"Forward secrecy via {spec.key_exchange_algorithm}",
            "AEAD mode prevents padding oracle attacks",
        ])
        
        weaknesses.extend([
            "Requires secure implementation to avoid side channels",
            "Random number generation is critical for security",
            "Certificate validation requires proper trust store management",
        ])
        
        self_criticism.extend([
            "Can authentication fail if certificates are expired or revoked? YES - "
            "implementation must check validity.",
            "Can keys be reused? Only if implementation allows; spec mandates ephemeral keys.",
            "Does forward secrecy actually exist? Only if ephemeral keys are truly ephemeral "
            "and not stored.",
            "Can replay succeed? If anti-replay is not implemented, YES.",
            "What assumptions are unrealistic? Perfect forward secrecy assumes keys are truly "
            "erased after use.",
            "Would this be difficult to implement safely? YES - constant-time operations, "
            "secure memory handling are challenging.",
        ])
        
        return {
            "advantages": advantages,
            "weaknesses": weaknesses,
            "self_criticism": self_criticism,
            "fatal_flaws": [],
        }
    
    def _compare_to_existing(self, spec: ProtocolSpec) -> Dict:
        """Compare to existing protocols."""
        comparisons = []
        
        # Compare to TLS 1.3
        tls_sim = 0
        if spec.handshake_rounds <= 1:
            tls_sim += 0.2
        if spec.forward_secrecy:
            tls_sim += 0.3
        if spec.encryption_mode == "aead":
            tls_sim += 0.2
        if spec.message_format == "tls_like":
            tls_sim += 0.3
        
        comparisons.append({
            "protocol": "TLS 1.3",
            "similarity": tls_sim,
            "differences": [
                "Novel handshake pattern" if spec.handshake_model != "full_duplex" else None,
                f"Alternative key exchange: {spec.key_exchange_algorithm}" if "X25519" not in spec.key_exchange_algorithm else None,
                "Post-quantum hybrid" if spec.supports_post_quantum else None,
            ],
        })
        
        # Compare to Noise
        noise_sim = 0
        if spec.message_format == "noise_like":
            noise_sim += 0.3
        if "pattern" in spec.handshake_model:
            noise_sim += 0.3
        
        comparisons.append({
            "protocol": "Noise Protocol Framework",
            "similarity": noise_sim,
            "differences": [
                "Fixed cipher suite" if noise_sim < 0.5 else "Similar pattern-based approach",
                f"Custom {spec.encryption_algorithm} integration",
            ],
        })
        
        # Compare to WireGuard
        wg_sim = 0
        if spec.handshake_model == "cookie":
            wg_sim += 0.3
        if not spec.authenticity or spec.authentication_model == "psk":
            wg_sim += 0.3
        
        comparisons.append({
            "protocol": "WireGuard",
            "similarity": wg_sim,
            "differences": [
                "Different key exchange mechanism",
                f"Certificate-based auth" if spec.authentication_model == "cert" else None,
                "No cryptokey routing model" if wg_sim < 0.5 else None,
            ],
        })
        
        return {
            "comparisons": comparisons,
            "novelty_score": 1.0 - max(tls_sim, noise_sim, wg_sim),
            "summary": f"This protocol is {int((1-max(tls_sim, noise_sim, wg_sim))*100)}% "
                      f"architecturally distinct from existing protocols.",
        }
    
    def generate(self) -> ProtocolSpec:
        """Generate a complete protocol specification."""
        # Initialize spec with protocol name
        spec = ProtocolSpec(name=self._generate_protocol_name())
        
        # Architecture choices
        spec.handshake_model = self._choose(PROTOCOL_PATTERNS["handshake_models"])
        spec.key_exchange_type = self._choose(PROTOCOL_PATTERNS["key_exchange_types"])
        spec.authentication_model = self._choose(PROTOCOL_PATTERNS["authentication_models"])
        spec.encryption_mode = "aead"  # Always AEAD for modern protocols
        spec.forward_secrecy_level = self._choose(
            PROTOCOL_PATTERNS["forward_secrecy_levels"],
            bias_towards=["strong"]  # Bias toward strong forward secrecy
        )
        spec.message_format = self._choose(PROTOCOL_PATTERNS["message_formats"])
        
        # Cryptographic primitives
        kex = self._select_key_exchange()
        spec.key_exchange_algorithm = kex[0]
        
        auth = self._select_authentication()
        spec.authentication_algorithm = auth[0]
        
        enc = self._select_encryption()
        spec.encryption_algorithm = enc[0]
        
        hash_alg = self._select_hash()
        spec.hash_algorithm = hash_alg[0]
        
        kdf = self._select_kdf()
        spec.kdf_algorithm = kdf[0]
        
        # Features
        spec.supports_0rtt = self.rng.random() < 0.3
        spec.supports_resumption = self.rng.random() < 0.7
        spec.supports_post_quantum = "Kyber" in spec.key_exchange_algorithm or "Dilithium" in spec.authentication_algorithm
        spec.connection_migration = self.rng.random() < 0.3
        
        # Security properties
        spec.forward_secrecy = spec.forward_secrecy_level in ["strong", "weak"]
        
        # Distinctive features
        if spec.supports_post_quantum:
            spec.distinct_features.append("Post-quantum key exchange")
        if spec.connection_migration:
            spec.distinct_features.append("Connection migration")
        if spec.handshake_rounds == 0:
            spec.distinct_features.append("Zero round-trip handshake")
        if spec.message_format == "tlv":
            spec.distinct_features.append("TLV extensibility")
        
        return spec
    
    def generate_full_specification(self) -> Dict:
        """Generate complete protocol specification with all details."""
        spec = self.generate()
        
        return {
            "protocol": asdict(spec),
            "handshake_sequence": self._generate_handshake_sequence(spec),
            "threat_model": self._generate_threat_model(spec),
            "state_machine": self._generate_state_machine(spec),
            "packet_structure": self._generate_packet_structure(spec),
            "security_analysis": self._generate_security_analysis(spec),
            "comparison": self._compare_to_existing(spec),
            "implementation_plan": self._generate_implementation_plan(spec),
        }
    
    def _generate_implementation_plan(self, spec: ProtocolSpec) -> Dict:
        """Generate implementation plan."""
        return {
            "language": "Python 3",
            "dependencies": [
                "cryptography (for crypto primitives)",
                " asyncio (for networking)",
            ],
            "phases": [
                {
                    "phase": 1,
                    "name": "Core Cryptography",
                    "tasks": [
                        f"Implement {spec.key_exchange_algorithm} key exchange",
                        f"Implement {spec.authentication_algorithm} signatures",
                        f"Implement {spec.encryption_algorithm} AEAD",
                        f"Implement {spec.kdf_algorithm} KDF",
                    ],
                },
                {
                    "phase": 2,
                    "name": "Protocol State Machine",
                    "tasks": [
                        "Implement state transitions",
                        "Handle connection lifecycle",
                        "Implement error handling",
                    ],
                },
                {
                    "phase": 3,
                    "name": "Message Framing",
                    "tasks": [
                        "Implement packet encoding/decoding",
                        "Implement record layer",
                        "Implement fragmentation",
                    ],
                },
                {
                    "phase": 4,
                    "name": "Handshake Protocol",
                    "tasks": [
                        "Implement handshake message types",
                        "Implement key schedule",
                        "Implement authentication",
                    ],
                },
                {
                    "phase": 5,
                    "name": "Testing",
                    "tasks": [
                        "Unit tests for crypto",
                        "Integration tests",
                        "Security tests",
                    ],
                },
            ],
        }
