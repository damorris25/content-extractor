// PDF extractor — extracts text and metadata from .pdf files.
// Uses pdfjs-dist for parsing (dynamic import for optional dependency).

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import {
  FileTooLargeError,
  UnsupportedFormatError,
  CorruptInputError,
  checkAborted,
} from '../extractor.js';
import {
  createMetadata,
  metadataSet,
  KEY_FORMAT,
  KEY_PAGE_COUNT,
} from '../metadata.js';

const MIME_PDF = 'application/pdf';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB
const MAX_PAGES = 10000; // Safety cap on page iteration

export class PdfExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_PDF];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_PDF) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_PDF);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_PDF, language: '' };
    }

    checkAborted(options?.signal);

    // Dynamic import of pdfjs-dist — it is an optional peer dependency.
    // Use variable to prevent TypeScript from resolving the module at compile time.
    const pdfModuleName = 'pdfjs-dist';
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let pdfjsLib: any;
    try {
      pdfjsLib = await import(/* @vite-ignore */ pdfModuleName);
    } catch {
      throw new Error(
        'pdfjs-dist is required for PDF extraction. Install it: npm install pdfjs-dist',
      );
    }

    let doc;
    try {
      const loadingTask = pdfjsLib.getDocument({
        data: data.slice(),
        useSystemFonts: true,
      });
      doc = await loadingTask.promise;
    } catch (err) {
      throw new CorruptInputError(
        `failed to parse PDF: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    const numPages = Math.min(doc.numPages, MAX_PAGES);
    metadataSet(metadata, KEY_PAGE_COUNT, String(doc.numPages));

    const pageTexts: string[] = [];

    for (let i = 1; i <= numPages; i++) {
      checkAborted(options?.signal);

      try {
        const page = await doc.getPage(i);
        const content = await page.getTextContent();
        let pageText = '';
        let lastY: number | null = null;

        for (const item of content.items) {
          if (!('str' in item)) continue;
          const textItem = item as { str: string; transform: number[] };
          const y = textItem.transform[5];

          // Detect line breaks by Y position changes.
          if (lastY !== null && y !== undefined && Math.abs(y - lastY) > 1) {
            pageText += '\n';
          }
          pageText += textItem.str;
          if (y !== undefined) lastY = y;
        }

        pageTexts.push(pageText);
      } catch {
        // Skip malformed pages — best-effort extraction.
        pageTexts.push('');
      }
    }

    const text = pageTexts.join('\n\n');

    return { text, metadata, contentType: MIME_PDF, language: '' };
  }
}
