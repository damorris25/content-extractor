// CONTRACT:C1-EXTRACTOR.1.0
package office

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	mimeDocx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	// DefaultMaxDocxSize is the maximum uncompressed size for a DOCX file.
	DefaultMaxDocxSize = 100 * 1024 * 1024 // 100MB
)

// DocxExtractor extracts text content and metadata from DOCX files.
// Safe for concurrent use.
type DocxExtractor struct {
	// MaxSize limits the total bytes read from the input. Zero means DefaultMaxDocxSize.
	MaxSize int64
}

// NewDocx creates a new DOCX extractor with default settings.
func NewDocx() *DocxExtractor {
	return &DocxExtractor{}
}

func (e *DocxExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxDocxSize
}

// Extract reads a DOCX file from r and returns extracted text and metadata.
func (e *DocxExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeDocx {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ZIP requires random access, so we must buffer the input.
	// Use a size limit to prevent memory exhaustion.
	limited := io.LimitReader(r, e.maxSize()+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	if int64(len(data)) > e.maxSize() {
		return nil, extractor.ErrFileTooLarge
	}
	if len(data) == 0 {
		return &extractor.Result{ContentType: mimeDocx, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ZIP: %v", extractor.ErrCorruptInput, err)
	}

	result := &extractor.Result{
		ContentType: mimeDocx,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeDocx)

	// Extract text from word/document.xml.
	text, err := extractDocumentText(ctx, zr)
	if err != nil {
		return result, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	result.Text = metadata.CleanUTF8(text)

	// Extract metadata from docProps/core.xml.
	extractCoreProperties(zr, result)

	// Extract metadata from docProps/app.xml.
	extractAppProperties(zr, result)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *DocxExtractor) SupportedMIMETypes() []string {
	return []string{mimeDocx}
}

// extractDocumentText parses word/document.xml and extracts text from <w:t> elements.
// NOTE: Go's encoding/xml does not process external entities (XXE-safe by default).
func extractDocumentText(ctx context.Context, zr *zip.Reader) (string, error) {
	f := findZipFile(zr, "word/document.xml")
	if f == nil {
		return "", nil
	}

	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	// Limit decompressed size to prevent zip bombs.
	lr := io.LimitReader(rc, maxDecompressedEntrySize)

	var b strings.Builder
	decoder := xml.NewDecoder(lr)
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
			switch t.Name.Local {
			case "p":
				if t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
					if paragraphCount > 0 {
						b.WriteString("\n")
					}
					inParagraph = true
					paragraphCount++
				}
			}
		case xml.EndElement:
			if t.Name.Local == "p" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
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

// extractCoreProperties reads docProps/core.xml for Dublin Core metadata.
func extractCoreProperties(zr *zip.Reader, result *extractor.Result) {
	f := findZipFile(zr, "docProps/core.xml")
	if f == nil {
		return
	}
	rc, err := f.Open()
	if err != nil {
		return
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 64*1024)) // 64KB limit for metadata
	if err != nil {
		return
	}

	var core coreProperties
	if err := xml.Unmarshal(data, &core); err != nil {
		return
	}

	result.Metadata.Set(metadata.KeyTitle, strings.TrimSpace(core.Title))
	result.Metadata.Set(metadata.KeyCreator, strings.TrimSpace(core.Creator))
	result.Metadata.Set(metadata.KeySubject, strings.TrimSpace(core.Subject))
	result.Metadata.Set(metadata.KeyDescription, strings.TrimSpace(core.Description))
	if core.Created != "" {
		result.Metadata.Set(metadata.KeyDate, metadata.NormalizeDate(core.Created))
	}
	if core.Modified != "" {
		result.Metadata.Set(metadata.KeyModified, metadata.NormalizeDate(core.Modified))
	}
	if core.Language != "" {
		result.Metadata.Set(metadata.KeyLanguage, strings.TrimSpace(core.Language))
		result.Language = strings.TrimSpace(core.Language)
	}
}

// extractAppProperties reads docProps/app.xml for Office-specific metadata.
func extractAppProperties(zr *zip.Reader, result *extractor.Result) {
	f := findZipFile(zr, "docProps/app.xml")
	if f == nil {
		return
	}
	rc, err := f.Open()
	if err != nil {
		return
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 64*1024))
	if err != nil {
		return
	}

	var app appProperties
	if err := xml.Unmarshal(data, &app); err != nil {
		return
	}

	result.Metadata.Set(metadata.KeyApplication, strings.TrimSpace(app.Application))
	result.Metadata.Set(metadata.KeyPageCount, strings.TrimSpace(app.Pages))
	result.Metadata.Set(metadata.KeyWordCount, strings.TrimSpace(app.Words))
	result.Metadata.Set(metadata.KeyCharCount, strings.TrimSpace(app.Characters))
}

// coreProperties models docProps/core.xml.
type coreProperties struct {
	XMLName     xml.Name `xml:"coreProperties"`
	Title       string   `xml:"title"`
	Creator     string   `xml:"creator"`
	Subject     string   `xml:"subject"`
	Description string   `xml:"description"`
	Created     string   `xml:"created"`
	Modified    string   `xml:"modified"`
	Language    string   `xml:"language"`
}

// appProperties models docProps/app.xml.
type appProperties struct {
	XMLName     xml.Name `xml:"Properties"`
	Application string   `xml:"Application"`
	Pages       string   `xml:"Pages"`
	Words       string   `xml:"Words"`
	Characters  string   `xml:"Characters"`
}

// findZipFile finds a file by name in a ZIP archive.
func findZipFile(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}
