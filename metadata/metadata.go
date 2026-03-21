// CONTRACT:I1-METADATA.1.0
package metadata

import "strings"

// Metadata provides typed access to extracted metadata key-value pairs.
// Keys follow a namespaced convention: "namespace:field" (e.g., "dc:title").
// All values are string slices to support multi-valued fields.
type Metadata map[string][]string

// Standard namespace prefixes.
const (
	NSDublinCore = "dc"        // Dublin Core: title, creator, subject, description, date
	NSOffice     = "office"    // Office-specific: revision, template, application
	NSEXIF       = "exif"      // EXIF image metadata: camera, GPS, dimensions
	NSAudio      = "audio"     // Audio metadata: artist, album, track, duration
	NSCAD        = "cad"       // CAD metadata: layers, units, version
	NSExtractor  = "extractor" // Extraction process: duration, extractor-version, warnings
)

// Well-known metadata keys.
const (
	KeyTitle       = "dc:title"
	KeyCreator     = "dc:creator"
	KeySubject     = "dc:subject"
	KeyDescription = "dc:description"
	KeyDate        = "dc:date"     // ISO 8601
	KeyModified    = "dc:modified" // ISO 8601
	KeyLanguage    = "dc:language" // BCP 47
	KeyFormat      = "dc:format"   // MIME type
	KeyPageCount   = "office:page-count"
	KeyWordCount   = "office:word-count"
	KeyCharCount   = "office:char-count"
	KeyApplication = "office:application"
	KeyRevision    = "office:revision"
)

// Get returns the first value for the given key, or empty string if not present.
// Safe to call on a nil Metadata.
func (m Metadata) Get(key string) string {
	if m == nil {
		return ""
	}
	vals := m[key]
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// GetAll returns all values for the given key, or nil if not present.
// Safe to call on a nil Metadata.
func (m Metadata) GetAll(key string) []string {
	if m == nil {
		return nil
	}
	return m[key]
}

// Set sets the key to a single value, replacing any existing values.
// Empty values are omitted per contract. NUL bytes are stripped.
func (m Metadata) Set(key, value string) {
	value = stripNUL(value)
	if value == "" {
		delete(m, key)
		return
	}
	m[key] = []string{value}
}

// Add appends a value to the key. Does not overwrite existing values.
// Empty values are omitted per contract. NUL bytes are stripped.
func (m Metadata) Add(key, value string) {
	value = stripNUL(value)
	if value == "" {
		return
	}
	m[key] = append(m[key], value)
}

// stripNUL removes NUL bytes from a string. NUL bytes in metadata values
// are never legitimate and can cause issues with downstream consumers.
func stripNUL(s string) string {
	if strings.IndexByte(s, 0) < 0 {
		return s
	}
	return strings.ReplaceAll(s, "\x00", "")
}

// Has returns true if the key exists and has at least one non-empty value.
// Safe to call on a nil Metadata.
func (m Metadata) Has(key string) bool {
	if m == nil {
		return false
	}
	return len(m[key]) > 0
}
