// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"io"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxBytes is the default maximum number of bytes to read from a text file.
	DefaultMaxBytes = 10 * 1024 * 1024 // 10MB

	mimeTextPlain = "text/plain"
)

// Extractor extracts text content from plain text and text-like files.
// Safe for concurrent use.
type Extractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxBytes.
	MaxBytes  int64
	mimeTypes map[string]bool
}

// New creates a new plain text extractor for text/plain.
func New() *Extractor {
	return &Extractor{mimeTypes: map[string]bool{mimeTextPlain: true}}
}

// NewTextLike creates a text extractor for additional text-like MIME types
// (JSON, CSV, YAML, Markdown, etc.) that are returned as-is without transformation.
func NewTextLike(mimes ...string) *Extractor {
	m := make(map[string]bool, len(mimes))
	for _, mime := range mimes {
		m[mime] = true
	}
	return &Extractor{mimeTypes: m}
}

func (e *Extractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxBytes
}

// Extract reads plain text from r and returns it as a Result.
func (e *Extractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if !e.supports(mime) {
		return nil, extractor.ErrUnsupportedFormat
	}

	// Check context before starting work.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Stream-read with a limit.
	limited := io.LimitReader(r, e.maxBytes())

	var b strings.Builder
	buf := make([]byte, 32*1024) // 32KB read buffer
	for {
		// Check context periodically.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		n, readErr := limited.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}

	text := metadata.CleanUTF8(b.String())

	result := &extractor.Result{
		Text:        text,
		ContentType: mime,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mime)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	if e.mimeTypes == nil {
		return []string{mimeTextPlain}
	}
	types := make([]string, 0, len(e.mimeTypes))
	for m := range e.mimeTypes {
		types = append(types, m)
	}
	return types
}

func (e *Extractor) supports(mime string) bool {
	if e.mimeTypes == nil {
		return mime == mimeTextPlain
	}
	return e.mimeTypes[mime]
}
