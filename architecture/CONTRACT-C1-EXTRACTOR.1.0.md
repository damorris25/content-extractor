# CONTRACT-C1-EXTRACTOR.1.0

**Version:** 1.0
**Status:** draft
**Owner:** dmorris

## Purpose

Defines the core extraction interface — the central abstraction that all
format-specific extractors implement. Any code that needs to extract text
or metadata from a file depends on this contract, never on a specific format.

## Interfaces

```go
// Extractor extracts text content and metadata from a specific file format.
// Implementations MUST be safe for concurrent use.
type Extractor interface {
    // Extract reads from r and returns extracted content and metadata.
    // The caller provides a context for cancellation and a MIME type hint.
    // Implementations MUST NOT buffer the entire input in memory.
    Extract(ctx context.Context, r io.Reader, mime string) (*Result, error)

    // SupportedMIMETypes returns the MIME types this extractor handles.
    SupportedMIMETypes() []string
}

// Result holds extracted content and metadata from a file.
type Result struct {
    // Text is the extracted textual content, UTF-8 encoded.
    Text string

    // Metadata contains key-value pairs extracted from the file.
    // Keys use a standardized namespace (see CONTRACT:I1-METADATA).
    Metadata map[string][]string

    // ContentType is the detected MIME type of the input.
    ContentType string

    // Language is the detected or declared language (BCP 47), if available.
    Language string
}
```

## Behavioral Contracts

| Behavior | Specification |
|----------|--------------|
| `Extract` with nil reader | Returns `ErrNilReader` |
| `Extract` with empty reader | Returns empty `Result` with no error (not an error condition) |
| `Extract` with unsupported MIME | Returns `ErrUnsupportedFormat` |
| `Extract` cancelled via context | Returns `ctx.Err()` promptly (within 100ms of cancellation) |
| `Extract` with corrupt/truncated input | Returns `ErrCorruptInput` with partial `Result` if any content was extracted |
| Memory usage | MUST NOT buffer entire input in memory. Use streaming. |
| Concurrent safety | All methods safe for concurrent use by multiple goroutines |
| `SupportedMIMETypes` | Returns non-nil, non-empty slice. Stable across calls. |
| `Result.Text` encoding | Always valid UTF-8. Invalid bytes replaced with U+FFFD. |
| `Result.Metadata` keys | Follow `I1-METADATA` contract namespace conventions |

## Error Contracts

| Error | When | Sentinel |
|-------|------|----------|
| `ErrNilReader` | Reader is nil | `extractor.ErrNilReader` |
| `ErrUnsupportedFormat` | MIME type not handled by this extractor | `extractor.ErrUnsupportedFormat` |
| `ErrCorruptInput` | Input is malformed or truncated | `extractor.ErrCorruptInput` |
| `ErrPasswordProtected` | File requires a password to read | `extractor.ErrPasswordProtected` |
| `ErrFileTooLarge` | Input exceeds configured size limit | `extractor.ErrFileTooLarge` |

## Dependencies

- Depends on: `CONTRACT:I1-METADATA.1.0` for metadata key conventions
- Depends on: `CONTRACT:I2-REGISTRY.1.0` for extractor registration and lookup
- External: none (stdlib only for this contract)

## Implementing Files

- `extractor/extractor.go` — interface and result type definitions
- `extractor/errors.go` — sentinel error definitions
- `extractor/extractor_test.go` — contract compliance test helpers

## Test Requirements

- [ ] Every method in the interface has at least one happy-path test
- [ ] Every behavioral contract in the table above has a test
- [ ] Every error contract has a test (trigger the error, verify sentinel)
- [ ] Concurrent safety verified with `-race` flag
- [ ] Contract compliance test helper that any format extractor can import and run
- [ ] Memory benchmark showing streaming behavior (no full-file buffering)

## Change History

| Version | Date | Change | Migration |
|---------|------|--------|-----------|
| 1.0 | 2026-03-21 | Initial contract | — |
