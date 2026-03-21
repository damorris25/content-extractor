// Audio metadata extractor.
// Port of formats/audio/audio.go + id3.go + vorbis.go

import type { Extractor, ExtractionResult, ExtractOptions } from '../extractor.js';
import { UnsupportedFormatError, FileTooLargeError, checkAborted } from '../extractor.js';
import {
  createMetadata, metadataSet, metadataGet,
  KEY_FORMAT, KEY_TITLE, KEY_CREATOR, KEY_DESCRIPTION, KEY_DATE,
  normalizeDate,
} from '../metadata.js';

const DEFAULT_MAX_AUDIO_SIZE = 100 * 1024 * 1024; // 100MB

const SUPPORTED_MIMES = new Set([
  'audio/mpeg',
  'audio/ogg',
  'audio/flac',
  'audio/wav',
]);

export class AudioExtractor implements Extractor {
  maxSize: number;

  constructor(maxSize?: number) {
    this.maxSize = maxSize ?? DEFAULT_MAX_AUDIO_SIZE;
  }

  supportedMIMETypes(): string[] {
    return [...SUPPORTED_MIMES];
  }

  async extract(
    data: Uint8Array,
    mime: string,
    options?: ExtractOptions,
  ): Promise<ExtractionResult> {
    if (!SUPPORTED_MIMES.has(mime)) {
      throw new UnsupportedFormatError();
    }
    checkAborted(options?.signal);

    if (data.byteLength > this.maxSize) {
      throw new FileTooLargeError();
    }

    const metadata = createMetadata();
    metadataSet(metadata, KEY_FORMAT, mime);

    if (data.byteLength === 0) {
      return { text: '', metadata, contentType: mime, language: '' };
    }

    checkAborted(options?.signal);

    switch (mime) {
      case 'audio/mpeg':
        parseID3v2(data, metadata);
        break;
      case 'audio/ogg':
        parseOggVorbis(data, metadata);
        break;
      case 'audio/flac':
        parseFLAC(data, metadata);
        break;
      case 'audio/wav':
        parseWAV(data, metadata);
        break;
    }

    const text = buildTextSummary(metadata);

    return { text, metadata, contentType: mime, language: '' };
  }
}

// ---------------------------------------------------------------------------
// Metadata mapping
// ---------------------------------------------------------------------------

type Metadata = Record<string, string[]>;

function setAudioMeta(m: Metadata, key: string, value: string): void {
  value = value.trim();
  if (value === '') return;
  switch (key.toUpperCase()) {
    case 'TITLE':
    case 'TIT2':
      metadataSet(m, KEY_TITLE, value);
      break;
    case 'ARTIST':
    case 'TPE1':
      metadataSet(m, KEY_CREATOR, value);
      break;
    case 'ALBUM':
    case 'TALB':
      metadataSet(m, 'audio:album', value);
      break;
    case 'TRACKNUMBER':
    case 'TRCK':
      metadataSet(m, 'audio:track', value);
      break;
    case 'DATE':
    case 'YEAR':
    case 'TDRC':
    case 'TYER':
      metadataSet(m, KEY_DATE, normalizeDate(value));
      break;
    case 'GENRE':
    case 'TCON':
      metadataSet(m, 'audio:genre', value);
      break;
    case 'COMMENT':
    case 'COMM':
      metadataSet(m, KEY_DESCRIPTION, value);
      break;
    case 'ALBUMARTIST':
    case 'TPE2':
      metadataSet(m, 'audio:album-artist', value);
      break;
    default:
      metadataSet(m, 'audio:' + key.toLowerCase(), value);
      break;
  }
}

function buildTextSummary(m: Metadata): string {
  const keys = [KEY_TITLE, KEY_CREATOR, 'audio:album', 'audio:genre'];
  const parts: string[] = [];
  for (const key of keys) {
    const v = metadataGet(m, key);
    if (v) parts.push(v);
  }
  return parts.join('\n');
}

// ---------------------------------------------------------------------------
// Binary helpers
// ---------------------------------------------------------------------------

function readUint32BE(view: DataView, offset: number): number {
  if (offset < 0 || offset + 4 > view.byteLength) return 0;
  return view.getUint32(offset, false);
}

