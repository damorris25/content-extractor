// CONTRACT:C1-EXTRACTOR.1.0
package cad

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	// DefaultMaxDXFSize is the maximum size for DXF files.
	DefaultMaxDXFSize = 200 * 1024 * 1024 // 200MB
)

// Extractor extracts text content and metadata from DXF files.
// Safe for concurrent use.
type Extractor struct {
	MaxSize int64
}

// New creates a new DXF extractor.
func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxDXFSize
}

// Extract reads a DXF file and returns extracted text and metadata.
func (e *Extractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != "image/vnd.dxf" {
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
		return &extractor.Result{ContentType: mime, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mime,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mime)

	parseDXF(data, result)

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	return []string{"image/vnd.dxf"}
}

// parseDXF parses a DXF file extracting text content and metadata.
// DXF files consist of group code/value pairs organized into sections.
func parseDXF(data []byte, result *extractor.Result) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Allow up to 1MB lines
	var texts []string
	var layers []string
	layerSeen := make(map[string]bool)
	const maxTexts = 100000 // Cap text entities to prevent memory exhaustion.
	const maxLayers = 10000

	// Track current parsing state.
	var section string
	var entityType string
	var headerVar string // current $VARIABLE name in HEADER

	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		if !scanner.Scan() {
			break
		}
		value := strings.TrimSpace(scanner.Text())

		// Track sections.
		if code == "0" && value == "SECTION" {
			continue
		}
		if code == "2" && section == "" {
			section = value
			continue
		}
		if code == "0" && value == "ENDSEC" {
			section = ""
			entityType = ""
			headerVar = ""
			continue
		}

		switch section {
		case "HEADER":
			if code == "9" {
				headerVar = value
			} else if headerVar != "" {
				setHeaderMeta(headerVar, value, result)
				headerVar = ""
			}
		case "TABLES":
			if code == "0" {
				entityType = value
			}
			// Collect layer names from the TABLES section.
			if entityType == "LAYER" && code == "2" && value != "" && len(layers) < maxLayers {
				if !layerSeen[value] {
					layerSeen[value] = true
					layers = append(layers, value)
				}
			}
		case "ENTITIES":
			if code == "0" {
				entityType = value
			}
			if len(texts) >= maxTexts {
				continue
			}
			switch entityType {
			case "TEXT", "ATTRIB", "ATTDEF":
				if code == "1" && value != "" {
					texts = append(texts, value)
				}
			case "MTEXT":
				if (code == "1" || code == "3") && value != "" {
					texts = append(texts, cleanMTEXT(value))
				}
			case "INSERT":
				if code == "2" && value != "" {
					texts = append(texts, value)
				}
			case "DIMENSION":
				if code == "1" && value != "" && value != "<>" {
					texts = append(texts, value)
				}
			}
		}
	}

	// scanner.Err() is intentionally not checked — a scan error on malformed
	// input should not prevent returning whatever text was already collected.

	if len(layers) > 0 {
		result.Metadata.Set("cad:layers", strings.Join(layers, ", "))
	}
	result.Text = strings.Join(texts, "\n")
}

// parseDXFHeader extracts metadata from header variable pairs.
// Called during the main parse with accumulated variable state.
func setHeaderMeta(varName, value string, result *extractor.Result) {
	switch varName {
	case "$ACADVER":
		result.Metadata.Set("cad:version", value)
	case "$INSUNITS":
		result.Metadata.Set("cad:units", dxfUnitName(value))
	case "$TDCREATE":
		// Julian date — just store raw for now.
		result.Metadata.Set(metadata.KeyDate, value)
	}
}

// dxfUnitName converts a DXF INSUNITS value to a human-readable name.
func dxfUnitName(code string) string {
	switch code {
	case "0":
		return "Unitless"
	case "1":
		return "Inches"
	case "2":
		return "Feet"
	case "3":
		return "Miles"
	case "4":
		return "Millimeters"
	case "5":
		return "Centimeters"
	case "6":
		return "Meters"
	case "7":
		return "Kilometers"
	default:
		return code
	}
}

// cleanMTEXT strips basic DXF MTEXT formatting codes.
// MTEXT uses codes like \P (newline), \fArial; (font), {\H1.5; (height), etc.
func cleanMTEXT(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case 'P', 'p':
				// Paragraph break.
				b.WriteByte('\n')
				i += 2
			case '\\':
				b.WriteByte('\\')
				i += 2
			case '{', '}':
				b.WriteByte(next)
				i += 2
			default:
				// Skip formatting code until ; or end.
				i += 2
				for i < len(s) && s[i] != ';' {
					i++
				}
				if i < len(s) {
					i++ // skip ;
				}
			}
		} else if ch == '{' || ch == '}' {
			// Grouping braces — skip.
			i++
		} else {
			b.WriteByte(ch)
			i++
		}
	}
	return b.String()
}
