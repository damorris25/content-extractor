// CONTRACT:C1-EXTRACTOR.1.0
// RTF extractor — strips RTF markup and returns plain text.

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT } from '../metadata.js';

const MIME_RTF = 'application/rtf';
const DEFAULT_MAX_BYTES = 10 * 1024 * 1024; // 10MB

/** Extracts text content from RTF files by stripping markup. */
export class RtfExtractor implements Extractor {
  private readonly maxBytes: number;

  constructor(maxBytes: number = DEFAULT_MAX_BYTES) {
    this.maxBytes = maxBytes;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_RTF) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_RTF);

    if (data.length === 0) {
      return { text: '', metadata, contentType: MIME_RTF, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const content = new TextDecoder('utf-8').decode(data);
    const text = extractRTFText(content);

    return { text, metadata, contentType: MIME_RTF, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [MIME_RTF];
  }
}

/**
 * Strip RTF markup and return plain text.
 * Handles: group nesting {}, control words \word, Unicode \uN,
 * hex escapes \'XX, destination groups {\*\...}.
 */
function extractRTFText(rtf: string): string {
  const parts: string[] = [];
  let i = 0;
  const n = rtf.length;
  let groupDepth = 0;
  let skipDepth = 0;

  while (i < n) {
    const ch = rtf[i]!;

    if (ch === '{') {
      groupDepth++;
      // Check for destination group: {\*\...}
      if (i + 3 < n && rtf[i + 1] === '\\' && rtf[i + 2] === '*') {
        skipDepth = groupDepth;
      }
      i++;
      continue;
    }

    if (ch === '}') {
      if (groupDepth === skipDepth) {
        skipDepth = 0;
      }
      if (groupDepth > 0) groupDepth--;
      i++;
      continue;
    }

    if (ch === '\\') {
      if (skipDepth > 0) {
        i = skipControlWord(rtf, i);
        continue;
      }

      i++; // skip backslash
      if (i >= n) break;

      const nextCh = rtf[i]!;

      // Unicode escape: \uN
      if (nextCh === 'u' && i + 1 < n && (rtf[i + 1]! >= '0' || rtf[i + 1] === '-')) {
        i++; // skip 'u'
        let neg = false;
        if (i < n && rtf[i] === '-') {
          neg = true;
          i++;
        }
        let val = 0;
        while (i < n && rtf[i]! >= '0' && rtf[i]! <= '9') {
          val = val * 10 + (rtf.charCodeAt(i) - 48);
          i++;
        }
        if (neg) val = -val;
        if (val < 0) val += 65536;
        parts.push(String.fromCodePoint(val));
        // Skip one subsequent character (replacement char).
        if (i < n && rtf[i] === ' ') {
          i++;
        } else if (i < n && rtf[i] !== '\\' && rtf[i] !== '{' && rtf[i] !== '}') {
          i++;
        }
        continue;
      }

      // Hex escape: \'XX
      // These are raw bytes in the document's codepage (often Windows-1252).
      // Without a codepage table, bytes > 0x7F are not valid single-byte UTF-8,
      // so we emit U+FFFD to match Go's CleanUTF8 behavior.
      if (nextCh === "'" && i + 2 < n) {
        i++; // skip apostrophe
        const hi = hexVal(rtf[i]!);
        const lo = hexVal(rtf[i + 1]!);
        if (hi >= 0 && lo >= 0) {
          const byte = (hi << 4) | lo;
          parts.push(byte > 0x7f ? '\uFFFD' : String.fromCharCode(byte));
        }
        i += 2;
        continue;
      }

      // Literal escapes
      if (nextCh === '\\' || nextCh === '{' || nextCh === '}') {
        parts.push(nextCh);
        i++;
        continue;
      }

      // Control word
      if ((nextCh >= 'a' && nextCh <= 'z') || (nextCh >= 'A' && nextCh <= 'Z')) {
        const [word, , end] = parseControlWord(rtf, i);
        i = end;

        switch (word) {
          case 'par':
          case 'pard':
            parts.push('\n');
            break;
          case 'line':
            parts.push('\n');
            break;
          case 'tab':
            parts.push('\t');
            break;
          case 'lquote':
          case 'rquote':
            parts.push("'");
            break;
          case 'ldblquote':
          case 'rdblquote':
            parts.push('"');
            break;
          case 'emdash':
            parts.push('\u2014');
            break;
          case 'endash':
            parts.push('\u2013');
            break;
          case 'bullet':
            parts.push('\u2022');
            break;
        }
        continue;
      }

      // Skip other control symbols
      i++;
      continue;
    }

    // Plain text
    if (skipDepth > 0) {
      i++;
      continue;
    }
    parts.push(ch);
    i++;
  }

  return cleanRTFText(parts.join(''));
}

function skipControlWord(rtf: string, i: number): number {
  i++; // skip backslash
  const n = rtf.length;
  if (i >= n) return i;
  // Skip word characters
  while (i < n && ((rtf[i]! >= 'a' && rtf[i]! <= 'z') || (rtf[i]! >= 'A' && rtf[i]! <= 'Z'))) {
    i++;
  }
  // Skip optional numeric parameter
  if (i < n && (rtf[i] === '-' || (rtf[i]! >= '0' && rtf[i]! <= '9'))) {
    if (rtf[i] === '-') i++;
    while (i < n && rtf[i]! >= '0' && rtf[i]! <= '9') i++;
  }
  // Skip optional trailing space
  if (i < n && rtf[i] === ' ') i++;
  return i;
}

function parseControlWord(rtf: string, i: number): [string, number, number] {
  const n = rtf.length;
  const start = i;
  while (i < n && ((rtf[i]! >= 'a' && rtf[i]! <= 'z') || (rtf[i]! >= 'A' && rtf[i]! <= 'Z'))) {
    i++;
  }
  const word = rtf.substring(start, i);

  let param = 0;
  if (i < n && (rtf[i] === '-' || (rtf[i]! >= '0' && rtf[i]! <= '9'))) {
    let neg = false;
    if (rtf[i] === '-') {
      neg = true;
      i++;
    }
    let val = 0;
    while (i < n && rtf[i]! >= '0' && rtf[i]! <= '9') {
      val = val * 10 + (rtf.charCodeAt(i) - 48);
      i++;
    }
    param = neg ? -val : val;
  }

  if (i < n && rtf[i] === ' ') i++;
  return [word, param, i];
}

function hexVal(ch: string): number {
  const c = ch.charCodeAt(0);
  if (c >= 48 && c <= 57) return c - 48;      // 0-9
  if (c >= 97 && c <= 102) return c - 87;      // a-f
  if (c >= 65 && c <= 70) return c - 55;        // A-F
  return -1;
}

/** Normalize whitespace: collapse multiple blank lines, trim. */
function cleanRTFText(s: string): string {
  const lines = s.split('\n');
  const result: string[] = [];
  let blankCount = 0;

  for (const line of lines) {
    const trimmed = line.replace(/[\t\r ]+$/, '');
    if (trimmed === '') {
      blankCount++;
      if (blankCount <= 1) {
        result.push('');
      }
    } else {
      blankCount = 0;
      result.push(trimmed);
    }
  }

  return result.join('\n').trim();
}
