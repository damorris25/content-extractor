// CONTRACT:C1-EXTRACTOR.1.0
// HTML extractor — extracts visible text content from HTML files.

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT, KEY_TITLE } from '../metadata.js';

const MIME_TEXT_HTML = 'text/html';
const DEFAULT_MAX_BYTES = 10 * 1024 * 1024; // 10MB

/** Elements whose content should be skipped entirely. */
const SKIP_ELEMENTS = new Set(['script', 'style', 'noscript']);

/** Extracts visible text content from HTML files. */
export class HTMLExtractor implements Extractor {
  private readonly maxBytes: number;

  constructor(maxBytes: number = DEFAULT_MAX_BYTES) {
    this.maxBytes = maxBytes;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_TEXT_HTML) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_TEXT_HTML);

    if (data.length === 0) {
      return { text: '', metadata, contentType: MIME_TEXT_HTML, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const decoder = new TextDecoder('utf-8');
    const content = decoder.decode(data);
    const { text, title } = extractHTMLText(content);

    if (title) {
      metadataSet(metadata, KEY_TITLE, title);
    }

    return { text, metadata, contentType: MIME_TEXT_HTML, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [MIME_TEXT_HTML];
  }
}

/**
 * Extract visible text from HTML using a simple tag-aware state machine.
 * Skips script, style, and noscript content. Extracts title from <title> tag.
 */
function extractHTMLText(html: string): { text: string; title: string } {
  const parts: string[] = [];
  let title = '';
  let i = 0;
  const len = html.length;

  // Stack of elements we're currently inside that should be skipped.
  let skipDepth = 0;
  let inTitle = false;
  let titleParts: string[] = [];

  while (i < len) {
    // Skip comments <!--...-->
    if (html[i] === '<' && html.startsWith('!--', i + 1)) {
      const end = html.indexOf('-->', i + 4);
      if (end < 0) break;
      i = end + 3;
      continue;
    }

    // Skip DOCTYPE
    if (html[i] === '<' && html.startsWith('!', i + 1)) {
      const end = html.indexOf('>', i + 1);
      if (end < 0) break;
      i = end + 1;
      continue;
    }

    // Handle tags
    if (html[i] === '<') {
      const end = html.indexOf('>', i + 1);
      if (end < 0) break;
      const tag = html.substring(i + 1, end).trim();
      i = end + 1;

      // Parse tag name
      const isClosing = tag.startsWith('/');
      const nameStr = isClosing ? tag.substring(1) : tag;
      // Extract tag name (before space or / for self-closing)
      const spaceIdx = nameStr.search(/[\s/]/);
      const tagName = (spaceIdx < 0 ? nameStr : nameStr.substring(0, spaceIdx)).toLowerCase();
      const isSelfClosing = tag.endsWith('/');

      if (!isSelfClosing) {
        if (isClosing) {
          if (SKIP_ELEMENTS.has(tagName)) {
            skipDepth = Math.max(0, skipDepth - 1);
          }
          if (tagName === 'title') {
            inTitle = false;
            title = titleParts.join(' ');
          }
        } else {
          if (SKIP_ELEMENTS.has(tagName)) {
            skipDepth++;
          }
          if (tagName === 'title') {
            inTitle = true;
            titleParts = [];
          }
        }
      }
      continue;
    }

    // Collect text content
    const tagStart = html.indexOf('<', i);
    const textEnd = tagStart < 0 ? len : tagStart;
    const chunk = html.substring(i, textEnd).trim();
    if (chunk) {
      const decoded = decodeHTMLEntities(chunk);
      if (inTitle) {
        titleParts.push(decoded);
      }
      if (skipDepth === 0) {
        parts.push(decoded);
      }
    }
    i = textEnd;
  }

  return { text: parts.join('\n'), title };
}

/** Decode common HTML entities. */
function decodeHTMLEntities(s: string): string {
  return s
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, ' ');
}
