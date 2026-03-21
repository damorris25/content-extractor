// GeoTIFF metadata extraction tests.
// Tests the pure-TypeScript GeoTIFF tag parser.

import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { createMetadata, metadataGet } from '../src/metadata.js';
import { extractGeoTIFF } from '../src/formats/geotiff.js';

const TESTDATA = join(__dirname, '..', '..', 'testdata');

describe('GeoTIFF metadata extraction', () => {
  it('extracts CRS, bounds, and pixel scale from GeoTIFF', () => {
    const data = new Uint8Array(readFileSync(join(TESTDATA, 'geotiff-sample.tif')));
    const metadata = createMetadata();
    extractGeoTIFF(data, metadata);

    expect(metadataGet(metadata, 'geo:crs')).toBe('WGS 84');
    expect(metadataGet(metadata, 'geo:pixel-scale')).toBe('0.001000,0.001000');
    expect(metadataGet(metadata, 'geo:bounds')).toBe('-122.000000,36.996000,-121.996000,37.000000');
    expect(metadataGet(metadata, 'exif:width')).toBe('4');
    expect(metadataGet(metadata, 'exif:height')).toBe('4');
  });

  it('produces no geo metadata for empty data', () => {
    const metadata = createMetadata();
    extractGeoTIFF(new Uint8Array(0), metadata);

    expect(metadataGet(metadata, 'geo:crs')).toBe('');
    expect(metadataGet(metadata, 'geo:bounds')).toBe('');
    expect(metadataGet(metadata, 'geo:pixel-scale')).toBe('');
  });

  it('handles truncated TIFF data gracefully', () => {
    const data = new Uint8Array(readFileSync(join(TESTDATA, 'geotiff-sample.tif')));
    const truncated = data.subarray(0, 12);
    const metadata = createMetadata();
    extractGeoTIFF(truncated, metadata);

    // Should not crash, and produce no geo metadata.
    expect(metadataGet(metadata, 'geo:crs')).toBe('');
  });

  it('handles big-endian TIFF header', () => {
    // Minimal big-endian TIFF: MM, 42, IFD offset=8, 0 entries
    const data = new Uint8Array([
      0x4D, 0x4D, 0x00, 0x2A,  // MM + magic
      0x00, 0x00, 0x00, 0x08,  // IFD0 offset = 8
      0x00, 0x00,              // 0 entries
    ]);
    const metadata = createMetadata();
    extractGeoTIFF(data, metadata);

    // Should not crash, produce no metadata.
    expect(metadataGet(metadata, 'geo:crs')).toBe('');
  });

  it('produces no geo metadata for regular TIFF without geo tags', () => {
    // A TIFF with just ImageWidth and ImageLength, no geo tags.
    const data = new Uint8Array([
      0x49, 0x49, 0x2A, 0x00,  // II + magic
      0x08, 0x00, 0x00, 0x00,  // IFD0 offset = 8
      0x02, 0x00,              // 2 entries
      // Entry 1: ImageWidth = 10 (SHORT, type 3, count 1)
      0x00, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00,
      // Entry 2: ImageLength = 20 (SHORT, type 3, count 1)
      0x01, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00,
      // Next IFD = 0
      0x00, 0x00, 0x00, 0x00,
    ]);
    const metadata = createMetadata();
    extractGeoTIFF(data, metadata);

    // Dimensions should be set, but no geo metadata.
    expect(metadataGet(metadata, 'exif:width')).toBe('10');
    expect(metadataGet(metadata, 'exif:height')).toBe('20');
    expect(metadataGet(metadata, 'geo:crs')).toBe('');
    expect(metadataGet(metadata, 'geo:bounds')).toBe('');
  });
});
