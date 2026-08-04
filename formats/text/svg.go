// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxSVGBytes is the default maximum bytes to read from an SVG file.
	DefaultMaxSVGBytes = 10 * 1024 * 1024 // 10MB

	mimeSVG = "image/svg+xml"
)

// SVGExtractor extracts text content and metadata from SVG files.
// Extracts <text>, <tspan>, <textPath> element content plus
// <title> and <desc> metadata. Safe for concurrent use.
type SVGExtractor struct {
	// MaxBytes limits how many bytes are read. Zero means DefaultMaxSVGBytes.
	MaxBytes int64
}

// NewSVG creates a new SVG extractor with default settings.
func NewSVG() *SVGExtractor {
	return &SVGExtractor{}
}

func (e *SVGExtractor) maxBytes() int64 {
	if e.MaxBytes > 0 {
		return e.MaxBytes
	}
	return DefaultMaxSVGBytes
}

// Extract reads SVG from r and returns text content and metadata.
func (e *SVGExtractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if mime != mimeSVG {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limited := io.LimitReader(r, e.maxBytes())
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mimeSVG,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mimeSVG)

	if len(data) == 0 {
		return result, nil
	}

	info := extractSVGContent(data)
	result.Text = metadata.CleanUTF8(info.text)
	if info.title != "" {
		result.Metadata.Set(metadata.KeyTitle, info.title)
	}
	if info.desc != "" {
		result.Metadata.Set(metadata.KeyDescription, info.desc)
	}

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *SVGExtractor) SupportedMIMETypes() []string {
	return []string{mimeSVG}
}

type svgInfo struct {
	text  string
	title string
	desc  string
}

// textElements are SVG elements that contain user-visible text.
var textElements = map[string]bool{
	"text":     true,
	"tspan":    true,
	"textPath": true,
}

// extractSVGContent parses SVG XML and extracts text content and metadata.
func extractSVGContent(data []byte) svgInfo {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false

	var info svgInfo
	var textParts []string
	var current strings.Builder

	// Track element stack for context.
	type frame struct {
		local    string
		inText   bool
		inTitle  bool
		inDesc   bool
		topLevel bool // direct child of <svg>
	}
	var stack []frame
	inText := false
	inTitle := false
	inDesc := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			f := frame{local: local}

			// Track depth — title/desc only at top level of SVG.
			parentIsSVG := len(stack) > 0 && stack[len(stack)-1].local == "svg"
			f.topLevel = parentIsSVG

			if textElements[local] {
				f.inText = true
				inText = true
			}
			if local == "title" && parentIsSVG {
				f.inTitle = true
				inTitle = true
				current.Reset()
			}
			if local == "desc" && parentIsSVG {
				f.inDesc = true
				inDesc = true
				current.Reset()
			}

			stack = append(stack, f)

		case xml.EndElement:
			if len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				if top.inText {
					inText = false
					// Check if parent is also a text element.
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].inText {
							inText = true
							break
						}
					}
				}
				if top.inTitle {
					inTitle = false
					info.title = strings.TrimSpace(current.String())
				}
				if top.inDesc {
					inDesc = false
					info.desc = strings.TrimSpace(current.String())
				}
			}

		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s == "" {
				continue
			}
			if inTitle {
				if current.Len() > 0 {
					current.WriteByte(' ')
				}
				current.WriteString(s)
			}
			if inDesc {
				if current.Len() > 0 {
					current.WriteByte(' ')
				}
				current.WriteString(s)
			}
			if inText {
				textParts = append(textParts, s)
			}
		}
	}

	info.text = strings.Join(textParts, "\n")
	return info
}
