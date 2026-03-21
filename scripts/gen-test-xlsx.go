//go:build ignore

// gen-test-xlsx generates minimal XLSX test fixtures for the test suite.
package main

import (
	"archive/zip"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createSimpleXlsx(filepath.Join(dir, "simple.xlsx"))
	createMultiSheetXlsx(filepath.Join(dir, "multisheet.xlsx"))
	createEmptyXlsx(filepath.Join(dir, "empty.xlsx"))
}

func createSimpleXlsx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", xlsxContentTypes)
	write(w, "_rels/.rels", xlsxRels)
	write(w, "xl/_rels/workbook.xml.rels", xlsxWorkbookRels)
	write(w, "xl/workbook.xml", xlsxWorkbook)
	write(w, "xl/sharedStrings.xml", simpleSharedStrings)
	write(w, "xl/worksheets/sheet1.xml", simpleSheet1)
	write(w, "docProps/core.xml", xlsxCore)
	write(w, "docProps/app.xml", xlsxApp)
}

func createMultiSheetXlsx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", xlsxContentTypes)
	write(w, "_rels/.rels", xlsxRels)
	write(w, "xl/_rels/workbook.xml.rels", multiSheetWorkbookRels)
	write(w, "xl/workbook.xml", multiSheetWorkbook)
	write(w, "xl/sharedStrings.xml", multiSheetSharedStrings)
	write(w, "xl/worksheets/sheet1.xml", multiSheet1)
	write(w, "xl/worksheets/sheet2.xml", multiSheet2)
	write(w, "docProps/core.xml", xlsxCore)
	write(w, "docProps/app.xml", xlsxApp)
}

func createEmptyXlsx(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	write(w, "[Content_Types].xml", xlsxContentTypes)
	write(w, "_rels/.rels", xlsxRels)
	write(w, "xl/_rels/workbook.xml.rels", xlsxWorkbookRels)
	write(w, "xl/workbook.xml", xlsxWorkbook)
	write(w, "xl/worksheets/sheet1.xml", emptySheet)
}

func write(w *zip.Writer, name, content string) {
	f, _ := w.Create(name)
	f.Write([]byte(content))
}

var xlsxContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
</Types>`

var xlsxRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`

var xlsxWorkbookRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`

var xlsxWorkbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`

// Simple: 2 rows, 3 columns. Row 1 = headers (shared strings), Row 2 = mixed.
var simpleSharedStrings = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="5" uniqueCount="5">
  <si><t>Name</t></si>
  <si><t>Age</t></si>
  <si><t>City</t></si>
  <si><t>Alice</t></si>
  <si><t>Springfield</t></si>
</sst>`

// Cell A1=Name(s:0), B1=Age(s:1), C1=City(s:2), A2=Alice(s:3), B2=30(num), C2=Springfield(s:4)
var simpleSheet1 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="s"><v>0</v></c>
      <c r="B1" t="s"><v>1</v></c>
      <c r="C1" t="s"><v>2</v></c>
    </row>
    <row r="2">
      <c r="A2" t="s"><v>3</v></c>
      <c r="B2"><v>30</v></c>
      <c r="C2" t="s"><v>4</v></c>
    </row>
  </sheetData>
</worksheet>`

// Multi-sheet
var multiSheetWorkbookRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`

var multiSheetWorkbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Data" sheetId="1" r:id="rId1"/>
    <sheet name="Summary" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`

var multiSheetSharedStrings = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="4" uniqueCount="4">
  <si><t>Product</t></si>
  <si><t>Price</t></si>
  <si><t>Widget</t></si>
  <si><t>Total</t></si>
</sst>`

var multiSheet1 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="s"><v>0</v></c>
      <c r="B1" t="s"><v>1</v></c>
    </row>
    <row r="2">
      <c r="A2" t="s"><v>2</v></c>
      <c r="B2"><v>9.99</v></c>
    </row>
  </sheetData>
</worksheet>`

var multiSheet2 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="s"><v>3</v></c>
      <c r="B1"><v>9.99</v></c>
    </row>
  </sheetData>
</worksheet>`

var emptySheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData/>
</worksheet>`

var xlsxCore = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Sales Data</dc:title>
  <dc:creator>Spreadsheet Author</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">2026-03-21T10:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2026-03-21T12:00:00Z</dcterms:modified>
</cp:coreProperties>`

var xlsxApp = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>content-extractor test</Application>
</Properties>`
