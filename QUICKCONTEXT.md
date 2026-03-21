# Quick Context

<!-- freshness: 2026-03-21 -->

**30-second orientation for new agent sessions.** Read this after README.md.
Verify claims against `git log --oneline -10`, then proceed to TODO.md.

## Project

Dual-language library (Go + TypeScript) for extracting text content and metadata from unstructured file formats (Office, ODF, PDF, images, audio, CAD). Commercial-friendly, minimal dependencies, low memory footprint. TypeScript port has proven output parity via golden file tests.

## Current Branch & State

| Field | Value |
|-------|-------|
| **Branch** | `master` |
| **Last significant commit** | README update with TypeScript implementation details |
| **Remote** | `https://github.com/damorris25/content-extractor.git` (master + main) |
| **Build status** | Go: Passing (`go build ./...`); TS: Passing (`npm run build`, 203ms) |
| **Test count** | Go: 139 tests / 12 packages; TS: 34 golden file parity tests |
| **Test status** | All passing (Go with `-race`; TS 20ms execution) |

## What's In Progress

- **Current focus:** Both implementations complete; all P0-P2 + TS done
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
- TypeScript port — browser-first, 34 golden file parity tests, all formats except PDF, 2026-03-21

## What's Blocked

_None currently._

## Key Numbers

| Metric | Go | TypeScript |
|--------|-----|-----------|
| Runtime deps | 2 (BSD) | 1 required (fflate, MIT) + 1 optional (pdfjs-dist) |
| Test count | 139 | 34 golden parity |
| Source lines | ~6,500 | ~4,200 |
| Bundle size | N/A (compiled) | 60 KB ESM / 42 KB CJS |

## What's Next

1. Additional edge-case test coverage
2. TypeScript PDF parity (requires pdfjs-dist golden files)
3. npm publish for TypeScript package

---
