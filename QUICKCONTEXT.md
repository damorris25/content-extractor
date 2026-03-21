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
| **Last significant commit** | Full P0 + P1: all Office, PDF, and ODF extractors |
| **Build status** | Passing (`go build ./...`) |
| **Test count** | 98 tests across 9 packages |
| **Test status** | All passing (including `-race`) |

## What's In Progress

- **Current focus:** P2 — image/audio/CAD metadata, legacy Office, benchmarks
- **Active branch(es):** `master`
- **Assigned agents:** none

## What's Recently Complete

- P0 foundation: core types, metadata, registry, detection, 2026-03-21
- Plain text extractor, 2026-03-21
- DOCX/XLSX/PPTX extractors (full OOXML suite), 2026-03-21
- PDF text extractor (ledongthuc/pdf, BSD-3-Clause), 2026-03-21
- ODT/ODS/ODP extractors (full ODF suite), 2026-03-21
- Contract compliance test harness + integration tests, 2026-03-21

## What's Blocked

_None currently._

## What's Next

1. Image metadata extraction (EXIF/PNG)
2. Audio metadata extraction (ID3/Vorbis)
3. CAD text extraction (DXF)
4. Legacy Office formats (DOC/XLS/PPT)
5. Benchmark suite

---
