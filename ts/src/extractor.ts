// Core extraction interface and types.
// Port of extractor/extractor.go + extractor/errors.go

import type { Metadata } from './metadata.js';

/** Options for extraction. */
export interface ExtractOptions {
  /** AbortSignal for cancellation (replaces Go's context.Context). */
  signal?: AbortSignal;
  /** Filename hint for format detection fallback. */
  filename?: string;
}

/** Result of content extraction. */
export interface ExtractionResult {
  /** Extracted text content (UTF-8). */
  text: string;
  /** Namespaced metadata key-value pairs. */
  metadata: Metadata;
  /** Detected or provided MIME type. */
  contentType: string;
  /** BCP 47 language code (empty if unknown). */
  language: string;
}

/** Extractor extracts text and metadata from a specific file format. */
export interface Extractor {
  /** Extract text and metadata from binary data. */
  extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult>;
  /** Returns the MIME types this extractor handles. */
  supportedMIMETypes(): string[];
}

// --- Error classes ---

/** Base class for all extraction errors. */
export class ExtractionError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ExtractionError';
  }
}

/** Input data is not in a supported format. */
export class UnsupportedFormatError extends ExtractionError {
  constructor(message = 'unsupported format') {
    super(message);
    this.name = 'UnsupportedFormatError';
  }
}

/** Input data is malformed or truncated. */
export class CorruptInputError extends ExtractionError {
  constructor(message = 'corrupt input') {
    super(message);
    this.name = 'CorruptInputError';
  }
}

/** Input exceeds the configured size limit. */
export class FileTooLargeError extends ExtractionError {
  constructor(message = 'file too large') {
    super(message);
    this.name = 'FileTooLargeError';
  }
}

/** Format auto-detection failed. */
export class DetectionFailedError extends ExtractionError {
  constructor(message = 'detection failed') {
    super(message);
    this.name = 'DetectionFailedError';
  }
}

/** File is password-protected. */
export class PasswordProtectedError extends ExtractionError {
  constructor(message = 'password protected') {
    super(message);
    this.name = 'PasswordProtectedError';
  }
}

/** Check if an AbortSignal has been aborted and throw if so. */
export function checkAborted(signal?: AbortSignal): void {
  if (signal?.aborted) {
    throw signal.reason ?? new DOMException('Aborted', 'AbortError');
  }
}
