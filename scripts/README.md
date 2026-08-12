# Scripts

This directory contains utility scripts for development, testing, and maintenance.

## Available Scripts

### Build Scripts

| Script | Platform | Purpose |
|--------|---------|---------|
| `build-workbench.ps1` | Windows | Build and run the DNP3 Workbench |

### Development

- [ ] `bootstrap.sh` - Initialize development environment
- [ ] `generate.sh` - Run code generators

### Testing

- [x] `verify-mvp.sh` - Run the DNP3 MVP verification gate (build + vet + unit/integration + race; exit 0 on a clean tree). See [DNP3-052](../active_work/DNP3_MASTER_ROADMAP.md).
- [x] `verify-external-mvp.sh` - Run the DNP3 **external** MVP verification gate (MEXT-021). Two tiers: Tier 1 internal real-TCP loopback tests (build + real-TCP transport tests); Tier 2 external/third-party (VEC-01) proof, **fail-closed** until genuine external interop proof lands (MEXT-022/MEXT-033). Exit 0 only when both tiers pass. Set `ALLOW_NO_EXTERNAL=1` to run Tier 1 only (does NOT satisfy the external claim). See [MEXT-021](../active_work/MEXT_MASTER_ROADMAP.md).
- [x] `run-conformance.sh` - Run the DLL/TL/AL layer conformance suites (plain + race; exit 0 on green). CI-runnable conformance gate. See [DNP3-098](../active_work/DNP3_MASTER_ROADMAP.md).
- [ ] `test-conformance.sh` - Run conformance tests (alias of `run-conformance.sh`)
- [ ] `test-interop.sh` - Run interoperability tests
- [ ] `test-fuzz.sh` - Run fuzzing tests

### Release

- [ ] `release.sh` - Create a release
- [ ] `changelog.sh` - Generate changelog

### Maintenance

- [ ] `format.sh` - Format code
- [ ] `lint.sh` - Run linters
- [ ] `update-deps.sh` - Update dependencies

## Usage

### Windows - PowerShell

```powershell
# Build the workbench
.\build-workbench.ps1 -Action build

# Build and run
.\build-workbench.ps1 -Action run

# Clean build artifacts
.\build-workbench.ps1 -Action clean

# Full rebuild
.\build-workbench.ps1 -Action all
```

### Linux/macOS - Bash

Make scripts executable first:
```bash
chmod +x scripts/*.sh
```

Run a script:
```bash
./scripts/<script-name>.sh
```

## Requirements

| Script | Requirements |
|--------|--------------|
| `build-workbench.ps1` | Go 1.22+, PowerShell 5.1+ |
| `bootstrap.sh` | Bash, Go |
| `test-*.sh` | Bash, Go, test dependencies |

Check individual script documentation for requirements.

> ⚠️ **Note**: Scripts will be implemented as needed during development.
