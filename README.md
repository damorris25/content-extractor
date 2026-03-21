# content-extractor

> **rebar v1.2.0** | **Tier 2: ADOPTED**

A high-performance library for extracting textual content and metadata from
unstructured file formats. Available in **Go** and **TypeScript** with proven
output parity. Inspired by [Apache Tika](https://tika.apache.org/),
built for commercial use with minimal dependencies and a small memory footprint.

**Supported format families:**
- Microsoft Office (DOCX, XLSX, PPTX)
- Open Document Format (ODT, ODS, ODP)
- PDF (text-layer extraction)
- Images (JPEG, PNG, TIFF, BMP, GIF, WebP — EXIF/metadata + optional OCR text extraction + GeoTIFF support)
- Audio (MP3, OGG, FLAC, WAV — ID3v2/Vorbis/RIFF metadata)
- CAD (DXF — text entities, layers, and metadata)

**Implementations:**

| | Go | TypeScript |
|---|---|---|
| **Runtime** | Go 1.26+ | Browser-first (also Node.js) |
| **Input** | `io.ReadSeeker` (streaming) | `Uint8Array` (buffered) |
| **Runtime deps** | 2 (BSD) + 1 optional CGO (`gosseract`, MIT) | 1 required (`fflate`, MIT), 2 optional peers (`pdfjs-dist`, `tesseract.js`) |
| **Bundle size** | N/A (compiled) | ~60 KB ESM / ~42 KB CJS (before gzip) |
| **Cancellation** | `context.Context` | `AbortSignal` |

**Design constraints:**
- No GPL-licensed dependencies (commercial-friendly: MIT, BSD, Apache 2.0 only)
- Low memory footprint — suitable for IoT, containers, browsers, and desktop environments
- Security-hardened: ZIP bomb protection, integer overflow prevention, NUL byte stripping

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

### TypeScript Accuracy

All extractors produce **identical output** to the Go library, verified by golden file parity tests.
Golden files are generated by Go (`scripts/gen-golden-json.go`) for all 37 test fixtures.
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

**Total: 37 parity tests passing, 40 golden files, 15 formats.**

Run tests yourself: `cd ts && npx vitest run`

### TypeScript Build & Bundle

| Metric | Value |
|--------|-------|
| Build time | 203 ms |
| ESM bundle | 60 KB (15 KB gzipped) |
| CJS bundle | 42 KB (13 KB gzipped) |
| Source lines | ~4,200 |
| Runtime dependencies | 1 (`fflate`, MIT, ~13 KB gzipped) |
| Test execution | 20 ms (34 tests) |

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
import "github.com/dmorris/content-extractor/formats/image"

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
| XXE injection | DOCTYPE stripping before XML parsing (xmldom protection) |
| Memory exhaustion | DXF generator-based line iteration, Latin-1 chunked decode |
| Unbounded iteration | PDF page cap (10K), text entity cap (100K), layer cap (10K) |
| NUL byte injection | All metadata values stripped of NUL bytes |

## Agent Cold Start

**Every agent session starts here.** Read these four files in order:

1. **This file** (`README.md`) — project orientation (you're here)
2. `QUICKCONTEXT.md` — what's true right now (branch, test counts, active work)
3. `TODO.md` — what needs doing + known issues + blockers
4. `AGENTS.md` — how we work (norms, testing cascade, contracts, collaboration)

Then: `CLAUDE.md` for Claude-specific configuration.

**Trust but verify:** After reading QUICKCONTEXT.md, run `git log --oneline -10`
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
import "github.com/dmorris/content-extractor/contentextractor"

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
├── architecture/       — contract documents
├── agents/             — subagent orchestration templates
└── practices/          — reference guides
```

## Core Tenets

1. **Pure Go, Minimal Dependencies** — Prefer stdlib. When third-party deps are necessary, only MIT/BSD/Apache 2.0 licensed. No CGO unless absolutely unavoidable.
2. **Contract-First** — Write the contract, then the code. Every extractor implements a defined behavioral contract.
3. **Memory-Efficient** — Stream large files. Never buffer an entire file in memory. Target environments include IoT devices with limited RAM.
4. **Accuracy Over Speed** — Correct extraction is more important than fast extraction. But do both when possible.
5. **Commercially Safe** — No GPL/LGPL/AGPL dependencies. Every dependency must be auditable for license compliance.

## Contributing

See `AGENTS.md` for development norms, testing cascade, and collaboration
patterns. All code must reference its architecture contract in the file header.

---
