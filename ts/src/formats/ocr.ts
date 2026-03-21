// OCR text extraction via tesseract.js (optional peer dependency).
// Port of formats/image/ocr.go

/** Result of OCR text recognition. */
export interface OCRResult {
  /** Recognized text content. */
  text: string;
  /** Mean confidence score (0-100). */
  confidence: number;
}

// Lazy-initialized worker singleton.
let workerPromise: Promise<any> | null = null;

/**
 * Recognize text in an image using Tesseract OCR.
 *
 * Requires `tesseract.js` to be installed as an optional peer dependency.
 * Throws an error if tesseract.js is not available.
 *
 * @param data - Image data as Uint8Array (JPEG, PNG, TIFF, BMP, etc.)
 * @param lang - Tesseract language code (default "eng")
 * @param signal - Optional AbortSignal for cancellation
 * @returns Recognized text and confidence score
 */
export async function recognizeImage(
  data: Uint8Array,
  lang = 'eng',
  signal?: AbortSignal,
): Promise<OCRResult> {
  if (data.byteLength === 0) {
    return { text: '', confidence: 0 };
  }

  if (signal?.aborted) {
    throw signal.reason ?? new DOMException('Aborted', 'AbortError');
  }

  const Tesseract = await loadTesseract();
  const worker = await getWorker(Tesseract, lang);

  if (signal?.aborted) {
    throw signal.reason ?? new DOMException('Aborted', 'AbortError');
  }

  const result = await worker.recognize(data);
  const text = (result.data.text ?? '').trim();
  const confidence = result.data.confidence ?? 0;

  return { text, confidence };
}

/**
 * Terminate the cached OCR worker and free resources.
 * Call this when OCR is no longer needed to release memory.
 */
export async function terminateOCR(): Promise<void> {
  if (workerPromise) {
    try {
      const worker = await workerPromise;
      await worker.terminate();
    } catch {
      // Ignore errors during cleanup.
    }
    workerPromise = null;
  }
}

/** Dynamically import tesseract.js. Throws descriptive error if not installed. */
async function loadTesseract(): Promise<any> {
  const moduleName = 'tesseract.js';
  try {
    return await import(/* @vite-ignore */ moduleName);
  } catch {
    throw new Error(
      'tesseract.js is required for OCR. Install it: npm install tesseract.js',
    );
  }
}

/** Get or create a cached Tesseract worker. */
async function getWorker(Tesseract: any, lang: string): Promise<any> {
  if (!workerPromise) {
    workerPromise = Tesseract.createWorker(lang);
  }
  return workerPromise;
}
