# Quick Context

<!-- freshness: 2026-03-24 -->

**30-second orientation for new agent sessions.** Read this after README.md.
Verify claims against `git log --oneline -10`, then proceed to TODO.md.

## Project

Dual-language library (Go + TypeScript) for extracting text content and metadata from unstructured file formats (Office, ODF, PDF, XML, HTML, SVG, Email, JSON, CSV, YAML, Markdown, RTF, images incl. HEIC/AVIF, audio, CAD). Commercial-friendly, minimal dependencies, low memory footprint. TypeScript port has proven output parity via golden file tests. Optional OCR text extraction from images via Tesseract (with auto-downscaling for large images). GeoTIFF and ISOBMFF metadata support.

## Current Branch & State

| Field | Value |
|-------|-------|
| **Branch** | `master` |
| **Last significant commit** | Complete all remaining TODOs: RIFF fix, PDF parity, npm prep, 86.1% coverage, OCR CI |
| **Remote** | `https://github.com/damorris25/content-extractor.git` (master + main) |
| **Build status** | Go: Passing (`go build ./...`); TS: Passing (`npm run build`) |
| **Test count** | Go: 290+ tests / 13 packages; TS: 65 tests (52 parity + 5 GeoTIFF + 7 OCR + 1 bench) |
| **Test status** | All passing (Go with `-race`; TS <600ms execution) |
| **Test coverage** | Go: 86.1% overall (extractor 100%, metadata 95.0%, email 93.3%, audio 91.9%, text 91.2%, cad 91.5%, detect 89.6%, image 86.5%, office 85.8%, odf 81.2%, pdf 78.9%) |

## What's In Progress

- **Current focus:** All TODO items resolved; 86.1% coverage; npm publish ready
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
- TypeScript port — browser-first, golden file parity tests, all formats except PDF, 2026-03-21
- OCR text extraction — optional Tesseract integration (Go: gosseract/CGO, TS: tesseract.js), 2026-03-21
- GeoTIFF metadata extraction — CRS, bounds, pixel scale (pure Go/TS), 2026-03-21
- XML extractor — tag stripping with CharData extraction (Go: encoding/xml, TS: state machine), 2026-03-24
- HTML extractor — visible text extraction, title metadata (Go: golang.org/x/net/html, TS: state machine), 2026-03-24
- EML extractor — RFC 2822 email parsing, multipart MIME, header metadata (Go: net/mail stdlib, TS: custom parser), 2026-03-24
- Security hardening — EML multipart recursion depth limit, DOCTYPE stripping, memory-safe HTML parsing, 2026-03-24
- JSON/CSV/TSV/YAML/Markdown — pass-through text extraction with format detection, 2026-03-24
- SVG extractor — text element extraction (text, tspan, textPath) with title/description metadata, 2026-03-24
- RTF extractor — control word parser with Unicode escape support, 2026-03-24
- HEIC/HEIF/AVIF — ISOBMFF container parsing for dimensions (ispe) and EXIF (iloc/iinf), zero new deps, 2026-03-24
- OCR downscaling — auto-resize large images (>2048px) before Tesseract via OffscreenCanvas, 2026-03-24
- Demo app — progress bar, elapsed time counter, 60s timeout, updated format list, 2026-03-24
- Security hardening round 2 — goexif panic recovery, PDF page cap, XLSX shared strings cap, TS prototype pollution defense, TS SVG depth limit, 2026-03-24
- Test coverage expansion — 71.7% → 82.5% overall; detect 48.8→89.6%, extractor 55.6→100%, email 58.3→93.3%, image 62.1→78.1%, audio 65.2→91.9%, 2026-03-24
- WebP/WAV RIFF detection fix — proper disambiguation via bytes 8-11, both Go and TS, 2026-03-24
- TypeScript PDF parity — pdfjs-dist integration, 3 PDF golden files, Node.js legacy build fix, 2026-03-24
- npm publish prep — repository/homepage/bugs/keywords, prepublishOnly script, exports order fix, 2026-03-24
- Edge-case test coverage — 82.5% → 86.1% (RTF 91.2%, DXF 91.5%, ISOBMFF 86.5%, GeoTIFF improved), 2026-03-24
- OCR CI script + multi-language testing — scripts/test-ocr.sh, German fixture + gated tests, 2026-03-24

## What's Blocked

_None currently._

## Key Numbers

| Metric | Go | TypeScript |
|--------|-----|-----------|
| Runtime deps | 3 (BSD) + 1 optional CGO (gosseract, MIT) | 1 required (fflate, MIT) + 2 optional (pdfjs-dist, tesseract.js) |
| Formats | 27 (DOCX, XLSX, PPTX, ODT, ODS, ODP, PDF, XML, HTML, SVG, EML, JSON, CSV, TSV, YAML, MD, RTF, JPEG, PNG, TIFF, BMP, GIF, WebP, HEIC, HEIF, AVIF, DXF) | Same |
| Test count | 290+ | 65 (52 parity + 5 GeoTIFF + 7 OCR + 1 bench) |
| Test coverage | 86.1% overall | N/A (golden file parity) |
| Golden files | 52 | 52 (49 from Go + 3 PDF from TS) |
| Source lines | ~11,000 | ~5,200 |
| Bundle size | N/A (compiled) | 60 KB ESM / 59 KB CJS |

## What's Next

1. `npm publish --access public` for `@content-extractor/core` (package ready, needs npm auth)
2. Browser-mode integration tests (deferred — high infra cost, manually verified via demo app)
3. Additional format support if needed

---
