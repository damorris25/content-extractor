//go:build ignore

// gen-test-pptx generates minimal PPTX test fixtures for the test suite.
package main

import (
	"archive/zip"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createSimplePptx(filepath.Join(dir, "simple.pptx"))
	createMultiSlidePptx(filepath.Join(dir, "multislide.pptx"))
	createEmptyPptx(filepath.Join(dir, "empty.pptx"))
}

func createSimplePptx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", pptxContentTypes)
	write(w, "_rels/.rels", pptxRels)
	write(w, "ppt/_rels/presentation.xml.rels", pptxPresRels1)
	write(w, "ppt/presentation.xml", pptxPresentation1)
	write(w, "ppt/slides/slide1.xml", slide1Simple)
	write(w, "docProps/core.xml", pptxCore)
	write(w, "docProps/app.xml", pptxApp)
}

func createMultiSlidePptx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", pptxContentTypes)
	write(w, "_rels/.rels", pptxRels)
	write(w, "ppt/_rels/presentation.xml.rels", pptxPresRels2)
	write(w, "ppt/presentation.xml", pptxPresentation2)
	write(w, "ppt/slides/slide1.xml", slide1Multi)
	write(w, "ppt/slides/slide2.xml", slide2Multi)
	write(w, "ppt/slides/slide3.xml", slide3Multi)
	write(w, "docProps/core.xml", pptxCore)
	write(w, "docProps/app.xml", pptxApp)
}

func createEmptyPptx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", pptxContentTypes)
	write(w, "_rels/.rels", pptxRels)
	write(w, "ppt/_rels/presentation.xml.rels", pptxPresRels1)
	write(w, "ppt/presentation.xml", pptxPresentation1)
	write(w, "ppt/slides/slide1.xml", slideEmpty)
}

func write(w *zip.Writer, name, content string) {
	f, _ := w.Create(name)
	f.Write([]byte(content))
}

var pptxContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
</Types>`

var pptxRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`

var pptxPresRels1 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
</Relationships>`

var pptxPresRels2 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide3.xml"/>
</Relationships>`

var pptxPresentation1 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId1"/>
  </p:sldIdLst>
</p:presentation>`

var pptxPresentation2 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId1"/>
    <p:sldId id="257" r:id="rId2"/>
    <p:sldId id="258" r:id="rId3"/>
  </p:sldIdLst>
</p:presentation>`

var slide1Simple = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Welcome to the Presentation</a:t></a:r></a:p>
          <a:p><a:r><a:t>By Test Author</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`

var slide1Multi = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Introduction</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`

var slide2Multi = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Main Content</a:t></a:r></a:p>
          <a:p><a:r><a:t>Bullet point one</a:t></a:r></a:p>
          <a:p><a:r><a:t>Bullet point two</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`

var slide3Multi = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Conclusion</a:t></a:r></a:p>
          <a:p><a:r><a:t>Thank you!</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`

var slideEmpty = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree/>
  </p:cSld>
</p:sld>`

var pptxCore = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Test Presentation</dc:title>
  <dc:creator>Slide Author</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">2026-03-21T10:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2026-03-21T12:00:00Z</dcterms:modified>
</cp:coreProperties>`

var pptxApp = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>content-extractor test</Application>
  <Slides>3</Slides>
</Properties>`
