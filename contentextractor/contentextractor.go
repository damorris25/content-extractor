// Package contentextractor provides a ready-to-use content extraction setup
// with all built-in extractors and format detection registered.
package contentextractor

import (
	"context"
	"io"

	"github.com/dmorris/content-extractor/detect"
	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/formats/audio"
	"github.com/dmorris/content-extractor/formats/image"
	"github.com/dmorris/content-extractor/formats/odf"
	"github.com/dmorris/content-extractor/formats/office"
	"github.com/dmorris/content-extractor/formats/pdf"
	"github.com/dmorris/content-extractor/formats/text"
)

// NewRegistry returns a registry pre-loaded with all built-in extractors
// and format detection.
func NewRegistry() *extractor.Registry {
	r := extractor.NewRegistry()
	r.SetDetector(detect.NewDetector())
	text.Register(r)
	office.Register(r)
	odf.Register(r)
	pdf.Register(r)
	image.Register(r)
	audio.Register(r)
	return r
}

// Extract is a convenience function that detects format and extracts content.
// If mime is empty, format auto-detection is used.
func Extract(ctx context.Context, r io.ReadSeeker, mime string) (*extractor.Result, error) {
	return NewRegistry().Extract(ctx, r, mime)
}
