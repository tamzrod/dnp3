# Contributing to go-dnp3

Thank you for your interest in contributing to go-dnp3!

This document provides guidelines and instructions for contributing to this project.

## Project Status

> ⚠️ **Important**: This project is currently in the **research and architecture phase**. 
> No protocol implementation exists yet. Implementation is prohibited until the 
> architecture has been defined and approved.

## How Can I Contribute?

### Research Phase Contributions

During this phase, we welcome contributions in:

- **Protocol Analysis**: Research into IEEE 1815 specifications
- **Architecture Discussion**: Feedback on proposed architectures
- **Documentation**: Improving documentation quality
- **Research Papers**: Sharing relevant research findings
- **Use Cases**: Describing real-world DNP3 deployment scenarios

### Non-Code Contributions

- Reporting issues
- Suggesting features
- Improving documentation
- Sharing expertise in DNP3/SCADA systems
- Providing feedback on architecture proposals

## Development Workflow

### 1. Fork and Clone

```bash
git clone https://github.com/your-username/go-dnp3.git
cd go-dnp3
```

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b research/your-research-topic
```

### 3. Make Your Changes

Follow the guidelines in this document.

### 4. Commit Your Changes

We follow conventional commit messages:

```
feat: add new architecture document
docs: update CONTRIBUTING.md
research: analyze DNP3 data link layer
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Commit Message Guidelines

- Use imperative mood ("add" not "added")
- Keep the first line under 72 characters
- Reference issues when applicable

## Pull Request Process

1. Fill out the PR template completely
2. Ensure all checks pass
3. Request review from maintainers
4. Address feedback constructively

## Architecture-First Methodology

This project follows an architecture-first development methodology:

1. **Research** → Document protocol invariants
2. **Architecture** → Define solution design
3. **Review** → Get approval on architecture
4. **Implement** → Write code per architecture
5. **Test** → Validate against protocol specifications

**No implementation without approved architecture.**

## Code Style

Once implementation begins, we will follow:

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Testing Guidelines

- All code must have tests
- Tests must pass before merging
- New features require documentation

## Documentation Standards

- Use clear, concise language
- Include examples where appropriate
- Update relevant documentation with changes

## License

By contributing, you agree that your contributions will be licensed 
under the Apache License 2.0.

## Questions?

Feel free to:

- Open an issue for questions
- Join project discussions
- Contact maintainers directly

## Thank You

Your contributions make this project better. Thank you for your time 
and effort!
