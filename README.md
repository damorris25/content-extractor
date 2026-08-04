# content-extractor

[![CI](https://github.com/damorris25/content-extractor/actions/workflows/ci.yml/badge.svg)](https://github.com/damorris25/content-extractor/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/damorris25/content-extractor.svg)](https://pkg.go.dev/github.com/damorris25/content-extractor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A high-performance library for extracting textual content and metadata from
unstructured file formats. Available in **Go** and **TypeScript** with proven
output parity. Inspired by [Apache Tika](https://tika.apache.org/),
built for commercial use with minimal dependencies and a small memory footprint.

**Supported format families:**
- Microsoft Office (DOCX, XLSX, PPTX)
- Open Document Format (ODT, ODS, ODP)
- PDF (text-layer extraction)
- Web/Markup (XML, HTML, SVG — tag stripping with visible text extraction)
- Structured Data (JSON, CSV, TSV, YAML — pass-through with format detection)
- Markdown (pass-through with format detection)
- Rich Text (RTF — control word parsing with Unicode support)
- Email (EML — RFC 2822 with multipart MIME support)
- Images (JPEG, PNG, TIFF, BMP, GIF, WebP, HEIC, HEIF, AVIF — EXIF/metadata + optional OCR + GeoTIFF + ISOBMFF)
- Audio (MP3, OGG, FLAC, WAV — ID3v2/Vorbis/RIFF metadata)
- CAD (DXF — text entities, layers, and metadata)

**Implementations:**

| | Go | TypeScript |
|---|---|---|
| **Runtime** | Go 1.26+ | Browser-first (also Node.js) |
| **Input** | `io.ReadSeeker` (streaming) | `Uint8Array` (buffered) |
| **Runtime deps** | 3 (BSD) + 1 optional CGO (`gosseract`, MIT) | 1 required (`fflate`, MIT), 2 optional peers (`pdfjs-dist`, `tesseract.js`) |
| **Bundle size** | N/A (compiled) | ~60 KB ESM / ~42 KB CJS (before gzip) |
| **Cancellation** | `context.Context` | `AbortSignal` |

**Design constraints:**
- No GPL-licensed dependencies (commercial-friendly: MIT, BSD, Apache 2.0 only)
- Low memory footprint — suitable for IoT, containers, browsers, and desktop environments
- Security-hardened: ZIP bomb protection, integer overflow prevention, NUL byte stripping, ISOBMFF box count limits

## Go Performance

Benchmarks run on Apple M5, Go 1.26.1. All times are per-operation on minimal test fixtures.

### Go Extraction Throughput

| Format | Throughput | Time/Op | Allocs/Op | Memory/Op |
|--------|-----------|---------|-----------|-----------|
| Plain Text (1KB) | 494 MB/s | 2.1 us | 7 | 34 KB |
| DOCX | 83 MB/s | 25 us | 342 | 51 KB |
| XLSX | 80 MB/s | 36 us | 516 | 68 KB |
| PPTX | 92 MB/s | 28 us | 382 | 54 KB |
| PDF | 54 MB/s | 11 us | 326 | 64 KB |
| ODT | 46 MB/s | 18 us | 252 | 39 KB |
| ODS | 34 MB/s | 28 us | 365 | 44 KB |
| ODP | 41 MB/s | 23 us | 298 | 42 KB |
| JPEG (metadata) | 216 MB/s | 3.3 us | 31 | 33 KB |
| PNG (metadata) | 154 MB/s | 2.0 us | 46 | 16 KB |
| MP3 (ID3v2) | 546 MB/s | 1.0 us | 48 | 3 KB |
| OGG (Vorbis) | 244 MB/s | 0.9 us | 47 | 2 KB |
| FLAC (Vorbis) | 205 MB/s | 0.9 us | 47 | 2 KB |
| WAV (RIFF) | 455 MB/s | 0.5 us | 23 | 1 KB |
| DXF (CAD) | 266 MB/s | 1.7 us | 53 | 6 KB |
| XML (tag stripping) | 69 MB/s | 6.6 us | 118 | 14 KB |
| HTML (text extraction) | 80 MB/s | 5.8 us | 100 | 20 KB |
| SVG (text extraction) | 72 MB/s | 7.1 us | 156 | 16 KB |
| EML (simple) | 84 MB/s | 3.4 us | 108 | 19 KB |
| EML (multipart) | 97 MB/s | 5.3 us | 140 | 27 KB |
| JSON (pass-through) | 59 MB/s | 2.7 us | 7 | 33 KB |
| CSV (pass-through) | 26 MB/s | 2.7 us | 7 | 33 KB |
| YAML (pass-through) | 65 MB/s | 2.7 us | 7 | 33 KB |
| Markdown (pass-through) | 68 MB/s | 2.7 us | 7 | 33 KB |
| RTF | 114 MB/s | 2.1 us | 37 | 11 KB |
| HEIC (metadata) | 612 MB/s | 0.8 us | 41 | 3 KB |
| AVIF (metadata) | 562 MB/s | 0.7 us | 39 | 2 KB |
| GeoTIFF (metadata) | — | 4.5 us | 206 | 18 KB |
| **OCR — PNG** | — | **12 ms** | 8 | 550 B |
| **OCR — JPEG** | — | **13 ms** | 8 | 546 B |
| OCR + metadata (PNG) | — | 12 ms | 30 | 8 KB |

> OCR benchmarks require Tesseract (`CGO_ENABLED=1 go test -tags ocr -bench=BenchmarkOCR`).
> OCR throughput is dominated by Tesseract recognition time, not Go overhead (30 allocs vs 8 for raw OCR).

### Text Scaling (Plain Text)

| Input Size | Throughput | Time/Op |
|------------|-----------|---------|
| 1 KB | 407 MB/s | 2.5 us |
| 10 KB | 2.9 GB/s | 3.6 us |
| 100 KB | 4.6 GB/s | 22 us |
| 1 MB | 2.4 GB/s | 442 us |

### Go Accuracy

All extractors pass accuracy tests against rich test fixtures with known expected content.

| Format | Text Extraction | Metadata | Unicode | Accuracy Tests |
|--------|----------------|----------|---------|----------------|
| DOCX | Full paragraphs, `&amp;` entities | Title, creator, subject, description, dates | CJK, Korean, French accents, symbols | Pass |
| XLSX | Headers, cell values (shared strings + numeric) | Title, creator | — | Pass |
| PPTX | All slides, bullet points, multi-shape | Title, creator | — | Pass |
| ODT | Paragraphs, headings | Title, creator, subject, description | — | Pass |
| ODS | Cell values, multi-table | Title, creator | — | Pass |
| ODP | Slide text (excludes notes) | Title, creator | — | Pass |
| PDF | Text-layer extraction | Format detection | — | Pass |
| JPEG | Optional OCR text | EXIF (dimensions, camera, GPS) | — | Pass |
| PNG | Optional OCR text | tEXt chunks (title, author, etc.) | — | Pass |
| TIFF/GeoTIFF | Optional OCR text | EXIF + GeoTIFF (CRS, bounds, pixel scale) | — | Pass |
| BMP | Optional OCR text | Dimensions | — | Pass |
| MP3 | Text summary | ID3v2 (title, artist, album, track, year, genre, album artist) | — | Pass |
| OGG | Text summary | Vorbis comments (title, artist, album, track, date, genre) | — | Pass |
| FLAC | Text summary | Vorbis comments (same as OGG) | — | Pass |
| WAV | Text summary | RIFF INFO + fmt (channels, sample rate, bits) | — | Pass |
| DXF | TEXT, MTEXT, ATTRIB, INSERT, DIMENSION entities | Version, units, layers | — | Pass |
| XML | CharData text content (tags stripped) | Format detection | — | Pass |
| HTML | Visible text (skips script/style/noscript) | Title extraction | — | Pass |
| EML | Plain text body (multipart: prefers text/plain) | Subject, From, To, Cc, Date | — | Pass |
| SVG | Text elements (text, tspan, textPath) | Title, description | — | Pass |
| JSON | Pass-through content | Format detection | — | Pass |
| CSV | Pass-through content | Format detection | — | Pass |
| TSV | Pass-through content | Format detection | — | Pass |
| YAML | Pass-through content | Format detection | — | Pass |
| Markdown | Pass-through content | Format detection | — | Pass |
| RTF | Paragraph text, Unicode escapes | Format detection | — | Pass |
| HEIC | Metadata text summary | ISOBMFF dimensions, EXIF (camera, GPS, dates) | — | Pass |
| HEIF | Metadata text summary | ISOBMFF dimensions, EXIF (camera, GPS, dates) | — | Pass |
| AVIF | Metadata text summary | ISOBMFF dimensions, EXIF (camera, GPS, dates) | — | Pass |

### Go OCR Accuracy

| Fixture | Expected Text | Recognized | Confidence | Status |
|---------|--------------|------------|------------|--------|
| ocr-text.png | "Hello World 12345" | "Hello World 12345" | >50% | Pass |
| ocr-text.jpg | "Hello World 12345" | "Hello World 12345" | >50% | Pass |

> OCR accuracy depends on image quality, font, resolution, and Tesseract model version.
> Bitmap-rendered test fixtures use a 5x7 pixel font at 200x50 resolution.

Run benchmarks yourself: `go test -bench=. -benchmem ./contentextractor/`

## TypeScript Performance

Benchmarks run on Apple M5, Node.js, using the same test fixtures as Go.
All times are per-operation averages over 500 iterations after warmup.

### TypeScript Extraction Throughput

| Format | File Size | Avg Time | Ops/sec |
|--------|-----------|----------|---------|
| DOCX | 2.0 KB | 69 us | 14,567 |
| DOCX (rich) | 2.2 KB | 78 us | 12,789 |
| XLSX | 2.8 KB | 93 us | 10,751 |
| PPTX | 2.5 KB | 66 us | 15,210 |
| ODT | 849 B | 36 us | 28,159 |
| ODS | 941 B | 57 us | 17,602 |
| ODP | 945 B | 44 us | 22,591 |
| JPEG (metadata) | 708 B | 914 ns | 1,093,891 |
| PNG (metadata) | 469 B | 706 ns | 1,416,932 |
| PNG (tEXt) | 312 B | 2 us | 609,942 |
| GIF (dimensions) | 63 B | 340 ns | 2,939,741 |
| MP3 (ID3v2) | 538 B | 3 us | 380,385 |
| MP3 (rich ID3v2) | 670 B | 2 us | 449,085 |
| OGG (Vorbis) | 225 B | 2 us | 448,062 |
| FLAC (Vorbis) | 183 B | 2 us | 524,521 |
| WAV (RIFF) | 240 B | 1 us | 722,152 |
| DXF (CAD) | 452 B | 3 us | 301,978 |
| DXF (rich) | 723 B | 4 us | 241,560 |
| XML | 459 B | 4 us | 273,585 |
| HTML | 462 B | 5 us | 191,890 |
| EML (simple) | 282 B | 3 us | 383,853 |
| EML (multipart) | 515 B | 3 us | 295,574 |
| JSON | 160 B | 2 us | 538,866 |
| CSV | 72 B | 679 ns | 1,473,657 |
| YAML | 173 B | 287 ns | 3,489,403 |
| Markdown | 185 B | 327 ns | 3,055,002 |
| SVG | 512 B | 3 us | 308,984 |
| RTF | 235 B | 5 us | 216,497 |
| HEIC (metadata) | 490 B | 5 us | 219,058 |
| AVIF (metadata) | 399 B | 3 us | 292,028 |
| **OCR — PNG** | 1.2 KB | **15 ms** | 67 |
| **OCR — JPEG** | 1.5 KB | **16 ms** | 63 |

> OCR requires `tesseract.js` peer dependency. First call includes ~127ms warm-up (WASM + model load).
> Subsequent calls average 15-16ms. Worker is reused across calls (singleton pattern).

### TypeScript Accuracy

All extractors produce **identical output** to the Go library, verified by golden file parity tests.
Golden files are generated by Go (`scripts/gen-golden-json.go`) for all test fixtures.
The TypeScript test suite asserts exact match on text, metadata, content type, and language.

| Format | Text Extraction | Metadata | Parity Tests | Status |
|--------|----------------|----------|--------------|--------|
| DOCX | Full paragraphs, `&amp;` entities | Title, creator, subject, description, dates | 5 (simple, multipar, empty, rich, unicode) | Pass |
| XLSX | Headers, cell values (shared strings + numeric) | Title, creator | 4 (simple, multisheet, empty, rich) | Pass |
| PPTX | All slides, bullet points, multi-shape | Title, creator | 4 (simple, multislide, empty, rich) | Pass |
| ODT | Paragraphs, headings | Title, creator, subject, description | 4 (simple, multipar, empty, rich) | Pass |
| ODS | Cell values, multi-table | Title, creator | 3 (simple, multisheet, empty) | Pass |
| ODP | Slide text (excludes notes) | Title, creator | 3 (simple, multislide, empty) | Pass |
| PDF | — | — | — | Requires `pdfjs-dist` peer dep |
| JPEG | Optional OCR text | EXIF (dimensions, camera, GPS) | 3 (simple, ocr-text) | Pass |
| PNG | Optional OCR text | tEXt chunks (title, author, etc.), dimensions | 4 (simple, text, ocr-text) | Pass |
| TIFF/GeoTIFF | Optional OCR text | EXIF + GeoTIFF (CRS, bounds, pixel scale) | 1 (geotiff-sample) | Pass |
| GIF | — | Dimensions | 1 (simple) | Pass |
| MP3 | Text summary | ID3v2 (title, artist, album, track, year, genre, album artist) | 2 (simple, rich) | Pass |
| OGG | Text summary | Vorbis comments (title, artist, album, track, date, genre) | 1 (simple) | Pass |
| FLAC | Text summary | Vorbis comments (same as OGG) | 1 (simple) | Pass |
| WAV | Text summary | RIFF INFO + fmt (channels, sample rate, bits) | 1 (simple) | Pass |
| DXF | TEXT, MTEXT, ATTRIB, INSERT, DIMENSION entities | Version, units, layers | 2 (simple, rich) | Pass |
| XML | CharData text content (tags stripped) | Format detection | 1 (simple) | Pass |
| HTML | Visible text (skips script/style/noscript) | Title extraction | 1 (simple) | Pass |
| EML | Plain text body (multipart support) | Subject, From, To, Cc, Date | 2 (simple, multipart) | Pass |
| SVG | Text elements (text, tspan, textPath) | Title, description | 1 (simple) | Pass |
| JSON | Pass-through content | Format detection | 1 (simple) | Pass |
| CSV | Pass-through content | Format detection | 1 (simple) | Pass |
| YAML | Pass-through content | Format detection | 1 (simple) | Pass |
| Markdown | Pass-through content | Format detection | 1 (simple) | Pass |
| RTF | Paragraph text, Unicode escapes | Format detection | 1 (simple) | Pass |
| HEIC | Metadata text summary | ISOBMFF dimensions, EXIF | 1 (simple) | Pass |
| AVIF | Metadata text summary | ISOBMFF dimensions, EXIF | 1 (simple) | Pass |

### TypeScript OCR Accuracy

| Fixture | Expected Text | Recognized | Confidence | Status |
|---------|--------------|------------|------------|--------|
| ocr-text.png | "Hello World 12345" | "Hello World 12343" | 76% | Pass |
| ocr-text.jpg | "Hello World 12345" | Contains "Hello" | >50% | Pass |

> TypeScript OCR uses tesseract.js (WebAssembly). Minor character-level differences from Go (gosseract/CGO)
> are expected — this is why OCR tests are separate from golden file parity tests.

**Total: 49 parity tests passing, 52 golden files, 27 formats + 6 OCR tests + 5 GeoTIFF tests.**

Run tests yourself: `cd ts && npx vitest run`

### TypeScript Build & Bundle

| Metric | Value |
|--------|-------|
| Build time | 203 ms |
| ESM bundle | 60 KB (15 KB gzipped) |
| CJS bundle | 42 KB (13 KB gzipped) |
| Source lines | ~4,600 |
| Runtime dependencies | 1 (`fflate`, MIT, ~13 KB gzipped) |
| Test execution | 25 ms (61 tests) |

### TypeScript Quick Start

```bash
cd ts
npm install
npm run build
npm test
```

### TypeScript Usage

```typescript
import { extract, createRegistry } from '@content-extractor/core';

// One-shot extraction (auto-detects format)
const result = await extract(fileBytes);
console.log(result.text);        // extracted text content
console.log(result.metadata);    // Record<string, string[]>
console.log(result.contentType); // detected MIME type

// Reusable registry (recommended for repeated use)
const registry = createRegistry();
const result2 = await registry.extract(data, 'application/pdf');

// With cancellation
const controller = new AbortController();
const result3 = await registry.extract(data, undefined, {
  signal: controller.signal,
});
```

### PDF Support (Optional)

PDF extraction requires `pdfjs-dist` as a peer dependency:

```bash
npm install pdfjs-dist
```

If not installed, all other formats work fine — PDF extraction throws a clear error.

### OCR Support (Optional)

OCR enables text extraction from images (scanned documents, photos of signs, etc.).
It is **opt-in** — disabled by default with zero overhead.

**Go:**
```bash
# Install Tesseract
brew install tesseract              # macOS
apt install libtesseract-dev        # Debian/Ubuntu

# Build with OCR enabled
CGO_ENABLED=1 go build -tags ocr ./...

# Run OCR tests
CGO_ENABLED=1 go test -tags ocr ./formats/image/
```

```go
import "github.com/damorris25/content-extractor/formats/image"

// Create extractor with OCR enabled
engine, err := image.NewOCREngine("eng")
if err != nil { /* Tesseract not available */ }
defer engine.Close()

ext := &image.Extractor{OCR: engine, OCRLanguage: "eng"}
result, err := ext.Extract(ctx, reader, "image/png")
// result.Text contains OCR-recognized text
// result.Metadata["ocr:confidence"] has confidence score
```

**TypeScript:**
```bash
npm install tesseract.js
```

```typescript
import { ImageExtractor } from '@content-extractor/core';

const ext = new ImageExtractor({ enableOCR: true, ocrLanguage: 'eng' });
const result = await ext.extract(imageBytes, 'image/png');
// result.text contains OCR-recognized text
// result.metadata['ocr:confidence'] has confidence score
```

Without Tesseract (Go) or tesseract.js (TypeScript), image extraction returns metadata only — no error.

### GeoTIFF Support

GeoTIFF metadata is automatically extracted from TIFF files containing geospatial tags:

- `geo:crs` — Coordinate Reference System (e.g., "WGS 84")
- `geo:bounds` — Bounding box as "minX,minY,maxX,maxY"
- `geo:pixel-scale` — Pixel dimensions as "scaleX,scaleY"

No additional dependencies required — pure Go/TypeScript TIFF tag parsing.

### TypeScript Security Hardening

| Protection | Implementation |
|-----------|---------------|
| Prototype pollution | `Object.create(null)` metadata + key rejection |
| ZIP bomb | Per-entry size limit (100 MB) + aggregate limit (500 MB) + entry count cap (10K) |
| Integer overflow | IFD count validation, syncsafe bounds checks, DataView bounds checks |
| XXE injection | DOCTYPE stripping before XML parsing (xmldom + XML extractor) |
| Memory exhaustion | DXF generator-based line iteration, Latin-1 chunked decode |
| Unbounded iteration | PDF page cap (10K), text entity cap (100K), layer cap (10K) |
| Multipart recursion | EML nested multipart depth limit (10) + part count limit (100) |
| NUL byte injection | All metadata values stripped of NUL bytes |
| ISOBMFF parsing | Box count cap (1K), item count cap (10K), bounds-checked offsets |

## Agent Cold Start

**Every agent session starts here.** Read these files in order:

1. **This file** (`README.md`) — project orientation (you're here)
2. `TODO.md` — what needs doing + known issues + blockers
3. `AGENTS.md` — how we work (norms, testing cascade, contracts, collaboration)

**Trust but verify:** After reading the docs, run `git log --oneline -10`
and compare. If docs and reality disagree, reality wins.

## Architecture

This project uses contract-driven development. All contracts live in
`architecture/`. See [DESIGN.md](DESIGN.md) for the full philosophy.

```bash
# Find all contracts
ls architecture/CONTRACT-*.md

# Find what contract a code file implements
head -10 path/to/file.go    # read the CONTRACT: header

# Find all code implementing a specific contract
grep -rn "CONTRACT:C1-EXTRACTOR" .
```

**Contract categories for this project:**

| Prefix | Meaning | Example |
|--------|---------|---------|
| `C` | Component | `C1-EXTRACTOR` (core interface), `C2-DOCX`, `C3-PDF` |
| `I` | Interface | `I1-METADATA` (metadata model), `I2-REGISTRY` (format registry) |
| `P` | Protocol | `P1-STREAM` (streaming extraction protocol) |

**Rule:** Don't implement without a contract. Don't modify code without
checking its contract. See `architecture/README.md` for details.

## Quick Start

### Go

```bash
git clone <repo-url>
cd content-extractor
go mod download
go test ./...
```

```go
import "github.com/damorris25/content-extractor/contentextractor"

// Auto-detect format and extract (reader must implement io.ReadSeeker)
result, err := contentextractor.Extract(ctx, reader, "")
fmt.Println(result.Text)            // extracted text content
fmt.Println(result.Metadata)        // structured metadata map

// Or specify MIME type explicitly
result, err = contentextractor.Extract(ctx, reader, "application/pdf")
```

### TypeScript

```bash
cd ts && npm install && npm run build && npm test
```

```typescript
import { extract } from '@content-extractor/core';

const result = await extract(fileBytes);
console.log(result.text, result.metadata, result.contentType);
```

## Project Structure

```
content-extractor/
├── extractor/          — core extraction interface and registry (Go)
├── formats/            — format-specific extractors (Go)
│   ├── office/         — Microsoft Office (OOXML + legacy)
│   ├── odf/            — Open Document Format
│   ├── pdf/            — PDF extraction
│   ├── text/           — plain text, XML, HTML extraction
│   ├── email/          — EML (RFC 2822) email extraction
│   ├── image/          — image metadata extraction
│   ├── audio/          — audio metadata extraction
│   └── cad/            — CAD text/metadata extraction
├── metadata/           — metadata model and normalization (Go)
├── detect/             — MIME type / format detection (Go)
├── ts/                 — TypeScript implementation (browser-first)
│   ├── src/            — source (extractor, metadata, detect, formats)
│   ├── __tests__/      — parity tests + golden files
│   └── dist/           — built ESM + CJS bundles
├── testdata/           — shared test fixture files (Go + TS)
├── scripts/            — enforcement, quality scanning, golden file generation
└── architecture/       — contract documents
```

## Core Tenets

1. **Pure Go, Minimal Dependencies** — Prefer stdlib. When third-party deps are necessary, only MIT/BSD/Apache 2.0 licensed. No CGO unless absolutely unavoidable.
2. **Contract-First** — Write the contract, then the code. Every extractor implements a defined behavioral contract.
3. **Memory-Efficient** — Stream large files. Never buffer an entire file in memory. Target environments include IoT devices with limited RAM.
4. **Accuracy Over Speed** — Correct extraction is more important than fast extraction. But do both when possible.
5. **Commercially Safe** — No GPL/LGPL/AGPL dependencies. Every dependency must be auditable for license compliance.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) to get started, and `AGENTS.md` for
development norms, testing cascade, and collaboration patterns. All code must
reference its architecture contract in the file header.

## License

[MIT](LICENSE) © 2026 Dana Morris

---
