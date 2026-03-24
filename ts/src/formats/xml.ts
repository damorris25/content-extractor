// CONTRACT:C1-EXTRACTOR.1.0
// XML extractor — strips XML tags and returns text content.

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT } from '../metadata.js';

const MIME_APPLICATION_XML = 'application/xml';
const MIME_TEXT_XML = 'text/xml';
const DEFAULT_MAX_BYTES = 10 * 1024 * 1024; // 10MB

/** Extracts text content from XML files by stripping tags. */
export class XMLExtractor implements Extractor {
  private readonly maxBytes: number;

  constructor(maxBytes: number = DEFAULT_MAX_BYTES) {
    this.maxBytes = maxBytes;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_APPLICATION_XML && mime !== MIME_TEXT_XML) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, mime);

    if (data.length === 0) {
      return { text: '', metadata, contentType: mime, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const decoder = new TextDecoder('utf-8');
    const content = decoder.decode(data);
    const text = extractXMLText(content);

    return { text, metadata, contentType: mime, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [MIME_APPLICATION_XML, MIME_TEXT_XML];
  }
}

/**
 * Extract text content from XML by parsing CharData between tags.
 * Uses a simple state machine to avoid DOM parser dependencies.
 */
function extractXMLText(xml: string): string {
  const parts: string[] = [];
  let i = 0;
  const len = xml.length;

  while (i < len) {
    // Skip processing instructions <?...?>
    if (xml[i] === '<' && i + 1 < len && xml[i + 1] === '?') {
      const end = xml.indexOf('?>', i + 2);
      if (end < 0) break;
      i = end + 2;
      continue;
    }

    // Skip comments <!--...-->
    if (xml[i] === '<' && xml.startsWith('!--', i + 1)) {
      const end = xml.indexOf('-->', i + 4);
      if (end < 0) break;
      i = end + 3;
      continue;
    }

    // Skip DOCTYPE declarations (XXE defense-in-depth).
    if (xml[i] === '<' && xml.startsWith('!DOCTYPE', i + 1)) {
      const end = xml.indexOf('>', i + 9);
      if (end < 0) break;
      i = end + 1;
      continue;
    }

    // Skip CDATA — extract content
    if (xml[i] === '<' && xml.startsWith('![CDATA[', i + 1)) {
      const end = xml.indexOf(']]>', i + 9);
      if (end < 0) break;
      const cdata = xml.substring(i + 9, end).trim();
      if (cdata) parts.push(cdata);
      i = end + 3;
      continue;
    }

    // Skip tags
    if (xml[i] === '<') {
      const end = xml.indexOf('>', i + 1);
      if (end < 0) break;
      i = end + 1;
      continue;
    }

    // Collect text content
    const tagStart = xml.indexOf('<', i);
    const textEnd = tagStart < 0 ? len : tagStart;
    const chunk = xml.substring(i, textEnd).trim();
    if (chunk) {
      // Decode basic XML entities
      parts.push(decodeXMLEntities(chunk));
    }
    i = textEnd;
  }

  return parts.join('\n');
}

/** Decode standard XML entities. */
function decodeXMLEntities(s: string): string {
  return s
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'");
}
