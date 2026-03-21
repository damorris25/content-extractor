// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	// DefaultMaxBytes is the default maximum number of bytes to read from a text file.
	DefaultMaxBytes = 10 * 1024 * 1024 // 10MB

	mimeTextPlain = "text/plain"
)

// Extractor extracts text content from plain text files.
// Safe for concurrent use.
type Extractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxBytes.
	MaxBytes int64
}

// New creates a new plain text extractor with default settings.
func New() *Extractor {
	return &Extractor{}
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
		ContentType: mimeTextPlain,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeTextPlain)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	return []string{mimeTextPlain}
}

func (e *Extractor) supports(mime string) bool {
	return mime == mimeTextPlain
}
