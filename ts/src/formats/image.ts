// Image metadata extractor.
// Port of formats/image/image.go

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import {
  createMetadata, metadataSet, metadataGet,
  KEY_FORMAT, KEY_TITLE, KEY_CREATOR, KEY_DESCRIPTION, KEY_APPLICATION, KEY_DATE,
  normalizeDate,
} from '../metadata.js';

const DEFAULT_MAX_IMAGE_SIZE = 100 * 1024 * 1024; // 100MB

const SUPPORTED_MIMES = new Set([
  'image/jpeg',
  'image/png',
  'image/tiff',
  'image/bmp',
  'image/gif',
  'image/webp',
]);

/** EXIF tag IDs. */
const TAG_MAKE = 0x010F;
const TAG_MODEL = 0x0110;
const TAG_ORIENTATION = 0x0112;
const TAG_IMAGE_DESCRIPTION = 0x010E;
const TAG_SOFTWARE = 0x0131;
const TAG_DATE_TIME = 0x0132;
const TAG_ARTIST = 0x013B;
const TAG_COPYRIGHT = 0x8298;
const TAG_EXPOSURE_TIME = 0x829A;
const TAG_FNUMBER = 0x829D;
const TAG_DATE_TIME_ORIGINAL = 0x9003;
const TAG_ISO_SPEED = 0x8827;
const TAG_FOCAL_LENGTH = 0x920A;
const TAG_GPS_IFD = 0x8825;

const TAG_GPS_LAT_REF = 0x0001;
const TAG_GPS_LAT = 0x0002;
const TAG_GPS_LON_REF = 0x0003;
const TAG_GPS_LON = 0x0004;

export class ImageExtractor implements Extractor {
  maxSize: number;

  constructor(maxSize?: number) {
    this.maxSize = maxSize ?? DEFAULT_MAX_IMAGE_SIZE;
  }

  supportedMIMETypes(): string[] {
    return [...SUPPORTED_MIMES];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (!SUPPORTED_MIMES.has(mime)) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, mime);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: mime, language: '' };
    }

    checkAborted(options?.signal);

    // Extract dimensions based on format.
    extractDimensions(data, mime, metadata);

    // Extract EXIF metadata (JPEG and TIFF only).
    if (mime === 'image/jpeg' || mime === 'image/tiff') {
      extractEXIF(data, mime, metadata);
    }

    // Extract PNG text chunks.
    if (mime === 'image/png') {
      extractPNGText(data, metadata);
    }

    const text = buildTextSummary(metadata);

    return { text, metadata, contentType: mime, language: '' };
  }
}

// ---------------------------------------------------------------------------
// Dimension extraction
// ---------------------------------------------------------------------------

function extractDimensions(
  data: Uint8Array,
  mime: string,
  metadata: Record<string, string[]>,
): void {
  let width = 0;
  let height = 0;

  switch (mime) {
    case 'image/jpeg':
      [width, height] = jpegDimensions(data);
      break;
    case 'image/png':
      [width, height] = pngDimensions(data);
      break;
    case 'image/gif':
      [width, height] = gifDimensions(data);
      break;
    case 'image/bmp':
      [width, height] = bmpDimensions(data);
      break;
    default:
      return;
  }

  if (width > 0) metadataSet(metadata, 'exif:width', String(width));
  if (height > 0) metadataSet(metadata, 'exif:height', String(height));
}

function jpegDimensions(data: Uint8Array): [number, number] {
  // Walk JPEG markers looking for SOF0 (0xFFC0) or SOF2 (0xFFC2).
  if (data.byteLength < 2 || data[0] !== 0xFF || data[1] !== 0xD8) {
    return [0, 0];
  }
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let pos = 2;
  while (pos + 4 < data.byteLength) {
    if (data[pos] !== 0xFF) {
      pos++;
      continue;
    }
    const marker = data[pos + 1]!;
    if (marker === 0xC0 || marker === 0xC2) {
      // SOF marker: length(2) + precision(1) + height(2) + width(2)
      if (pos + 9 <= data.byteLength) {
        const height = view.getUint16(pos + 5, false);
        const width = view.getUint16(pos + 7, false);
        return [width, height];
      }
      return [0, 0];
    }
    // Skip this marker segment.
    if (pos + 3 >= data.byteLength) break;
    const segLen = view.getUint16(pos + 2, false);
    pos += 2 + segLen;
  }
  return [0, 0];
}

