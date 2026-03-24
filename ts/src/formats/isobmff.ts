// CONTRACT:C1-EXTRACTOR.1.0
// ISOBMFF container parser for HEIC/HEIF/AVIF metadata extraction.
// Extracts dimensions from ispe box and EXIF from embedded Exif item.

import { metadataSet } from '../metadata.js';
import { parseExifBlock } from './image.js';

/** Safety limits. */
const MAX_BOX_COUNT = 1000;
const MAX_ITEM_COUNT = 10000;

interface Box {
  boxType: string;
  dataOffset: number;
  dataSize: number;
}

/**
 * Extract ISOBMFF metadata (dimensions + EXIF) into the metadata record.
 */
export function extractISOBMFF(
  data: Uint8Array,
  metadata: Record<string, string[]>,
): void {
  const boxes = walkBoxes(data, 0, data.byteLength);

  // Find the meta box.
  const metaBox = boxes.find(b => b.boxType === 'meta');
  if (!metaBox) return;

  // meta is a FullBox — skip 4 bytes of version/flags before children.
  const metaStart = metaBox.dataOffset + 4;
  const metaEnd = metaBox.dataOffset + metaBox.dataSize;
  if (metaStart >= metaEnd || metaStart >= data.byteLength) return;

  const metaChildren = walkBoxes(data, metaStart, metaEnd);

  // Extract dimensions from ispe (meta > iprp > ipco > ispe).
  extractISPE(data, metaChildren, metadata);

  // Extract EXIF via iloc + iinf.
  extractISOBMFFExif(data, metaChildren, metadata);
}

/** Walk ISOBMFF boxes between start and end offsets. */
function walkBoxes(data: Uint8Array, start: number, end: number): Box[] {
  if (end > data.byteLength) end = data.byteLength;
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  const boxes: Box[] = [];
  let pos = start;

  while (pos + 8 <= end && boxes.length < MAX_BOX_COUNT) {
    const size = view.getUint32(pos, false);
    const boxType = String.fromCharCode(
      data[pos + 4]!, data[pos + 5]!, data[pos + 6]!, data[pos + 7]!,
    );

    let headerSize = 8;
    let boxEnd = pos + size;

    if (size === 1) {
      // Extended size (8-byte).
      if (pos + 16 > end) break;
      const hi = view.getUint32(pos + 8, false);
      const lo = view.getUint32(pos + 12, false);
      const extSize = hi * 0x100000000 + lo;
      if (extSize > end - pos) break;
      boxEnd = pos + extSize;
      headerSize = 16;
    } else if (size === 0) {
      boxEnd = end;
    } else if (size < 8) {
      break;
    }

    if (boxEnd > end) boxEnd = end;

    boxes.push({
      boxType,
      dataOffset: pos + headerSize,
      dataSize: boxEnd - pos - headerSize,
    });

    pos = boxEnd;
  }

  return boxes;
}

/** Extract dimensions from ispe box at meta > iprp > ipco > ispe. */
function extractISPE(
  data: Uint8Array,
  metaChildren: Box[],
  metadata: Record<string, string[]>,
): void {
  const iprp = metaChildren.find(b => b.boxType === 'iprp');
  if (!iprp) return;

  const iprpChildren = walkBoxes(data, iprp.dataOffset, iprp.dataOffset + iprp.dataSize);
  const ipco = iprpChildren.find(b => b.boxType === 'ipco');
  if (!ipco) return;

  const ipcoChildren = walkBoxes(data, ipco.dataOffset, ipco.dataOffset + ipco.dataSize);
  for (const box of ipcoChildren) {
    if (box.boxType === 'ispe' && box.dataSize >= 12) {
      const off = box.dataOffset;
      if (off + 12 > data.byteLength) continue;
      const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
      // ispe: 4 bytes version/flags + 4 bytes width + 4 bytes height
      const w = view.getUint32(off + 4, false);
      const h = view.getUint32(off + 8, false);
      if (w > 0) metadataSet(metadata, 'exif:width', String(w));
      if (h > 0) metadataSet(metadata, 'exif:height', String(h));
      return;
    }
  }
}

interface ILocItem {
  itemID: number;
  offset: number;
  size: number;
}

/** Extract EXIF from ISOBMFF: find Exif item via iinf, locate via iloc, parse. */
function extractISOBMFFExif(
  data: Uint8Array,
  metaChildren: Box[],
  metadata: Record<string, string[]>,
): void {
  const exifItemID = findExifItemID(data, metaChildren);
  if (exifItemID === 0) return;

  const items = parseILOC(data, metaChildren);
  if (!items) return;

  const item = items.get(exifItemID);
  if (!item) return;

  if (item.offset + item.size > data.byteLength || item.size < 10) return;

  let exifBlob = data.subarray(item.offset, item.offset + item.size);

  // HEIC/AVIF EXIF items have a 4-byte prefix (offset to TIFF header).
  if (exifBlob.byteLength > 4) {
    const view = new DataView(exifBlob.buffer, exifBlob.byteOffset, exifBlob.byteLength);
    const tiffOffset = view.getUint32(0, false);
    const headerStart = 4 + tiffOffset;
    if (headerStart < exifBlob.byteLength) {
      exifBlob = exifBlob.subarray(headerStart);
    }
  }

  parseExifBlock(exifBlob, metadata);
}

