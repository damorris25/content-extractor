# CONTRACT-I2-REGISTRY.1.0

**Version:** 1.0
**Status:** draft
**Owner:** dmorris

## Purpose

Defines the extractor registry — how format-specific extractors register
themselves and how callers look up the right extractor for a given MIME type.
The registry is the central dispatch mechanism: callers never import format
packages directly.

## Interfaces

```go
// Registry maps MIME types to Extractor implementations.
// Thread-safe for concurrent reads after initialization.
type Registry struct {
    // unexported fields
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry

// Register adds an extractor for the given MIME types.
// Panics if a MIME type is already registered (fail-fast at startup).
func (r *Registry) Register(ext Extractor)

// Lookup returns the extractor for the given MIME type, or nil if none registered.
func (r *Registry) Lookup(mime string) Extractor

// Extract detects the format (if mime is empty) and dispatches to the
// appropriate extractor. This is the primary entry point for callers.
func (r *Registry) Extract(ctx context.Context, reader io.ReadSeeker, mime string) (*Result, error)

// SupportedTypes returns all registered MIME types, sorted alphabetically.
func (r *Registry) SupportedTypes() []string

// DefaultRegistry returns a registry pre-loaded with all built-in extractors.
// Uses explicit registration (no init() functions).
func DefaultRegistry() *Registry
```

## Behavioral Contracts

| Behavior | Specification |
|----------|--------------|
| `Register` duplicate MIME | Panics with descriptive message (fail-fast, caught at startup) |
| `Register` nil extractor | Panics |
| `Lookup` unregistered MIME | Returns nil (not an error — caller decides what to do) |
| `Extract` with empty MIME | Detects format via `CONTRACT:C2-DETECT` first, then dispatches |
| `Extract` unregistered MIME | Returns `ErrUnsupportedFormat` |
| `SupportedTypes` | Returns sorted, non-nil slice. Stable after init. |
| Concurrent `Lookup` | Safe for concurrent use after all `Register` calls complete |
| `Register` after first `Lookup` | Not supported. Registration happens at startup only. |
| `DefaultRegistry` | Returns a new registry each call. Caller owns it. |

## Error Contracts

| Error | When | Sentinel |
|-------|------|----------|
| `ErrUnsupportedFormat` | No extractor registered for MIME type | `extractor.ErrUnsupportedFormat` |
| `ErrDetectionFailed` | Auto-detection could not determine format | `extractor.ErrDetectionFailed` |

## Dependencies

- Depends on: `CONTRACT:C1-EXTRACTOR.1.0` for the `Extractor` interface and `Result` type
- Depends on: `CONTRACT:C2-DETECT.1.0` for MIME type detection (when MIME is empty)
- External: none (stdlib only)

## Implementing Files

- `extractor/registry.go` — registry implementation
- `extractor/default.go` — DefaultRegistry with all built-in extractors
- `extractor/registry_test.go` — contract tests

## Test Requirements

- [ ] Register and lookup round-trip
- [ ] Duplicate registration panics
- [ ] Nil extractor registration panics
- [ ] Unregistered MIME returns nil from Lookup
- [ ] Extract dispatches to correct extractor based on MIME
- [ ] Extract with empty MIME triggers detection
- [ ] SupportedTypes returns sorted list
- [ ] Concurrent Lookup safety (run with -race)

## Change History

| Version | Date | Change | Migration |
|---------|------|--------|-----------|
| 1.0 | 2026-03-21 | Initial contract | — |
