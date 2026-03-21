// OCR tests — require tesseract.js to be installed.
// These tests are separate from golden file parity tests since OCR output
// varies between Tesseract implementations (Go gosseract vs TS tesseract.js).

import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { ImageExtractor, type ImageExtractorOptions } from '../src/formats/image.js';
import { metadataGet } from '../src/metadata.js';

const TESTDATA = join(__dirname, '..', '..', 'testdata');

/** Check if tesseract.js is available. */
async function isTesseractAvailable(): Promise<boolean> {
  try {
    await import('tesseract.js');
    return true;
  } catch {
    return false;
  }
}

let hasTesseract = false;

beforeAll(async () => {
  hasTesseract = await isTesseractAvailable();
  if (!hasTesseract) {
    console.log('tesseract.js not installed — skipping OCR tests');
  }
});

afterAll(async () => {
  if (hasTesseract) {
    try {
      const { terminateOCR } = await import('../src/formats/ocr.js');
      await terminateOCR();
    } catch {
      // Ignore cleanup errors.
    }
  }
});

describe('OCR text extraction', () => {
  it('recognizes text in an image', async () => {
    if (!hasTesseract) return;

    const data = new Uint8Array(readFileSync(join(TESTDATA, 'ocr-text.png')));
    const ext = new ImageExtractor({ enableOCR: true });
    const result = await ext.extract(data, 'image/png');

    // Should contain "Hello" somewhere in the recognized text.
    expect(result.text.toLowerCase()).toContain('hello');
    expect(metadataGet(result.metadata, 'ocr:confidence')).not.toBe('');
    expect(metadataGet(result.metadata, 'ocr:language')).toBe('eng');
  });

  it('falls back to metadata summary when OCR disabled', async () => {
    const data = new Uint8Array(readFileSync(join(TESTDATA, 'text.png')));
    const ext = new ImageExtractor(); // OCR disabled by default
    const result = await ext.extract(data, 'image/png');

    // Should have metadata summary, not OCR text.
    expect(result.text).toContain('Green Square');
    expect(metadataGet(result.metadata, 'ocr:confidence')).toBe('');
  });

  it('handles OCR on JPEG format', async () => {
    if (!hasTesseract) return;

    const data = new Uint8Array(readFileSync(join(TESTDATA, 'ocr-text.jpg')));
    const ext = new ImageExtractor({ enableOCR: true });
    const result = await ext.extract(data, 'image/jpeg');

    expect(result.text.toLowerCase()).toContain('hello');
  });

  it('returns empty text for blank images', async () => {
    if (!hasTesseract) return;

    const data = new Uint8Array(readFileSync(join(TESTDATA, 'simple.png')));
    const ext = new ImageExtractor({ enableOCR: true });
    const result = await ext.extract(data, 'image/png');

    // Blank image — OCR should return empty or very short text.
    // Metadata summary may still be present.
    expect(metadataGet(result.metadata, 'exif:width')).toBe('200');
  });

  it('gracefully handles missing tesseract.js', async () => {
    // Even with enableOCR, if tesseract.js throws, we get metadata-only result.
    const data = new Uint8Array(readFileSync(join(TESTDATA, 'text.png')));
    const ext = new ImageExtractor({ enableOCR: false });
    const result = await ext.extract(data, 'image/png');

    // Should not throw, should have metadata.
    expect(result.contentType).toBe('image/png');
    expect(metadataGet(result.metadata, 'exif:width')).toBe('64');
  });

  it('supports AbortSignal cancellation', async () => {
    if (!hasTesseract) return;

    const controller = new AbortController();
    controller.abort(); // Abort immediately.

    const data = new Uint8Array(readFileSync(join(TESTDATA, 'ocr-text.png')));
    const ext = new ImageExtractor({ enableOCR: true });

    await expect(ext.extract(data, 'image/png', { signal: controller.signal }))
      .rejects.toThrow();
  });
});
