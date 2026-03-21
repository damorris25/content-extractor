// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const mimeODT = "application/vnd.oasis.opendocument.text"

// ODTExtractor extracts text content and metadata from ODT files.
// Safe for concurrent use.
type ODTExtractor struct {
	MaxSize int64
}

// NewODT creates a new ODT extractor with default settings.
func NewODT() *ODTExtractor {
	return &ODTExtractor{}
}

func (e *ODTExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxODFSize
}

// Extract reads an ODT file from r and returns extracted text and metadata.
func (e *ODTExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeODT {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, _, emptyResult, err := openODF(r, e.maxSize(), mimeODT)
	if err != nil {
		return nil, err
	}
	if emptyResult != nil {
		return emptyResult, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mimeODT,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeODT)

	// Extract text from content.xml.
	f := findZipFile(zr, "content.xml")
	if f == nil {
		extractODFMeta(zr, result)
		return result, nil
	}

	rc, err := f.Open()
	if err != nil {
		return result, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	defer rc.Close()

	text, err := extractODTText(ctx, io.LimitReader(rc, maxDecompressedEntrySize))
	if err != nil {
		result.Text = metadata.CleanUTF8(text)
		extractODFMeta(zr, result)
		return result, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	result.Text = metadata.CleanUTF8(text)

	extractODFMeta(zr, result)
	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *ODTExtractor) SupportedMIMETypes() []string {
	return []string{mimeODT}
}

// extractODTText parses content.xml for text:p elements.
func extractODTText(ctx context.Context, r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var b strings.Builder
	inParagraph := false
	paragraphCount := 0

	for {
		if err := ctx.Err(); err != nil {
			return b.String(), err
		}

		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return b.String(), err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == textNS && (t.Name.Local == "p" || t.Name.Local == "h") {
				if paragraphCount > 0 {
					b.WriteString("\n")
				}
				inParagraph = true
				paragraphCount++
			}
			// Handle text:tab and text:line-break.
			if t.Name.Space == textNS {
				switch t.Name.Local {
				case "tab":
					b.WriteString("\t")
				case "line-break":
					b.WriteString("\n")
				}
			}
		case xml.EndElement:
			if t.Name.Space == textNS && (t.Name.Local == "p" || t.Name.Local == "h") {
				inParagraph = false
			}
		case xml.CharData:
			if inParagraph {
				b.Write(t)
			}
		}
	}

	return b.String(), nil
}
