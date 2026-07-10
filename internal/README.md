# Internal Packages

This directory contains internal implementation packages that should not be imported by external users.

## Purpose

Internal packages contain implementation details that may change without notice. Users should only import packages from `pkg/`.

## Structure

```
internal/
├── dll/           # Data Link Layer
├── tl/            # Transport Layer
├── al/            # Application Layer
├── sa/            # Secure Authentication
└── master/        # Master implementation
```

## Why Internal?

Go's internal package mechanism prevents external imports:

> "An internal package is importable only by code that lives within a tree rooted at the parent of the internal directory."

This provides encapsulation and allows us to change internal implementation without breaking user code.

## Public API

The public API is defined in `pkg/`. Only types and interfaces in `pkg/` are considered public API.

> ⚠️ **Note**: Packages will be implemented once architecture is approved.
