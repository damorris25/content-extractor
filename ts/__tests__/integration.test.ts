// Integration tests — golden file parity with Go.
// Each test extracts a fixture and asserts exact match with Go's output.

import { describe, it, beforeAll } from 'vitest';
import { createRegistry, Registry } from '../src/index.js';
import { assertGoldenMatch } from './helpers/golden.js';

let registry: Registry;

beforeAll(() => {
  registry = createRegistry();
});

describe('Golden file parity — DOCX', () => {
  it('simple.docx', () => assertGoldenMatch(registry, 'simple.docx'));
  it('multipar.docx', () => assertGoldenMatch(registry, 'multipar.docx'));
  it('empty.docx', () => assertGoldenMatch(registry, 'empty.docx'));
  it('rich.docx', () => assertGoldenMatch(registry, 'rich.docx'));
  it('unicode.docx', () => assertGoldenMatch(registry, 'unicode.docx'));
});

describe('Golden file parity — XLSX', () => {
  it('simple.xlsx', () => assertGoldenMatch(registry, 'simple.xlsx'));
  it('multisheet.xlsx', () => assertGoldenMatch(registry, 'multisheet.xlsx'));
  it('empty.xlsx', () => assertGoldenMatch(registry, 'empty.xlsx'));
  it('rich.xlsx', () => assertGoldenMatch(registry, 'rich.xlsx'));
});

describe('Golden file parity — PPTX', () => {
  it('simple.pptx', () => assertGoldenMatch(registry, 'simple.pptx'));
  it('multislide.pptx', () => assertGoldenMatch(registry, 'multislide.pptx'));
  it('empty.pptx', () => assertGoldenMatch(registry, 'empty.pptx'));
  it('rich.pptx', () => assertGoldenMatch(registry, 'rich.pptx'));
});

describe('Golden file parity — ODT', () => {
  it('simple.odt', () => assertGoldenMatch(registry, 'simple.odt'));
  it('multipar.odt', () => assertGoldenMatch(registry, 'multipar.odt'));
  it('empty.odt', () => assertGoldenMatch(registry, 'empty.odt'));
  it('rich.odt', () => assertGoldenMatch(registry, 'rich.odt'));
});

describe('Golden file parity — ODS', () => {
  it('simple.ods', () => assertGoldenMatch(registry, 'simple.ods'));
  it('multisheet.ods', () => assertGoldenMatch(registry, 'multisheet.ods'));
  it('empty.ods', () => assertGoldenMatch(registry, 'empty.ods'));
});

describe('Golden file parity — ODP', () => {
  it('simple.odp', () => assertGoldenMatch(registry, 'simple.odp'));
  it('multislide.odp', () => assertGoldenMatch(registry, 'multislide.odp'));
  it('empty.odp', () => assertGoldenMatch(registry, 'empty.odp'));
});

describe('Golden file parity — DXF', () => {
  it('simple.dxf', () => assertGoldenMatch(registry, 'simple.dxf'));
  it('rich.dxf', () => assertGoldenMatch(registry, 'rich.dxf'));
});

describe('Golden file parity — Images', () => {
  it('simple.jpg', () => assertGoldenMatch(registry, 'simple.jpg'));
  it('simple.png', () => assertGoldenMatch(registry, 'simple.png'));
  it('simple.gif', () => assertGoldenMatch(registry, 'simple.gif'));
  it('text.png', () => assertGoldenMatch(registry, 'text.png'));
});

describe('Golden file parity — Audio', () => {
  it('simple.mp3', () => assertGoldenMatch(registry, 'simple.mp3'));
  it('simple.ogg', () => assertGoldenMatch(registry, 'simple.ogg'));
  it('simple.flac', () => assertGoldenMatch(registry, 'simple.flac'));
  it('simple.wav', () => assertGoldenMatch(registry, 'simple.wav'));
  it('rich.mp3', () => assertGoldenMatch(registry, 'rich.mp3'));
});

// PDF tests skipped unless pdfjs-dist is installed.
// describe('Golden file parity — PDF', () => { ... });
