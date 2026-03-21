// DOCX extractor — extracts text and metadata from .docx files.
// Port of formats/office/docx.go

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
import { safeUnzip, readZipEntry } from './zip-utils.js';
import { parseXML, getFirstTextNS } from './xml-utils.js';

const MIME_DOCX =
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB
const WML_NS = 'http://schemas.openxmlformats.org/wordprocessingml/2006/main';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const DCTERMS_NS = 'http://purl.org/dc/terms/';
const EP_NS =
  'http://schemas.openxmlformats.org/officeDocument/2006/extended-properties';

export class DocxExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_DOCX];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_DOCX) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_DOCX);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_DOCX, language: '' };
    }

    checkAborted(options?.signal);

    let entries;
    try {
      entries = safeUnzip(data, this.maxSize);
    } catch (err) {
      if (err instanceof FileTooLargeError) throw err;
      throw new CorruptInputError(
        `invalid ZIP: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    // Extract text from word/document.xml.
    const text = await extractDocumentText(entries);

    // Extract metadata from docProps/core.xml and docProps/app.xml.
    let language = '';
    const coreLang = await extractCoreProperties(entries, metadata);
    if (coreLang) language = coreLang;
    await extractAppProperties(entries, metadata);

    return { text, metadata, contentType: MIME_DOCX, language };
  }
}

/** Parse word/document.xml and extract text from w:p / w:t elements. */
async function extractDocumentText(entries: Record<string, Uint8Array>): Promise<string> {
  const xml = readZipEntry(entries, 'word/document.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);
  const paragraphs = doc.getElementsByTagNameNS(WML_NS, 'p');
  const parts: string[] = [];

  for (let i = 0; i < paragraphs.length; i++) {
    const p = paragraphs[i]!;
    const runs = p.getElementsByTagNameNS(WML_NS, 't');
    let pText = '';
    for (let j = 0; j < runs.length; j++) {
      pText += runs[j]!.textContent ?? '';
    }
    parts.push(pText);
  }

  return parts.join('\n');
}

/**
 * Extract Dublin Core metadata from docProps/core.xml.
 * Returns language string if found.
 */
async function extractCoreProperties(
  entries: Record<string, Uint8Array>,
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
  entries: Record<string, Uint8Array>,
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
