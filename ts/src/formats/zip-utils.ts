// ZIP utility wrapper with safety limits.
// Port of shared ZIP handling from Go office/register.go and odf/common.go

import { unzipSync } from 'fflate';
import { FileTooLargeError, CorruptInputError } from '../extractor.js';

const MAX_DECOMPRESSED_ENTRY_SIZE = 100 * 1024 * 1024; // 100MB per entry
const MAX_ZIP_ENTRIES = 10000;
const MAX_TOTAL_DECOMPRESSED_SIZE = 500 * 1024 * 1024; // 500MB aggregate

export interface ZipEntries {
  [name: string]: Uint8Array;
}

/**
 * Safely unzip data with size and entry count limits.
 * Throws FileTooLargeError if compressed data exceeds maxSize.
 * Throws CorruptInputError if the ZIP is malformed.
 */
export function safeUnzip(data: Uint8Array, maxSize: number): ZipEntries {
  // Check compressed size.
  if (data.byteLength > maxSize) {
    throw new FileTooLargeError();
  }

  let entries: ZipEntries;
  try {
    entries = unzipSync(data, {
      filter(file) {
        // Reject individual entries that are too large (decompressed).
        return file.size <= MAX_DECOMPRESSED_ENTRY_SIZE;
      },
    });
  } catch (err) {
    if (err instanceof FileTooLargeError) throw err;
    throw new CorruptInputError(
      `invalid ZIP: ${err instanceof Error ? err.message : String(err)}`,
    );
  }

  // Check entry count and aggregate decompressed size.
  const names = Object.keys(entries);
  if (names.length > MAX_ZIP_ENTRIES) {
    throw new CorruptInputError(
      `ZIP contains too many entries: ${names.length} > ${MAX_ZIP_ENTRIES}`,
    );
  }

  let totalSize = 0;
  for (const name of names) {
    const entry = entries[name];
    if (entry) {
      totalSize += entry.byteLength;
      if (totalSize > MAX_TOTAL_DECOMPRESSED_SIZE) {
        throw new FileTooLargeError(
          `ZIP total decompressed size exceeds ${MAX_TOTAL_DECOMPRESSED_SIZE} bytes`,
        );
      }
    }
  }

  return entries;
}

/** Read a named entry from a ZIP as a UTF-8 string. */
export function readZipEntry(entries: ZipEntries, name: string): string | null {
  const data = entries[name];
  if (!data) return null;
  return new TextDecoder().decode(data);
}

/** Read a named entry as raw bytes. */
export function getZipEntry(entries: ZipEntries, name: string): Uint8Array | null {
  return entries[name] ?? null;
}
