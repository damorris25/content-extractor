# CONTRACT-C2-DETECT.1.0

**Version:** 1.0
**Status:** draft
**Owner:** dmorris

## Purpose

Defines the format detection system — identifies file formats from magic bytes
(file signatures) and file extensions. This is the first step in the extraction
pipeline: before we can extract, we must know what we're looking at.

## Interfaces

```go
// Detector identifies file formats from content and/or filename.
type Detector interface {
    // Detect determines the MIME type by reading magic bytes from r.
    // r must be seekable — Detect reads a prefix and seeks back to start.
    // filename is optional (empty string OK) — used as a fallback hint.
    // Returns the MIME type and nil error, or ErrDetectionFailed.
    Detect(r io.ReadSeeker, filename string) (string, error)
}

// DetectFunc allows using a function as a Detector.
type DetectFunc func(r io.ReadSeeker, filename string) (string, error)

// MagicSignature defines a file format's magic bytes.
type MagicSignature struct {
    MIME    string   // MIME type this signature identifies
    Offset int64    // Byte offset where signature starts
    Magic  []byte   // The magic byte sequence
    Mask   []byte   // Optional mask (nil = exact match)
}
```

## Behavioral Contracts

| Behavior | Specification |
|----------|--------------|
| Detection priority | Magic bytes take precedence over file extension |
| `Detect` with empty reader | Returns `ErrDetectionFailed` |
| `Detect` with nil reader | Returns `ErrNilReader` |
| Reader position after `Detect` | Seek back to original position (caller's reader is unchanged) |
| ZIP-based formats | Must distinguish DOCX/XLSX/PPTX/ODT/ODS/ODP by inspecting ZIP contents |
| Unknown format | Returns `ErrDetectionFailed`, never guesses |
| Extension-only fallback | If magic bytes match nothing AND filename has extension, use extension mapping |
| Buffer size | Read at most 8KB for detection (sufficient for all known signatures) |
| Concurrent safety | Stateless — safe for concurrent use |

## Error Contracts

| Error | When | Sentinel |
|-------|------|----------|
| `ErrNilReader` | Reader is nil | `extractor.ErrNilReader` |
| `ErrDetectionFailed` | Could not determine format | `extractor.ErrDetectionFailed` |

## Dependencies

- External: none (stdlib only — `archive/zip` for ZIP inspection)

## Implementing Files

- `detect/detect.go` — detector implementation
- `detect/signatures.go` — magic byte signature table
- `detect/extensions.go` — extension-to-MIME mapping table
- `detect/detect_test.go` — contract tests

## Test Requirements

- [ ] Detect each supported format by magic bytes (one test per format)
- [ ] Detect ZIP-based Office/ODF formats correctly (not just "application/zip")
- [ ] Extension fallback works when magic bytes don't match
- [ ] Reader position restored after detection
- [ ] Unknown format returns ErrDetectionFailed
- [ ] Nil/empty reader handling
- [ ] Concurrent safety (run with -race)

## Change History

| Version | Date | Change | Migration |
|---------|------|--------|-----------|
| 1.0 | 2026-03-21 | Initial contract | — |
