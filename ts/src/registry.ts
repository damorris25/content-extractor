// Registry maps MIME types to Extractor implementations.
// Port of extractor/registry.go

import type { Extractor, ExtractionResult, ExtractOptions } from './extractor.js';
import { UnsupportedFormatError, DetectionFailedError } from './extractor.js';
import type { Detector } from './detect.js';

/**
 * Registry maps MIME types to Extractor implementations.
 * Safe for concurrent reads after initialization (JS is single-threaded).
 */
export class Registry {
  private extractors = new Map<string, Extractor>();
  private detector: Detector | null = null;

  /** Configure the format detector for auto-detection. */
  setDetector(d: Detector): void {
    this.detector = d;
  }

  /**
   * Register an extractor for all MIME types it supports.
   * Throws on duplicate MIME or null extractor (fail-fast at startup).
   */
  register(ext: Extractor): void {
    if (!ext) {
      throw new Error('register called with null/undefined extractor');
    }
    for (const mime of ext.supportedMIMETypes()) {
      const existing = this.extractors.get(mime);
      if (existing) {
        throw new Error(`MIME type "${mime}" already registered`);
      }
      this.extractors.set(mime, ext);
    }
  }

  /** Returns the extractor for the given MIME type, or null. */
  lookup(mime: string): Extractor | null {
    return this.extractors.get(mime) ?? null;
  }

  /**
   * Extract detects the format (if mime is empty) and dispatches to the
   * appropriate extractor.
   */
  async extract(
    data: Uint8Array,
    mime?: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (!mime) {
      if (!this.detector) {
        throw new DetectionFailedError();
      }
      mime = this.detector.detect(data, options?.filename);
    }

    const ext = this.lookup(mime);
    if (!ext) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    return ext.extract(data, mime, options);
  }

  /** Returns all registered MIME types, sorted alphabetically. */
  supportedTypes(): string[] {
    return [...this.extractors.keys()].sort();
  }
}
