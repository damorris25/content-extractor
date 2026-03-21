// PPTX extractor — extracts text and metadata from .pptx files.
// Port of formats/office/pptx.go

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
  normalizeDate,
  KEY_FORMAT,
  KEY_TITLE,
  KEY_CREATOR,
  KEY_SUBJECT,
  KEY_DESCRIPTION,
  KEY_DATE,
  KEY_MODIFIED,
  KEY_LANGUAGE,
  KEY_APPLICATION,
  KEY_PAGE_COUNT,
  KEY_WORD_COUNT,
  KEY_CHAR_COUNT,
} from '../metadata.js';
import type { ZipEntries } from './zip-utils.js';
import { safeUnzip, readZipEntry } from './zip-utils.js';
import { parseXML, getFirstTextNS } from './xml-utils.js';

const MIME_PPTX =
  'application/vnd.openxmlformats-officedocument.presentationml.presentation';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB
const DRAWING_ML_NS = 'http://schemas.openxmlformats.org/drawingml/2006/main';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const DCTERMS_NS = 'http://purl.org/dc/terms/';
const EP_NS =
  'http://schemas.openxmlformats.org/officeDocument/2006/extended-properties';

export class PptxExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_PPTX];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_PPTX) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_PPTX);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_PPTX, language: '' };
    }

    checkAborted(options?.signal);

    let entries: ZipEntries;
    try {
      entries = safeUnzip(data, this.maxSize);
    } catch (err) {
      if (err instanceof FileTooLargeError) throw err;
      throw new CorruptInputError(
        `invalid ZIP: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    // Extract text from all slides.
    const text = await extractSlideText(entries);

    // Extract metadata.
    let language = '';
    const coreLang = await extractCoreProperties(entries, metadata);
    if (coreLang) language = coreLang;
    await extractAppProperties(entries, metadata);

    return { text, metadata, contentType: MIME_PPTX, language };
  }
}

/** Extract text from all slide XML files, sorted by name. */
async function extractSlideText(entries: ZipEntries): Promise<string> {
  const slideNames = Object.keys(entries)
    .filter(
      (name) =>
        name.startsWith('ppt/slides/slide') && name.endsWith('.xml'),
    )
    .sort();

  const slideTexts: string[] = [];

  for (const name of slideNames) {
    const xml = readZipEntry(entries, name);
    if (!xml) continue;
    slideTexts.push(await extractSingleSlide(xml));
  }

  return slideTexts.join('\n\n');
}

/** Extract text from a single slide XML. Text is in a:p / a:t elements. */
async function extractSingleSlide(xml: string): Promise<string> {
  const doc = await parseXML(xml);
  const paragraphs = doc.getElementsByTagNameNS(DRAWING_ML_NS, 'p');
  const parts: string[] = [];

  for (let i = 0; i < paragraphs.length; i++) {
    const p = paragraphs[i]!;
    const runs = p.getElementsByTagNameNS(DRAWING_ML_NS, 't');
    let pText = '';
    for (let j = 0; j < runs.length; j++) {
      pText += runs[j]!.textContent ?? '';
    }
    parts.push(pText);
  }

  return parts.join('\n');
}

/** Extract Dublin Core metadata from docProps/core.xml. */
async function extractCoreProperties(
  entries: ZipEntries,
  metadata: Record<string, string[]>,
): Promise<string> {
  const xml = readZipEntry(entries, 'docProps/core.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);

  metadataSet(metadata, KEY_TITLE, getFirstTextNS(doc, DC_NS, 'title'));
  metadataSet(metadata, KEY_CREATOR, getFirstTextNS(doc, DC_NS, 'creator'));
  metadataSet(metadata, KEY_SUBJECT, getFirstTextNS(doc, DC_NS, 'subject'));
  metadataSet(metadata, KEY_DESCRIPTION, getFirstTextNS(doc, DC_NS, 'description'));

  const created = getFirstTextNS(doc, DCTERMS_NS, 'created');
  if (created) metadataSet(metadata, KEY_DATE, normalizeDate(created));

  const modified = getFirstTextNS(doc, DCTERMS_NS, 'modified');
  if (modified) metadataSet(metadata, KEY_MODIFIED, normalizeDate(modified));

  const language = getFirstTextNS(doc, DC_NS, 'language');
  if (language) metadataSet(metadata, KEY_LANGUAGE, language);

  return language;
}

/** Extract app properties from docProps/app.xml. */
async function extractAppProperties(
  entries: ZipEntries,
  metadata: Record<string, string[]>,
): Promise<void> {
  const xml = readZipEntry(entries, 'docProps/app.xml');
  if (!xml) return;

  const doc = await parseXML(xml);

  metadataSet(metadata, KEY_APPLICATION, getFirstTextNS(doc, EP_NS, 'Application'));
  metadataSet(metadata, KEY_PAGE_COUNT, getFirstTextNS(doc, EP_NS, 'Pages'));
  metadataSet(metadata, KEY_WORD_COUNT, getFirstTextNS(doc, EP_NS, 'Words'));
  metadataSet(metadata, KEY_CHAR_COUNT, getFirstTextNS(doc, EP_NS, 'Characters'));
}
