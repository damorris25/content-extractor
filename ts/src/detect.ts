// Format detection from magic bytes and file extension.
// Port of detect/detect.go + detect/signatures.go + detect/extensions.go

import { DetectionFailedError } from './extractor.js';
import { safeUnzip } from './formats/zip-utils.js';

/** Detector identifies file formats from content and/or filename. */
export interface Detector {
  detect(data: Uint8Array, filename?: string): string;
}

interface MagicSignature {
  mime: string;
  offset: number;
  magic: Uint8Array;
  mask?: Uint8Array;
}

const MAX_DETECT_BYTES = 8192;

// Magic byte signatures, ordered by specificity.
const signatures: MagicSignature[] = [
  // PDF
  { mime: 'application/pdf', offset: 0, magic: new Uint8Array([0x25, 0x50, 0x44, 0x46]) }, // %PDF
  // ZIP (DOCX, XLSX, PPTX, ODT, ODS, ODP)
  { mime: 'application/zip', offset: 0, magic: new Uint8Array([0x50, 0x4b, 0x03, 0x04]) }, // PK\x03\x04
  // PNG
  { mime: 'image/png', offset: 0, magic: new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]) },
  // JPEG
  { mime: 'image/jpeg', offset: 0, magic: new Uint8Array([0xff, 0xd8, 0xff]) },
  // TIFF (little-endian)
  { mime: 'image/tiff', offset: 0, magic: new Uint8Array([0x49, 0x49, 0x2a, 0x00]) },
  // TIFF (big-endian)
  { mime: 'image/tiff', offset: 0, magic: new Uint8Array([0x4d, 0x4d, 0x00, 0x2a]) },
  // BMP
  { mime: 'image/bmp', offset: 0, magic: new Uint8Array([0x42, 0x4d]) }, // BM
  // WebP (RIFF)
  { mime: 'image/webp', offset: 0, magic: new Uint8Array([0x52, 0x49, 0x46, 0x46]) }, // RIFF
  // GIF87a
  { mime: 'image/gif', offset: 0, magic: new Uint8Array([0x47, 0x49, 0x46, 0x38, 0x37, 0x61]) },
  // GIF89a
  { mime: 'image/gif', offset: 0, magic: new Uint8Array([0x47, 0x49, 0x46, 0x38, 0x39, 0x61]) },
  // MP3 (ID3v2 tag)
  { mime: 'audio/mpeg', offset: 0, magic: new Uint8Array([0x49, 0x44, 0x33]) }, // ID3
  // MP3 (sync word)
  { mime: 'audio/mpeg', offset: 0, magic: new Uint8Array([0xff, 0xe0]), mask: new Uint8Array([0xff, 0xe0]) },
  // OGG
  { mime: 'audio/ogg', offset: 0, magic: new Uint8Array([0x4f, 0x67, 0x67, 0x53]) }, // OggS
  // FLAC
  { mime: 'audio/flac', offset: 0, magic: new Uint8Array([0x66, 0x4c, 0x61, 0x43]) }, // fLaC
  // WAV (RIFF)
  { mime: 'audio/wav', offset: 0, magic: new Uint8Array([0x52, 0x49, 0x46, 0x46]) }, // RIFF
  // OLE2 (legacy Office)
  { mime: 'application/x-ole2', offset: 0, magic: new Uint8Array([0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1]) },
  // DWG
  { mime: 'image/vnd.dwg', offset: 0, magic: new Uint8Array([0x41, 0x43, 0x31, 0x30]) }, // AC10
  // ISOBMFF container (HEIC, HEIF, AVIF) — ftyp box type at offset 4.
  { mime: 'application/x-isobmff', offset: 4, magic: new Uint8Array([0x66, 0x74, 0x79, 0x70]) }, // ftyp
];

const validODFMIME = new Set([
  'application/vnd.oasis.opendocument.text',
  'application/vnd.oasis.opendocument.spreadsheet',
  'application/vnd.oasis.opendocument.presentation',
  'application/vnd.oasis.opendocument.graphics',
  'application/vnd.oasis.opendocument.chart',
  'application/vnd.oasis.opendocument.formula',
  'application/vnd.oasis.opendocument.text-template',
  'application/vnd.oasis.opendocument.spreadsheet-template',
  'application/vnd.oasis.opendocument.presentation-template',
]);

