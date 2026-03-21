// ODP extractor — extracts text and metadata from .odp files.
// Port of formats/odf/odp.go

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
} from '../metadata.js';
import type { ZipEntries } from './zip-utils.js';
import { safeUnzip, readZipEntry } from './zip-utils.js';
import { parseXML, getFirstTextNS } from './xml-utils.js';

const MIME_ODP = 'application/vnd.oasis.opendocument.presentation';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB

const DRAW_NS = 'urn:oasis:names:tc:opendocument:xmlns:drawing:1.0';
const PRESENTATION_NS =
  'urn:oasis:names:tc:opendocument:xmlns:presentation:1.0';
const TEXT_NS = 'urn:oasis:names:tc:opendocument:xmlns:text:1.0';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const META_NS = 'urn:oasis:names:tc:opendocument:xmlns:meta:1.0';

export class OdpExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_ODP];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_ODP) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_ODP);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_ODP, language: '' };
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

    // Extract text from content.xml.
    const text = await extractODPText(entries);

    // Extract metadata from meta.xml.
    const language = await extractODFMeta(entries, metadata);

    return { text, metadata, contentType: MIME_ODP, language };
  }
}

/**
 * Parse content.xml for presentation page text.
 * Walk: draw:page > text elements (text:p, text:h).
 * Skip presentation:notes elements.
 * Pages separated by double newlines, paragraphs by newlines.
 */
async function extractODPText(entries: ZipEntries): Promise<string> {
  const xml = readZipEntry(entries, 'content.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);

  const pages = doc.getElementsByTagNameNS(DRAW_NS, 'page');
  const pageTexts: string[] = [];

  for (let i = 0; i < pages.length; i++) {
    const page = pages[i]!;
    const paragraphs: string[] = [];
    extractPageParagraphs(page, paragraphs);
    if (paragraphs.length > 0) {
      pageTexts.push(paragraphs.join('\n'));
    }
  }

  return pageTexts.join('\n\n');
}

/**
 * Recursively collect paragraph text from a page element,
 * skipping presentation:notes elements.
 */
function extractPageParagraphs(el: Element, result: string[]): void {
  const children = el.childNodes;
  for (let i = 0; i < children.length; i++) {
    const child = children[i]!;
    if (child.nodeType !== 1) continue;
    const elem = child as Element;

    // Skip presenter notes.
    if (
      elem.namespaceURI === PRESENTATION_NS &&
      elem.localName === 'notes'
    ) {
      continue;
    }

    if (
      elem.namespaceURI === TEXT_NS &&
      (elem.localName === 'p' || elem.localName === 'h')
    ) {
      result.push(extractODFParagraphText(elem));
    } else {
      extractPageParagraphs(elem, result);
    }
  }
}

/**
 * Extract text content from an ODF paragraph element.
 * Handles text:tab and text:line-break special elements.
 */
function extractODFParagraphText(el: Element): string {
  let text = '';
  const children = el.childNodes;
  for (let i = 0; i < children.length; i++) {
    const child = children[i]!;
    if (child.nodeType === 3) {
      // Text node.
      text += child.nodeValue ?? '';
    } else if (child.nodeType === 1) {
      const elem = child as Element;
      if (elem.namespaceURI === TEXT_NS && elem.localName === 'tab') {
        text += '\t';
      } else if (
        elem.namespaceURI === TEXT_NS &&
        elem.localName === 'line-break'
      ) {
        text += '\n';
      } else {
        // Recurse into child elements (e.g. text:span).
        text += extractODFParagraphText(elem);
      }
    }
  }
  return text;
}

/**
 * Extract ODF metadata from meta.xml.
 * Returns language string if found.
 */
async function extractODFMeta(
  entries: ZipEntries,
  metadata: Record<string, string[]>,
): Promise<string> {
  const xml = readZipEntry(entries, 'meta.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);

  metadataSet(metadata, KEY_TITLE, getFirstTextNS(doc, DC_NS, 'title'));
  metadataSet(metadata, KEY_CREATOR, getFirstTextNS(doc, DC_NS, 'creator'));
  metadataSet(metadata, KEY_SUBJECT, getFirstTextNS(doc, DC_NS, 'subject'));
  metadataSet(metadata, KEY_DESCRIPTION, getFirstTextNS(doc, DC_NS, 'description'));

  const dcDate = getFirstTextNS(doc, DC_NS, 'date');
  if (dcDate) metadataSet(metadata, KEY_MODIFIED, normalizeDate(dcDate));

  const language = getFirstTextNS(doc, DC_NS, 'language');
  if (language) metadataSet(metadata, KEY_LANGUAGE, language);

  const creationDate = getFirstTextNS(doc, META_NS, 'creation-date');
  if (creationDate) metadataSet(metadata, KEY_DATE, normalizeDate(creationDate));

  const generator = getFirstTextNS(doc, META_NS, 'generator');
  if (generator) metadataSet(metadata, KEY_APPLICATION, generator);

  return language;
}
