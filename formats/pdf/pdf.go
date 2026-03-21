// CONTRACT:C1-EXTRACTOR.1.0
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	lpdf "github.com/ledongthuc/pdf"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	mimePDF = "application/pdf"

	// DefaultMaxPDFSize is the maximum size for a PDF file.
	DefaultMaxPDFSize = 100 * 1024 * 1024 // 100MB
)

// Extractor extracts text content and metadata from PDF files.
// Uses text-layer extraction only (no OCR).
// Safe for concurrent use.
type Extractor struct {
	// MaxSize limits the total bytes read from the input. Zero means DefaultMaxPDFSize.
	MaxSize int64
}

// New creates a new PDF extractor with default settings.
func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxPDFSize
}

// Extract reads a PDF file from r and returns extracted text and metadata.
func (e *Extractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimePDF {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// PDF requires random access, so buffer the input.
	limited := io.LimitReader(r, e.maxSize()+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	if int64(len(data)) > e.maxSize() {
		return nil, extractor.ErrFileTooLarge
	}
	if len(data) == 0 {
		return &extractor.Result{ContentType: mimePDF, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	pdfReader, err := lpdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}

	result := &extractor.Result{
		ContentType: mimePDF,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimePDF)

	// Extract text from all pages.
	var b strings.Builder
	numPages := pdfReader.NumPage()

	if numPages > 0 {
		result.Metadata.Set(metadata.KeyPageCount, fmt.Sprintf("%d", numPages))
	}

	for i := 1; i <= numPages; i++ {
		if err := ctx.Err(); err != nil {
			result.Text = metadata.CleanUTF8(b.String())
			return result, err
		}

		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // Skip pages that fail to parse.
		}

		if i > 1 {
			b.WriteString("\n")
		}
		b.WriteString(text)
	}

	result.Text = metadata.CleanUTF8(b.String())

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	return []string{mimePDF}
}
