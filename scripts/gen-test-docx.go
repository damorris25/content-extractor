//go:build ignore

// gen-test-docx generates minimal DOCX test fixtures for the test suite.
package main

import (
	"archive/zip"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join("testdata")
	os.MkdirAll(dir, 0o755)

	createDocx(filepath.Join(dir, "simple.docx"), "Hello, World!", "Test Title", "Test Author")
	createDocx(filepath.Join(dir, "multipar.docx"), "", "Multi Paragraph", "Author")
	createEmptyDocx(filepath.Join(dir, "empty.docx"))
}

func createDocx(path, text, title, creator string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	writeEntry(w, "[Content_Types].xml", contentTypes)
	writeEntry(w, "_rels/.rels", rels)
	writeEntry(w, "word/_rels/document.xml.rels", wordRels)

	if text != "" {
		writeEntry(w, "word/document.xml", makeDocumentXML(text))
	} else {
		writeEntry(w, "word/document.xml", multiParDocumentXML)
	}

	writeEntry(w, "docProps/core.xml", makeCoreXML(title, creator))
	writeEntry(w, "docProps/app.xml", appXML)
}

func createEmptyDocx(path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	writeEntry(w, "[Content_Types].xml", contentTypes)
	writeEntry(w, "_rels/.rels", rels)
	writeEntry(w, "word/_rels/document.xml.rels", wordRels)
	writeEntry(w, "word/document.xml", emptyDocumentXML)
}

func writeEntry(w *zip.Writer, name, content string) {
	f, err := w.Create(name)
	if err != nil {
		panic(err)
	}
	f.Write([]byte(content))
}

func makeDocumentXML(text string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>` + text + `</w:t></w:r></w:p>
  </w:body>
</w:document>`
}

var multiParDocumentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>First paragraph.</w:t></w:r></w:p>
    <w:p><w:r><w:t>Second paragraph.</w:t></w:r></w:p>
    <w:p><w:r><w:t>Third paragraph.</w:t></w:r></w:p>
  </w:body>
</w:document>`

var emptyDocumentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
  </w:body>
</w:document>`

func makeCoreXML(title, creator string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>` + title + `</dc:title>
  <dc:creator>` + creator + `</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">2026-03-21T10:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2026-03-21T12:00:00Z</dcterms:modified>
</cp:coreProperties>`
}

var appXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>content-extractor test</Application>
  <Pages>1</Pages>
  <Words>2</Words>
  <Characters>13</Characters>
</Properties>`

var contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

var rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`

var wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`
