// GeoTIFF metadata extraction.
// Port of formats/image/geotiff.go

import { metadataSet, metadataGet, metadataHas } from '../metadata.js';
import type { Metadata } from '../metadata.js';

// GeoTIFF tag IDs.
const TAG_MODEL_PIXEL_SCALE = 33550;
const TAG_MODEL_TIEPOINT = 33922;
const TAG_GEO_KEY_DIRECTORY = 34735;
const TAG_GEO_ASCII_PARAMS = 34737;

// TIFF image dimension tags.
const TAG_IMAGE_WIDTH = 256;
const TAG_IMAGE_LENGTH = 257;

// GeoKey IDs.
const GEO_KEY_GEOG_CITATION = 2049;

/**
 * Extract GeoTIFF metadata from TIFF IFD tags.
 * Extracts CRS, bounds, and pixel scale from ModelTiepointTag,
 * ModelPixelScaleTag, GeoKeyDirectoryTag, and GeoAsciiParamsTag.
 * Non-GeoTIFF files silently produce no geo metadata.
 */
export function extractGeoTIFF(data: Uint8Array, metadata: Metadata): void {
  if (data.byteLength < 8) return;

  // Determine byte order.
  const bo0 = data[0]!;
  const bo1 = data[1]!;
  let littleEndian: boolean;
  if (bo0 === 0x49 && bo1 === 0x49) {
    littleEndian = true;
  } else if (bo0 === 0x4D && bo1 === 0x4D) {
    littleEndian = false;
  } else {
    return;
  }

  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);

  // Verify TIFF magic (42).
  if (view.getUint16(2, littleEndian) !== 42) return;

  // Read IFD0 offset.
  const ifdOffset = view.getUint32(4, littleEndian);
  if (ifdOffset + 2 > data.byteLength) return;

  const numEntries = view.getUint16(ifdOffset, littleEndian);
  const entryStart = ifdOffset + 2;

  let pixelScale: [number, number, number] | null = null;
  let tiepoints: number[] = [];
  let geoAscii = '';
  let geoKeys: number[] = [];
  let tiffWidth = 0;
  let tiffHeight = 0;

  for (let i = 0; i < numEntries; i++) {
    const off = entryStart + i * 12;
    if (off + 12 > data.byteLength) break;

    const tag = view.getUint16(off, littleEndian);
    const typ = view.getUint16(off + 2, littleEndian);
    const count = view.getUint32(off + 4, littleEndian);
    const valueOff = off + 8;

    switch (tag) {
      case TAG_IMAGE_WIDTH:
        tiffWidth = readShortOrLongValue(data, view, typ, valueOff, littleEndian);
        break;
      case TAG_IMAGE_LENGTH:
        tiffHeight = readShortOrLongValue(data, view, typ, valueOff, littleEndian);
        break;

      case TAG_MODEL_PIXEL_SCALE: {
        const doubles = readDoubles(data, view, typ, count, valueOff, littleEndian);
        if (doubles.length >= 3) {
          pixelScale = [doubles[0]!, doubles[1]!, doubles[2]!];
        }
        break;
      }
      case TAG_MODEL_TIEPOINT:
        tiepoints = readDoubles(data, view, typ, count, valueOff, littleEndian);
        break;

      case TAG_GEO_KEY_DIRECTORY:
        geoKeys = readShorts(data, view, typ, count, valueOff, littleEndian);
        break;

      case TAG_GEO_ASCII_PARAMS:
        geoAscii = readASCII(data, view, typ, count, valueOff, littleEndian);
        break;
    }
  }

  // Set TIFF dimensions if not already present.
  if (tiffWidth > 0 && !metadataHas(metadata, 'exif:width')) {
    metadataSet(metadata, 'exif:width', String(tiffWidth));
  }
  if (tiffHeight > 0 && !metadataHas(metadata, 'exif:height')) {
    metadataSet(metadata, 'exif:height', String(tiffHeight));
  }

  // Extract CRS.
  const crs = extractCRS(geoKeys, geoAscii);
  if (crs) {
    metadataSet(metadata, 'geo:crs', crs);
  }

  // Extract pixel scale.
  if (pixelScale) {
    metadataSet(metadata, 'geo:pixel-scale', `${pixelScale[0].toFixed(6)},${pixelScale[1].toFixed(6)}`);
  }

  // Compute bounds from first tiepoint + pixel scale + image dimensions.
  if (tiepoints.length >= 6 && pixelScale && tiffWidth > 0 && tiffHeight > 0) {
    const originX = tiepoints[3]!;
    const originY = tiepoints[4]!;
    const minX = originX;
    const maxY = originY;
    const maxX = originX + tiffWidth * pixelScale[0];
    const minY = originY - tiffHeight * pixelScale[1];
    metadataSet(metadata, 'geo:bounds',
      `${minX.toFixed(6)},${minY.toFixed(6)},${maxX.toFixed(6)},${maxY.toFixed(6)}`);
  }
}

