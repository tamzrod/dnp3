# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue,
please report it responsibly.

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via:

1. **Email**: Send an email to the project maintainers
2. **Private Security Advisory**: Use GitHub's [Private Vulnerability Reporting](https://github.com/your-username/go-dnp3/security/advisories/new) feature

### What to Include

When reporting, please include:

- Type of vulnerability
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact assessment of the vulnerability

### Response Timeline

We aim to:

- Acknowledge receipt of vulnerability reports within **48 hours**
- Provide an estimated timeline for a fix within **7 days**
- Release a patch as quickly as possible, depending on complexity

## Security Considerations

### DNP3 Protocol Security

This implementation will support IEEE 1815-2012 Secure Authentication, 
which provides:

- Challenge-response authentication
- Key change procedures
- Session security

### Best Practices

When using this library:

1. **Always use TLS** for network transport when possible
2. **Enable Secure Authentication** for production deployments
3. **Rotate keys regularly** per your security policy
4. **Monitor logs** for authentication failures
5. **Follow the principle of least privilege** for access control

### Known Limitations

- This project is in early development
- Security features are not yet implemented
- Do not use in production until a stable release is available

## Security Updates

Security updates will be released as patch versions. Users will be notified
through:

- GitHub Security Advisories
- Release notes
- Project documentation

## Attribution

We thank the security community for their efforts in improving the security
of this project. Security researchers who report valid vulnerabilities will
be acknowledged (unless anonymity is requested).
