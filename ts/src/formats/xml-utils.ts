// XML parsing utility — cross-environment DOMParser wrapper.
// Provides helpers for namespace-aware element traversal.

let _domParser: DOMParser | null = null;
let _domParserPromise: Promise<DOMParser> | null = null;

async function resolveDOMParser(): Promise<DOMParser> {
  if (_domParser) return _domParser;
  if (typeof DOMParser !== 'undefined') {
    _domParser = new DOMParser();
    return _domParser;
  }
  // Node.js fallback — try @xmldom/xmldom.
  try {
    const mod = await import('@xmldom/xmldom');
    _domParser = new mod.DOMParser() as unknown as DOMParser;
    return _domParser;
  } catch {
    throw new Error(
      'DOMParser not available. Install @xmldom/xmldom for Node.js support.',
    );
  }
}

function getDOMParser(): Promise<DOMParser> {
  if (_domParser) return Promise.resolve(_domParser);
  if (!_domParserPromise) {
    _domParserPromise = resolveDOMParser();
  }
  return _domParserPromise;
}

/** Parse an XML string into a Document. */
export async function parseXML(xmlString: string): Promise<Document> {
  // Strip BOM.
  xmlString = xmlString.replace(/^\uFEFF/, '');
  // Strip <?xml ...?> declarations that appear inside element content.
  // @xmldom/xmldom throws fatalError on these, but Go's encoding/xml treats
  // them as processing instructions (ignored). Browser DOMParser also ignores them.
  // We strip them to match Go's behavior and xmldom compatibility.
  xmlString = stripInnerXMLDeclarations(xmlString);
  const parser = await getDOMParser();
  return parser.parseFromString(xmlString, 'application/xml');
}

/** Remove <?xml ...?> PIs that appear after the document prolog. */
function stripInnerXMLDeclarations(s: string): string {
  // Preserve the first XML declaration (at position 0 after trimming).
  const trimmed = s.trimStart();
  if (!trimmed.startsWith('<?xml')) {
    // No XML declaration at all — just remove any occurrences.
    return s.replaceAll(/<\?xml[^?]*\?>/g, '');
  }
  // Find end of first declaration.
  const firstEnd = s.indexOf('?>');
  if (firstEnd < 0) return s;
  const after = firstEnd + 2;
  // Remove any subsequent XML declarations.
  return s.substring(0, after) + s.substring(after).replace(/<\?xml[^?]*\?>/g, '');
}

/** Get text content from all matching elements by namespace and local name. */
export function getTextByTagNS(
  parent: Document | Element,
  ns: string,
  localName: string,
): string[] {
  const elements = parent.getElementsByTagNameNS(ns, localName);
  const results: string[] = [];
  for (let i = 0; i < elements.length; i++) {
    const el = elements[i]!;
    const text = textContent(el);
    if (text) results.push(text);
  }
  return results;
}

/** Get first element by namespace and local name. */
export function getFirstElementNS(
  parent: Document | Element,
  ns: string,
  localName: string,
): Element | null {
  const elements = parent.getElementsByTagNameNS(ns, localName);
  return elements.length > 0 ? elements[0]! : null;
}

/**
 * Recursively collect all text content from an element.
 * Concatenates all text nodes, trimming each.
 */
export function textContent(el: Element): string {
  return el.textContent ?? '';
}

/**
 * Get the trimmed text content of the first element matching ns:localName.
 * Returns empty string if not found.
 */
export function getFirstTextNS(
  parent: Document | Element,
  ns: string,
  localName: string,
): string {
  const el = getFirstElementNS(parent, ns, localName);
  if (!el) return '';
  return (el.textContent ?? '').trim();
}
