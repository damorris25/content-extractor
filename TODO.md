# TODO

<!-- freshness: 2026-03-24 -->
<!-- last-synced: 2026-03-24 — date this file was verified against code -->

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
- [x] XML/HTML/EML extractors (tag stripping, email parsing)
- [x] JSON/CSV/TSV/YAML/Markdown pass-through extractors
- [x] SVG text extractor (text, tspan, textPath elements + title/desc metadata)
- [x] RTF text extractor (control word parser with Unicode support)
- [x] HEIC/HEIF/AVIF image support (ISOBMFF parser for dimensions + EXIF, zero new deps)

## P3 — Future

- [ ] TypeScript PDF parity (generate PDF-specific golden files with pdfjs-dist)
- [ ] npm publish for `@content-extractor/core`
- [ ] TypeScript browser-mode tests (Vitest browser environment)
- [ ] Additional edge-case test coverage (both Go and TS)
- [ ] OCR end-to-end testing with Tesseract installed (CI job with CGO + libtesseract)
- [ ] OCR multi-language testing (beyond English)
- [ ] HEIC/AVIF OCR support (requires image decoder — Tesseract/Leptonica can't decode these formats natively)

## Known Issues & Blockers

### Active Blockers

_None currently._

### Gotchas

- TypeScript PDF extractor requires `pdfjs-dist` peer dep — not included in golden parity tests yet (Go uses `ledongthuc/pdf` which may produce slightly different text)
- `@xmldom/xmldom` needed as dev dep for Node.js tests (browser DOMParser not available in Node)
- RTF font table text leaks into output (e.g. "Times New Roman;") — both Go and TS produce identical output so parity is maintained, accepted as-is

### Workarounds In Place

- `stripInnerXMLDeclarations()` in `ts/src/formats/xml-utils.ts` — removes `<?xml...?>` PIs after document prolog because `@xmldom/xmldom` throws fatalError on them (Go's `encoding/xml` and browser DOMParser both ignore them)
- RTF hex escapes `\'XX` where byte > 0x7F emit `\uFFFD` in TS to match Go's `CleanUTF8` behavior

## Discoveries

_None yet._

## Code Debt

- [ ] WebP detection should check bytes 8-11 for "WEBP" (currently matches any RIFF)
  <!-- TRACKED-TASK in detect/signatures.go:37 -->
- [ ] WAV/WebP disambiguation needs bytes 8-11 check (RIFF + WAVE vs RIFF + WEBP)
  <!-- TRACKED-TASK in detect/signatures.go:57 -->
- [ ] OCR not supported for HEIC/AVIF — Tesseract/Leptonica can't decode these formats; would need a full image decoder
  <!-- TRACKED-TASK in formats/image/image.go:100 and ts/src/formats/image.ts:114 -->

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
- [x] 2026-03-21: OCR text extraction from images (optional, both Go + TS):
  - Go: `gosseract` (MIT, CGO) with build tags (`//go:build cgo && ocr`)
  - TypeScript: `tesseract.js` (Apache 2.0) as optional peer dependency
  - Graceful degradation — metadata-only extraction when OCR unavailable
  - GeoTIFF metadata extraction (CRS, bounds, pixel scale) — pure Go/TS, no deps
  - New test fixtures: geotiff-sample.tif, ocr-text.png, ocr-text.jpg
  - 48 TypeScript tests (37 parity + 5 GeoTIFF + 6 OCR), 40 golden files
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
- [x] 2026-03-24: XML extractor — encoding/xml token parser (Go), state machine (TS)
- [x] 2026-03-24: HTML extractor — golang.org/x/net/html (Go), state machine (TS)
- [x] 2026-03-24: EML extractor — net/mail+mime/multipart (Go), custom parser (TS), multipart depth limit
- [x] 2026-03-24: Text format extractors — JSON, CSV, TSV, YAML, Markdown (pass-through with detection)
- [x] 2026-03-24: SVG extractor — text/tspan/textPath extraction, title/desc metadata
- [x] 2026-03-24: RTF extractor — control word state machine, Unicode escapes, hex byte handling
- [x] 2026-03-24: HEIC/HEIF/AVIF support — custom ISOBMFF parser (~300 lines Go, ~270 lines TS), zero new deps:
  - ftyp brand detection (heic/heix/heim/heis/mif1/msf1/avif/avis)
  - Dimensions from ispe box (meta > iprp > ipco > ispe)
  - EXIF via iloc + iinf/infe (4-byte header prefix skip, TIFF blob to existing parser)
  - Safety: box count cap (1K), item count cap (10K), bounds-checked offsets
- [x] 2026-03-24: OCR downscaling — auto-resize images >2048px via OffscreenCanvas before Tesseract
- [x] 2026-03-24: Demo app updates — progress bar, elapsed time counter, 60s timeout, all 27 formats listed
- [x] 2026-03-24: README updated — 27 formats, new benchmarks, accuracy tables, 52 golden files, 61 TS tests
- [x] 2026-03-24: Benchmarks added for all new formats (Go + TS): HEIC 612 MB/s, AVIF 562 MB/s, RTF 114 MB/s

---