/** Find the Exif item ID from the iinf box. */
function findExifItemID(data: Uint8Array, metaChildren: Box[]): number {
  const iinf = metaChildren.find(b => b.boxType === 'iinf');
  if (!iinf) return 0;

  const off = iinf.dataOffset;
  const end = iinf.dataOffset + iinf.dataSize;
  if (off + 6 > end || off + 6 > data.byteLength) return 0;

  const version = data[off]!;
  const entryCountSize = version === 0 ? 6 : 8;

  const infeBoxes = walkBoxes(data, off + entryCountSize, end);

  for (const box of infeBoxes) {
    if (box.boxType !== 'infe' || box.dataSize < 8) continue;

    const infeOff = box.dataOffset;
    if (infeOff + 12 > data.byteLength) continue;

    const infeVersion = data[infeOff]!;
    if (infeVersion >= 2) {
      if (infeOff + 12 > box.dataOffset + box.dataSize) continue;
      const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
      const itemID = view.getUint16(infeOff + 4, false);
      const itemType = String.fromCharCode(
        data[infeOff + 8]!, data[infeOff + 9]!, data[infeOff + 10]!, data[infeOff + 11]!,
      );
      if (itemType === 'Exif') return itemID;
    }
  }

  return 0;
}

/** Parse the iloc box and return a map of item_id -> location. */
function parseILOC(data: Uint8Array, metaChildren: Box[]): Map<number, ILocItem> | null {
  const iloc = metaChildren.find(b => b.boxType === 'iloc');
  if (!iloc) return null;

  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let off = iloc.dataOffset;
  const end = iloc.dataOffset + iloc.dataSize;
  if (off + 8 > end || off + 8 > data.byteLength) return null;

  const version = data[off]!;
  const sizeByte1 = data[off + 4]!;
  const sizeByte2 = data[off + 5]!;
  const offsetSize = sizeByte1 >> 4;
  const lengthSize = sizeByte1 & 0x0F;
  const baseOffsetSize = sizeByte2 >> 4;

  let pos = off + 6;

  let itemCount: number;
  if (version < 2) {
    if (pos + 2 > end) return null;
    itemCount = view.getUint16(pos, false);
    pos += 2;
  } else {
    if (pos + 4 > end) return null;
    itemCount = view.getUint32(pos, false);
    pos += 4;
  }

  if (itemCount > MAX_ITEM_COUNT) return null;

  const items = new Map<number, ILocItem>();

  for (let i = 0; i < itemCount; i++) {
    let itemID: number;
    if (version < 2) {
      if (pos + 2 > end) break;
      itemID = view.getUint16(pos, false);
      pos += 2;
    } else {
      if (pos + 4 > end) break;
      itemID = view.getUint32(pos, false);
      pos += 4;
    }

    // construction_method (v1/v2 only)
    if (version === 1 || version === 2) pos += 2;

    // data_reference_index
    pos += 2;

    // base_offset
    const baseOffset = readILOCSize(view, pos, end, baseOffsetSize);
    pos += baseOffsetSize;

    // extent_count
    if (pos + 2 > end) break;
    const extentCount = view.getUint16(pos, false);
    pos += 2;

    for (let j = 0; j < extentCount; j++) {
      // extent_index (v1/v2 only)
      if (version === 1 || version === 2) {
        const indexSize = sizeByte2 & 0x0F;
        pos += indexSize;
      }

      const extentOffset = readILOCSize(view, pos, end, offsetSize);
      pos += offsetSize;

      const extentLength = readILOCSize(view, pos, end, lengthSize);
      pos += lengthSize;

      if (j === 0) {
        items.set(itemID, {
          itemID,
          offset: baseOffset + extentOffset,
          size: extentLength,
        });
      }
    }
  }

  return items;
}

/** Read a variable-size integer (0, 2, 4, or 8 bytes) from iloc data. */
function readILOCSize(view: DataView, pos: number, end: number, size: number): number {
  switch (size) {
    case 2:
      return pos + 2 <= end ? view.getUint16(pos, false) : 0;
    case 4:
      return pos + 4 <= end ? view.getUint32(pos, false) : 0;
    case 8:
      if (pos + 8 > end) return 0;
      return view.getUint32(pos, false) * 0x100000000 + view.getUint32(pos + 4, false);
    default:
      return 0;
  }
}
