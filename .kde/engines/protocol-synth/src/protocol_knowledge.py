"""
Protocol Knowledge Base
Studies existing protocols and extracts architectural patterns.
"""

# Protocol architectural patterns
PROTOCOL_PATTERNS = {
    "handshake_models": [
        "simplex",       # One-way initiation
        "full_duplex",   # Mutual simultaneous initiation
        "hello_request", # Server-initiated resumption
        "resumption",    # Session ticket based
    ],
    "key_exchange_types": [
        "static_static",  # Both parties have fixed keys
        "static_ephemeral",  # One party has fixed, one ephemeral
        "ephemeral_ephemeral",  # Both ephemeral (forward secret)
    ],
    "authentication_models": [
        "psk",           # Pre-shared key
        "cert",          # Certificate-based
        "auth_ephemeral", # Auth + ephemeral
        "ppk",           # Post-quantum key exchange
    ],
    "encryption_modes": [
        "aead",          # Authenticated encryption with associated data
        " Encrypt_then_MAC",  # EtM
        "MAC_then_encrypt",   # MtE
    ],
    "forward_secrecy_levels": [
        "none",          # No forward secrecy
        "weak",          # Partial forward secrecy
        "strong",        # Full forward secrecy with ratcheting
    ],
    "message_formats": [
        "tls_like",      # TLS record structure
        "noise_like",    # Noise protocol format
        "binary_fixed",   # Fixed-size binary
        "tlv",           # Type-Length-Value encoding
        "variable",      # Variable-length header
    ],
}

# Known protocol architectures
EXISTING_PROTOCOLS = {
    "TLS_1_3": {
        "handshake": "full_duplex",
        "key_exchange": "ephemeral_ephemeral",
        "auth": "cert",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "tls_like",
        "resumption": "psk_ticket",
        "distinctive_features": [
            "1-RTT handshake (1 round trip)",
            "AEAD-only cipher suites",
            "Encrypted handshake after ServerHello",
            "HelloRetryRequest for key share selection",
        ],
    },
    "QUIC": {
        "handshake": "full_duplex",
        "key_exchange": "ephemeral_ephemeral",
        "auth": "cert",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "binary_fixed",
        "resumption": "token_0rtt",
        "distinctive_features": [
            "UDP-based transport",
            "0-RTT data support",
            "Connection migration",
            "Multiplexed streams",
            "Packet number space separation",
        ],
    },
    "Noise": {
        "handshake": "pattern_defined",
        "key_exchange": "pattern_defined",
        "auth": "pattern_defined",
        "encryption": "aead",
        "forward_secrecy": "pattern_defined",
        "message_format": "noise_like",
        "resumption": "handshake_pattern",
        "distinctive_features": [
            "Pattern-based handshake",
            "XX, IK, XXfallback, etc.",
            "DH functions:25519, 448",
            "Lightweight specification",
        ],
    },
    "Signal_Double_Ratchet": {
        "handshake": "session_init",
        "key_exchange": "dh_ratchet",
        "auth": "auth_ephemeral",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "binary_fixed",
        "resumption": "sealed_sender",
        "distinctive_features": [
            "Symmetric-key ratcheting",
            "DH ratcheting",
            "Sealed sender encryption",
            "Double Ratchet algorithm",
        ],
    },
    "WireGuard": {
        "handshake": "cookie",
        "key_exchange": "ephemeral_ephemeral",
        "auth": "psk",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "binary_fixed",
        "resumption": "initiator_rekey",
        "distinctive_features": [
            "Cryptokey routing",
            "Timer-based rekey",
            " cookie messages for DoS protection",
            "Very minimal codebase",
        ],
    },
    "SSH": {
        "handshake": "full_duplex",
        "key_exchange": "ephemeral_ephemeral",
        "auth": "multiple",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "binary_packet",
        "resumption": "rekey",
        "distinctive_features": [
            "Key exchange (KEX) algorithms",
            "Multiple authentication methods",
            "Channel-based multiplexing",
            "Extensible protocol",
        ],
    },
    "HPKE": {
        "handshake": "mode_defined",
        "key_exchange": "hybrid",
        "auth": "mode_defined",
        "encryption": "aead",
        "forward_secrecy": "mode_defined",
        "message_format": "tls_like",
        "resumption": "external",
        "distinctive_features": [
            "Basic mode (base)",
            "PSK mode",
            "Auth mode",
            "Auth+PSK mode",
            "Post-quantum hybrid",
        ],
    },
    "DTLS": {
        "handshake": "full_duplex",
        "key_exchange": "ephemeral_ephemeral",
        "auth": "cert",
        "encryption": "aead",
        "forward_secrecy": "strong",
        "message_format": "tls_like",
        "resumption": "psk_ticket",
        "distinctive_features": [
            "UDP-based DTLS",
            "Record layer fragmentation",
            "cookie exchange",
            "Handshake timeouts",
        ],
    },
}

# Attack vectors to consider
ATTACK_VECTORS = [
    "replay_attack",
    "mitm_attack",
    "reflection_attack",
    "downgrade_attack",
    "key_compromise",
    "session_hijacking",
    "identity_spoofing",
    "packet_modification",
    "state_desynchronization",
    "malformed_packets",
    "timing_attacks",
    "side_channel",
    "oracle_attacks",
    "birthday_attacks",
]


def get_protocol_distance(protocol_a: dict, protocol_b: dict) -> float:
    """Calculate architectural distance between two protocols."""
    differences = 0
    total = 0
    
    for key in ["handshake", "key_exchange", "auth", "encryption", "forward_secrecy"]:
        if protocol_a.get(key) != protocol_b.get(key):
            differences += 1
        total += 1
    
    return differences / total if total > 0 else 0.0


def analyze_architectural_gaps() -> list:
    """Identify gaps in existing protocols that could be novel targets."""
    gaps = [
        {
            "gap": "Simultaneous bidirectional encryption initiation",
            "existing": ["TLS 1.3", "Noise"],
            "novel_approach": "Symmetric initiation handshake",
        },
        {
            "gap": "Lightweight post-quantum forward secrecy",
            "existing": ["Kyber+X25519 hybrids"],
            "novel_approach": "Optimized PQ ephemeral exchange",
        },
        {
            "gap": "Minimal round-trip session establishment",
            "existing": ["QUIC 0-RTT"],
            "novel_approach": "Predictive handshake caching",
        },
        {
            "gap": "Formal verification-friendly protocol",
            "existing": ["TLS 1.3 (with difficulty)"],
            "novel_approach": "State-machine minimalism",
        },
        {
            "gap": "Connection migration with forward secrecy",
            "existing": ["QUIC"],
            "novel_approach": "Token-based migration",
        },
    ]
    return gaps
