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
	"strconv"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	mimeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	// DefaultMaxXlsxSize is the maximum uncompressed size for an XLSX file.
	DefaultMaxXlsxSize = 100 * 1024 * 1024 // 100MB
)

// XlsxExtractor extracts text content and metadata from XLSX files.
// Safe for concurrent use.
type XlsxExtractor struct {
	// MaxSize limits the total bytes read from the input. Zero means DefaultMaxXlsxSize.
	MaxSize int64
}

// NewXlsx creates a new XLSX extractor with default settings.
func NewXlsx() *XlsxExtractor {
	return &XlsxExtractor{}
}

func (e *XlsxExtractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxXlsxSize
}

// Extract reads an XLSX file from r and returns extracted text and metadata.
func (e *XlsxExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeXlsx {
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
		return &extractor.Result{ContentType: mimeXlsx, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ZIP: %v", extractor.ErrCorruptInput, err)
	}

	result := &extractor.Result{
		ContentType: mimeXlsx,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeXlsx)

	// Load shared strings table.
	sharedStrings, err := loadSharedStrings(ctx, zr)
	if err != nil {
		return result, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}

	// Extract text from all worksheets.
	text, err := extractSheetText(ctx, zr, sharedStrings)
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
func (e *XlsxExtractor) SupportedMIMETypes() []string {
	return []string{mimeXlsx}
}

// loadSharedStrings parses xl/sharedStrings.xml into a string slice indexed by position.
func loadSharedStrings(ctx context.Context, zr *zip.Reader) ([]string, error) {
	f := findZipFile(zr, "xl/sharedStrings.xml")
	if f == nil {
		return nil, nil // No shared strings is valid (e.g., all inline or numeric values).
	}

	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	lr := io.LimitReader(rc, maxDecompressedEntrySize)

	var strings []string
	decoder := xml.NewDecoder(lr)
	var current stringBuilder
	inSI := false
	inT := false

	// Cap shared string count to prevent memory exhaustion from crafted files.
	const maxSharedStrings = 1_000_000

	for {
		if err := ctx.Err(); err != nil {
			return strings, err
		}

		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				inSI = true
				current.Reset()
			case "t":
				if inSI {
					inT = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "si":
				inSI = false
				strings = append(strings, current.String())
				if len(strings) >= maxSharedStrings {
					return strings, nil
				}
			case "t":
				inT = false
			}
		case xml.CharData:
			if inT {
				current.Write(t)
			}
		}
	}

	return strings, nil
}

// stringBuilder is a minimal wrapper to collect text across rich text runs.
type stringBuilder struct {
	b builderImpl
}

type builderImpl = strings.Builder

func (s *stringBuilder) Reset()           { s.b.Reset() }
func (s *stringBuilder) Write(p []byte)   { s.b.Write(p) }
func (s *stringBuilder) String() string   { return s.b.String() }

// extractSheetText reads all worksheets and extracts cell text.
func extractSheetText(ctx context.Context, zr *zip.Reader, sharedStrings []string) (string, error) {
	// Find all sheet files, sorted by name for deterministic order.
	// Limit entry count to prevent resource exhaustion.
	var sheetFiles []*zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			sheetFiles = append(sheetFiles, f)
			if len(sheetFiles) >= maxZIPEntries {
				break
			}
		}
	}
	sort.Slice(sheetFiles, func(i, j int) bool {
		return sheetFiles[i].Name < sheetFiles[j].Name
	})

	var result strings.Builder
	for i, sf := range sheetFiles {
		if err := ctx.Err(); err != nil {
			return result.String(), err
		}
		if i > 0 {
			result.WriteString("\n\n") // Separate sheets with blank line.
		}
		text, err := extractSingleSheet(ctx, sf, sharedStrings)
		if err != nil {
			return result.String(), err
		}
		result.WriteString(text)
	}

	return result.String(), nil
}

// extractSingleSheet extracts text from a single worksheet XML file.
func extractSingleSheet(ctx context.Context, f *zip.File, sharedStrings []string) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	lr := io.LimitReader(rc, maxDecompressedEntrySize)
	decoder := xml.NewDecoder(lr)
	var result strings.Builder
	var cellValue strings.Builder
	var cellType string
	inRow := false
	inV := false
	inInlineT := false
	firstRow := true
	cellsInRow := 0

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
			switch t.Name.Local {
			case "row":
				inRow = true
				if !firstRow {
					result.WriteString("\n")
				}
				firstRow = false
				cellsInRow = 0
			case "c":
				if inRow {
					cellType = ""
					for _, attr := range t.Attr {
						if attr.Name.Local == "t" {
							cellType = attr.Value
						}
					}
					cellValue.Reset()
				}
			case "v":
				inV = true
			case "t":
				// Inline string value.
				if inRow {
					inInlineT = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "row":
				inRow = false
			case "c":
				if inRow {
					val := cellValue.String()
					if cellType == "s" && len(sharedStrings) > 0 {
						// Shared string reference.
						idx, err := strconv.Atoi(val)
						if err == nil && idx >= 0 && idx < len(sharedStrings) {
							val = sharedStrings[idx]
						}
					}
					if val != "" {
						if cellsInRow > 0 {
							result.WriteString("\t")
						}
						result.WriteString(val)
						cellsInRow++
					}
				}
			case "v":
				inV = false
			case "t":
				inInlineT = false
			}
		case xml.CharData:
			if inV || inInlineT {
				cellValue.Write(t)
			}
		}
	}

	return result.String(), nil
}
