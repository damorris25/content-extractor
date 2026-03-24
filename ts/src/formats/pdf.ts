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

// Track whether we've already configured the worker.
let workerConfigured = false;

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
    // The string literal (not a variable) is required for bundlers like Vite
    // to resolve the module. The @ts-ignore suppresses the missing-types error
    // since pdfjs-dist is an optional peer dep (not in devDependencies).
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let pdfjsLib: any;
    try {
      // @ts-ignore — optional peer dependency
      pdfjsLib = await import('pdfjs-dist');
    } catch {
      throw new Error(
        'pdfjs-dist is required for PDF extraction. Install it: npm install pdfjs-dist',
      );
    }

    // Configure the PDF.js worker (once per runtime).
    // In browser environments, pdfjs-dist needs a worker to avoid blocking the
    // main thread. The caller can set GlobalWorkerOptions.workerSrc before
    // calling extract (recommended). If not set, we skip — pdfjs-dist will
    // fall back to running on the main thread.
    if (!workerConfigured && pdfjsLib.GlobalWorkerOptions) {
      workerConfigured = true;
    }

    let doc;
    try {
      const loadingTask = pdfjsLib.getDocument({
        data: data.slice(),
        useSystemFonts: true,
        isEvalSupported: false,
        useWorkerFetch: false,
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
