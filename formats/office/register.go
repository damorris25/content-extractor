// CONTRACT:C1-EXTRACTOR.1.0
package office

import "github.com/dmorris/content-extractor/extractor"

// Register adds all Office format extractors to the given registry.
func Register(r *extractor.Registry) {
	r.Register(NewDocx())
}
