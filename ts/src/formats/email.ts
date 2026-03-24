// CONTRACT:C1-EXTRACTOR.1.0
// EML extractor — extracts text and metadata from RFC 2822 email files.

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, CorruptInputError, checkAborted } from '../extractor.js';
import {
  createMetadata,
  metadataSet,
  KEY_FORMAT,
  KEY_TITLE,
  KEY_CREATOR,
  KEY_DATE,
  KEY_EMAIL_TO,
  KEY_EMAIL_CC,
} from '../metadata.js';

const MIME_EML = 'message/rfc822';
const DEFAULT_MAX_BYTES = 50 * 1024 * 1024; // 50MB
const MAX_MULTIPART_DEPTH = 10; // Prevent stack overflow from nested multipart
const MAX_PARTS = 100; // Max parts per multipart level

/** HTML elements whose text content should be skipped. */
const SKIP_ELEMENTS = new Set(['script', 'style', 'noscript']);

/** Extracts text content and metadata from EML (RFC 2822) email files. */
export class EMLExtractor implements Extractor {
  private readonly maxBytes: number;

  constructor(maxBytes: number = DEFAULT_MAX_BYTES) {
    this.maxBytes = maxBytes;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_EML) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_EML);

    if (data.length === 0) {
      return { text: '', metadata, contentType: MIME_EML, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const content = new TextDecoder('utf-8').decode(data);

    // Split headers from body at the first blank line.
    const { headers, body } = parseMessage(content);
    if (!headers) {
      throw new CorruptInputError('failed to parse EML headers');
    }

    // Extract metadata from headers.
    const subject = headers.get('subject');
    if (subject) metadataSet(metadata, KEY_TITLE, subject);

    const from = headers.get('from');
    if (from) metadataSet(metadata, KEY_CREATOR, from);

    const to = headers.get('to');
    if (to) metadataSet(metadata, KEY_EMAIL_TO, to);

    const cc = headers.get('cc');
    if (cc) metadataSet(metadata, KEY_EMAIL_CC, cc);

    const date = headers.get('date');
    if (date) {
      const parsed = parseRFC2822Date(date);
      if (parsed) metadataSet(metadata, KEY_DATE, parsed);
    }

    // Extract body text.
    const contentType = headers.get('content-type') || 'text/plain';
    const text = extractBody(body, contentType);

    return { text: text.trim(), metadata, contentType: MIME_EML, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [MIME_EML];
  }
}

/** Parse RFC 2822 message into headers and body. */
function parseMessage(content: string): { headers: Map<string, string> | null; body: string } {
  // Find the blank line separating headers from body.
  // RFC 2822 uses \r\n\r\n, but we also handle \n\n.
  let splitIdx = content.indexOf('\r\n\r\n');
  let bodyOffset = 4;
  if (splitIdx < 0) {
    splitIdx = content.indexOf('\n\n');
    bodyOffset = 2;
  }

  if (splitIdx < 0) {
    // No body — entire content is headers.
    return { headers: parseHeaders(content), body: '' };
  }

  const headerSection = content.substring(0, splitIdx);
  const body = content.substring(splitIdx + bodyOffset);
  return { headers: parseHeaders(headerSection), body };
}

/** Parse RFC 2822 headers, handling folded (continuation) lines. */
function parseHeaders(headerText: string): Map<string, string> {
  const headers = new Map<string, string>();
  // Unfold continuation lines: lines starting with whitespace are part of the previous header.
  const unfolded = headerText.replace(/\r?\n[ \t]+/g, ' ');
  const lines = unfolded.split(/\r?\n/);

  for (const line of lines) {
    const colonIdx = line.indexOf(':');
    if (colonIdx < 0) continue;
    const key = line.substring(0, colonIdx).trim().toLowerCase();
    const value = line.substring(colonIdx + 1).trim();
    if (key && value) {
      headers.set(key, value);
    }
  }

  return headers;
}

/** Extract body text, handling multipart and HTML content. */
function extractBody(body: string, contentType: string): string {
  const { mediaType, params } = parseContentType(contentType);

  // Multipart: prefer text/plain parts.
  if (mediaType.startsWith('multipart/')) {
    const boundary = params.get('boundary');
    if (!boundary) return '';
    return extractMultipart(body, boundary);
  }

  // Single part.
  if (mediaType.startsWith('text/html')) {
    return stripHTMLTags(body);
  }
  return body;
}

/** Parse a Content-Type header into media type and parameters. */
function parseContentType(ct: string): { mediaType: string; params: Map<string, string> } {
  const params = new Map<string, string>();
  const parts = ct.split(';');
  const mediaType = (parts[0] || 'text/plain').trim().toLowerCase();

  for (let i = 1; i < parts.length; i++) {
    const param = parts[i]!.trim();
    const eqIdx = param.indexOf('=');
    if (eqIdx < 0) continue;
    const key = param.substring(0, eqIdx).trim().toLowerCase();
    let value = param.substring(eqIdx + 1).trim();
    // Remove quotes.
    if (value.startsWith('"') && value.endsWith('"')) {
      value = value.substring(1, value.length - 1);
    }
    params.set(key, value);
  }

  return { mediaType, params };
}

/** Extract text from multipart body, preferring text/plain parts. */
function extractMultipart(body: string, boundary: string, depth = 0): string {
  if (depth >= MAX_MULTIPART_DEPTH) return '';

  const delimiter = '--' + boundary;
  const endDelimiter = delimiter + '--';
  const parts = body.split(delimiter);

  const plainParts: string[] = [];
  const htmlParts: string[] = [];
  let partCount = 0;

  for (const part of parts) {
    if (partCount >= MAX_PARTS) break;

    const trimmed = part.trim();
    if (!trimmed || trimmed === '--' || trimmed.startsWith('--')) continue;

    partCount++;

    // Split part headers from part body.
    let partSplitIdx = trimmed.indexOf('\r\n\r\n');
    let partBodyOffset = 4;
    if (partSplitIdx < 0) {
      partSplitIdx = trimmed.indexOf('\n\n');
      partBodyOffset = 2;
    }
    if (partSplitIdx < 0) continue;

    const partHeaders = trimmed.substring(0, partSplitIdx);
    let partBody = trimmed.substring(partSplitIdx + partBodyOffset);

    // Remove trailing end delimiter.
    if (partBody.endsWith(endDelimiter)) {
      partBody = partBody.substring(0, partBody.length - endDelimiter.length);
    }
    // Remove trailing delimiter (next boundary).
    if (partBody.endsWith(delimiter)) {
      partBody = partBody.substring(0, partBody.length - delimiter.length);
    }
    partBody = partBody.trim();

    // Get part content type.
    const headers = parseHeaders(partHeaders);
    const partCT = headers.get('content-type') || 'text/plain';
    const { mediaType: partMedia, params: partParams } = parseContentType(partCT);

    // Handle nested multipart.
    if (partMedia.startsWith('multipart/')) {
      const nestedBoundary = partParams.get('boundary');
      if (nestedBoundary) {
        const nested = extractMultipart(partBody, nestedBoundary, depth + 1);
        if (nested) plainParts.push(nested);
      }
      continue;
    }

    if (!partMedia.startsWith('text/')) continue;

    if (partMedia.startsWith('text/plain')) {
      plainParts.push(partBody);
    } else if (partMedia.startsWith('text/html')) {
      htmlParts.push(partBody);
    }
  }

  // Prefer plain text over HTML.
  if (plainParts.length > 0) {
    return plainParts.join('\n');
  }
  if (htmlParts.length > 0) {
    return htmlParts.map(stripHTMLTags).join('\n');
  }
  return '';
}

/**
 * Strip HTML tags and extract visible text.
 * Uses regex-based approach for efficiency (avoids O(n^2) string concatenation).
 * Skips script/style/noscript content.
 */
function stripHTMLTags(s: string): string {
  // Remove script, style, noscript blocks entirely.
  let cleaned = s.replace(/<(script|style|noscript)\b[^>]*>[\s\S]*?<\/\1>/gi, '');

  // Remove all remaining tags.
  cleaned = cleaned.replace(/<[^>]*>/g, '');

  // Decode common HTML entities.
  cleaned = cleaned
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, ' ');

  // Normalize whitespace: collapse blank lines.
  const lines = cleaned.split('\n');
  const result: string[] = [];
  for (const line of lines) {
    const trimmed = line.trim();
    if (trimmed) result.push(trimmed);
  }
  return result.join('\n');
}

/** Parse RFC 2822 date to ISO 8601 UTC string. */
function parseRFC2822Date(dateStr: string): string | null {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return null;
  return d.toISOString().replace(/\.\d{3}Z$/, 'Z');
}
