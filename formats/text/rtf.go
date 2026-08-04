// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxRTFBytes is the default maximum bytes to read from an RTF file.
	DefaultMaxRTFBytes = 10 * 1024 * 1024 // 10MB

	mimeRTF = "application/rtf"
)

// RTFExtractor extracts text content from RTF files by stripping markup.
// Safe for concurrent use.
type RTFExtractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxRTFBytes.
	MaxBytes int64
}

// NewRTF creates a new RTF extractor with default settings.
func NewRTF() *RTFExtractor {
	return &RTFExtractor{}
}

func (e *RTFExtractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxRTFBytes
}

// Extract reads RTF from r and returns plain text with markup stripped.
func (e *RTFExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeRTF {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limited := io.LimitReader(r, e.maxBytes())
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mimeRTF,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeRTF)

	if len(data) == 0 {
		return result, nil
	}

	text := extractRTFText(data)
	result.Text = metadata.CleanUTF8(text)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *RTFExtractor) SupportedMIMETypes() []string {
	return []string{mimeRTF}
}

// extractRTFText strips RTF markup and returns plain text.
// Handles: group nesting {}, control words \word, Unicode \uN,
// hex escapes \'XX, destination groups {\*\...}.
func extractRTFText(data []byte) string {
	var b strings.Builder
	i := 0
	n := len(data)
	groupDepth := 0
	skipDepth := 0 // depth at which we started skipping a destination group

	for i < n {
		ch := data[i]

		switch ch {
		case '{':
			groupDepth++
			// Check for destination group: {\*\...}
			if i+3 < n && data[i+1] == '\\' && data[i+2] == '*' {
				skipDepth = groupDepth
			}
			i++

		case '}':
			if groupDepth == skipDepth {
				skipDepth = 0
			}
			if groupDepth > 0 {
				groupDepth--
			}
			i++

		case '\\':
			if skipDepth > 0 {
				// Inside a skip-destination, just advance past the control word.
				i = skipControlWord(data, i)
				continue
			}
			i++ // skip backslash
			if i >= n {
				break
			}

			nextCh := data[i]

			// Unicode escape: \uN
			if nextCh == 'u' && i+1 < n && (data[i+1] >= '0' || data[i+1] == '-') {
				i++ // skip 'u'
				neg := false
				if i < n && data[i] == '-' {
					neg = true
					i++
				}
				val := 0
				for i < n && data[i] >= '0' && data[i] <= '9' {
					val = val*10 + int(data[i]-'0')
					i++
				}
				if neg {
					val = -val
				}
				// RTF uses signed 16-bit; convert to unsigned.
				if val < 0 {
					val += 65536
				}
				r := rune(val)
				if utf8.ValidRune(r) {
					b.WriteRune(r)
				}
				// Skip one subsequent character (the "replacement" char).
				if i < n && data[i] == ' ' {
					i++
				} else if i < n && data[i] != '\\' && data[i] != '{' && data[i] != '}' {
					i++
				}
				continue
			}

			// Hex escape: \'XX
			if nextCh == '\'' && i+2 < n {
				i++ // skip apostrophe
				hi := hexVal(data[i])
				lo := hexVal(data[i+1])
				if hi >= 0 && lo >= 0 {
					b.WriteByte(byte(hi<<4 | lo))
				}
				i += 2
				continue
			}

			// Special character escapes.
			if nextCh == '\\' || nextCh == '{' || nextCh == '}' {
				b.WriteByte(nextCh)
				i++
				continue
			}

			// Control word: \word[N][ ]
			if nextCh >= 'a' && nextCh <= 'z' || nextCh >= 'A' && nextCh <= 'Z' {
				word, param, end := parseControlWord(data, i)
				i = end

				switch word {
				case "par", "pard":
					b.WriteByte('\n')
				case "line":
					b.WriteByte('\n')
				case "tab":
					b.WriteByte('\t')
				case "lquote", "rquote":
					b.WriteByte('\'')
				case "ldblquote", "rdblquote":
					b.WriteByte('"')
				case "emdash":
					b.WriteString("—")
				case "endash":
					b.WriteString("–")
				case "bullet":
					b.WriteString("•")
				case "u":
					// Handled above, but just in case.
					_ = param
				}
				continue
			}

			// Skip other control symbols (e.g., \~, \-, etc.)
			i++

		default:
			if skipDepth > 0 {
				i++
				continue
			}
			// Plain text character.
			b.WriteByte(ch)
			i++
		}
	}

	return cleanRTFText(b.String())
}

// skipControlWord advances past a control word starting at position i (which points to \).
func skipControlWord(data []byte, i int) int {
	i++ // skip backslash
	n := len(data)
	if i >= n {
		return i
	}
	// Skip the word characters.
	for i < n && ((data[i] >= 'a' && data[i] <= 'z') || (data[i] >= 'A' && data[i] <= 'Z')) {
		i++
	}
	// Skip optional numeric parameter.
	if i < n && (data[i] == '-' || (data[i] >= '0' && data[i] <= '9')) {
		if data[i] == '-' {
			i++
		}
		for i < n && data[i] >= '0' && data[i] <= '9' {
			i++
		}
	}
	// Skip optional trailing space.
	if i < n && data[i] == ' ' {
		i++
	}
	return i
}

// parseControlWord parses a control word starting at position i.
// Returns the word, optional numeric parameter, and the position after.
func parseControlWord(data []byte, i int) (word string, param int, end int) {
	n := len(data)
	start := i
	for i < n && ((data[i] >= 'a' && data[i] <= 'z') || (data[i] >= 'A' && data[i] <= 'Z')) {
		i++
	}
	word = string(data[start:i])

	// Optional numeric parameter.
	if i < n && (data[i] == '-' || (data[i] >= '0' && data[i] <= '9')) {
		neg := false
		if data[i] == '-' {
			neg = true
			i++
		}
		val := 0
		for i < n && data[i] >= '0' && data[i] <= '9' {
			val = val*10 + int(data[i]-'0')
			i++
		}
		if neg {
			val = -val
		}
		param = val
	}

	// Skip optional trailing space delimiter.
	if i < n && data[i] == ' ' {
		i++
	}

	return word, param, i
}

// hexVal returns the value of a hex digit, or -1 if invalid.
func hexVal(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	default:
		return -1
	}
}

// cleanRTFText normalizes whitespace in extracted RTF text.
func cleanRTFText(s string) string {
	// Collapse multiple consecutive newlines to max 2.
	lines := strings.Split(s, "\n")
	var result []string
	blankCount := 0
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if trimmed == "" {
			blankCount++
			if blankCount <= 1 {
				result = append(result, "")
			}
		} else {
			blankCount = 0
			result = append(result, trimmed)
		}
	}

	text := strings.Join(result, "\n")
	text = strings.TrimSpace(text)
	return text
}

// Ensure RTFExtractor implements extractor.Extractor.
var _ extractor.Extractor = (*RTFExtractor)(nil)

// Ensure SVGExtractor implements extractor.Extractor.
var _ extractor.Extractor = (*SVGExtractor)(nil)

// Ensure HTMLExtractor implements extractor.Extractor.
var _ extractor.Extractor = (*HTMLExtractor)(nil)

// Ensure XMLExtractor implements extractor.Extractor.
var _ extractor.Extractor = (*XMLExtractor)(nil)

// Ensure Extractor (plain text) implements extractor.Extractor.
var _ extractor.Extractor = (*Extractor)(nil)