function pngDimensions(data: Uint8Array): [number, number] {
  // IHDR chunk starts at offset 8 (after 8-byte signature).
  // Chunk: length(4) + type(4) + data... First 8 bytes of IHDR data = width(4) + height(4).
  if (data.byteLength < 24) return [0, 0];
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  const width = view.getUint32(16, false);
  const height = view.getUint32(20, false);
  return [width, height];
}

function gifDimensions(data: Uint8Array): [number, number] {
  if (data.byteLength < 10) return [0, 0];
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  const width = view.getUint16(6, true);
  const height = view.getUint16(8, true);
  return [width, height];
}

function bmpDimensions(data: Uint8Array): [number, number] {
  if (data.byteLength < 26) return [0, 0];
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  const width = view.getInt32(18, true);
  const height = Math.abs(view.getInt32(22, true));
  return [width, height];
}

// ---------------------------------------------------------------------------
// EXIF parsing
// ---------------------------------------------------------------------------

function extractEXIF(
  data: Uint8Array,
  mime: string,
  metadata: Record<string, string[]>,
): void {
  let exifData: Uint8Array;

  if (mime === 'image/jpeg') {
    const offset = findAPP1(data);
    if (offset < 0) return;
    exifData = data.subarray(offset);
  } else {
    // TIFF: the file itself is a TIFF header.
    exifData = data;
  }

  parseExifBlock(exifData, metadata);
}

/** Find APP1 marker in JPEG and return offset to the Exif TIFF header. */
function findAPP1(data: Uint8Array): number {
  if (data.byteLength < 2 || data[0] !== 0xFF || data[1] !== 0xD8) return -1;
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let pos = 2;
  while (pos + 4 < data.byteLength) {
    if (data[pos] !== 0xFF) { pos++; continue; }
    const marker = data[pos + 1]!;
    if (marker === 0xE1) {
      // APP1 marker. Check for "Exif\0\0" header.
      const segLen = view.getUint16(pos + 2, false);
      if (pos + 4 + 6 <= data.byteLength &&
          data[pos + 4] === 0x45 && // E
          data[pos + 5] === 0x78 && // x
          data[pos + 6] === 0x69 && // i
          data[pos + 7] === 0x66 && // f
          data[pos + 8] === 0x00 &&
          data[pos + 9] === 0x00) {
        // Return offset to TIFF header (after "Exif\0\0").
        return pos + 10;
      }
      pos += 2 + segLen;
      continue;
    }
    // Skip marker segment.
    if (marker === 0xDA) break; // Start of scan — stop.
    if (pos + 3 >= data.byteLength) break;
    const segLen = view.getUint16(pos + 2, false);
    pos += 2 + segLen;
  }
  return -1;
}

