# content-extractor

> **rebar v1.2.0** | **Tier 2: ADOPTED**

A high-performance Go library for extracting textual content and metadata from
unstructured file formats. Inspired by [Apache Tika](https://tika.apache.org/),
built for commercial use with minimal dependencies and a small memory footprint.

**Supported format families:**
- Microsoft Office (DOCX, XLSX, PPTX, DOC, XLS, PPT)
- Open Document Format (ODT, ODS, ODP)
- PDF
- Images (JPEG, PNG, TIFF, BMP, WebP — EXIF/metadata extraction)
- Audio (MP3, WAV, FLAC, OGG — ID3/metadata extraction)
- CAD (DXF, DWG — text/metadata extraction)

**Design constraints:**
- Pure Go with minimal third-party dependencies
- No GPL-licensed dependencies (commercial-friendly: MIT, BSD, Apache 2.0 only)
- Low memory footprint — suitable for IoT, containers, and desktop environments
- Streaming-capable for large files

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
import "github.com/dmorris/content-extractor/extractor"

result, err := extractor.Extract(ctx, reader, "application/pdf")
fmt.Println(result.Text)
fmt.Println(result.Metadata)
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
