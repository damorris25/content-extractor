// CONTRACT:C1-EXTRACTOR.1.0
package email

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxBytes is the default maximum bytes to read from an EML file.
	DefaultMaxBytes = 50 * 1024 * 1024 // 50MB

	mimeEML = "message/rfc822"

	// maxMultipartDepth prevents stack overflow from deeply nested multipart.
	maxMultipartDepth = 10

	// maxParts limits iteration per multipart level.
	maxParts = 100
)

// skipElements contains HTML elements whose text content should be skipped.
var skipElements = map[atom.Atom]bool{
	atom.Script:   true,
	atom.Style:    true,
	atom.Noscript: true,
}

// Extractor extracts text content and metadata from EML (RFC 2822) email files.
// Safe for concurrent use.
type Extractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxBytes.
	MaxBytes int64
}

// New creates a new EML extractor with default settings.
func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxBytes
}

// Extract reads an EML file from r and returns extracted text and metadata.
func (e *Extractor) Extract(ctx context.Context, r io.Reader, mimeType string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mimeType != mimeEML {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limited := io.LimitReader(r, e.maxBytes())

	msg, err := mail.ReadMessage(limited)
	if err != nil {
		// Empty input — return empty result (compliance requirement).
		if err == io.EOF {
			return &extractor.Result{
				ContentType: mimeEML,
				Metadata:    make(metadata.Metadata),
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mimeEML,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeEML)

	// Extract metadata from headers.
	if subject := msg.Header.Get("Subject"); subject != "" {
		result.Metadata.Set(metadata.KeyTitle, subject)
	}
	if from := msg.Header.Get("From"); from != "" {
		result.Metadata.Set(metadata.KeyCreator, from)
	}
	if to := msg.Header.Get("To"); to != "" {
		result.Metadata.Set(metadata.KeyEmailTo, to)
	}
	if cc := msg.Header.Get("Cc"); cc != "" {
		result.Metadata.Set(metadata.KeyEmailCc, cc)
	}
	if date := msg.Header.Get("Date"); date != "" {
		if t, err := mail.ParseDate(date); err == nil {
			result.Metadata.Set(metadata.KeyDate, t.UTC().Format("2006-01-02T15:04:05Z"))
		}
	}

	// Extract body text.
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	body, err := extractBody(msg.Body, contentType)
	if err != nil {
		// Best-effort: return what we have without body text.
		return result, nil
	}

	result.Text = metadata.CleanUTF8(strings.TrimSpace(body))
	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	return []string{mimeEML}
}

// extractBody extracts text from the message body, handling multipart messages.
func extractBody(r io.Reader, contentType string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Fall back to reading as plain text.
		data, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Multipart message: prefer text/plain parts over text/html.
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return "", fmt.Errorf("multipart message without boundary")
		}
		return extractMultipart(r, boundary, 0)
	}

	// Single-part message.
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	text := string(data)
	if strings.HasPrefix(mediaType, "text/html") {
		text = stripHTMLTagsProper(data)
	}
	return text, nil
}

// extractMultipart reads multipart content, preferring text/plain parts.
// depth limits recursion to prevent stack overflow from nested multipart.
func extractMultipart(r io.Reader, boundary string, depth int) (string, error) {
	if depth >= maxMultipartDepth {
		return "", nil
	}

	mr := multipart.NewReader(r, boundary)

	var plainParts []string
	var htmlParts []string

	for i := 0; i < maxParts; i++ {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		partType := part.Header.Get("Content-Type")
		if partType == "" {
			partType = "text/plain"
		}

		partMedia, partParams, _ := mime.ParseMediaType(partType)

		// Handle nested multipart.
		if strings.HasPrefix(partMedia, "multipart/") {
			nested, err := extractMultipart(part, partParams["boundary"], depth+1)
			if err == nil && nested != "" {
				plainParts = append(plainParts, nested)
			}
			part.Close()
			continue
		}

		if !strings.HasPrefix(partMedia, "text/") {
			part.Close()
			continue
		}

		data, err := io.ReadAll(part)
		part.Close()
		if err != nil {
			continue
		}

		if strings.HasPrefix(partMedia, "text/plain") {
			plainParts = append(plainParts, string(data))
		} else if strings.HasPrefix(partMedia, "text/html") {
			htmlParts = append(htmlParts, string(data))
		}
	}

	// Prefer plain text over HTML.
	if len(plainParts) > 0 {
		return strings.Join(plainParts, "\n"), nil
	}
	if len(htmlParts) > 0 {
		var stripped []string
		for _, h := range htmlParts {
			stripped = append(stripped, stripHTMLTagsProper([]byte(h)))
		}
		return strings.Join(stripped, "\n"), nil
	}

	return "", nil
}

// stripHTMLTagsProper uses golang.org/x/net/html to properly parse and extract
// visible text from HTML content. This is safer than a naive tag stripper
// because it correctly handles malformed HTML, script/style content, and entities.
func stripHTMLTagsProper(data []byte) string {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return string(data) // fallback to raw text
	}

	var b strings.Builder
	needSpace := false

	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inSkip bool) {
		if n.Type == html.ElementNode {
			if skipElements[n.DataAtom] {
				inSkip = true
			}
		}

		if n.Type == html.TextNode && !inSkip {
			s := strings.TrimSpace(n.Data)
			if s != "" {
				if needSpace && b.Len() > 0 {
					b.WriteByte('\n')
				}
				b.WriteString(s)
				needSpace = true
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inSkip)
		}
	}

	walk(doc, false)
	return b.String()
}