const extMap: Record<string, string> = {
  '.txt': 'text/plain',
  '.csv': 'text/csv',
  '.tsv': 'text/tab-separated-values',
  '.html': 'text/html',
  '.htm': 'text/html',
  '.xml': 'application/xml',
  '.json': 'application/json',
  '.docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  '.xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  '.pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  '.doc': 'application/msword',
  '.xls': 'application/vnd.ms-excel',
  '.ppt': 'application/vnd.ms-powerpoint',
  '.odt': 'application/vnd.oasis.opendocument.text',
  '.ods': 'application/vnd.oasis.opendocument.spreadsheet',
  '.odp': 'application/vnd.oasis.opendocument.presentation',
  '.pdf': 'application/pdf',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.png': 'image/png',
  '.gif': 'image/gif',
  '.bmp': 'image/bmp',
  '.tif': 'image/tiff',
  '.tiff': 'image/tiff',
  '.webp': 'image/webp',
  '.mp3': 'audio/mpeg',
  '.wav': 'audio/wav',
  '.ogg': 'audio/ogg',
  '.flac': 'audio/flac',
  '.yaml': 'application/x-yaml',
  '.yml': 'application/x-yaml',
  '.md': 'text/markdown',
  '.markdown': 'text/markdown',
  '.svg': 'image/svg+xml',
  '.rtf': 'application/rtf',
  '.log': 'text/plain',
  '.heic': 'image/heic',
  '.heif': 'image/heif',
  '.avif': 'image/avif',
  '.dxf': 'image/vnd.dxf',
  '.dwg': 'image/vnd.dwg',
  '.eml': 'message/rfc822',
};

/** Default format detector using magic bytes and extension fallback. */
export class DefaultDetector implements Detector {
  detect(data: Uint8Array, filename?: string): string {
    if (data.length === 0) {
      throw new DetectionFailedError();
    }

    const buf = data.length > MAX_DETECT_BYTES ? data.subarray(0, MAX_DETECT_BYTES) : data;

    // Try magic byte signatures.
    const mime = matchSignature(buf);
    if (mime) {
      // ZIP-based formats need further inspection.
      if (mime === 'application/zip') {
        const zipMIME = detectZIPFormat(data);
        if (zipMIME) return zipMIME;
      }
      // ISOBMFF-based formats need brand inspection.
      if (mime === 'application/x-isobmff') {
        const bmffMIME = detectISOBMFF(buf);
        if (bmffMIME) return bmffMIME;
        // Unknown ISOBMFF brand — fall through to text/extension detection.
      } else {
        return mime;
      }
    }

    // Check for text-based formats.
    const textMIME = detectTextFormat(buf);
    if (textMIME) return textMIME;

    // Fall back to extension.
    if (filename) {
      const extMIME = extensionToMIME(filename);
      if (extMIME) return extMIME;
    }

    throw new DetectionFailedError();
  }
}

function matchSignature(buf: Uint8Array): string | null {
  for (const sig of signatures) {
    if (buf.length < sig.offset + sig.magic.length) continue;
    let match = true;
    for (let i = 0; i < sig.magic.length; i++) {
      const b = buf[sig.offset + i]!;
      const m = sig.magic[i]!;
      if (sig.mask) {
        if ((b & sig.mask[i]!) !== (m & sig.mask[i]!)) {
          match = false;
          break;
        }
      } else {
        if (b !== m) {
          match = false;
          break;
        }
      }
    }
    if (match) return sig.mime;
  }
  return null;
}

/** Max compressed size for ZIP detection (2MB). */
const MAX_ZIP_DETECT_SIZE = 2 * 1024 * 1024;