function readShortOrLongValue(
  data: Uint8Array, view: DataView, typ: number, valueOff: number, littleEndian: boolean,
): number {
  if (valueOff + 4 > data.byteLength) return 0;
  if (typ === 3) return view.getUint16(valueOff, littleEndian); // SHORT
  if (typ === 4) return view.getUint32(valueOff, littleEndian); // LONG
  return 0;
}

function readDoubles(
  data: Uint8Array, view: DataView, typ: number, count: number,
  valueOff: number, littleEndian: boolean,
): number[] {
  if (typ !== 12) return []; // DOUBLE
  if (count > 10000) return [];
  if (valueOff + 4 > data.byteLength) return [];
  const dataOff = view.getUint32(valueOff, littleEndian);
  if (dataOff + count * 8 > data.byteLength) return [];
  const result: number[] = [];
  for (let i = 0; i < count; i++) {
    result.push(view.getFloat64(dataOff + i * 8, littleEndian));
  }
  return result;
}

function readShorts(
  data: Uint8Array, view: DataView, typ: number, count: number,
  valueOff: number, littleEndian: boolean,
): number[] {
  if (typ !== 3) return []; // SHORT
  if (count > 100000) return [];
  const totalBytes = count * 2;
  let dataOff: number;
  if (totalBytes <= 4) {
    dataOff = valueOff;
  } else {
    if (valueOff + 4 > data.byteLength) return [];
    dataOff = view.getUint32(valueOff, littleEndian);
  }
  if (dataOff + totalBytes > data.byteLength) return [];
  const result: number[] = [];
  for (let i = 0; i < count; i++) {
    result.push(view.getUint16(dataOff + i * 2, littleEndian));
  }
  return result;
}

function readASCII(
  data: Uint8Array, view: DataView, typ: number, count: number,
  valueOff: number, littleEndian: boolean,
): string {
  if (typ !== 2) return ''; // ASCII
  if (count > 100000) return '';
  let dataOff: number;
  if (count <= 4) {
    dataOff = valueOff;
  } else {
    if (valueOff + 4 > data.byteLength) return '';
    dataOff = view.getUint32(valueOff, littleEndian);
  }
  if (dataOff + count > data.byteLength) return '';
  let s = '';
  for (let i = 0; i < count; i++) {
    const ch = data[dataOff + i]!;
    if (ch === 0) break;
    s += String.fromCharCode(ch);
  }
  // Strip trailing pipe characters (GeoASCII uses | as separator).
  while (s.length > 0 && s[s.length - 1] === '|') {
    s = s.slice(0, -1);
  }
  return s;
}

function extractCRS(geoKeys: number[], geoAscii: string): string {
  if (geoKeys.length < 4) return '';
  const numKeys = geoKeys[3]!;
  for (let i = 0; i < numKeys; i++) {
    const base = 4 + i * 4;
    if (base + 3 >= geoKeys.length) break;
    const keyID = geoKeys[base]!;
    const tiffTagLoc = geoKeys[base + 1]!;
    const count = geoKeys[base + 2]!;
    const valueOff = geoKeys[base + 3]!;

    if (keyID === GEO_KEY_GEOG_CITATION && tiffTagLoc === TAG_GEO_ASCII_PARAMS) {
      if (valueOff + count <= geoAscii.length) {
        let s = geoAscii.slice(valueOff, valueOff + count);
        while (s.length > 0 && (s[s.length - 1] === '|' || s.charCodeAt(s.length - 1) === 0)) {
          s = s.slice(0, -1);
        }
        return s;
      }
      return geoAscii;
    }
  }
  if (geoAscii) return geoAscii;
  return '';
}
