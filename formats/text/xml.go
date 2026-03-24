// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	// DefaultMaxXMLBytes is the default maximum bytes to read from an XML file.
	DefaultMaxXMLBytes = 10 * 1024 * 1024 // 10MB

	mimeApplicationXML = "application/xml"
	mimeTextXML        = "text/xml"
)

// XMLExtractor extracts text content from XML files by stripping tags.
// Safe for concurrent use.
type XMLExtractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxXMLBytes.
	MaxBytes int64
}

// NewXML creates a new XML extractor with default settings.
func NewXML() *XMLExtractor {
	return &XMLExtractor{}
}

func (e *XMLExtractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxXMLBytes
}

// Extract reads XML from r and returns the text content with tags stripped.
func (e *XMLExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if !e.supportsXML(mime) {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limited := io.LimitReader(r, e.maxBytes())

	// Read all data within limit for XML decoding.
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mime,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mime)

	if len(data) == 0 {
		return result, nil
	}

	text := extractXMLText(data)
	result.Text = metadata.CleanUTF8(text)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *XMLExtractor) SupportedMIMETypes() []string {
	return []string{mimeApplicationXML, mimeTextXML}
}

func (e *XMLExtractor) supportsXML(mime string) bool {
	return mime == mimeApplicationXML || mime == mimeTextXML
}

// extractXMLText parses XML and returns concatenated CharData text.
func extractXMLText(data []byte) string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	// Be lenient with encoding issues.
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose

	var b strings.Builder
	needSpace := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				if needSpace && b.Len() > 0 {
					b.WriteByte('\n')
				}
				b.WriteString(s)
				needSpace = true
			}
		}
	}

	return b.String()
}
