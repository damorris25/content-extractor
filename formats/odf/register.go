// CONTRACT:C1-EXTRACTOR.1.0
package odf

import "github.com/damorris25/content-extractor/extractor"

// Register adds all ODF format extractors to the given registry.
func Register(r *extractor.Registry) {
	r.Register(NewODT())
	r.Register(NewODS())
	r.Register(NewODP())
}
