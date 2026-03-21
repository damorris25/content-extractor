# Quick Context

<!-- freshness: 2026-03-21 -->

**30-second orientation for new agent sessions.** Read this after README.md.
Verify claims against `git log --oneline -10`, then proceed to TODO.md.

## Project

Go library for extracting text content and metadata from unstructured file formats (Office, ODF, PDF, images, audio, CAD). Commercial-friendly, minimal dependencies, low memory footprint.

## Current Branch & State

| Field | Value |
|-------|-------|
| **Branch** | `master` |
| **Last significant commit** | Security hardening across all extractors |
| **Build status** | Passing (`go build ./...`) |
| **Test count** | 139 tests across 12 packages |
| **Test status** | All passing (including `-race`) |

## What's In Progress

- **Current focus:** Security hardening complete; all P0-P2 work done
- **Active branch(es):** `master`
- **Assigned agents:** none

## What's Recently Complete

- P0 foundation: core types, metadata, registry, detection, 2026-03-21
- Plain text extractor, 2026-03-21
- DOCX/XLSX/PPTX extractors (full OOXML suite), 2026-03-21
- PDF text extractor (ledongthuc/pdf, BSD-3-Clause), 2026-03-21
- ODT/ODS/ODP extractors (full ODF suite), 2026-03-21
- Contract compliance test harness + integration tests, 2026-03-21
- Image metadata extractor (EXIF, PNG tEXt, dimensions), 2026-03-21
- Audio metadata extractor (ID3v2, Vorbis, WAV RIFF — pure Go), 2026-03-21
- DXF text/metadata extractor (TEXT, MTEXT, layers, header vars), 2026-03-21
- Benchmark suite + license audit + accuracy tests, 2026-03-21
- Security hardening (18 files, decompression limits, overflow fixes, panic recovery), 2026-03-21

## What's Blocked

_None currently._

## What's Next

1. TypeScript/WASM bindings (future)
2. Additional edge-case test coverage

---
