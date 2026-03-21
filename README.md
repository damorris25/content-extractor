# content-extractor

> **rebar v1.2.0** | **Tier 2: ADOPTED**

A high-performance Go library for extracting textual content and metadata from
unstructured file formats. Inspired by [Apache Tika](https://tika.apache.org/),
built for commercial use with minimal dependencies and a small memory footprint.

**Supported format families:**
- Microsoft Office (DOCX, XLSX, PPTX)
- Open Document Format (ODT, ODS, ODP)
- PDF (text-layer extraction)
- Images (JPEG, PNG, TIFF, BMP, GIF, WebP — EXIF/metadata extraction)
- Audio (MP3, OGG, FLAC, WAV — ID3v2/Vorbis/RIFF metadata)
- CAD (DXF — text entities, layers, and metadata)

**Design constraints:**
- Pure Go with minimal third-party dependencies (2 total, both BSD-licensed)
- No GPL-licensed dependencies (commercial-friendly: MIT, BSD, Apache 2.0 only)
- Low memory footprint — suitable for IoT, containers, and desktop environments
- Streaming-capable for large files

## Performance

Benchmarks run on Apple M5, Go 1.26.1. All times are per-operation on minimal test fixtures.

### Extraction Throughput

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

### Accuracy

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
| JPEG | — | EXIF (dimensions, camera, GPS) | — | Pass |
| PNG | — | tEXt chunks (title, author, etc.) | — | Pass |
| MP3 | Text summary | ID3v2 (title, artist, album, track, year, genre, album artist) | — | Pass |
| OGG | Text summary | Vorbis comments (title, artist, album, track, date, genre) | — | Pass |
| FLAC | Text summary | Vorbis comments (same as OGG) | — | Pass |
| WAV | Text summary | RIFF INFO + fmt (channels, sample rate, bits) | — | Pass |
| DXF | TEXT, MTEXT, ATTRIB, INSERT, DIMENSION entities | Version, units, layers | — | Pass |

Run benchmarks yourself: `go test -bench=. -benchmem ./contentextractor/`

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

```bash
git clone <repo-url>
cd content-extractor
go mod download
go test ./...
```

### Usage

```go
import "github.com/dmorris/content-extractor/contentextractor"

// Auto-detect format and extract (reader must implement io.ReadSeeker)
result, err := contentextractor.Extract(ctx, reader, "")
fmt.Println(result.Text)            // extracted text content
fmt.Println(result.Metadata)        // structured metadata map

// Or specify MIME type explicitly
result, err = contentextractor.Extract(ctx, reader, "application/pdf")
```

## Project Structure

```
content-extractor/
├── extractor/          — core extraction interface and registry
├── formats/            — format-specific extractors
│   ├── office/         — Microsoft Office (OOXML + legacy)
│   ├── odf/            — Open Document Format
│   ├── pdf/            — PDF extraction
│   ├── image/          — image metadata extraction
│   ├── audio/          — audio metadata extraction
│   └── cad/            — CAD text/metadata extraction
├── metadata/           — metadata model and normalization
├── detect/             — MIME type / format detection
├── architecture/       — contract documents
├── agents/             — subagent orchestration templates
├── practices/          — reference guides
├── scripts/            — enforcement and quality scanning
└── testdata/           — test fixture files
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
