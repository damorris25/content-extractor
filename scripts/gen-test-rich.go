//go:build ignore

// gen-test-rich generates content-rich test fixtures with known expected values.
// These fixtures are more realistic than the simple ones and include:
// - Multi-paragraph text with special characters
// - Rich metadata (multiple fields populated)
// - Edge cases (Unicode, long text, special formatting)
//
// The expected extraction values are defined in testdata/expected.go (test file)
// to serve as the ground truth for accuracy testing.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createRichDOCX(filepath.Join(dir, "rich.docx"))
	createRichXLSX(filepath.Join(dir, "rich.xlsx"))
	createRichPPTX(filepath.Join(dir, "rich.pptx"))
	createRichODT(filepath.Join(dir, "rich.odt"))
	createRichDXF(filepath.Join(dir, "rich.dxf"))
	createRichMP3(filepath.Join(dir, "rich.mp3"))
	createUnicodeDOCX(filepath.Join(dir, "unicode.docx"))

	fmt.Println("Generated rich test fixtures in testdata/")
}

// --- DOCX ---

func createRichDOCX(path string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Content types
	addZipFile(zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
</Types>`)

	// Relationships
	addZipFile(zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`)

	addZipFile(zw, "word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`)

	// Document with multiple paragraphs, styles, and content
	addZipFile(zw, "word/document.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Quarterly Financial Report - Q4 2025</w:t></w:r></w:p>
    <w:p><w:r><w:t>Prepared by the Finance Department</w:t></w:r></w:p>
    <w:p><w:r><w:t>Revenue for Q4 reached $4.2 million, representing a 15% increase over Q3. Operating expenses were $3.1 million, yielding an operating margin of 26.2%.</w:t></w:r></w:p>
    <w:p><w:r><w:t>Key highlights include:</w:t></w:r></w:p>
    <w:p><w:r><w:t>The enterprise segment grew 22% year-over-year, driven by three major contract wins in the healthcare and financial services verticals.</w:t></w:r></w:p>
    <w:p><w:r><w:t>Customer retention rate improved to 94.3%, up from 91.7% in the previous quarter.</w:t></w:r></w:p>
    <w:p><w:r><w:t>R&amp;D investment totaled $890,000, focused on next-generation platform capabilities.</w:t></w:r></w:p>
  </w:body>
</w:document>`)

	// Rich metadata
	addZipFile(zw, "docProps/core.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:dcterms="http://purl.org/dc/terms/"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Quarterly Financial Report</dc:title>
  <dc:creator>Jane Smith</dc:creator>
  <dc:subject>Finance, Q4 2025, Revenue Analysis</dc:subject>
  <dc:description>Comprehensive quarterly financial analysis with revenue breakdown and projections.</dc:description>
  <dcterms:created xsi:type="dcterms:W3CDTF">2025-12-15T09:30:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2025-12-18T16:45:00Z</dcterms:modified>
  <cp:lastModifiedBy>John Davis</cp:lastModifiedBy>
</cp:coreProperties>`)

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0o644)
}

// --- Unicode DOCX ---

