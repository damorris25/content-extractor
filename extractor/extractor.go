// CONTRACT:C1-EXTRACTOR.1.0
package extractor

import (
	"context"
	"io"

	"github.com/damorris25/content-extractor/metadata"
)

// Extractor extracts text content and metadata from a specific file format.
// Implementations MUST be safe for concurrent use.
type Extractor interface {
	// Extract reads from r and returns extracted content and metadata.
	// The caller provides a context for cancellation and a MIME type hint.
	// Implementations MUST NOT buffer the entire input in memory.
	Extract(ctx context.Context, r io.Reader, mime string) (*Result, error)

	// SupportedMIMETypes returns the MIME types this extractor handles.
	SupportedMIMETypes() []string
}

// Result holds extracted content and metadata from a file.
type Result struct {
	// Text is the extracted textual content, UTF-8 encoded.
	Text string

	// Metadata contains key-value pairs extracted from the file.
	// Keys use a standardized namespace (see CONTRACT:I1-METADATA).
	Metadata metadata.Metadata

	// ContentType is the detected MIME type of the input.
	ContentType string

	// Language is the detected or declared language (BCP 47), if available.
	Language string
}
