# TODO

<!-- freshness: 2026-03-21 -->
<!-- last-synced: 2026-03-21 — date this file was verified against code -->

Consolidated task tracking for the project. All `TRACKED-TASK:` comments in
code reference sections of this file.

**Two-Tag System:** See AGENTS.md "TODO Tracking" for full rules. Short version:
- `TODO:` in code = untracked = **blocks commit**
- `TRACKED-TASK:` in code = tracked here = commit allowed

---

## P0 — Immediate (do next)

- [x] Define core contracts: C1-EXTRACTOR, I1-METADATA, I2-REGISTRY, C2-DETECT
- [x] Initialize Go module (`go.mod`) and base directory structure
- [x] Implement core types (Extractor interface, Result, errors, metadata model, registry)
- [x] Implement format detection (`detect/` package) — MIME type identification from magic bytes and file extension
- [x] Implement plain text extractor — proof-of-concept for the extraction pipeline
- [x] Implement DOCX extractor (OOXML ZIP-based, `word/document.xml` text extraction)

## P1 — Soon (this sprint / this week)

- [x] Implement XLSX extractor (shared strings + sheet XML)
- [x] Implement PPTX extractor (slide XML text extraction)
- [x] Implement PDF text extractor (text-layer PDF, not OCR) — first third-party dep: `ledongthuc/pdf` (BSD-3-Clause)
- [x] Implement ODT/ODS/ODP extractors (ZIP-based, `content.xml`)

## P2 — Backlog (when time permits)

- [x] Image metadata extraction (EXIF from JPEG/TIFF, basic PNG metadata)
- [x] Audio metadata extraction (ID3v2 from MP3, Vorbis comments from OGG/FLAC, WAV RIFF INFO)
- [x] CAD text extraction (DXF text entities — TEXT, MTEXT, ATTRIB, DIMENSION, INSERT)
- [~] ~~Legacy Office formats (DOC, XLS, PPT)~~ — **won't do** (low demand, OLE2 complexity not justified)
- [~] ~~DWG support~~ — **won't do** (proprietary format, no pure-Go parser, DXF covers the use case)
- [x] Benchmark suite for memory usage and throughput
- [x] License audit tooling (automated check of dependency licenses)
- [x] Quality test content for accuracy testing (rich fixtures + accuracy test suite)
- [x] TypeScript library (browser-first, golden file parity with Go)

## P3 — Future

- [ ] TypeScript PDF parity (generate PDF-specific golden files with pdfjs-dist)
- [ ] npm publish for `@content-extractor/core`
- [ ] TypeScript browser-mode tests (Vitest browser environment)
- [ ] Additional edge-case test coverage (both Go and TS)

## Known Issues & Blockers

### Active Blockers

_None currently._

### Gotchas

- TypeScript PDF extractor requires `pdfjs-dist` peer dep — not included in golden parity tests yet (Go uses `ledongthuc/pdf` which may produce slightly different text)
- `@xmldom/xmldom` needed as dev dep for Node.js tests (browser DOMParser not available in Node)

### Workarounds In Place

- `stripInnerXMLDeclarations()` in `ts/src/formats/xml-utils.ts` — removes `<?xml...?>` PIs after document prolog because `@xmldom/xmldom` throws fatalError on them (Go's `encoding/xml` and browser DOMParser both ignore them)

## Discoveries

_None yet._

## Code Debt

- [ ] WebP detection should check bytes 8-11 for "WEBP" (currently matches any RIFF)
  <!-- TRACKED-TASK in detect/signatures.go:37 -->
- [ ] WAV/WebP disambiguation needs bytes 8-11 check (RIFF + WAVE vs RIFF + WEBP)
  <!-- TRACKED-TASK in detect/signatures.go:57 -->

## Completed

- [x] 2026-03-21: Core types (Extractor, Result, errors, metadata, registry)
- [x] 2026-03-21: Format detection (magic bytes + extension fallback)
- [x] 2026-03-21: Plain text extractor
- [x] 2026-03-21: DOCX extractor with metadata extraction
- [x] 2026-03-21: Integration wiring (contentextractor convenience package)
- [x] 2026-03-21: Contract compliance test harness (extractortest package)
- [x] 2026-03-21: XLSX extractor with shared strings and multi-sheet support
- [x] 2026-03-21: PPTX extractor with DrawingML text and multi-slide support
- [x] 2026-03-21: PDF text extractor (ledongthuc/pdf, BSD-3-Clause)
- [x] 2026-03-21: ODF extractors — ODT, ODS, ODP with shared helpers
- [x] 2026-03-21: Image metadata extractor (JPEG EXIF, PNG tEXt, dimensions for all formats)
- [x] 2026-03-21: Audio metadata extractor (MP3 ID3v2, OGG/FLAC Vorbis, WAV RIFF — pure Go)
- [x] 2026-03-21: DXF text/metadata extractor (TEXT, MTEXT, layers, header vars)
- [x] 2026-03-21: Benchmark suite (per-format throughput, text scaling, detection overhead)
- [x] 2026-03-21: License audit tooling (scripts/check-licenses.sh, all deps pass)
- [x] 2026-03-21: Rich test fixtures + accuracy test suite (7 formats, Unicode, known expected values)
- [x] 2026-03-21: Security hardening — decompression limits, uint32 overflow fixes, panic recovery, NUL stripping, type confusion prevention
- [x] 2026-03-21: TypeScript port — browser-first, 34 golden file parity tests, 1 runtime dep (fflate)
- [x] 2026-03-21: README updated with dual-language docs, parity metrics, TS usage examples
- [x] 2026-03-21: GitHub repo created and pushed (damorris25/content-extractor, master + main)
- [x] 2026-03-21: TypeScript security hardening (9 fixes):
  - High: prototype pollution — `Object.create(null)` + key rejection for `__proto__`/`constructor`/`prototype`
  - Medium: ZIP bomb in detect.ts — replaced raw `unzipSync` with `safeUnzip`
  - Medium: aggregate ZIP decompressed size limit (500MB) in zip-utils.ts
  - Medium: DXF memory reduction — generator-based line iteration instead of `.split()`
  - Medium: `readSyncsafe` bounds check in audio.ts
  - Medium: IFD `dataOffset` validation with count cap in image.ts
  - Low: O(n²) Latin-1 decode → chunked `fromCharCode` in audio.ts
  - Low: PDF page iteration capped at 10K in pdf.ts
  - Low: XXE protection via DOCTYPE stripping for xmldom in xml-utils.ts

---
