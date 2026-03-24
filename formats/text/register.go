// CONTRACT:C1-EXTRACTOR.1.0
package text

import "github.com/dmorris/content-extractor/extractor"

// Register adds all text-based extractors (plain text, XML, HTML) to the given registry.
func Register(r *extractor.Registry) {
	r.Register(New())
	r.Register(NewXML())
	r.Register(NewHTML())
}