/** Parse a TIFF/EXIF block starting with the TIFF header (II or MM). */
function parseExifBlock(
  tiffData: Uint8Array,
  metadata: Record<string, string[]>,
): void {
  if (tiffData.byteLength < 8) return;

  // Determine byte order.
  const bo0 = tiffData[0]!;
  const bo1 = tiffData[1]!;
  let littleEndian: boolean;
  if (bo0 === 0x49 && bo1 === 0x49) {
    littleEndian = true;
  } else if (bo0 === 0x4D && bo1 === 0x4D) {
    littleEndian = false;
  } else {
    return;
  }

  const view = new DataView(tiffData.buffer, tiffData.byteOffset, tiffData.byteLength);

  // Verify TIFF magic (42).
  const magic = view.getUint16(2, littleEndian);
  if (magic !== 42) return;

  // Read IFD0 offset.
  const ifd0Offset = view.getUint32(4, littleEndian);
  if (ifd0Offset >= tiffData.byteLength) return;

  // Parse IFD0.
  let gpsIfdOffset = -1;
  const entries = readIFDEntries(tiffData, view, ifd0Offset, littleEndian);
  for (const entry of entries) {
    switch (entry.tag) {
      case TAG_MAKE:
        metadataSet(metadata, 'exif:make', readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_MODEL:
        metadataSet(metadata, 'exif:model', readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_ORIENTATION:
        metadataSet(metadata, 'exif:orientation', String(readShortOrLong(view, entry, littleEndian)));
        break;
      case TAG_IMAGE_DESCRIPTION:
        metadataSet(metadata, KEY_DESCRIPTION, readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_SOFTWARE:
        metadataSet(metadata, KEY_APPLICATION, readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_DATE_TIME:
        if (!metadataGet(metadata, KEY_DATE)) {
          metadataSet(metadata, KEY_DATE, normalizeDate(readStringTag(tiffData, view, entry, littleEndian)));
        }
        break;
      case TAG_DATE_TIME_ORIGINAL:
        metadataSet(metadata, KEY_DATE, normalizeDate(readStringTag(tiffData, view, entry, littleEndian)));
        break;
      case TAG_ARTIST:
        metadataSet(metadata, KEY_CREATOR, readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_COPYRIGHT:
        metadataSet(metadata, 'exif:copyright', readStringTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_EXPOSURE_TIME:
        metadataSet(metadata, 'exif:exposure-time', readRationalTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_FNUMBER:
        metadataSet(metadata, 'exif:f-number', readRationalTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_ISO_SPEED:
        metadataSet(metadata, 'exif:iso', String(readShortOrLong(view, entry, littleEndian)));
        break;
      case TAG_FOCAL_LENGTH:
        metadataSet(metadata, 'exif:focal-length', readRationalTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_GPS_IFD:
        gpsIfdOffset = readShortOrLong(view, entry, littleEndian);
        break;
    }
  }

  // Also try to find EXIF sub-IFD (tag 0x8769) for exposure/date tags.
  for (const entry of entries) {
    if (entry.tag === 0x8769) {
      const exifIfdOffset = readShortOrLong(view, entry, littleEndian);
      if (exifIfdOffset > 0 && exifIfdOffset < tiffData.byteLength) {
        parseExifSubIFD(tiffData, view, exifIfdOffset, littleEndian, metadata);
      }
      break;
    }
  }

  // Parse GPS IFD.
  if (gpsIfdOffset >= 0 && gpsIfdOffset < tiffData.byteLength) {
    parseGPSIFD(tiffData, view, gpsIfdOffset, littleEndian, metadata);
  }
}

/** Parse the EXIF sub-IFD for tags that live there (DateTimeOriginal, exposure, etc). */
function parseExifSubIFD(
  tiffData: Uint8Array,
  view: DataView,
  offset: number,
  littleEndian: boolean,
  metadata: Record<string, string[]>,
): void {
  const entries = readIFDEntries(tiffData, view, offset, littleEndian);
  for (const entry of entries) {
    switch (entry.tag) {
      case TAG_DATE_TIME_ORIGINAL:
        metadataSet(metadata, KEY_DATE, normalizeDate(readStringTag(tiffData, view, entry, littleEndian)));
        break;
      case TAG_EXPOSURE_TIME:
        metadataSet(metadata, 'exif:exposure-time', readRationalTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_FNUMBER:
        metadataSet(metadata, 'exif:f-number', readRationalTag(tiffData, view, entry, littleEndian));
        break;
      case TAG_ISO_SPEED:
        metadataSet(metadata, 'exif:iso', String(readShortOrLong(view, entry, littleEndian)));
        break;
      case TAG_FOCAL_LENGTH:
        metadataSet(metadata, 'exif:focal-length', readRationalTag(tiffData, view, entry, littleEndian));
        break;
    }
  }
}

interface IFDEntry {
  tag: number;
  type: number;
  count: number;
  /** The 4-byte value/offset field (raw offset in the IFD). */
  valueOffset: number;
}

function readIFDEntries(
  tiffData: Uint8Array,
  view: DataView,
  ifdOffset: number,
  littleEndian: boolean,
): IFDEntry[] {
  if (ifdOffset + 2 > tiffData.byteLength) return [];
  const count = view.getUint16(ifdOffset, littleEndian);
  const entries: IFDEntry[] = [];
  let pos = ifdOffset + 2;
  for (let i = 0; i < count && pos + 12 <= tiffData.byteLength; i++) {
    entries.push({
      tag: view.getUint16(pos, littleEndian),
      type: view.getUint16(pos + 2, littleEndian),
      count: view.getUint32(pos + 4, littleEndian),
      valueOffset: pos + 8, // offset of the 4-byte value/offset field within tiffData
    });
    pos += 12;
  }
  return entries;
}

/** Size in bytes of each TIFF type. */
function tiffTypeSize(type: number): number {
  switch (type) {
    case 1: return 1; // BYTE
    case 2: return 1; // ASCII
    case 3: return 2; // SHORT
    case 4: return 4; // LONG
    case 5: return 8; // RATIONAL
    case 7: return 1; // UNDEFINED
    case 9: return 4; // SLONG
    case 10: return 8; // SRATIONAL
    default: return 1;
  }
}

/** Get the offset where the actual data lives (inline or pointed-to). Returns -1 on invalid. */
function dataOffset(
  tiffData: Uint8Array,
  view: DataView,
  entry: IFDEntry,
  littleEndian: boolean,
): number {
  // Validate count to prevent overflow (count is uint32, max type size is 8).
  if (entry.count > 0xFFFFFF) return -1; // ~16M entries max — generous but prevents overflow
  const totalBytes = entry.count * tiffTypeSize(entry.type);
  if (totalBytes <= 4) {
    // Data is inline in the value/offset field.
    return entry.valueOffset;
  }
  // Data is at the offset stored in the value/offset field.
  if (entry.valueOffset + 4 > tiffData.byteLength) return -1;
  const off = view.getUint32(entry.valueOffset, littleEndian);
  if (off >= tiffData.byteLength) return -1;
  return off;
}

function readStringTag(
  tiffData: Uint8Array,
  view: DataView,
  entry: IFDEntry,
  littleEndian: boolean,
): string {
  if (entry.type !== 2 && entry.type !== 7) return '';
  const off = dataOffset(tiffData, view, entry, littleEndian);
  if (off < 0 || off + entry.count > tiffData.byteLength) return '';
  let str = '';
  for (let i = 0; i < entry.count; i++) {
    const ch = tiffData[off + i]!;
    if (ch === 0) break;
    str += String.fromCharCode(ch);
  }
  return str.trim();
}

function readShortOrLong(
  view: DataView,
  entry: IFDEntry,
  littleEndian: boolean,
): number {
  if (entry.type === 3) {
    return view.getUint16(entry.valueOffset, littleEndian);
  }
  return view.getUint32(entry.valueOffset, littleEndian);
}

function readRationalTag(
  tiffData: Uint8Array,
  view: DataView,
  entry: IFDEntry,
  littleEndian: boolean,
): string {
  if (entry.type !== 5 && entry.type !== 10) return '';
  const off = dataOffset(tiffData, view, entry, littleEndian);
  if (off < 0 || off + 8 > tiffData.byteLength) return '';
  const num = view.getUint32(off, littleEndian);
  const den = view.getUint32(off + 4, littleEndian);
  if (den === 0) return String(num);
  return `${num}/${den}`;
}

/** Read an array of RATIONAL values (type 5) from an IFD entry. */
function readRationals(
  tiffData: Uint8Array,
  view: DataView,
  entry: IFDEntry,
  littleEndian: boolean,
): Array<[number, number]> {
  if (entry.type !== 5 && entry.type !== 10) return [];
  const off = dataOffset(tiffData, view, entry, littleEndian);
  if (off < 0) return [];
  const result: Array<[number, number]> = [];
  for (let i = 0; i < entry.count; i++) {
    const rOff = off + i * 8;
    if (rOff + 8 > tiffData.byteLength) break;
    result.push([
      view.getUint32(rOff, littleEndian),
      view.getUint32(rOff + 4, littleEndian),
    ]);
  }
  return result;
}

/** Parse GPS IFD and extract latitude/longitude. */
function parseGPSIFD(
  tiffData: Uint8Array,
  view: DataView,
  offset: number,
  littleEndian: boolean,
  metadata: Record<string, string[]>,
): void {
  const entries = readIFDEntries(tiffData, view, offset, littleEndian);

  let latRef = '';
  let lonRef = '';
  let latRationals: Array<[number, number]> = [];
  let lonRationals: Array<[number, number]> = [];

  for (const entry of entries) {
    switch (entry.tag) {
      case TAG_GPS_LAT_REF:
        latRef = readStringTag(tiffData, view, entry, littleEndian);
        break;
      case TAG_GPS_LAT:
        latRationals = readRationals(tiffData, view, entry, littleEndian);
        break;
      case TAG_GPS_LON_REF:
        lonRef = readStringTag(tiffData, view, entry, littleEndian);
        break;
      case TAG_GPS_LON:
        lonRationals = readRationals(tiffData, view, entry, littleEndian);
        break;
    }
  }

  if (latRationals.length >= 3 && lonRationals.length >= 3) {
    let lat = dmsToDecimal(latRationals);
    let lon = dmsToDecimal(lonRationals);
    if (latRef === 'S') lat = -lat;
    if (lonRef === 'W') lon = -lon;
    metadataSet(metadata, 'exif:gps-latitude', lat.toFixed(6));
    metadataSet(metadata, 'exif:gps-longitude', lon.toFixed(6));
  }
}

/** Convert degrees/minutes/seconds rationals to decimal degrees. */
function dmsToDecimal(rationals: Array<[number, number]>): number {
  const deg = rationals[0]![1] !== 0 ? rationals[0]![0] / rationals[0]![1] : 0;
  const min = rationals[1]![1] !== 0 ? rationals[1]![0] / rationals[1]![1] : 0;
  const sec = rationals[2]![1] !== 0 ? rationals[2]![0] / rationals[2]![1] : 0;
  return deg + min / 60 + sec / 3600;
}

// ---------------------------------------------------------------------------
// PNG tEXt chunk parsing
// ---------------------------------------------------------------------------

function extractPNGText(
  data: Uint8Array,
  metadata: Record<string, string[]>,
): void {
  // PNG signature is 8 bytes, then chunks follow.
  if (data.byteLength < 8) return;
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let pos = 8;

  while (pos + 12 <= data.byteLength) {
    const chunkLen = view.getUint32(pos, false);
    const chunkType = String.fromCharCode(
      data[pos + 4]!, data[pos + 5]!, data[pos + 6]!, data[pos + 7]!,
    );

    if (chunkLen > data.byteLength - pos - 12) break;

    if (chunkType === 'tEXt') {
      const chunkData = data.subarray(pos + 8, pos + 8 + chunkLen);
      const nulIdx = chunkData.indexOf(0);
      if (nulIdx >= 0) {
        const key = textDecoder.decode(chunkData.subarray(0, nulIdx));
        const val = textDecoder.decode(chunkData.subarray(nulIdx + 1));
        const pngKey = 'png:' + key.toLowerCase();
        metadataSet(metadata, pngKey, val);

        // Map well-known PNG keys to standard metadata.
        switch (key.toLowerCase()) {
          case 'title':
            metadataSet(metadata, KEY_TITLE, val);
            break;
          case 'author':
            metadataSet(metadata, KEY_CREATOR, val);
            break;
          case 'description':
            metadataSet(metadata, KEY_DESCRIPTION, val);
            break;
          case 'software':
            metadataSet(metadata, KEY_APPLICATION, val);
            break;
        }
      }
    }

    pos += 12 + chunkLen; // length(4) + type(4) + data + crc(4)
  }
}

const textDecoder = new TextDecoder('utf-8');

// ---------------------------------------------------------------------------
// Text summary
// ---------------------------------------------------------------------------

function buildTextSummary(metadata: Record<string, string[]>): string {
  const keys = [KEY_TITLE, KEY_DESCRIPTION, KEY_CREATOR, 'exif:make', 'exif:model'];
  const parts: string[] = [];
  for (const key of keys) {
    const v = metadataGet(metadata, key);
    if (v) parts.push(v);
  }
  return parts.join('\n');
}
