// CONTRACT:C1-EXTRACTOR.1.0
// DXF CAD extractor — port of formats/cad/dxf.go

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT, KEY_DATE } from '../metadata.js';

/** Default maximum DXF file size (200 MB). */
const DEFAULT_MAX_DXF_SIZE = 200 * 1024 * 1024;

const MIME_DXF = 'image/vnd.dxf';

/** Cap on collected text entities to prevent memory exhaustion. */
const MAX_TEXTS = 100_000;
/** Cap on collected layers. */
const MAX_LAYERS = 10_000;

/** Extracts text content and metadata from DXF CAD files. */
export class DXFExtractor implements Extractor {
  private readonly maxSize: number;

  constructor(maxSize: number = DEFAULT_MAX_DXF_SIZE) {
    this.maxSize = maxSize;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_DXF) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    if (data.length === 0) {
      const metadata = createMetadata();
      return {
        text: '',
        metadata,
        contentType: mime,
        language: '',
      };
    }

    if (data.length > this.maxSize) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const decoder = new TextDecoder('utf-8');
    const content = decoder.decode(data);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, mime);

    const result: ExtractionResult = {
      text: '',
      metadata,
      contentType: mime,
      language: '',
    };

    parseDXF(content, result);

    return result;
  }

  supportedMIMETypes(): string[] {
    return [MIME_DXF];
  }
}

/**
 * Parse a DXF file extracting text content and metadata.
 * DXF files consist of group code/value pairs organized into sections.
 */
function parseDXF(content: string, result: ExtractionResult): void {
  const lines = content.split(/\r?\n/);
  const texts: string[] = [];
  const layers: string[] = [];
  const layerSeen = new Set<string>();

  let section = '';
  let entityType = '';
  let headerVar = '';

  let i = 0;
  while (i < lines.length) {
    const code = lines[i]!.trim();
    i++;
    if (i >= lines.length) break;
    const value = lines[i]!.trim();
    i++;

    // Track sections.
    if (code === '0' && value === 'SECTION') {
      continue;
    }
    if (code === '2' && section === '') {
      section = value;
      continue;
    }
    if (code === '0' && value === 'ENDSEC') {
      section = '';
      entityType = '';
      headerVar = '';
      continue;
    }

    switch (section) {
      case 'HEADER':
        if (code === '9') {
          headerVar = value;
        } else if (headerVar !== '') {
          setHeaderMeta(headerVar, value, result);
          headerVar = '';
        }
        break;

      case 'TABLES':
        if (code === '0') {
          entityType = value;
        }
        // Collect layer names from the TABLES section.
        if (entityType === 'LAYER' && code === '2' && value !== '' && layers.length < MAX_LAYERS) {
          if (!layerSeen.has(value)) {
            layerSeen.add(value);
            layers.push(value);
          }
        }
        break;

      case 'ENTITIES':
        if (code === '0') {
          entityType = value;
        }
        if (texts.length >= MAX_TEXTS) {
          continue;
        }
        switch (entityType) {
          case 'TEXT':
          case 'ATTRIB':
          case 'ATTDEF':
            if (code === '1' && value !== '') {
              texts.push(value);
            }
            break;
          case 'MTEXT':
            if ((code === '1' || code === '3') && value !== '') {
              texts.push(cleanMTEXT(value));
            }
            break;
          case 'INSERT':
            if (code === '2' && value !== '') {
              texts.push(value);
            }
            break;
          case 'DIMENSION':
            if (code === '1' && value !== '' && value !== '<>') {
              texts.push(value);
            }
            break;
        }
        break;
    }
  }

  if (layers.length > 0) {
    metadataSet(result.metadata, 'cad:layers', layers.join(', '));
  }
  result.text = texts.join('\n');
}

/** Extract metadata from DXF header variable pairs. */
function setHeaderMeta(varName: string, value: string, result: ExtractionResult): void {
  switch (varName) {
    case '$ACADVER':
      metadataSet(result.metadata, 'cad:version', value);
      break;
    case '$INSUNITS':
      metadataSet(result.metadata, 'cad:units', dxfUnitName(value));
      break;
    case '$TDCREATE':
      // Julian date — store raw for now.
      metadataSet(result.metadata, KEY_DATE, value);
      break;
  }
}

/** Convert a DXF INSUNITS value to a human-readable name. */
function dxfUnitName(code: string): string {
  switch (code) {
    case '0': return 'Unitless';
    case '1': return 'Inches';
    case '2': return 'Feet';
    case '3': return 'Miles';
    case '4': return 'Millimeters';
    case '5': return 'Centimeters';
    case '6': return 'Meters';
    case '7': return 'Kilometers';
    default: return code;
  }
}

/**
 * Strip basic DXF MTEXT formatting codes.
 * MTEXT uses codes like \P (newline), \fArial; (font), {\H1.5; (height), etc.
 */
function cleanMTEXT(s: string): string {
  const parts: string[] = [];
  let i = 0;
  while (i < s.length) {
    const ch = s[i]!;
    if (ch === '\\' && i + 1 < s.length) {
      const next = s[i + 1]!;
      switch (next) {
        case 'P':
        case 'p':
          // Paragraph break.
          parts.push('\n');
          i += 2;
          break;
        case '\\':
          parts.push('\\');
          i += 2;
          break;
        case '{':
        case '}':
          parts.push(next);
          i += 2;
          break;
        default:
          // Skip formatting code until ; or end.
          i += 2;
          while (i < s.length && s[i] !== ';') {
            i++;
          }
          if (i < s.length) {
            i++; // skip ;
          }
          break;
      }
    } else if (ch === '{' || ch === '}') {
      // Grouping braces — skip.
      i++;
    } else {
      parts.push(ch);
      i++;
    }
  }
  return parts.join('');
}
