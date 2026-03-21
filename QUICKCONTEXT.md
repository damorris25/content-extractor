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
| **Last significant commit** | P0 foundation + plain text + DOCX extractors |
| **Build status** | Passing (`go build ./...`) |
| **Test count** | 47 tests across 6 packages |
| **Test status** | All passing (including `-race`) |

## What's In Progress

- **Current focus:** P1 — additional format extractors (XLSX, PPTX, PDF, ODF)
- **Active branch(es):** `master`
- **Assigned agents:** none

## What's Recently Complete

- Core types: Extractor interface, Result, sentinel errors, 2026-03-21
- Metadata model with namespace constants and normalization helpers, 2026-03-21
- Registry with MIME-based dispatch and format auto-detection, 2026-03-21
- Format detection (magic bytes + ZIP inspection + extension fallback), 2026-03-21
- Plain text extractor with streaming and UTF-8 cleanup, 2026-03-21
- DOCX extractor with text and metadata extraction, 2026-03-21
- Contract compliance test harness (extractortest package), 2026-03-21
- Integration wiring (contentextractor convenience package), 2026-03-21

## What's Blocked

_None currently._

## What's Next

1. Implement XLSX extractor (shared strings + sheet XML)
2. Implement PPTX extractor (slide XML text extraction)
3. Implement PDF text extractor (text-layer, not OCR)
4. Implement ODT/ODS/ODP extractors (ZIP-based, `content.xml`)

---
