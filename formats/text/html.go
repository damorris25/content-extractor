// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	// DefaultMaxHTMLBytes is the default maximum bytes to read from an HTML file.
	DefaultMaxHTMLBytes = 10 * 1024 * 1024 // 10MB

	mimeTextHTML = "text/html"
)

// skipElements contains elements whose text content should be skipped.
var skipElements = map[atom.Atom]bool{
	atom.Script:   true,
	atom.Style:    true,
	atom.Noscript: true,
}

// HTMLExtractor extracts visible text content from HTML files.
// Safe for concurrent use.
type HTMLExtractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxHTMLBytes.
	MaxBytes int64
}

// NewHTML creates a new HTML extractor with default settings.
func NewHTML() *HTMLExtractor {
	return &HTMLExtractor{}
}

func (e *HTMLExtractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxHTMLBytes
}

// Extract reads HTML from r and returns the visible text content.
func (e *HTMLExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeTextHTML {
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
		ContentType: mimeTextHTML,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeTextHTML)

	if len(data) == 0 {
		return result, nil
	}

	text, title := extractHTMLText(data)
	result.Text = metadata.CleanUTF8(text)
	if title != "" {
		result.Metadata.Set(metadata.KeyTitle, title)
	}

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *HTMLExtractor) SupportedMIMETypes() []string {
	return []string{mimeTextHTML}
}

// extractHTMLText parses HTML and returns visible text content and the title.
func extractHTMLText(data []byte) (text string, title string) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return "", ""
	}

	var b strings.Builder
	var titleBuf strings.Builder
	needSpace := false

	var walk func(*html.Node, bool, bool)
	walk = func(n *html.Node, inSkip bool, inTitle bool) {
		if n.Type == html.ElementNode {
			if skipElements[n.DataAtom] {
				inSkip = true
			}
			if n.DataAtom == atom.Title {
				inTitle = true
			}
		}

		if n.Type == html.TextNode {
			s := strings.TrimSpace(n.Data)
			if s != "" {
				if inTitle {
					if titleBuf.Len() > 0 {
						titleBuf.WriteByte(' ')
					}
					titleBuf.WriteString(s)
				}
				if !inSkip {
					if needSpace && b.Len() > 0 {
						b.WriteByte('\n')
					}
					b.WriteString(s)
					needSpace = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inSkip, inTitle)
		}
	}

	walk(doc, false, false)
	return b.String(), titleBuf.String()
}
