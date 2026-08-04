// Package contentextractor provides a ready-to-use content extraction setup
// with all built-in extractors and format detection registered.
package contentextractor

import (
	"context"
	"io"

	"github.com/damorris25/content-extractor/detect"
	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/formats/audio"
	"github.com/damorris25/content-extractor/formats/cad"
	"github.com/damorris25/content-extractor/formats/email"
	"github.com/damorris25/content-extractor/formats/image"
	"github.com/damorris25/content-extractor/formats/odf"
	"github.com/damorris25/content-extractor/formats/office"
	"github.com/damorris25/content-extractor/formats/pdf"
	"github.com/damorris25/content-extractor/formats/text"
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
	cad.Register(r)
	email.Register(r)
	return r
}

// Extract is a convenience function that detects format and extracts content.
// If mime is empty, format auto-detection is used.
// Note: this creates a new registry on each call. For repeated use, call
// NewRegistry() once and reuse it.
func Extract(ctx context.Context, r io.ReadSeeker, mime string) (*extractor.Result, error) {
	return NewRegistry().Extract(ctx, r, mime)
}
