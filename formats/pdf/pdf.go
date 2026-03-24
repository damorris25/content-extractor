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

	// maxPages caps page iteration to prevent DoS from PDFs claiming millions of pages.
	maxPages = 10000
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

	result, err := e.extractPDF(ctx, data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// extractPDF performs the actual PDF parsing with panic recovery.
// The ledongthuc/pdf library may panic on malformed input.
func (e *Extractor) extractPDF(ctx context.Context, data []byte) (result *extractor.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("%w: PDF parser panic: %v", extractor.ErrCorruptInput, r)
		}
	}()

	reader := bytes.NewReader(data)
	pdfReader, pdfErr := lpdf.NewReader(reader, int64(len(data)))
	if pdfErr != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, pdfErr)
	}

	result = &extractor.Result{
		ContentType: mimePDF,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimePDF)

	// Extract text from all pages.
	var b strings.Builder
	numPages := pdfReader.NumPage()
	if numPages > maxPages {
		numPages = maxPages
	}

	if numPages > 0 {
		result.Metadata.Set(metadata.KeyPageCount, fmt.Sprintf("%d", numPages))
	}

	for i := 1; i <= numPages; i++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			result.Text = metadata.CleanUTF8(b.String())
			return result, ctxErr
		}

		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, textErr := page.GetPlainText(nil)
		if textErr != nil {
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
