//go:build ignore

// gen-test-odf generates minimal ODT, ODS, ODP test fixtures.
package main

import (
	"archive/zip"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	// ODT
	createODF(filepath.Join(dir, "simple.odt"), "application/vnd.oasis.opendocument.text", odtContent, odfMeta)
	createODF(filepath.Join(dir, "multipar.odt"), "application/vnd.oasis.opendocument.text", odtMultiPar, odfMeta)
	createEmptyODF(filepath.Join(dir, "empty.odt"), "application/vnd.oasis.opendocument.text")

	// ODS
	createODF(filepath.Join(dir, "simple.ods"), "application/vnd.oasis.opendocument.spreadsheet", odsContent, odfMeta)
	createODF(filepath.Join(dir, "multisheet.ods"), "application/vnd.oasis.opendocument.spreadsheet", odsMultiSheet, odfMeta)
	createEmptyODF(filepath.Join(dir, "empty.ods"), "application/vnd.oasis.opendocument.spreadsheet")

	// ODP
	createODF(filepath.Join(dir, "simple.odp"), "application/vnd.oasis.opendocument.presentation", odpContent, odfMeta)
	createODF(filepath.Join(dir, "multislide.odp"), "application/vnd.oasis.opendocument.presentation", odpMultiSlide, odfMeta)
	createEmptyODF(filepath.Join(dir, "empty.odp"), "application/vnd.oasis.opendocument.presentation")
}

func createODF(path, mimeType, content, meta string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	// mimetype must be first entry, uncompressed.
	mw, _ := w.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	mw.Write([]byte(mimeType))

	write(w, "content.xml", content)
	write(w, "meta.xml", meta)
}

func createEmptyODF(path, mimeType string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	mw, _ := w.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	mw.Write([]byte(mimeType))

	write(w, "content.xml", emptyContent)
}

func write(w *zip.Writer, name, content string) {
	f, _ := w.Create(name)
	f.Write([]byte(content))
}

// ODT content
var odtContent = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>Hello, ODT World!</text:p>
    </office:text>
  </office:body>
</office:document-content>`

var odtMultiPar = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>First paragraph.</text:p>
      <text:p>Second paragraph.</text:p>
      <text:h>A heading</text:h>
      <text:p>Third paragraph.</text:p>
    </office:text>
  </office:body>
</office:document-content>`

// ODS content
var odsContent = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:spreadsheet>
      <table:table table:name="Sheet1">
        <table:table-row>
          <table:table-cell><text:p>Name</text:p></table:table-cell>
          <table:table-cell><text:p>Age</text:p></table:table-cell>
          <table:table-cell><text:p>City</text:p></table:table-cell>
        </table:table-row>
        <table:table-row>
          <table:table-cell><text:p>Alice</text:p></table:table-cell>
          <table:table-cell><text:p>30</text:p></table:table-cell>
          <table:table-cell><text:p>Springfield</text:p></table:table-cell>
        </table:table-row>
      </table:table>
    </office:spreadsheet>
  </office:body>
</office:document-content>`

var odsMultiSheet = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:spreadsheet>
      <table:table table:name="Data">
        <table:table-row>
          <table:table-cell><text:p>Product</text:p></table:table-cell>
          <table:table-cell><text:p>Price</text:p></table:table-cell>
        </table:table-row>
        <table:table-row>
          <table:table-cell><text:p>Widget</text:p></table:table-cell>
          <table:table-cell><text:p>9.99</text:p></table:table-cell>
        </table:table-row>
      </table:table>
      <table:table table:name="Summary">
        <table:table-row>
          <table:table-cell><text:p>Total</text:p></table:table-cell>
          <table:table-cell><text:p>9.99</text:p></table:table-cell>
        </table:table-row>
      </table:table>
    </office:spreadsheet>
  </office:body>
</office:document-content>`

// ODP content
var odpContent = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:presentation="urn:oasis:names:tc:opendocument:xmlns:presentation:1.0">
  <office:body>
    <office:presentation>
      <draw:page draw:name="Slide1">
        <draw:frame>
          <draw:text-box>
            <text:p>Welcome to the Presentation</text:p>
            <text:p>By Test Author</text:p>
          </draw:text-box>
        </draw:frame>
      </draw:page>
    </office:presentation>
  </office:body>
</office:document-content>`

var odpMultiSlide = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:presentation="urn:oasis:names:tc:opendocument:xmlns:presentation:1.0">
  <office:body>
    <office:presentation>
      <draw:page draw:name="Slide1">
        <draw:frame>
          <draw:text-box>
            <text:p>Introduction</text:p>
          </draw:text-box>
        </draw:frame>
      </draw:page>
      <draw:page draw:name="Slide2">
        <draw:frame>
          <draw:text-box>
            <text:p>Main Content</text:p>
            <text:p>Details here</text:p>
          </draw:text-box>
        </draw:frame>
      </draw:page>
      <draw:page draw:name="Slide3">
        <draw:frame>
          <draw:text-box>
            <text:p>Conclusion</text:p>
          </draw:text-box>
        </draw:frame>
      </draw:page>
    </office:presentation>
  </office:body>
</office:document-content>`

var emptyContent = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:body/>
</office:document-content>`

var odfMeta = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0">
  <office:meta>
    <dc:title>Test Document</dc:title>
    <dc:creator>ODF Author</dc:creator>
    <dc:subject>Testing</dc:subject>
    <meta:creation-date>2026-03-21T10:00:00Z</meta:creation-date>
    <dc:date>2026-03-21T12:00:00Z</dc:date>
    <meta:generator>content-extractor test</meta:generator>
  </office:meta>
</office:document-meta>`
