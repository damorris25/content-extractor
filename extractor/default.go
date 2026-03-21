// CONTRACT:I2-REGISTRY.1.0
package extractor

import "io"

// RegistryBuilder is a function that registers extractors into a registry.
// Format packages provide these to avoid circular imports.
type RegistryBuilder func(r *Registry)

// DetectorFactory is a function that creates a Detector.
// The detect package provides this to avoid circular imports.
type DetectorFactory func() Detector

var (
	builders        []RegistryBuilder
	detectorFactory DetectorFactory
)

// RegisterBuilder adds a builder function that will be called by DefaultRegistry.
// Call this from format package init-free setup functions.
func RegisterBuilder(b RegistryBuilder) {
	builders = append(builders, b)
}

// RegisterDetector sets the detector factory for DefaultRegistry.
func RegisterDetector(f DetectorFactory) {
	detectorFactory = f
}

// DefaultRegistry returns a new registry pre-loaded with all built-in extractors.
// Returns a new registry each call — caller owns it.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	if detectorFactory != nil {
		r.SetDetector(detectorFactory())
	}
	for _, b := range builders {
		b(r)
	}
	return r
}

// Extract is a convenience function that uses the DefaultRegistry to extract
// content from the given reader. If mime is empty, format detection is used.
func Extract(r io.ReadSeeker, mime string) (*Result, error) {
	return DefaultRegistry().Extract(nil, r, mime)
}
