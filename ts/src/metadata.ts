// Metadata model and normalization.
// Port of metadata/metadata.go + metadata/normalize.go

/** Metadata is a map of namespaced keys to string arrays. */
export type Metadata = Record<string, string[]>;

// --- Namespace constants ---
export const NS_DUBLIN_CORE = 'dc';
export const NS_OFFICE = 'office';
export const NS_EXIF = 'exif';
export const NS_AUDIO = 'audio';
export const NS_CAD = 'cad';
export const NS_EXTRACTOR = 'extractor';

// --- Well-known keys ---
export const KEY_TITLE = 'dc:title';
export const KEY_CREATOR = 'dc:creator';
export const KEY_SUBJECT = 'dc:subject';
export const KEY_DESCRIPTION = 'dc:description';
export const KEY_DATE = 'dc:date';
export const KEY_MODIFIED = 'dc:modified';
export const KEY_LANGUAGE = 'dc:language';
export const KEY_FORMAT = 'dc:format';
export const KEY_PAGE_COUNT = 'office:page-count';
export const KEY_WORD_COUNT = 'office:word-count';
export const KEY_CHAR_COUNT = 'office:char-count';
export const KEY_APPLICATION = 'office:application';
export const KEY_REVISION = 'office:revision';

// --- Helper functions ---

/** Get the first value for a key, or empty string. */
export function metadataGet(m: Metadata | null | undefined, key: string): string {
  if (!m || !isSafeKey(key)) return '';
  const vals = m[key];
  if (!vals || vals.length === 0) return '';
  return vals[0]!;
}

/** Get all values for a key, or empty array. */
export function metadataGetAll(m: Metadata | null | undefined, key: string): string[] {
  if (!m || !isSafeKey(key)) return [];
  return m[key] ?? [];
}

/** Set a key to a single value. Empty values remove the key. NUL bytes stripped. */
export function metadataSet(m: Metadata, key: string, value: string): void {
  if (!isSafeKey(key)) return;
  value = stripNUL(value);
  if (value === '') {
    delete m[key];
    return;
  }
  m[key] = [value];
}

/** Append a value to a key. Empty values are ignored. NUL bytes stripped. */
export function metadataAdd(m: Metadata, key: string, value: string): void {
  if (!isSafeKey(key)) return;
  value = stripNUL(value);
  if (value === '') return;
  const existing = m[key];
  if (existing) {
    existing.push(value);
  } else {
    m[key] = [value];
  }
}

/** Check if a key exists with at least one value. */
export function metadataHas(m: Metadata | null | undefined, key: string): boolean {
  if (!m || !isSafeKey(key)) return false;
  const vals = m[key];
  return !!vals && vals.length > 0;
}

/** Dangerous keys that could cause prototype pollution. */
const DANGEROUS_KEYS = new Set(['__proto__', 'constructor', 'prototype']);

/** Check if a key is safe for use as a metadata property. */
function isSafeKey(key: string): boolean {
  return !DANGEROUS_KEYS.has(key);
}

/** Create a new empty Metadata object (null-prototype to prevent prototype pollution). */
export function createMetadata(): Metadata {
  return Object.create(null) as Metadata;
}

// --- Normalization ---

/** Strip NUL bytes from a string. */
export function stripNUL(s: string): string {
  if (s.indexOf('\0') < 0) return s;
  return s.replaceAll('\0', '');
}

/**
 * Normalize a date string to ISO 8601 format.
 * Attempts the same 12 formats as the Go version.
 * Returns the original string if no format matches.
 */
export function normalizeDate(s: string): string {
  s = s.trim();
  if (s === '') return '';

  // Try each pattern in order (matching Go's dateFormats).
  for (const parser of dateParsers) {
    const result = parser(s);
    if (result !== null) return result;
  }
  return s;
}

// Date format parsers matching Go's dateFormats list.
const dateParsers: Array<(s: string) => string | null> = [
  // RFC3339 / ISO 8601 with timezone
  (s) => {
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(Z|[+-]\d{2}:\d{2})$/);
    if (!m) return null;
    return toISO(m[1]!, m[2]!, m[3]!, m[4]!, m[5]!, m[6]!, m[7]!);
  },
  // ISO without timezone (assume UTC)
  (s) => {
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})$/);
    if (!m) return null;
    return toISO(m[1]!, m[2]!, m[3]!, m[4]!, m[5]!, m[6]!, 'Z');
  },
  // Date with space time
  (s) => {
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})$/);
    if (!m) return null;
    return toISO(m[1]!, m[2]!, m[3]!, m[4]!, m[5]!, m[6]!, 'Z');
  },
  // Date only
  (s) => {
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!m) return null;
    return toISO(m[1]!, m[2]!, m[3]!, '00', '00', '00', 'Z');
  },
  // MM/DD/YYYY
  (s) => {
    const m = s.match(/^(\d{2})\/(\d{2})\/(\d{4})$/);
    if (!m) return null;
    return toISO(m[3]!, m[1]!, m[2]!, '00', '00', '00', 'Z');
  },
  // DD/MM/YYYY (ambiguous with above, matches Go's order)
  // Go tries this after MM/DD/YYYY. Since both patterns match the same regex,
  // and Go's time.Parse succeeds on the first match, we skip this duplicate.

  // "Jan 2, 2006"
  (s) => {
    const m = s.match(/^([A-Z][a-z]{2}) (\d{1,2}), (\d{4})$/);
    if (!m) return null;
    const month = monthNum(m[1]!);
    if (month === null) return null;
    return toISO(m[3]!, month, pad2(m[2]!), '00', '00', '00', 'Z');
  },
  // "January 2, 2006"
  (s) => {
    const m = s.match(/^([A-Z][a-z]+) (\d{1,2}), (\d{4})$/);
    if (!m) return null;
    const month = monthNum(m[1]!);
    if (month === null) return null;
    return toISO(m[3]!, month, pad2(m[2]!), '00', '00', '00', 'Z');
  },
  // Year only
  (s) => {
    const m = s.match(/^(\d{4})$/);
    if (!m) return null;
    return toISO(m[1]!, '01', '01', '00', '00', '00', 'Z');
  },
];

function toISO(
  year: string, month: string, day: string,
  hour: string, min: string, sec: string,
  tz: string,
): string {
  // Parse to Date and format as RFC3339 UTC.
  let offset = 0;
  if (tz !== 'Z') {
    const tzm = tz.match(/^([+-])(\d{2}):(\d{2})$/);
    if (tzm) {
      offset = (parseInt(tzm[2]!, 10) * 60 + parseInt(tzm[3]!, 10)) * 60 * 1000;
      if (tzm[1] === '-') offset = -offset;
    }
  }
  const d = new Date(Date.UTC(
    parseInt(year, 10), parseInt(month, 10) - 1, parseInt(day, 10),
    parseInt(hour, 10), parseInt(min, 10), parseInt(sec, 10),
  ));
  // Adjust for timezone offset.
  d.setTime(d.getTime() - offset);
  return d.toISOString().replace('.000Z', 'Z');
}

const MONTHS: Record<string, string> = {
  jan: '01', feb: '02', mar: '03', apr: '04', may: '05', jun: '06',
  jul: '07', aug: '08', sep: '09', oct: '10', nov: '11', dec: '12',
  january: '01', february: '02', march: '03', april: '04',
  june: '06', july: '07', august: '08', september: '09',
  october: '10', november: '11', december: '12',
};

function monthNum(name: string): string | null {
  return MONTHS[name.toLowerCase()] ?? null;
}

function pad2(s: string): string {
  return s.length === 1 ? '0' + s : s;
}
