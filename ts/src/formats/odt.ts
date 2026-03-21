// ODT extractor — extracts text and metadata from .odt files.
// Port of formats/odf/odt.go

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

const MIME_ODT = 'application/vnd.oasis.opendocument.text';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB

const TEXT_NS = 'urn:oasis:names:tc:opendocument:xmlns:text:1.0';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const META_NS = 'urn:oasis:names:tc:opendocument:xmlns:meta:1.0';

export class OdtExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_ODT];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_ODT) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_ODT);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_ODT, language: '' };
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
    const text = await extractODTText(entries);

    // Extract metadata from meta.xml.
    const language = await extractODFMeta(entries, metadata);

    return { text, metadata, contentType: MIME_ODT, language };
  }
}

/**
 * Parse content.xml for text:p and text:h elements.
 * Handles text:tab -> \t and text:line-break -> \n.
 */
async function extractODTText(entries: ZipEntries): Promise<string> {
  const xml = readZipEntry(entries, 'content.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);

  // Collect all text:p and text:h elements.
  const paragraphs: Element[] = [];
  collectParagraphs(doc.documentElement, paragraphs);

  const parts: string[] = [];
  for (const p of paragraphs) {
    parts.push(extractODFParagraphText(p));
  }

  return parts.join('\n');
}

/** Recursively collect text:p and text:h elements (in document order). */
function collectParagraphs(el: Element, result: Element[]): void {
  if (!el) return;
  const children = el.childNodes;
  for (let i = 0; i < children.length; i++) {
    const child = children[i]!;
    if (child.nodeType === 1) {
      // Element node.
      const elem = child as Element;
      if (
        elem.namespaceURI === TEXT_NS &&
        (elem.localName === 'p' || elem.localName === 'h')
      ) {
        result.push(elem);
      } else {
        collectParagraphs(elem, result);
      }
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
      } else if (elem.namespaceURI === TEXT_NS && elem.localName === 'line-break') {
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
