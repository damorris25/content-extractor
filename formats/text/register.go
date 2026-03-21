// CONTRACT:C1-EXTRACTOR.1.0
package text

import "github.com/dmorris/content-extractor/extractor"

// Register adds the plain text extractor to the given registry.
func Register(r *extractor.Registry) {
	r.Register(New())
}
