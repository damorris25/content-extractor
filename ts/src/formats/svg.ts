// CONTRACT:C1-EXTRACTOR.1.0
// SVG extractor — extracts text content and metadata from SVG files.

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import { createMetadata, metadataSet, KEY_FORMAT, KEY_TITLE, KEY_DESCRIPTION } from '../metadata.js';

const MIME_SVG = 'image/svg+xml';
const DEFAULT_MAX_BYTES = 10 * 1024 * 1024; // 10MB

/** SVG elements that contain user-visible text. */
const TEXT_ELEMENTS = new Set(['text', 'tspan', 'textpath']);

/** Extracts text content and metadata from SVG files. */
export class SvgExtractor implements Extractor {
  private readonly maxBytes: number;

  constructor(maxBytes: number = DEFAULT_MAX_BYTES) {
    this.maxBytes = maxBytes;
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (mime !== MIME_SVG) {
      throw new UnsupportedFormatError(`unsupported format: ${mime}`);
    }

    checkAborted(options?.signal);

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, MIME_SVG);

    if (data.length === 0) {
      return { text: '', metadata, contentType: MIME_SVG, language: '' };
    }

    if (data.length > this.maxBytes) {
      throw new FileTooLargeError();
    }

    checkAborted(options?.signal);

    const content = new TextDecoder('utf-8').decode(data);
    const info = extractSVGContent(content);

    if (info.title) {
      metadataSet(metadata, KEY_TITLE, info.title);
    }
    if (info.desc) {
      metadataSet(metadata, KEY_DESCRIPTION, info.desc);
    }

    return { text: info.text, metadata, contentType: MIME_SVG, language: '' };
  }

  supportedMIMETypes(): string[] {
    return [MIME_SVG];
  }
}

interface SVGInfo {
  text: string;
  title: string;
  desc: string;
}

/**
 * Extract text content and metadata from SVG using a state machine parser.
 * Extracts text from <text>, <tspan>, <textPath> elements.
 * Extracts <title> and <desc> from direct children of <svg>.
 */
function extractSVGContent(svg: string): SVGInfo {
  const textParts: string[] = [];
  let title = '';
  let desc = '';

  interface Frame {
    local: string;
    inText: boolean;
    inTitle: boolean;
    inDesc: boolean;
  }

  const MAX_STACK_DEPTH = 1000; // Prevent memory exhaustion from deeply nested SVG.
  const stack: Frame[] = [];
  let inText = false;
  let inTitle = false;
  let inDesc = false;
  let currentBuf = '';

  let i = 0;
  const len = svg.length;

  while (i < len) {
    // Skip processing instructions <?...?>
    if (svg[i] === '<' && i + 1 < len && svg[i + 1] === '?') {
      const end = svg.indexOf('?>', i + 2);
      if (end < 0) break;
      i = end + 2;
      continue;
    }

    // Skip comments <!--...-->
    if (svg[i] === '<' && svg.startsWith('!--', i + 1)) {
      const end = svg.indexOf('-->', i + 4);
      if (end < 0) break;
      i = end + 3;
      continue;
    }

    // Skip DOCTYPE
    if (svg[i] === '<' && svg.startsWith('!DOCTYPE', i + 1)) {
      const end = svg.indexOf('>', i + 9);
      if (end < 0) break;
      i = end + 1;
      continue;
    }

    // Handle tags
    if (svg[i] === '<') {
      const end = svg.indexOf('>', i + 1);
      if (end < 0) break;
      const tag = svg.substring(i + 1, end).trim();
      i = end + 1;

      const isClosing = tag.startsWith('/');
      const nameStr = isClosing ? tag.substring(1) : tag;
      const spaceIdx = nameStr.search(/[\s/]/);
      const tagName = (spaceIdx < 0 ? nameStr : nameStr.substring(0, spaceIdx)).toLowerCase();
      const isSelfClosing = tag.endsWith('/');

      if (isClosing || (!isSelfClosing && !isClosing)) {
        if (isClosing) {
          // Closing tag
          if (stack.length > 0) {
            const top = stack[stack.length - 1]!;
            if (top.local === tagName) {
              stack.pop();

              if (top.inText) {
                inText = false;
                for (let j = stack.length - 1; j >= 0; j--) {
                  if (stack[j]!.inText) {
                    inText = true;
                    break;
                  }
                }
              }
              if (top.inTitle) {
                inTitle = false;
                title = currentBuf.trim();
              }
              if (top.inDesc) {
                inDesc = false;
                desc = currentBuf.trim();
              }
            }
          }
        } else {
          // Opening tag
          const parentIsSVG = stack.length > 0 && stack[stack.length - 1]!.local === 'svg';
          const frame: Frame = {
            local: tagName,
            inText: TEXT_ELEMENTS.has(tagName),
            inTitle: tagName === 'title' && parentIsSVG,
            inDesc: tagName === 'desc' && parentIsSVG,
          };

          if (frame.inText) inText = true;
          if (frame.inTitle) {
            inTitle = true;
            currentBuf = '';
          }
          if (frame.inDesc) {
            inDesc = true;
            currentBuf = '';
          }

          if (stack.length < MAX_STACK_DEPTH) {
            stack.push(frame);
          }
        }
      }
      continue;
    }

    // Collect text content
    const tagStart = svg.indexOf('<', i);
    const textEnd = tagStart < 0 ? len : tagStart;
    const chunk = svg.substring(i, textEnd).trim();
    if (chunk) {
      const decoded = decodeXMLEntities(chunk);
      if (inTitle) {
        currentBuf += (currentBuf ? ' ' : '') + decoded;
      }
      if (inDesc) {
        currentBuf += (currentBuf ? ' ' : '') + decoded;
      }
      if (inText) {
        textParts.push(decoded);
      }
    }
    i = textEnd;
  }

  return { text: textParts.join('\n'), title, desc };
}

/** Decode standard XML entities. */
function decodeXMLEntities(s: string): string {
  return s
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'");
}