function readUint32LE(view: DataView, offset: number): number {
  if (offset < 0 || offset + 4 > view.byteLength) return 0;
  return view.getUint32(offset, true);
}

function readUint16LE(view: DataView, offset: number): number {
  if (offset < 0 || offset + 2 > view.byteLength) return 0;
  return view.getUint16(offset, true);
}

function readSyncsafe(data: Uint8Array, offset: number): number {
  if (offset + 3 >= data.byteLength) return 0;
  // Mask to 7 bits per byte (syncsafe encoding requires MSB=0).
  return (
    ((data[offset]! & 0x7F) << 21) |
    ((data[offset + 1]! & 0x7F) << 14) |
    ((data[offset + 2]! & 0x7F) << 7) |
    (data[offset + 3]! & 0x7F)
  );
}

// ---------------------------------------------------------------------------
// ID3v2 parsing (MP3)
// ---------------------------------------------------------------------------

function parseID3v2(data: Uint8Array, metadata: Metadata): void {
  if (data.byteLength < 10) return;
  // Check "ID3" header.
  if (data[0] !== 0x49 || data[1] !== 0x44 || data[2] !== 0x33) return;

  const majorVersion = data[3]!;
  // const flags = data[5]!;
  const tagSize = readSyncsafe(data, 6);

  const end = Math.min(10 + tagSize, data.byteLength);
  const pos = 10;

  switch (majorVersion) {
    case 3:
    case 4:
      parseID3v2Frames(data, pos, end, majorVersion, metadata);
      break;
    case 2:
      parseID3v22Frames(data, pos, end, metadata);
      break;
  }
}

/** Parse ID3v2.3/v2.4 frames. Frame: ID(4) + size(4) + flags(2) + data. */
function parseID3v2Frames(
  data: Uint8Array,
  start: number,
  end: number,
  version: number,
  metadata: Metadata,
): void {
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let pos = start;

  while (pos + 10 <= end) {
    const frameID = String.fromCharCode(
      data[pos]!, data[pos + 1]!, data[pos + 2]!, data[pos + 3]!,
    );
    if (frameID.charCodeAt(0) === 0) break; // Padding.

    let frameSize: number;
    if (version === 4) {
      frameSize = readSyncsafe(data, pos + 4);
    } else {
      frameSize = readUint32BE(view, pos + 4);
    }

    pos += 10; // Skip header (ID + size + flags).

    if (frameSize === 0 || frameSize > end - pos) break;

    const frameData = data.subarray(pos, pos + frameSize);
    const value = decodeTextFrame(frameData);
    setAudioMeta(metadata, frameID, value);

    pos += frameSize;
  }
}

/** Parse ID3v2.2 frames. Frame: ID(3) + size(3) + data. */
function parseID3v22Frames(
  data: Uint8Array,
  start: number,
  end: number,
  metadata: Metadata,
): void {
  let pos = start;

  while (pos + 6 <= end) {
    const frameID = String.fromCharCode(
      data[pos]!, data[pos + 1]!, data[pos + 2]!,
    );
    if (frameID.charCodeAt(0) === 0) break;

    const frameSize = (data[pos + 3]! << 16) | (data[pos + 4]! << 8) | data[pos + 5]!;
    pos += 6;

    if (frameSize === 0 || pos + frameSize > end) break;

    const frameData = data.subarray(pos, pos + frameSize);
    const mapped = mapV22Frame(frameID);
    const value = decodeTextFrame(frameData);
    setAudioMeta(metadata, mapped, value);

    pos += frameSize;
  }
}

/** Map ID3v2.2 3-char frame IDs to v2.3+ 4-char equivalents. */
function mapV22Frame(id: string): string {
  switch (id) {
    case 'TT2': return 'TIT2';
    case 'TP1': return 'TPE1';
    case 'TAL': return 'TALB';
    case 'TRK': return 'TRCK';
    case 'TYE': return 'TYER';
    case 'TCO': return 'TCON';
    case 'COM': return 'COMM';
    default: return id;
  }
}

