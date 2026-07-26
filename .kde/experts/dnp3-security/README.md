# DNP3 Security Expert

**Expert ID**: DNP3-SEC-EXPERT-001  
**Domain**: DNP3 Protocol Security  
**Version**: 1.0.0  
**Status**: Active  

---

## Overview

This expert contains security domain knowledge for DNP3 protocol implementation and Secure Authentication (SA).

## Domain Knowledge

### Secure Authentication Overview

| Aspect | Details |
|--------|---------|
| Standard | IEC 62351-6 |
| Challenge Size | 64-bit |
| Key Size | AES-128 |
| MAC Size | 128-bit |
| Version | DNP3-SA v1, v2, v3 |

### Security Levels

| Level | Description | MAC | Challenge | Encryption |
|-------|-------------|-----|-----------|------------|
| 0 | None | No | No | No |
| 1 | Authentication | Yes | No | No |
| 2 | Authentication + Challenge | Yes | Yes | No |
| 3 | Authentication + Encryption | Yes | Yes | Yes |

### Key Concepts

- **Challenge/Response**: Nonce-based authentication
- **Key Change**: Periodic key rotation
- **MAC Calculation**: HMAC-SHA-256 derived
- **Session Management**: Keep-alive and timeout

## Rules and Constraints

### Security Implementation Rules

1. **No Plaintext Keys**: Keys must never appear in logs or errors
2. **Key Storage**: Use secure storage (keyring, vault)
3. **Timeout Enforcement**: Sessions must timeout
4. **Failed Auth Tracking**: Log and limit failed attempts
5. **SA Version Negotiation**: Must match peer capability

### Protocol Constraints

| Constraint | Requirement |
|------------|--------------|
| Challenge Expiry | Must reject stale challenges |
| MAC Verification | Must verify before processing |
| Key Change Interval | Configurable per policy |
| Session Timeout | Default 5 minutes |

## Best Practices

### Implementation

1. Use established crypto libraries (not custom)
2. Generate cryptographically random nonces
3. Validate all MAC calculations
4. Implement rate limiting on authentication failures
5. Secure key derivation from master key

### Operations

1. Regular key rotation (per policy)
2. Monitor authentication failures
3. Alert on security events
4. Audit trail for security decisions

## Reference Standards

- IEC 62351-6: Security for SCADA
- IEEE 1815: DNP3 Protocol
- NIST SP 800-108: Key Derivation

## Related Artifacts

| Artifact | Purpose |
|----------|---------|
| internal/sa/ | Secure Authentication implementation |
| DNP3-EXPERT-001 | General DNP3 protocol |

---

**Expert Status**: ACTIVE  
**Last Updated**: 2026-07-26
