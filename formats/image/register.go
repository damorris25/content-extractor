// CONTRACT:C1-EXTRACTOR.1.0
package image

import "github.com/damorris25/content-extractor/extractor"

// Register adds the image metadata extractor to the given registry.
func Register(r *extractor.Registry) {
	r.Register(New())
}
