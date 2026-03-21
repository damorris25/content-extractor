// CONTRACT:C1-EXTRACTOR.1.0
package office

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	mimePptx = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// DefaultMaxPptxSize is the maximum uncompressed size for a PPTX file.
	DefaultMaxPptxSize = 100 * 1024 * 1024 // 100MB
)

// PptxExtractor extracts text content and metadata from PPTX files.
// Safe for concurrent use.
type PptxExtractor struct {
	// MaxSize limits the total bytes read from the input. Zero means DefaultMaxPptxSize.
	MaxSize int64
}

// NewPptx creates a new PPTX extractor with default settings.
func NewPptx() *PptxExtractor {
	return &PptxExtractor{}
}

func (e *PptxExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxPptxSize
}

// Extract reads a PPTX file from r and returns extracted text and metadata.
func (e *PptxExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimePptx {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limited := io.LimitReader(r, e.maxSize()+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	if int64(len(data)) > e.maxSize() {
		return nil, extractor.ErrFileTooLarge
	}
	if len(data) == 0 {
		return &extractor.Result{ContentType: mimePptx, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ZIP: %v", extractor.ErrCorruptInput, err)
	}

	result := &extractor.Result{
		ContentType: mimePptx,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimePptx)

	// Extract text from all slides.
	text, err := extractSlideText(ctx, zr)
	if err != nil {
		return result, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	result.Text = metadata.CleanUTF8(text)

	// Extract metadata.
	extractCoreProperties(zr, result)
	extractAppProperties(zr, result)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *PptxExtractor) SupportedMIMETypes() []string {
	return []string{mimePptx}
}

// extractSlideText reads all slide XML files and extracts text from <a:t> elements.
func extractSlideText(ctx context.Context, zr *zip.Reader) (string, error) {
	// Collect slide files sorted by name for deterministic order.
	var slideFiles []*zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideFiles = append(slideFiles, f)
			if len(slideFiles) >= maxZIPEntries {
				break
			}
		}
	}
	sort.Slice(slideFiles, func(i, j int) bool {
		return slideFiles[i].Name < slideFiles[j].Name
	})

	var result strings.Builder
	for i, sf := range slideFiles {
		if err := ctx.Err(); err != nil {
			return result.String(), err
		}
		if i > 0 {
			result.WriteString("\n\n") // Separate slides with blank line.
		}
		text, err := extractSingleSlide(ctx, sf)
		if err != nil {
			return result.String(), err
		}
		result.WriteString(text)
	}

	return result.String(), nil
}

// extractSingleSlide extracts text from a single slide XML.
// Text lives in <a:t> elements within <a:p> paragraphs (DrawingML).
func extractSingleSlide(ctx context.Context, f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	lr := io.LimitReader(rc, maxDecompressedEntrySize)
	const drawingMLNS = "http://schemas.openxmlformats.org/drawingml/2006/main"

	decoder := xml.NewDecoder(lr)
	var result strings.Builder
	inParagraph := false
	inText := false
	paragraphCount := 0

	for {
		if err := ctx.Err(); err != nil {
			return result.String(), err
		}

		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result.String(), err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == drawingMLNS {
				switch t.Name.Local {
				case "p":
					if paragraphCount > 0 {
						result.WriteString("\n")
					}
					inParagraph = true
					paragraphCount++
				case "t":
					if inParagraph {
						inText = true
					}
				}
			}
		case xml.EndElement:
			if t.Name.Space == drawingMLNS {
				switch t.Name.Local {
				case "p":
					inParagraph = false
				case "t":
					inText = false
				}
			}
		case xml.CharData:
			if inText {
				result.Write(t)
			}
		}
	}

	return result.String(), nil
}
