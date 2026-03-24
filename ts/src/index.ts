// Public API for content-extractor TypeScript library.
// Port of contentextractor/contentextractor.go

export type { Extractor, ExtractionResult, ExtractOptions } from './extractor.js';
export {
  ExtractionError,
  UnsupportedFormatError,
  CorruptInputError,
  FileTooLargeError,
  DetectionFailedError,
  PasswordProtectedError,
  checkAborted,
} from './extractor.js';

export type { Metadata } from './metadata.js';
export {
  metadataGet, metadataGetAll, metadataSet, metadataAdd, metadataHas,
  createMetadata, stripNUL, normalizeDate,
  NS_DUBLIN_CORE, NS_OFFICE, NS_EXIF, NS_AUDIO, NS_CAD, NS_EXTRACTOR, NS_OCR, NS_GEO, NS_EMAIL,
  KEY_TITLE, KEY_CREATOR, KEY_SUBJECT, KEY_DESCRIPTION, KEY_DATE,
  KEY_MODIFIED, KEY_LANGUAGE, KEY_FORMAT, KEY_PAGE_COUNT, KEY_WORD_COUNT,
  KEY_CHAR_COUNT, KEY_APPLICATION, KEY_REVISION, KEY_EMAIL_TO, KEY_EMAIL_CC,
} from './metadata.js';

export type { Detector } from './detect.js';
export { DefaultDetector } from './detect.js';
export { Registry } from './registry.js';
export { ImageExtractor } from './formats/image.js';
export type { ImageExtractorOptions } from './formats/image.js';

import type { ExtractionResult, ExtractOptions } from './extractor.js';
import { Registry } from './registry.js';
import { DefaultDetector } from './detect.js';
import { TextExtractor } from './formats/text.js';
import { SvgExtractor } from './formats/svg.js';
import { RtfExtractor } from './formats/rtf.js';
import { DXFExtractor as DxfExtractor } from './formats/dxf.js';
import { ImageExtractor } from './formats/image.js';
import { AudioExtractor } from './formats/audio.js';
import { DocxExtractor } from './formats/docx.js';
import { XlsxExtractor } from './formats/xlsx.js';
import { PptxExtractor } from './formats/pptx.js';
import { OdtExtractor } from './formats/odt.js';
import { OdsExtractor } from './formats/ods.js';
import { OdpExtractor } from './formats/odp.js';
import { PdfExtractor } from './formats/pdf.js';
import { XMLExtractor } from './formats/xml.js';
import { HTMLExtractor } from './formats/html.js';
import { EMLExtractor } from './formats/email.js';

/** Options for createRegistry(). */
export interface CreateRegistryOptions {
  /** Enable OCR for image extraction. Requires tesseract.js. */
  enableOCR?: boolean;
  /** Tesseract language code for OCR (default "eng"). */
  ocrLanguage?: string;
}

/**
 * Create a registry pre-loaded with all built-in extractors and format detection.
 */
export function createRegistry(options?: CreateRegistryOptions): Registry {
  const r = new Registry();
  r.setDetector(new DefaultDetector());
  r.register(new TextExtractor());
  r.register(new TextExtractor([
    'application/json',
    'text/csv',
    'text/tab-separated-values',
    'application/x-yaml',
    'text/markdown',
  ]));
  r.register(new XMLExtractor());
  r.register(new HTMLExtractor());
  r.register(new SvgExtractor());
  r.register(new RtfExtractor());
  r.register(new DxfExtractor());
  r.register(new ImageExtractor({
    enableOCR: options?.enableOCR ?? false,
    ocrLanguage: options?.ocrLanguage ?? 'eng',
  }));
  r.register(new AudioExtractor());
  r.register(new DocxExtractor());
  r.register(new XlsxExtractor());
  r.register(new PptxExtractor());
  r.register(new OdtExtractor());
  r.register(new OdsExtractor());
  r.register(new OdpExtractor());
  r.register(new PdfExtractor());
  r.register(new EMLExtractor());
  return r;
}

/**
 * Convenience function: detects format and extracts content.
 * Note: creates a new registry on each call. For repeated use, call
 * createRegistry() once and reuse it.
 */
export async function extract(
  data: Uint8Array,
  mime?: string,
  options?: ExtractOptions,
): Promise<ExtractionResult> {
  return createRegistry().extract(data, mime, options);
}
