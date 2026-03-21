# CONTRACT-I1-METADATA.1.0

**Version:** 1.0
**Status:** draft
**Owner:** dmorris

## Purpose

Defines the metadata model — standardized key namespaces, normalization rules,
and the common metadata fields that extractors populate. Inspired by Dublin Core
and Apache Tika's metadata model, adapted for Go idioms.

## Interfaces

```go
// Metadata provides typed access to extracted metadata key-value pairs.
// Keys follow a namespaced convention: "namespace:field" (e.g., "dc:title").
// All values are string slices to support multi-valued fields.
type Metadata map[string][]string

// Standard namespace prefixes
const (
    NSDublinCore   = "dc"       // Dublin Core: title, creator, subject, description, date
    NSOffice       = "office"   // Office-specific: revision, template, application
    NSEXIF         = "exif"     // EXIF image metadata: camera, GPS, dimensions
    NSAudio        = "audio"    // Audio metadata: artist, album, track, duration
    NSCAD          = "cad"      // CAD metadata: layers, units, version
    NSExtractor    = "extractor" // Extraction process: duration, extractor-version, warnings
)

// Well-known metadata keys (not exhaustive — extractors may add format-specific keys)
const (
    KeyTitle       = "dc:title"
    KeyCreator     = "dc:creator"
    KeySubject     = "dc:subject"
    KeyDescription = "dc:description"
    KeyDate        = "dc:date"         // ISO 8601
    KeyModified    = "dc:modified"     // ISO 8601
    KeyLanguage    = "dc:language"     // BCP 47
    KeyFormat      = "dc:format"       // MIME type
    KeyPageCount   = "office:page-count"
    KeyWordCount   = "office:word-count"
    KeyCharCount   = "office:char-count"
    KeyApplication = "office:application"  // Creating application
    KeyRevision    = "office:revision"
)
```

## Behavioral Contracts

| Behavior | Specification |
|----------|--------------|
| Key format | `namespace:field-name`, lowercase, hyphens for multi-word fields |
| Unknown namespace | Allowed. Extractors may introduce format-specific keys with a custom namespace prefix |
| Nil metadata | Treat as empty. Never panic on nil Metadata |
| Multi-valued fields | Append, don't overwrite. `dc:creator` may have multiple authors |
| Date values | Always ISO 8601 (`2026-03-21T10:30:00Z`). Convert from format-specific representations |
| String encoding | All values valid UTF-8. Invalid bytes replaced with U+FFFD |
| Empty values | Omit keys with no value. Never store `"dc:title": [""]` |
| Case sensitivity | Keys are case-sensitive and MUST be lowercase |

## Error Contracts

No errors defined — metadata is a data model, not an operation. Metadata
normalization functions may return errors for invalid date formats, etc.

## Dependencies

- External: none (stdlib only)

## Implementing Files

- `metadata/metadata.go` — type definitions and constants
- `metadata/normalize.go` — normalization helpers (date parsing, encoding cleanup)
- `metadata/metadata_test.go` — contract tests

## Test Requirements

- [ ] Key format validation (namespace:field-name pattern)
- [ ] Date normalization from common formats to ISO 8601
- [ ] UTF-8 enforcement (invalid bytes replaced)
- [ ] Multi-valued field append behavior
- [ ] Nil safety (operations on nil Metadata don't panic)

## Change History

| Version | Date | Change | Migration |
|---------|------|--------|-----------|
| 1.0 | 2026-03-21 | Initial contract | — |