/** Decode an ID3v2 text frame. First byte is encoding. */
function decodeTextFrame(data: Uint8Array): string {
  if (data.byteLength < 2) return '';

  const encoding = data[0]!;
  const payload = data.subarray(1);

  switch (encoding) {
    case 0: // ISO-8859-1
      return cleanString(decodeLatin1(payload));
    case 3: // UTF-8
      return cleanString(utf8Decoder.decode(payload));
    case 1: // UTF-16 with BOM
      return decodeUTF16WithBOM(payload);
    case 2: // UTF-16BE without BOM
      return decodeUTF16BE(payload);
    default:
      return cleanString(decodeLatin1(payload));
  }
}

const utf8Decoder = new TextDecoder('utf-8');
const utf16leDecoder = new TextDecoder('utf-16le');
const utf16beDecoder = new TextDecoder('utf-16be');

function decodeLatin1(data: Uint8Array): string {
  // Decode in chunks to avoid O(n²) string concatenation.
  const CHUNK = 8192;
  const parts: string[] = [];
  for (let i = 0; i < data.byteLength; i += CHUNK) {
    const end = Math.min(i + CHUNK, data.byteLength);
    const codes: number[] = [];
    for (let j = i; j < end; j++) {
      codes.push(data[j]!);
    }
    parts.push(String.fromCharCode(...codes));
  }
  return parts.join('');
}

function decodeUTF16WithBOM(data: Uint8Array): string {
  if (data.byteLength < 2) return '';
  let payload: Uint8Array;
  let decoder: TextDecoder;
  if (data[0] === 0xFF && data[1] === 0xFE) {
    decoder = utf16leDecoder;
    payload = data.subarray(2);
  } else if (data[0] === 0xFE && data[1] === 0xFF) {
    decoder = utf16beDecoder;
    payload = data.subarray(2);
  } else {
    decoder = utf16beDecoder;
    payload = data;
  }
  const s = decoder.decode(payload);
  // Trim at first NUL.
  const nulIdx = s.indexOf('\0');
  return nulIdx >= 0 ? s.substring(0, nulIdx) : s;
}

function decodeUTF16BE(data: Uint8Array): string {
  const s = utf16beDecoder.decode(data);
  const nulIdx = s.indexOf('\0');
  return nulIdx >= 0 ? s.substring(0, nulIdx) : s;
}

/** Remove NUL terminators and trim whitespace. */
function cleanString(s: string): string {
  const nulIdx = s.indexOf('\0');
  if (nulIdx >= 0) s = s.substring(0, nulIdx);
  return s.trim();
}

// ---------------------------------------------------------------------------
// Vorbis comment parsing (OGG / FLAC)
// ---------------------------------------------------------------------------

function parseOggVorbis(data: Uint8Array, metadata: Metadata): void {
  // Scan for \x03vorbis marker in first 64KB.
  const marker = [0x03, 0x76, 0x6F, 0x72, 0x62, 0x69, 0x73]; // \x03vorbis
  const searchLimit = Math.min(data.byteLength, 65536);
  const idx = findBytes(data, marker, searchLimit);
  if (idx < 0) return;

  parseVorbisComment(data, idx + marker.length, metadata);
}

function parseFLAC(data: Uint8Array, metadata: Metadata): void {
  if (data.byteLength < 4) return;
  // Check "fLaC" magic.
  if (data[0] !== 0x66 || data[1] !== 0x4C || data[2] !== 0x61 || data[3] !== 0x43) return;

  let pos = 4;
  while (pos + 4 <= data.byteLength) {
    const header = data[pos]!;
    const isLast = (header & 0x80) !== 0;
    const blockType = header & 0x7F;
    const blockSize = (data[pos + 1]! << 16) | (data[pos + 2]! << 8) | data[pos + 3]!;
    pos += 4;

    if (pos + blockSize > data.byteLength) break;

    if (blockType === 4) { // Vorbis comment
      parseVorbisComment(data, pos, metadata);
      return;
    }

    pos += blockSize;
    if (isLast) break;
  }
}

