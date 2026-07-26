# DNP3-EXP-005: Resolve Windows Build Environment for Workbench

## Issue Summary

Investigate and resolve Windows build blockers preventing `cmd/workbench` from compiling.

## Task 1 - Source Investigation

### File: `cmd/workbench/internal/session/session.go`

**Imports found:**
```go
import (
    "context"
    "fmt"
    "sync"
    "time"

    "dnp3/pkg/dnp3"       // Line 10 - UNUSED
    "dnp3/pkg/dnp3/types" // Line 11 - USED
    "dnp3/pkg/transport"   // Line 12 - USED
)
```

**Usage Analysis:**
| Import | Used | Evidence |
|--------|------|----------|
| `dnp3/pkg/dnp3` | ❌ No | Only imported, never referenced |
| `dnp3/pkg/dnp3/types` | ✅ Yes | `types.GroupRequest`, `types.BinaryInput`, `types.QualityOnline` |
| `dnp3/pkg/transport` | ✅ Yes | `transport.Handler`, `transport.TCPConfig` |

**Root Cause:** Stale import from earlier development phase.

## Task 2 - Windows Environment Investigation

### Evidence: GLFW Requires CGO

```c
// github.com/go-gl/glfw/v3.3/glfw/c_glfw.go
/*
#include "glfw/src/context.c"
#include "glfw/src/init.c"
...
*/
import "C"
```

**Finding:** GLFW uses `import "C"` which requires CGO.

### Evidence: Makefile Windows Build

```makefile
# cmd/workbench/Makefile
windows:
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
        go build -ldflags="-s -w" -o workbench.exe .
```

**Finding:** Windows cross-compilation requires MinGW-w64 compiler.

### Evidence: CGO Cannot Be Disabled

- Fyne uses GLFW for window management
- GLFW requires native OpenGL bindings
- CGO is required for Fyne GUI applications
- `CGO_ENABLED=0` would prevent GLFW from compiling

## Task 3 - Environment Validation

### Current Environment (Linux amd64)
```
CGO_ENABLED=1
CC=gcc
CXX=g++
GOOS=linux
GOARCH=amd64
PKG_CONFIG=pkg-config
```

**Finding:** Environment is correctly configured for Linux CGO builds.

## Task 4 - Recommendation

### Comparison: MSYS2 vs MinGW-w64

| Aspect | MSYS2 | MinGW-w64 |
|--------|-------|-----------|
| Package Size | ~3GB | ~200MB |
| Components | MinGW + pacman + bash | Compiler only |
| Windows Integration | Full POSIX emulation | Direct Windows syscalls |
| Go Cross-compile | Overkill | Optimal |

### Recommendation: MinGW-w64

**Rationale:**
1. Go cross-compilation only needs the C compiler
2. MSYS2's pacman and bash are unnecessary for `go build`
3. MinGW-w64 is explicitly referenced in the Makefile (`x86_64-w64-mingw32-gcc`)
4. Lightweight and purpose-specific

**Minimal Windows Toolchain:**
```bash
# Install MinGW-w64 on Linux for cross-compilation
sudo apt install mingw-w64

# Verify installation
x86_64-w64-mingw32-gcc --version
```

## Task 5 - Implementation

### Repository Changes

**File:** `cmd/workbench/internal/session/session.go`

**Before:**
```go
import (
    "context"
    "fmt"
    "sync"
    "time"

    "dnp3/pkg/dnp3"
    "dnp3/pkg/dnp3/types"
    "dnp3/pkg/transport"
)
```

**After:**
```go
import (
    "context"
    "fmt"
    "sync"
    "time"

    "dnp3/pkg/dnp3/types"
    "dnp3/pkg/transport"
)
```

## Verification Results

### Pre-fix Error
```
internal/session/session.go:10:2:
"dnp3/pkg/dnp3" imported and not used
```

### Post-fix Build
```
$ go build ./cmd/workbench/internal/session/
# SUCCESS - no errors
```

### Full Build Status
The session package now compiles successfully. Full workbench build encounters:
1. **X11 libraries missing** (Linux environment - expected)
2. **Fyne API incompatibility** (separate issue - not in scope)

## Findings Summary

| Finding | Status |
|---------|--------|
| Unused import `dnp3/pkg/dnp3` | ✅ Fixed |
| CGO required for Fyne/GLFW | ✅ Confirmed |
| CGO cannot be disabled | ✅ Confirmed |
| MinGW-w64 recommended | ✅ Documented |
| Environment correctly configured | ✅ Confirmed |

## Root Cause

**Primary:** Stale unused import `dnp3/pkg/dnp3` in `session.go`

**Secondary:** Windows build requires MinGW-w64 toolchain (not installed in Linux CI)

## Environment Findings

1. **Current Linux environment:** Correctly configured for native Linux CGO builds
2. **Windows cross-compilation:** Requires MinGW-w64 (`x86_64-w64-mingw32-gcc`)
3. **CGO is mandatory:** Fyne GUI applications cannot disable CGO

## Remaining Issues

| Issue | Priority | Notes |
|-------|----------|-------|
| Fyne API incompatibility | High | `widget.RadioButton` not in v2.4.0 |
| X11 libraries missing | Low | Linux GUI only, expected |
| MinGW-w64 not installed | Medium | Required for Windows builds |

## Related Experiments

- DNP3-EXP-003: Stale `fyne-io/fyne` dependency
- DNP3-EXP-004: Stale `go-text/types` dependency

## Conclusion

**Status:** Partial Resolution

The unused import issue has been fixed. Windows build environment requires MinGW-w64 installation, which is outside the scope of repository changes.
