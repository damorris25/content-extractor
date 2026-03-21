// XLSX extractor — extracts text and metadata from .xlsx files.
// Port of formats/office/xlsx.go

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

const MIME_XLSX =
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB
const SPREADSHEET_NS =
  'http://schemas.openxmlformats.org/spreadsheetml/2006/main';
const DC_NS = 'http://purl.org/dc/elements/1.1/';
const DCTERMS_NS = 'http://purl.org/dc/terms/';
const EP_NS =
  'http://schemas.openxmlformats.org/officeDocument/2006/extended-properties';

export class XlsxExtractor implements Extractor {
  readonly maxSize: number;

  constructor(maxSize = DEFAULT_MAX_SIZE) {
    this.maxSize = maxSize;
  }

  supportedMIMETypes(): string[] {
    return [MIME_XLSX];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_XLSX) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_XLSX);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: MIME_XLSX, language: '' };
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

    // Load shared strings table.
    const sharedStrings = await loadSharedStrings(entries);

    // Extract text from all worksheets.
    const text = await extractSheetText(entries, sharedStrings);

    // Extract metadata.
    let language = '';
    const coreLang = await extractCoreProperties(entries, metadata);
    if (coreLang) language = coreLang;
    await extractAppProperties(entries, metadata);

    return { text, metadata, contentType: MIME_XLSX, language };
  }
}

/** Parse xl/sharedStrings.xml into a string array indexed by position. */
async function loadSharedStrings(entries: ZipEntries): Promise<string[]> {
  const xml = readZipEntry(entries, 'xl/sharedStrings.xml');
  if (!xml) return [];

  const doc = await parseXML(xml);
  const siElements = doc.getElementsByTagNameNS(SPREADSHEET_NS, 'si');
  const strings: string[] = [];

  for (let i = 0; i < siElements.length; i++) {
    const si = siElements[i]!;
    // Collect text from all <t> descendants within this <si>.
    const tElements = si.getElementsByTagNameNS(SPREADSHEET_NS, 't');
    let text = '';
    for (let j = 0; j < tElements.length; j++) {
      text += tElements[j]!.textContent ?? '';
    }
    strings.push(text);
  }

  return strings;
}

/** Extract text from all worksheets, sorted by name. */
async function extractSheetText(entries: ZipEntries, sharedStrings: string[]): Promise<string> {
  // Find all sheet files, sorted by name for deterministic order.
  const sheetNames = Object.keys(entries)
    .filter(
      (name) =>
        name.startsWith('xl/worksheets/sheet') && name.endsWith('.xml'),
    )
    .sort();

  const sheetTexts: string[] = [];

  for (const name of sheetNames) {
    const xml = readZipEntry(entries, name);
    if (!xml) continue;
    sheetTexts.push(await extractSingleSheet(xml, sharedStrings));
  }

  return sheetTexts.join('\n\n');
}

/** Extract text from a single worksheet XML. */
async function extractSingleSheet(xml: string, sharedStrings: string[]): Promise<string> {
  const doc = await parseXML(xml);
  const rows = doc.getElementsByTagNameNS(SPREADSHEET_NS, 'row');
  const rowTexts: string[] = [];

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i]!;
    const cells = row.getElementsByTagNameNS(SPREADSHEET_NS, 'c');
    const cellTexts: string[] = [];

    for (let j = 0; j < cells.length; j++) {
      const cell = cells[j]!;
      const cellType = cell.getAttribute('t') ?? '';

      // Get value from <v> element or inline <t> element.
      let val = '';
      const vEl = cell.getElementsByTagNameNS(SPREADSHEET_NS, 'v');
      if (vEl.length > 0) {
        val = vEl[0]!.textContent ?? '';
      } else {
        // Check for inline string <is><t>...</t></is>
        const tEl = cell.getElementsByTagNameNS(SPREADSHEET_NS, 't');
        if (tEl.length > 0) {
          val = tEl[0]!.textContent ?? '';
        }
      }

      if (cellType === 's' && sharedStrings.length > 0) {
        // Shared string reference.
        const idx = parseInt(val, 10);
        if (!isNaN(idx) && idx >= 0 && idx < sharedStrings.length) {
          val = sharedStrings[idx]!;
        }
      }

      if (val !== '') {
        cellTexts.push(val);
      }
    }

    if (cellTexts.length > 0) {
      rowTexts.push(cellTexts.join('\t'));
    }
  }

  return rowTexts.join('\n');
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
