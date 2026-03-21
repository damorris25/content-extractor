// CONTRACT:I2-REGISTRY.1.0
package extractor

import (
	"context"
	"fmt"
	"io"
	"sort"
)

// Detector identifies file formats from content and/or filename.
// Defined here to avoid a circular import with the detect package.
type Detector interface {
	Detect(r io.ReadSeeker, filename string) (string, error)
}

// Registry maps MIME types to Extractor implementations.
// Safe for concurrent reads (Lookup, Extract, SupportedTypes) after initialization.
// Register and SetDetector must only be called during initialization (not concurrently).
type Registry struct {
	extractors map[string]Extractor
	detector   Detector
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]Extractor),
	}
}

// SetDetector configures the format detector used when Extract is called
// with an empty MIME type.
func (r *Registry) SetDetector(d Detector) {
	r.detector = d
}

// Register adds an extractor for all MIME types it supports.
// Panics if a MIME type is already registered or if ext is nil (fail-fast at startup).
func (r *Registry) Register(ext Extractor) {
	if ext == nil {
		panic("extractor: Register called with nil Extractor")
	}
	for _, mime := range ext.SupportedMIMETypes() {
		if existing, ok := r.extractors[mime]; ok {
			panic(fmt.Sprintf("extractor: MIME type %q already registered by %T", mime, existing))
		}
		r.extractors[mime] = ext
	}
}

// Lookup returns the extractor for the given MIME type, or nil if none registered.
func (r *Registry) Lookup(mime string) Extractor {
	return r.extractors[mime]
}

// Extract detects the format (if mime is empty) and dispatches to the
// appropriate extractor. This is the primary entry point for callers.
func (r *Registry) Extract(ctx context.Context, reader io.ReadSeeker, mime string) (*Result, error) {
	if reader == nil {
		return nil, ErrNilReader
	}

	if mime == "" {
		if r.detector == nil {
			return nil, ErrDetectionFailed
		}
		detected, err := r.detector.Detect(reader, "")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDetectionFailed, err)
		}
		mime = detected
	}

	ext := r.Lookup(mime)
	if ext == nil {
		return nil, ErrUnsupportedFormat
	}

	return ext.Extract(ctx, reader, mime)
}

// SupportedTypes returns all registered MIME types, sorted alphabetically.
func (r *Registry) SupportedTypes() []string {
	types := make([]string, 0, len(r.extractors))
	for mime := range r.extractors {
		types = append(types, mime)
	}
	sort.Strings(types)
	return types
}
