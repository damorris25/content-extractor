// CONTRACT:C1-EXTRACTOR.1.0
package office

import "github.com/dmorris/content-extractor/extractor"

const (
	// maxDecompressedEntrySize limits the decompressed size of individual ZIP entries
	// to prevent zip bomb attacks. Applies to content XML files (document.xml, slide XML, etc.).
	maxDecompressedEntrySize = 100 * 1024 * 1024 // 100MB

	// maxZIPEntries limits the number of ZIP entries processed to prevent resource
	// exhaustion from archives with millions of entries.
	maxZIPEntries = 10000
)

// Register adds all Office format extractors to the given registry.
func Register(r *extractor.Registry) {
	r.Register(NewDocx())
	r.Register(NewXlsx())
	r.Register(NewPptx())
}
