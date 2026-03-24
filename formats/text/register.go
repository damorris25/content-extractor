// CONTRACT:C1-EXTRACTOR.1.0
package text

import "github.com/dmorris/content-extractor/extractor"

// Register adds all text-based extractors to the given registry.
func Register(r *extractor.Registry) {
	r.Register(New())
	r.Register(NewTextLike(
		"application/json",
		"text/csv",
		"text/tab-separated-values",
		"application/x-yaml",
		"text/markdown",
	))
	r.Register(NewXML())
	r.Register(NewHTML())
	r.Register(NewSVG())
	r.Register(NewRTF())
}