function parseVorbisComment(
  data: Uint8Array,
  startOffset: number,
  metadata: Metadata,
): void {
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  if (startOffset + 8 > data.byteLength) return;

  // Skip vendor string.
  const vendorLen = readUint32LE(view, startOffset);
  if (vendorLen > data.byteLength - startOffset - 4) return;
  let pos = startOffset + 4 + vendorLen;
  if (pos + 4 > data.byteLength) return;

  let commentCount = readUint32LE(view, pos);
  pos += 4;

  const maxComments = 65536;
  if (commentCount > maxComments) commentCount = maxComments;

  for (let i = 0; i < commentCount && pos + 4 <= data.byteLength; i++) {
    const commentLen = readUint32LE(view, pos);
    pos += 4;

    if (commentLen > data.byteLength - pos) break;

    const comment = utf8Decoder.decode(data.subarray(pos, pos + commentLen));
    pos += commentLen;

    const eqIdx = comment.indexOf('=');
    if (eqIdx > 0) {
      const key = comment.substring(0, eqIdx);
      const value = comment.substring(eqIdx + 1);
      setAudioMeta(metadata, key, value);
    }
  }
}

// ---------------------------------------------------------------------------
// WAV RIFF parsing
// ---------------------------------------------------------------------------

function parseWAV(data: Uint8Array, metadata: Metadata): void {
  if (data.byteLength < 12) return;
  // Check RIFF header.
  if (data[0] !== 0x52 || data[1] !== 0x49 || data[2] !== 0x46 || data[3] !== 0x46) return;
  // Check WAVE.
  if (data[8] !== 0x57 || data[9] !== 0x41 || data[10] !== 0x56 || data[11] !== 0x45) return;

  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let pos = 12;

  while (pos + 8 <= data.byteLength) {
    const chunkID = String.fromCharCode(
      data[pos]!, data[pos + 1]!, data[pos + 2]!, data[pos + 3]!,
    );
    const chunkSize = readUint32LE(view, pos + 4);
    pos += 8;

    if (chunkSize > data.byteLength - pos) break;

    switch (chunkID) {
      case 'fmt ':
        if (chunkSize >= 16) {
          const channels = readUint16LE(view, pos + 2);
          const sampleRate = readUint32LE(view, pos + 4);
          const bitsPerSample = readUint16LE(view, pos + 14);
          metadataSet(metadata, 'audio:channels', String(channels));
          metadataSet(metadata, 'audio:sample-rate', String(sampleRate));
          metadataSet(metadata, 'audio:bits-per-sample', String(bitsPerSample));
        }
        break;
      case 'LIST':
        if (chunkSize >= 4) {
          const subType = String.fromCharCode(
            data[pos]!, data[pos + 1]!, data[pos + 2]!, data[pos + 3]!,
          );
          if (subType === 'INFO') {
            parseRIFFInfo(data, view, pos + 4, pos + chunkSize, metadata);
          }
        }
        break;
    }

    pos += chunkSize;
    if (chunkSize % 2 === 1) pos++; // Word-aligned.
  }
}

function parseRIFFInfo(
  data: Uint8Array,
  view: DataView,
  start: number,
  end: number,
  metadata: Metadata,
): void {
  let pos = start;

  while (pos + 8 <= end) {
    const tag = String.fromCharCode(
      data[pos]!, data[pos + 1]!, data[pos + 2]!, data[pos + 3]!,
    );
    const size = readUint32LE(view, pos + 4);
    pos += 8;

    if (size > end - pos) break;

    const value = cleanString(utf8Decoder.decode(data.subarray(pos, pos + size)));

    switch (tag) {
      case 'INAM': setAudioMeta(metadata, 'TITLE', value); break;
      case 'IART': setAudioMeta(metadata, 'ARTIST', value); break;
      case 'IPRD': setAudioMeta(metadata, 'ALBUM', value); break;
      case 'IGNR': setAudioMeta(metadata, 'GENRE', value); break;
      case 'ICRD': setAudioMeta(metadata, 'DATE', value); break;
      case 'ICMT': setAudioMeta(metadata, 'COMMENT', value); break;
    }

    pos += size;
    if (size % 2 === 1) pos++; // Word-aligned.
  }
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

function findBytes(data: Uint8Array, needle: number[], limit: number): number {
  const searchEnd = Math.min(data.byteLength, limit) - needle.length;
  outer:
  for (let i = 0; i <= searchEnd; i++) {
    for (let j = 0; j < needle.length; j++) {
      if (data[i + j] !== needle[j]) continue outer;
    }
    return i;
  }
  return -1;
}