func createUnicodeDOCX(path string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipFile(zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)

	addZipFile(zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`)

	addZipFile(zw, "word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`)

	// Unicode content: CJK, Arabic, emoji, accented Latin
	addZipFile(zw, "word/document.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>`+xml.Header+`</w:t></w:r></w:p>
    <w:p><w:r><w:t>English: The quick brown fox jumps over the lazy dog.</w:t></w:r></w:p>
    <w:p><w:r><w:t>French: Les na`+"\u00ef"+`fs `+"\u00E6"+`githers fran`+"\u00E7"+`ais r`+"\u00EA"+`vent de cr`+"\u00E8"+`me br`+"\u00FB"+`l`+"\u00E9"+`e.</w:t></w:r></w:p>
    <w:p><w:r><w:t>Japanese: `+"\u6771\u4EAC\u306F\u65E5\u672C\u306E\u9996\u90FD\u3067\u3059\u3002"+`</w:t></w:r></w:p>
    <w:p><w:r><w:t>Chinese: `+"\u4E2D\u6587\u6D4B\u8BD5\u6587\u672C\u3002"+`</w:t></w:r></w:p>
    <w:p><w:r><w:t>Korean: `+"\uD55C\uAD6D\uC5B4\u0020\uD14C\uC2A4\uD2B8\u3002"+`</w:t></w:r></w:p>
    <w:p><w:r><w:t>Symbols: `+"\u00A9\u00AE\u2122\u2603\u2764"+`</w:t></w:r></w:p>
  </w:body>
</w:document>`)

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0o644)
}

// --- XLSX ---

func createRichXLSX(path string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipFile(zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
</Types>`)

	addZipFile(zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`)

	addZipFile(zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`)

	addZipFile(zw, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
    xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Employee Data" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`)

	// Shared strings: Employee, Department, Salary, Alice Chen, Engineering, Bob Martinez, Marketing, Carol Tanaka, Sales
	addZipFile(zw, "xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="9" uniqueCount="9">
  <si><t>Employee</t></si>
  <si><t>Department</t></si>
  <si><t>Salary</t></si>
  <si><t>Alice Chen</t></si>
  <si><t>Engineering</t></si>
  <si><t>Bob Martinez</t></si>
  <si><t>Marketing</t></si>
  <si><t>Carol Tanaka</t></si>
  <si><t>Sales</t></si>
</sst>`)

	// Sheet with header row + 3 data rows with shared strings and numeric values
	addZipFile(zw, "xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="s"><v>0</v></c>
      <c r="B1" t="s"><v>1</v></c>
      <c r="C1" t="s"><v>2</v></c>
    </row>
    <row r="2">
      <c r="A2" t="s"><v>3</v></c>
      <c r="B2" t="s"><v>4</v></c>
      <c r="C2"><v>125000</v></c>
    </row>
    <row r="3">
      <c r="A3" t="s"><v>5</v></c>
      <c r="B3" t="s"><v>6</v></c>
      <c r="C3"><v>98000</v></c>
    </row>
    <row r="4">
      <c r="A4" t="s"><v>7</v></c>
      <c r="B4" t="s"><v>8</v></c>
      <c r="C4"><v>105000</v></c>
    </row>
  </sheetData>
</worksheet>`)

	addZipFile(zw, "docProps/core.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:dcterms="http://purl.org/dc/terms/"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Employee Compensation Data</dc:title>
  <dc:creator>HR Department</dc:creator>
</cp:coreProperties>`)

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0o644)
}

// --- PPTX ---

func createRichPPTX(path string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipFile(zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
  <Override PartName="/ppt/slides/slide2.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
  <Override PartName="/ppt/slides/slide3.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
</Types>`)

	addZipFile(zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`)

	addZipFile(zw, "ppt/_rels/presentation.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide3.xml"/>
</Relationships>`)

	addZipFile(zw, "ppt/presentation.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
    xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId1"/>
    <p:sldId id="257" r:id="rId2"/>
    <p:sldId id="258" r:id="rId3"/>
  </p:sldIdLst>
</p:presentation>`)

	// Slide 1: Title slide
	addZipFile(zw, "ppt/slides/slide1.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
    xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Product Launch Strategy 2026</a:t></a:r></a:p>
          <a:p><a:r><a:t>Confidential - Internal Use Only</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`)

	// Slide 2: Content slide with bullet points
	addZipFile(zw, "ppt/slides/slide2.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
    xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Market Analysis</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Total addressable market estimated at $12 billion</a:t></a:r></a:p>
          <a:p><a:r><a:t>Primary competitors hold 45% market share</a:t></a:r></a:p>
          <a:p><a:r><a:t>Growth opportunity in emerging markets of Southeast Asia</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`)

	// Slide 3: Conclusion
	addZipFile(zw, "ppt/slides/slide3.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
    xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Next Steps and Timeline</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Phase 1: Beta launch in Q1 2026</a:t></a:r></a:p>
          <a:p><a:r><a:t>Phase 2: General availability in Q2 2026</a:t></a:r></a:p>
          <a:p><a:r><a:t>Phase 3: International expansion in Q3 2026</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`)

	addZipFile(zw, "docProps/core.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
    xmlns:dc="http://purl.org/dc/elements/1.1/">
  <dc:title>Product Launch Strategy</dc:title>
  <dc:creator>Strategy Team</dc:creator>
</cp:coreProperties>`)

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0o644)
}

// --- ODT ---

func createRichODT(path string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipFile(zw, "mimetype", "application/vnd.oasis.opendocument.text")

	addZipFile(zw, "META-INF/manifest.xml", `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
  <manifest:file-entry manifest:media-type="application/vnd.oasis.opendocument.text" manifest:full-path="/"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="content.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="meta.xml"/>
</manifest:manifest>`)

	addZipFile(zw, "content.xml", `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
    xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>Technical Specification: Cloud Migration Framework</text:p>
      <text:p>Version 2.1 - Last Updated December 2025</text:p>
      <text:p>This document outlines the technical architecture for migrating on-premises infrastructure to a hybrid cloud environment. The framework supports AWS, Azure, and GCP as target platforms.</text:p>
      <text:h>1. Architecture Overview</text:h>
      <text:p>The migration framework consists of three core components: the Discovery Agent, the Migration Engine, and the Validation Suite. Each component operates independently and communicates via a message queue.</text:p>
      <text:h>2. Data Transfer Protocol</text:h>
      <text:p>Data transfer uses AES-256 encryption in transit with checksums verified at both source and destination. Maximum transfer rate is configurable up to 10 Gbps per stream.</text:p>
    </office:text>
  </office:body>
</office:document-content>`)

	addZipFile(zw, "meta.xml", `<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0">
  <office:meta>
    <dc:title>Cloud Migration Framework</dc:title>
    <dc:creator>Platform Engineering</dc:creator>
    <dc:subject>Cloud Architecture, Migration, Infrastructure</dc:subject>
    <dc:description>Technical specification for hybrid cloud migration.</dc:description>
  </office:meta>
</office:document-meta>`)

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0o644)
}

// --- DXF ---

func createRichDXF(path string) {
	dxf := `  0
SECTION
  2
HEADER
  9
$ACADVER
  1
AC1032
  9
$INSUNITS
 70
6
  0
ENDSEC
  0
SECTION
  2
TABLES
  0
TABLE
  2
LAYER
  0
LAYER
  2
Foundation
 70
0
  0
LAYER
  2
Structural
 70
0
  0
LAYER
  2
Electrical
 70
0
  0
LAYER
  2
Plumbing
 70
0
  0
LAYER
  2
HVAC
 70
0
  0
ENDTAB
  0
ENDSEC
  0
SECTION
  2
ENTITIES
  0
TEXT
  1
Building A - Ground Floor Plan
  0
TEXT
  1
Scale: 1:100
  0
TEXT
  1
North Elevation
  0
MTEXT
  1
Load-bearing wall\PThickness: 300mm\PMaterial: Reinforced concrete
  0
TEXT
  1
Fire Exit - Do Not Block
  0
TEXT
  1
Electrical Panel EP-01
  0
DIMENSION
  1
15.750m
  0
DIMENSION
  1
8.200m
  0
INSERT
  2
WindowType-A
  0
INSERT
  2
DoorType-B
  0
TEXT
  1
Stairwell Access
  0
ENDSEC
  0
EOF
`
	os.WriteFile(path, []byte(dxf), 0o644)
}

// --- MP3 with rich ID3v2 ---

func createRichMP3(path string) {
	var buf []byte

	// ID3v2.3 header.
	buf = append(buf, 'I', 'D', '3')
	buf = append(buf, 3, 0) // Version 2.3.0
	buf = append(buf, 0)    // Flags
	buf = append(buf, 0, 0, 0, 0)

	tagStart := len(buf)

	buf = appendID3Frame(buf, "TIT2", "Midnight in the Garden of Sound")
	buf = appendID3Frame(buf, "TPE1", "The Electronica Collective")
	buf = appendID3Frame(buf, "TALB", "Synthetic Dreams Vol. 3")
	buf = appendID3Frame(buf, "TRCK", "7/12")
	buf = appendID3Frame(buf, "TYER", "2025")
	buf = appendID3Frame(buf, "TCON", "Electronic")
	buf = appendID3Frame(buf, "TPE2", "Various Artists")
	buf = appendID3Frame(buf, "COMM", "Recorded live at Berlin Warehouse Sessions")

	tagSize := len(buf) - tagStart
	buf[6] = byte((tagSize >> 21) & 0x7F)
	buf[7] = byte((tagSize >> 14) & 0x7F)
	buf[8] = byte((tagSize >> 7) & 0x7F)
	buf[9] = byte(tagSize & 0x7F)

	// Minimal MP3 frame.
	buf = append(buf, 0xFF, 0xFB, 0x90, 0x00)
	buf = append(buf, make([]byte, 413)...)

	os.WriteFile(path, buf, 0o644)
}

func appendID3Frame(buf []byte, id, value string) []byte {
	data := []byte{0x03} // UTF-8 encoding
	data = append(data, []byte(value)...)

	buf = append(buf, []byte(id)...)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))
	buf = append(buf, size...)
	buf = append(buf, 0, 0) // Flags
	buf = append(buf, data...)
	return buf
}

// --- Helpers ---

func addZipFile(zw *zip.Writer, name, content string) {
	w, _ := zw.Create(name)
	w.Write([]byte(content))
}
