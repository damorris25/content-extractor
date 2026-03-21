//go:build ignore

// gen-test-pdf generates minimal PDF test fixtures with correct byte offsets.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createSimplePDF(filepath.Join(dir, "simple.pdf"))
	createMultiPagePDF(filepath.Join(dir, "multipage.pdf"))
	createEmptyPDF(filepath.Join(dir, "empty.pdf"))
}

type pdfWriter struct {
	buf     []byte
	offsets []int // byte offset of each object (1-indexed, index 0 unused)
}

func (w *pdfWriter) write(s string) {
	w.buf = append(w.buf, []byte(s)...)
}

func (w *pdfWriter) writef(format string, args ...any) {
	w.write(fmt.Sprintf(format, args...))
}

func (w *pdfWriter) startObj(num int) {
	for len(w.offsets) <= num {
		w.offsets = append(w.offsets, 0)
	}
	w.offsets[num] = len(w.buf)
	w.writef("%d 0 obj\n", num)
}

func (w *pdfWriter) endObj() {
	w.write("endobj\n\n")
}

func (w *pdfWriter) writeXref() int {
	xrefOffset := len(w.buf)
	w.write("xref\n")
	w.writef("0 %d\n", len(w.offsets))
	w.writef("0000000000 65535 f \n")
	for i := 1; i < len(w.offsets); i++ {
		w.writef("%010d 00000 n \n", w.offsets[i])
	}
	return xrefOffset
}

func (w *pdfWriter) writeTrailer(size, root, xrefOffset int) {
	w.write("trailer\n")
	w.writef("<< /Size %d /Root %d 0 R >>\n", size, root)
	w.write("startxref\n")
	w.writef("%d\n", xrefOffset)
	w.write("%%EOF\n")
}

func (w *pdfWriter) save(path string) {
	os.WriteFile(path, w.buf, 0o644)
}

func createSimplePDF(path string) {
	w := &pdfWriter{}
	w.write("%PDF-1.4\n")

	// Obj 1: Catalog
	w.startObj(1)
	w.write("<< /Type /Catalog /Pages 2 0 R >>\n")
	w.endObj()

	// Obj 2: Pages
	w.startObj(2)
	w.write("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	w.endObj()

	// Obj 3: Page
	w.startObj(3)
	w.write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]\n")
	w.write("   /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	w.endObj()

	// Obj 4: Content stream
	content := "BT /F1 12 Tf 100 700 Td (Hello, PDF World!) Tj ET"
	w.startObj(4)
	w.writef("<< /Length %d >>\n", len(content))
	w.write("stream\n")
	w.write(content)
	w.write("\nendstream\n")
	w.endObj()

	// Obj 5: Font
	w.startObj(5)
	w.write("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	w.endObj()

	xref := w.writeXref()
	w.writeTrailer(6, 1, xref)
	w.save(path)
}

func createMultiPagePDF(path string) {
	w := &pdfWriter{}
	w.write("%PDF-1.4\n")

	w.startObj(1)
	w.write("<< /Type /Catalog /Pages 2 0 R >>\n")
	w.endObj()

	w.startObj(2)
	w.write("<< /Type /Pages /Kids [3 0 R 6 0 R] /Count 2 >>\n")
	w.endObj()

	// Page 1
	w.startObj(3)
	w.write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]\n")
	w.write("   /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	w.endObj()

	content1 := "BT /F1 12 Tf 100 700 Td (Page one content.) Tj ET"
	w.startObj(4)
	w.writef("<< /Length %d >>\n", len(content1))
	w.write("stream\n")
	w.write(content1)
	w.write("\nendstream\n")
	w.endObj()

	w.startObj(5)
	w.write("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	w.endObj()

	// Page 2
	w.startObj(6)
	w.write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]\n")
	w.write("   /Contents 7 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	w.endObj()

	content2 := "BT /F1 12 Tf 100 700 Td (Page two content.) Tj ET"
	w.startObj(7)
	w.writef("<< /Length %d >>\n", len(content2))
	w.write("stream\n")
	w.write(content2)
	w.write("\nendstream\n")
	w.endObj()

	xref := w.writeXref()
	w.writeTrailer(8, 1, xref)
	w.save(path)
}

func createEmptyPDF(path string) {
	w := &pdfWriter{}
	w.write("%PDF-1.4\n")

	w.startObj(1)
	w.write("<< /Type /Catalog /Pages 2 0 R >>\n")
	w.endObj()

	w.startObj(2)
	w.write("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	w.endObj()

	w.startObj(3)
	w.write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\n")
	w.endObj()

	xref := w.writeXref()
	w.writeTrailer(4, 1, xref)
	w.save(path)
}
