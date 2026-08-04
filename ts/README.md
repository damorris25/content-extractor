# @damorris25/content-extractor

[![CI](https://github.com/damorris25/content-extractor/actions/workflows/ci.yml/badge.svg)](https://github.com/damorris25/content-extractor/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/damorris25/content-extractor/blob/master/LICENSE)

Extract text content and metadata from unstructured file formats — Office
documents, OpenDocument, PDF, images, audio, email, CAD, and more. Inspired by
[Apache Tika](https://tika.apache.org/), built browser-first with a single
runtime dependency ([`fflate`](https://www.npmjs.com/package/fflate), MIT).

This is the TypeScript implementation of
[content-extractor](https://github.com/damorris25/content-extractor). A Go
implementation lives in the same repository; both produce identical output,
enforced by golden-file parity tests.

## Install

```bash
npm install @damorris25/content-extractor
```

Optional peer dependencies:

```bash
npm install pdfjs-dist    # PDF text extraction
npm install tesseract.js  # OCR for images
```

## Quick Start

```ts
import { extract } from '@damorris25/content-extractor';

const bytes = new Uint8Array(await file.arrayBuffer()); // or fs.readFileSync
const result = await extract(bytes, { filename: 'report.docx' });

console.log(result.text);        // extracted text content
console.log(result.contentType); // detected MIME type
console.log(result.metadata);    // Record<string, string[]> — title, creator, dates, ...
```

Works in browsers and Node.js. Input is always a `Uint8Array`; extraction is
async and accepts an optional `AbortSignal`:

```ts
const controller = new AbortController();
const result = await extract(bytes, {
  filename: 'big.xlsx',
  signal: controller.signal,
});
```

## Supported Formats

| Family | Formats | Extracted |
|--------|---------|-----------|
| Microsoft Office | DOCX, XLSX, PPTX | Text + document metadata |
| OpenDocument | ODT, ODS, ODP | Text + document metadata |
| PDF | PDF (text layer) | Text (via `pdfjs-dist` peer) |
| Web/Markup | XML, HTML, SVG | Visible text, titles |
| Structured data | JSON, CSV, TSV, YAML, Markdown | Pass-through + detection |
| Rich text | RTF | Text with Unicode support |
| Email | EML (RFC 2822) | Body text + headers |
| Images | JPEG, PNG, TIFF/GeoTIFF, BMP, GIF, WebP, HEIC, HEIF, AVIF | EXIF/metadata, optional OCR |
| Audio | MP3, OGG, FLAC, WAV | ID3v2 / Vorbis / RIFF metadata |
| CAD | DXF | Text entities, layers, metadata |

## Security

Built for untrusted input: ZIP bomb protection (compression ratio + size
caps), integer overflow checks, NUL byte stripping, and bounded parsing for
ISOBMFF containers. Report vulnerabilities via
[GitHub Security Advisories](https://github.com/damorris25/content-extractor/security/advisories/new).

## Documentation

Full documentation, architecture contracts, benchmarks, and the Go
implementation: [github.com/damorris25/content-extractor](https://github.com/damorris25/content-extractor)

## License

[MIT](https://github.com/damorris25/content-extractor/blob/master/LICENSE) © 2026 Dana Morris
