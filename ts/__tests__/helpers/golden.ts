// Golden file test utility for parity testing with Go.

import { readFileSync } from 'fs';
import { resolve } from 'path';
import type { ExtractionResult } from '../../src/extractor.js';
import type { Registry } from '../../src/registry.js';

interface GoldenResult {
  text: string;
  metadata: Record<string, string[]>;
  contentType: string;
  language: string;
}

const testdataDir = resolve(__dirname, '../../../testdata');
const goldenDir = resolve(__dirname, '../golden');

/** Load a fixture file as Uint8Array. */
export function loadFixture(name: string): Uint8Array {
  return new Uint8Array(readFileSync(resolve(testdataDir, name)));
}

/** Load a golden JSON file. */
export function loadGolden(fixtureName: string): GoldenResult {
  const path = resolve(goldenDir, fixtureName + '.json');
  return JSON.parse(readFileSync(path, 'utf-8')) as GoldenResult;
}

/** Normalize metadata for comparison: sort keys and sort values within each key. */
function normalizeMetadata(m: Record<string, string[]>): Record<string, string[]> {
  const result: Record<string, string[]> = {};
  for (const key of Object.keys(m).sort()) {
    const vals = m[key];
    if (vals && vals.length > 0) {
      result[key] = [...vals].sort();
    }
  }
  return result;
}

/**
 * Assert that extraction of a fixture produces results matching the golden file.
 * Uses the provided registry and optional explicit MIME type.
 */
export async function assertGoldenMatch(
  registry: Registry,
  fixtureName: string,
  mime?: string,
): Promise<void> {
  const data = loadFixture(fixtureName);
  const golden = loadGolden(fixtureName);

  const result = await registry.extract(data, mime);

  // Compare text.
  expect(result.text).toBe(golden.text);

  // Compare contentType.
  expect(result.contentType).toBe(golden.contentType);

  // Compare language.
  expect(result.language).toBe(golden.language);

  // Compare metadata (normalized for deterministic comparison).
  expect(normalizeMetadata(result.metadata)).toEqual(normalizeMetadata(golden.metadata));
}

/**
 * Assert that extraction of a fixture matches golden file using explicit MIME.
 */
export async function assertGoldenMatchMIME(
  registry: Registry,
  fixtureName: string,
  mime: string,
): Promise<void> {
  return assertGoldenMatch(registry, fixtureName, mime);
}
