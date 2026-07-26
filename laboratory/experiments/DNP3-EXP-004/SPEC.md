# DNP3-EXP-004: Fyne v2.4.0 Resolves Non-Existent `github.com/go-text/types`

## Issue Summary

**Problem**: `cmd/workbench/go.mod` contains a stale indirect dependency entry for `github.com/go-text/types v0.0.0-20240903004611-7a47d0d1c6ba` that causes build failures.

**Root Cause**: This is a stale indirect require entry, not an actual transitive dependency. The module `github.com/go-text/types` does not exist on GitHub (404 Not Found), and this entry was likely added by Go's dependency resolution at some point but the module has since been deleted/archived.

## Investigation Findings

### 1. Module Status

| Module | Status |
|--------|--------|
| `github.com/go-text/types` | ❌ Deleted/Not Found (404) |
| `github.com/go-text/shaping` | ⚠️ Archived |
| `github.com/go-text/di` | ⚠️ Archived |
| `github.com/go-text/font` | ⚠️ Archived |
| `github.com/go-text/typesetting` | ✅ Active |
| `github.com/go-text/render` | ✅ Active |
| `github.com/go-text/typesetting-utils` | ✅ Active |

### 2. Dependency Analysis

- **No code imports `github.com/go-text/types`** - This is a stale entry in go.mod
- **Fyne v2.4.0 uses `go-text/typesetting v0.0.0-20230616162802-9c17dd34aa4a`** which does not require `go-text/types`
- **All go-text modules use `golang.org/x/text`** instead of `github.com/go-text/types`

### 3. Classification

**Category**: Stale Dependency Entry

This is similar to DNP3-EXP-003 where a stale indirect require entry was present in the go.mod file. The entry was likely added during an earlier `go mod tidy` run when the module existed, but the module has since been deleted.

## Resolution

### Changes Made

1. **Removed stale entry from `cmd/workbench/go.mod`**:
   - Removed: `github.com/go-text/types v0.0.0-20240903004611-7a47d0d1c6ba // indirect`

2. **Created `go.work` file** at project root to enable multi-module workspace:
   - Includes both root module (`dnp3`) and `cmd/workbench` module
   - Enables cross-module imports between workbench and core packages

3. **Ran `go mod tidy`** to regenerate clean dependency list

### Verification

```
$ cd cmd/workbench && go mod verify
all modules verified
```

### Build Status

The build now proceeds past the `go-text/types` issue but requires X11 development libraries for the GUI components. This is expected for Fyne applications.

```
$ cd cmd/workbench && make build
# github.com/go-gl/glfw/v3.3/glfw
In file included from ./glfw/src/x11_platform.h:33:
./glfw/src/x11_platform.h:33:10: fatal error: X11/Xlib.h: No such file or directory
```

## Files Modified

| File | Change |
|------|--------|
| `cmd/workbench/go.mod` | Removed stale `github.com/go-text/types` entry |
| `go.work` | Created to enable multi-module workspace |

## Related Issues

- DNP3-EXP-003: Similar stale dependency entry for `github.com/fyne-io/fyne v1.4.2`

## Conclusion

**Status**: ✅ Resolved

The `github.com/go-text/types` dependency issue was caused by a stale indirect require entry in `cmd/workbench/go.mod`. The module `github.com/go-text/types` no longer exists and was never actually required by any code in the repository. Removing this stale entry and creating a `go.work` file fixes the build issue.

**Note**: The build still requires X11 development libraries for the Fyne GUI components, but this is expected for GUI applications and is not a dependency resolution issue.
