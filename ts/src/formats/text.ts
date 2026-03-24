// CONTRACT:C1-EXTRACTOR.1.0
// Plain text extractor — port of formats/text/text.go

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT } from '../metadata.js';

/** Default maximum bytes to read from a text file (10 MB). */
const DEFAULT_MAX_BYTES = 10 * 1024 * 1024;

const MIME_TEXT_PLAIN = 'text/plain';

/** Extracts text content from plain text and text-like files. */
export class TextExtractor implements Extractor {
  private readonly maxBytes: number;
  private readonly mimeTypes: Set<string>;

  constructor(maxBytes?: number);
  constructor(mimeTypes: string[], maxBytes?: number);
  constructor(first?: number | string[], second?: number) {
    if (Array.isArray(first)) {
      this.mimeTypes = new Set(first);
      this.maxBytes = second ?? DEFAULT_MAX_BYTES;
    } else {
      this.mimeTypes = new Set([MIME_TEXT_PLAIN]);
      this.maxBytes = first ?? DEFAULT_MAX_BYTES;
    }
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (!this.mimeTypes.has(mime)) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    if (data.length === 0) {
      const metadata = createMetadata();
      metadataSet(metadata, KEY_FORMAT, mime);
      return { text: '', metadata, contentType: mime, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const decoder = new TextDecoder('utf-8');
    const text = decoder.decode(data);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, mime);

    return { text, metadata, contentType: mime, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [...this.mimeTypes];
  }
}