function detectZIPFormat(data: Uint8Array): string | null {
  try {
    // Limit to 2MB for detection. safeUnzip handles per-entry size + entry count limits.
    const limited = data.length > MAX_ZIP_DETECT_SIZE ? data.subarray(0, MAX_ZIP_DETECT_SIZE) : data;
    const entries = safeUnzip(limited, MAX_ZIP_DETECT_SIZE);
    const names = Object.keys(entries);

    const fileSet = new Set(names);

    // Check ODF formats (mimetype file).
    if (fileSet.has('mimetype')) {
      const mimeData = entries['mimetype'];
      if (mimeData) {
        const mime = new TextDecoder().decode(mimeData).trim();
        if (validODFMIME.has(mime)) return mime;
      }
    }

    // Check OOXML formats.
    if (fileSet.has('word/document.xml')) {
      return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
    }
    if (fileSet.has('xl/workbook.xml')) {
      return 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
    }
    if (fileSet.has('ppt/presentation.xml')) {
      return 'application/vnd.openxmlformats-officedocument.presentationml.presentation';
    }

    return 'application/zip';
  } catch {
    return null;
  }
}

/** ISOBMFF brand-to-MIME mapping. */
const isobmffBrands: Record<string, string> = {
  heic: 'image/heic',
  heix: 'image/heic',
  heim: 'image/heic',
  heis: 'image/heic',
  mif1: 'image/heif',
  msf1: 'image/heif',
  avif: 'image/avif',
  avis: 'image/avif',
};

/**
 * Detect specific ISOBMFF format from ftyp box brands.
 * ftyp layout: [4B size][4B "ftyp"][4B major_brand][4B minor_version][4B compatible_brands...]
 */
function detectISOBMFF(buf: Uint8Array): string | null {
  if (buf.length < 12) return null;

  // Read box size (big-endian uint32).
  let boxSize = (buf[0]! << 24) | (buf[1]! << 16) | (buf[2]! << 8) | buf[3]!;
  if (boxSize < 12 || boxSize > buf.length) {
    boxSize = buf.length;
  }

  // Major brand at bytes 8-11.
  const major = String.fromCharCode(buf[8]!, buf[9]!, buf[10]!, buf[11]!);
  if (isobmffBrands[major]) return isobmffBrands[major]!;

  // Scan compatible brands starting at byte 16.
  for (let i = 16; i + 4 <= boxSize; i += 4) {
    const brand = String.fromCharCode(buf[i]!, buf[i + 1]!, buf[i + 2]!, buf[i + 3]!);
    if (isobmffBrands[brand]) return isobmffBrands[brand]!;
  }

  return null;
}

function detectTextFormat(buf: Uint8Array): string | null {
  const s = new TextDecoder().decode(buf).trimStart();

  // DXF files start with a section header.
  if (s.startsWith('0\nSECTION') || s.startsWith('0\r\nSECTION')) {
    return 'image/vnd.dxf';
  }

  // RTF: starts with {\rtf
  if (s.startsWith('{\\rtf')) {
    return 'application/rtf';
  }

  // XML: starts with <?xml. Check for SVG before generic XML.
  if (s.startsWith('<?xml')) {
    if (s.includes('<svg')) {
      return 'image/svg+xml';
    }
    return 'application/xml';
  }

  // SVG without XML declaration.
  if (s.startsWith('<svg')) {
    return 'image/svg+xml';
  }

  // HTML: starts with <!DOCTYPE html or <html (case-insensitive).
  const lower = s.toLowerCase();
  if (lower.startsWith('<!doctype html') || lower.startsWith('<html')) {
    return 'text/html';
  }

  // EML (RFC 2822): starts with common email headers.
  if (looksLikeEML(s)) {
    return 'message/rfc822';
  }

  // JSON: starts with { or [
  if (s.length > 0 && (s[0] === '{' || s[0] === '[')) {
    return 'application/json';
  }

  return null;
}

/** Check if text starts with RFC 2822 email headers. */
function looksLikeEML(s: string): boolean {
  const prefixes = [
    'From: ',
    'Return-Path: ',
    'Received: ',
    'MIME-Version: ',
    'Message-ID: ',
    'Message-Id: ',
    'Date: ',
    'Subject: ',
  ];
  for (const p of prefixes) {
    if (s.startsWith(p)) return true;
  }
  return false;
}

function extensionToMIME(filename: string): string | null {
  const dot = filename.lastIndexOf('.');
  if (dot < 0) return null;
  const ext = filename.slice(dot).toLowerCase();
  return extMap[ext] ?? null;
}
