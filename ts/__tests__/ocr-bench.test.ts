import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { join } from 'path';

// Performance measurement for OCR — not a parity test.
// Run with: npx vitest run -t "OCR bench"

function loadFixture(name: string): Uint8Array {
  return new Uint8Array(readFileSync(join(__dirname, '../../testdata', name)));
}

describe('OCR bench', () => {
  it('measure OCR recognition time (PNG)', async () => {
    let recognizeImage: (data: Uint8Array, lang: string) => Promise<{ text: string; confidence: number }>;
    try {
      const mod = await import('../src/formats/ocr.js');
      recognizeImage = mod.recognizeImage;
    } catch {
      console.log('tesseract.js not available, skipping OCR bench');
      return;
    }

    const data = loadFixture('ocr-text.png');

    // Warm up (first call loads WASM + model data)
    const warmStart = performance.now();
    const warmResult = await recognizeImage(data, 'eng');
    const warmMs = performance.now() - warmStart;
    console.log(`OCR warm-up (PNG): ${warmMs.toFixed(0)}ms, text=${JSON.stringify(warmResult.text)}, confidence=${warmResult.confidence.toFixed(1)}%`);
    expect(warmResult.text).toContain('Hello');

    // Subsequent calls (worker already loaded)
    const times: number[] = [];
    for (let i = 0; i < 3; i++) {
      const start = performance.now();
      await recognizeImage(data, 'eng');
      times.push(performance.now() - start);
    }
    const avg = times.reduce((a, b) => a + b, 0) / times.length;
    console.log(`OCR subsequent calls (PNG): avg=${avg.toFixed(0)}ms, runs=[${times.map(t => t.toFixed(0)).join(', ')}]ms`);

    // Also bench JPEG
    const jpegData = loadFixture('ocr-text.jpg');
    const jpegTimes: number[] = [];
    for (let i = 0; i < 3; i++) {
      const start = performance.now();
      await recognizeImage(jpegData, 'eng');
      jpegTimes.push(performance.now() - start);
    }
    const jpegAvg = jpegTimes.reduce((a, b) => a + b, 0) / jpegTimes.length;
    console.log(`OCR subsequent calls (JPEG): avg=${jpegAvg.toFixed(0)}ms, runs=[${jpegTimes.map(t => t.toFixed(0)).join(', ')}]ms`);

    // Cleanup
    const { terminateOCR } = await import('../src/formats/ocr.js');
    await terminateOCR();
  }, 120000);
});
