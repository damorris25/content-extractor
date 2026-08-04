// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	mimeODP   = "application/vnd.oasis.opendocument.presentation"
	officeNS  = "urn:oasis:names:tc:opendocument:xmlns:office:1.0"
	presentNS = "urn:oasis:names:tc:opendocument:xmlns:presentation:1.0"
)

// ODPExtractor extracts text content and metadata from ODP files.
// Safe for concurrent use.
type ODPExtractor struct {
	MaxSize int64
}

// NewODP creates a new ODP extractor with default settings.
func NewODP() *ODPExtractor {
	return &ODPExtractor{}
}

func (e *ODPExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxODFSize
}

// Extract reads an ODP file from r and returns extracted text and metadata.
func (e *ODPExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeODP {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, _, emptyResult, err := openODF(r, e.maxSize(), mimeODP)
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
		ContentType: mimeODP,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeODP)

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

	text, err := extractODPText(ctx, io.LimitReader(rc, maxDecompressedEntrySize))
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
func (e *ODPExtractor) SupportedMIMETypes() []string {
	return []string{mimeODP}
}

// extractODPText parses content.xml for presentation page text.
// Structure: draw:page > draw:frame > draw:text-box > text:p
// Text is in text:p elements nested within draw elements on each page.
func extractODPText(ctx context.Context, r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var b strings.Builder
	inPage := false
	inTextP := false
	pageCount := 0
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
			switch {
			case t.Name.Space == drawNS && t.Name.Local == "page":
				if pageCount > 0 {
					b.WriteString("\n\n")
				}
				inPage = true
				pageCount++
				paragraphCount = 0
			case t.Name.Space == presentNS && t.Name.Local == "notes":
				// Skip presenter notes — only extract slide content.
				decoder.Skip()
			case t.Name.Space == textNS && (t.Name.Local == "p" || t.Name.Local == "h") && inPage:
				if paragraphCount > 0 {
					b.WriteString("\n")
				}
				inTextP = true
				paragraphCount++
			case t.Name.Space == textNS && t.Name.Local == "tab":
				b.WriteString("\t")
			case t.Name.Space == textNS && t.Name.Local == "line-break":
				b.WriteString("\n")
			}
		case xml.EndElement:
			switch {
			case t.Name.Space == drawNS && t.Name.Local == "page":
				inPage = false
			case t.Name.Space == textNS && (t.Name.Local == "p" || t.Name.Local == "h"):
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
