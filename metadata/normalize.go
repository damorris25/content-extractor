// CONTRACT:I1-METADATA.1.0
package metadata

import (
	"strings"
	"time"
	"unicode/utf8"
)

// Common date formats encountered in document metadata.
var dateFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"01/02/2006",
	"02/01/2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"2006",
}

// NormalizeDate attempts to parse a date string in various common formats
// and returns it in ISO 8601 format. Returns the original string if parsing fails.
func NormalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, fmt := range dateFormats {
		if t, err := time.Parse(fmt, s); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return s
}

// CleanUTF8 replaces invalid UTF-8 bytes with the Unicode replacement character U+FFFD.
func CleanUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		b.WriteRune(r)
		i += size
	}
	return b.String()
}
