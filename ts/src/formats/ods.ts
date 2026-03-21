// ODS extractor — extracts text and metadata from .ods files.
// Port of formats/odf/ods.go

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

const MIME_ODS = 'application/vnd.oasis.opendocument.spreadsheet';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB

const TABLE_NS = 'urn:oasis:names:tc:opendocument:xmlns:table:1.0';
const TEXT_NS = 'urn:oasis:names:tc:opendocument:xmlns:text:1.0';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const META_NS = 'urn:oasis:names:tc:opendocument:xmlns:meta:1.0';

export class OdsExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_ODS];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_ODS) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_ODS);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_ODS, language: '' };
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
    const text = await extractODSText(entries);

    // Extract metadata from meta.xml.
    const language = await extractODFMeta(entries, metadata);

    return { text, metadata, contentType: MIME_ODS, language };
  }
}

/**
 * Parse content.xml for table data.
 * Structure: table:table > table:table-row > table:table-cell > text:p
 * Cells separated by tabs, rows by newlines, tables by double newlines.
 */
async function extractODSText(entries: ZipEntries): Promise<string> {
  const xml = readZipEntry(entries, 'content.xml');
  if (!xml) return '';

  const doc = await parseXML(xml);

  const tables = doc.getElementsByTagNameNS(TABLE_NS, 'table');
  const tableTexts: string[] = [];

  for (let t = 0; t < tables.length; t++) {
    const table = tables[t]!;
    const rows = table.getElementsByTagNameNS(TABLE_NS, 'table-row');
    const rowTexts: string[] = [];

    for (let r = 0; r < rows.length; r++) {
      const row = rows[r]!;
      // Only get direct table-cell children (not nested table cells).
      const cells = getDirectChildrenNS(row, TABLE_NS, 'table-cell');
      const cellTexts: string[] = [];

      for (const cell of cells) {
        const pElements = cell.getElementsByTagNameNS(TEXT_NS, 'p');
        let cellText = '';
        for (let p = 0; p < pElements.length; p++) {
          if (p > 0) cellText += ' ';
          cellText += pElements[p]!.textContent ?? '';
        }
        if (cellText !== '') {
          cellTexts.push(cellText);
        }
      }

      if (cellTexts.length > 0) {
        rowTexts.push(cellTexts.join('\t'));
      }
    }

    if (rowTexts.length > 0) {
      tableTexts.push(rowTexts.join('\n'));
    }
  }

  return tableTexts.join('\n\n');
}

/** Get direct child elements matching a namespace and local name. */
function getDirectChildrenNS(
  parent: Element,
  ns: string,
  localName: string,
): Element[] {
  const result: Element[] = [];
  const children = parent.childNodes;
  for (let i = 0; i < children.length; i++) {
    const child = children[i]!;
    if (
      child.nodeType === 1 &&
      (child as Element).namespaceURI === ns &&
      (child as Element).localName === localName
    ) {
      result.push(child as Element);
    }
  }
  return result;
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
