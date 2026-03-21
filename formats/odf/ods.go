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

const mimeODS = "application/vnd.oasis.opendocument.spreadsheet"

// ODSExtractor extracts text content and metadata from ODS files.
// Safe for concurrent use.
type ODSExtractor struct {
	MaxSize int64
}

// NewODS creates a new ODS extractor with default settings.
func NewODS() *ODSExtractor {
	return &ODSExtractor{}
}

func (e *ODSExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxODFSize
}

// Extract reads an ODS file from r and returns extracted text and metadata.
func (e *ODSExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeODS {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, _, emptyResult, err := openODF(r, e.maxSize(), mimeODS)
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
		ContentType: mimeODS,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeODS)

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

	text, err := extractODSText(ctx, rc)
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
func (e *ODSExtractor) SupportedMIMETypes() []string {
	return []string{mimeODS}
}

// extractODSText parses content.xml for table data.
// Structure: table:table > table:table-row > table:table-cell > text:p
func extractODSText(ctx context.Context, r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var b strings.Builder
	inCell := false
	inTextP := false
	rowCount := 0
	cellCount := 0
	tableCount := 0

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
			switch {
			case t.Name.Space == tableNS && t.Name.Local == "table":
				if tableCount > 0 {
					b.WriteString("\n\n")
				}
				tableCount++
				rowCount = 0
			case t.Name.Space == tableNS && t.Name.Local == "table-row":
				if rowCount > 0 {
					b.WriteString("\n")
				}
				rowCount++
				cellCount = 0
			case t.Name.Space == tableNS && t.Name.Local == "table-cell":
				inCell = true
			case t.Name.Space == textNS && t.Name.Local == "p" && inCell:
				if cellCount > 0 {
					b.WriteString("\t")
				}
				inTextP = true
			}
		case xml.EndElement:
			switch {
			case t.Name.Space == tableNS && t.Name.Local == "table-cell":
				if inCell {
					cellCount++
				}
				inCell = false
				inTextP = false
			case t.Name.Space == textNS && t.Name.Local == "p":
				inTextP = false
			}
		case xml.CharData:
			if inTextP {
				b.Write(t)
			}
		}
	}

	return b.String(), nil
}
