// CONTRACT:C2-DETECT.1.0
package detect

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
)

const maxDetectBytes = 8192 // 8KB is sufficient for all known signatures

// validODFMIME is a whitelist of known ODF MIME types that can be trusted
// from a ZIP file's "mimetype" entry. Prevents type confusion attacks where
// an attacker sets the mimetype entry to route their file to an arbitrary extractor.
var validODFMIME = map[string]bool{
	"application/vnd.oasis.opendocument.text":                  true,
	"application/vnd.oasis.opendocument.spreadsheet":           true,
	"application/vnd.oasis.opendocument.presentation":          true,
	"application/vnd.oasis.opendocument.graphics":              true,
	"application/vnd.oasis.opendocument.chart":                 true,
	"application/vnd.oasis.opendocument.formula":               true,
	"application/vnd.oasis.opendocument.text-template":         true,
	"application/vnd.oasis.opendocument.spreadsheet-template":  true,
	"application/vnd.oasis.opendocument.presentation-template": true,
}

// Detector identifies file formats from content and/or filename.
type Detector struct{}

// NewDetector creates a new format detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect determines the MIME type by reading magic bytes from r.
// r must be seekable — Detect reads a prefix and seeks back to start.
// filename is optional (empty string OK) — used as a fallback hint.
func (d *Detector) Detect(r io.ReadSeeker, filename string) (string, error) {
	if r == nil {
		return "", extractor.ErrNilReader
	}

	// Save current position.
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", extractor.ErrDetectionFailed
	}

	// Read header bytes for magic detection.
	buf := make([]byte, maxDetectBytes)
	n, readErr := r.Read(buf)
	buf = buf[:n]

	// Seek back to original position.
	if _, err := r.Seek(pos, io.SeekStart); err != nil {
		return "", extractor.ErrDetectionFailed
	}

	if n == 0 {
		if readErr != nil {
			return "", extractor.ErrDetectionFailed
		}
		return "", extractor.ErrDetectionFailed
	}

	// Try magic byte signatures.
	if mime := matchSignature(buf); mime != "" {
		// ZIP-based formats need further inspection.
		if mime == "application/zip" {
			if zipMIME := detectZIPFormat(r, pos); zipMIME != "" {
				return zipMIME, nil
			}
		}
		return mime, nil
	}

	// Check for text-based formats that don't have magic bytes.
	if mime := detectTextFormat(buf); mime != "" {
		return mime, nil
	}

	// Fall back to extension.
	if filename != "" {
		if mime := extensionToMIME(filename); mime != "" {
			return mime, nil
		}
	}

	return "", extractor.ErrDetectionFailed
}

// matchSignature checks the buffer against known magic byte signatures.
func matchSignature(buf []byte) string {
	for _, sig := range signatures {
		if int64(len(buf)) < sig.Offset+int64(len(sig.Magic)) {
			continue
		}
		start := sig.Offset
		candidate := buf[start : start+int64(len(sig.Magic))]
		if sig.Mask != nil {
			match := true
			for i := range sig.Magic {
				if candidate[i]&sig.Mask[i] != sig.Magic[i]&sig.Mask[i] {
					match = false
					break
				}
			}
			if match {
				return sig.MIME
			}
		} else {
			if bytes.Equal(candidate, sig.Magic) {
				return sig.MIME
			}
		}
	}
	return ""
}

// detectZIPFormat inspects ZIP contents to distinguish Office/ODF formats.
func detectZIPFormat(r io.ReadSeeker, originalPos int64) string {
	// Seek to start for ZIP reading.
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return ""
	}

	// Read enough for ZIP inspection.
	data, err := io.ReadAll(io.LimitReader(r, 2*1024*1024)) // 2MB limit for ZIP detection
	if err != nil {
		return ""
	}

	// Seek back to original position.
	if _, err := r.Seek(originalPos, io.SeekStart); err != nil {
		return ""
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}

	// Limit file set size to prevent resource exhaustion from ZIP bombs
	// with millions of entries.
	const maxEntries = 10000
	if len(zr.File) > maxEntries {
		return "application/zip" // Too many entries to inspect safely.
	}
	fileSet := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		fileSet[f.Name] = true
	}

	// Check for ODF formats first (they have a mimetype file).
	// Validate against known ODF MIME types to prevent type confusion.
	if fileSet["mimetype"] {
		mime := strings.TrimSpace(readZIPEntry(zr, "mimetype"))
		if validODFMIME[mime] {
			return mime
		}
	}

	// Check for OOXML Office formats.
	switch {
	case fileSet["word/document.xml"]:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case fileSet["xl/workbook.xml"]:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case fileSet["ppt/presentation.xml"]:
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}

	return "application/zip"
}

// readZIPEntry reads the content of a named file from a ZIP archive.
func readZIPEntry(zr *zip.Reader, name string) string {
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return ""
			}
			defer rc.Close()
			data, err := io.ReadAll(io.LimitReader(rc, 256))
			if err != nil {
				return ""
			}
			return string(data)
		}
	}
	return ""
}

// detectTextFormat checks if the buffer looks like a known text-based format.
func detectTextFormat(buf []byte) string {
	s := strings.TrimSpace(string(buf))

	// DXF files start with a section header.
	if strings.HasPrefix(s, "0\nSECTION") ||
		strings.HasPrefix(s, "0\r\nSECTION") {
		return "image/vnd.dxf"
	}

	// XML: starts with <?xml or a tag starting with <.
	// Check for XML declaration first (most reliable).
	if strings.HasPrefix(s, "<?xml") {
		return "application/xml"
	}

	// HTML: starts with <!DOCTYPE html (case-insensitive) or <html.
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "<!doctype html") ||
		strings.HasPrefix(lower, "<html") {
		return "text/html"
	}

	// EML (RFC 2822): starts with common email headers.
	if looksLikeEML(s) {
		return "message/rfc822"
	}

	return ""
}

// looksLikeEML checks if the text starts with RFC 2822 email headers.
func looksLikeEML(s string) bool {
	// Check for common first-line headers in email files.
	prefixes := []string{
		"From: ",
		"Return-Path: ",
		"Received: ",
		"MIME-Version: ",
		"Message-ID: ",
		"Message-Id: ",
		"Date: ",
		"Subject: ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// DetectFunc allows using a function as an extractor.Detector.
type DetectFunc func(r io.ReadSeeker, filename string) (string, error)

// Detect implements extractor.Detector.
func (f DetectFunc) Detect(r io.ReadSeeker, filename string) (string, error) {
	return f(r, filename)
}

// extensionToMIME looks up a MIME type by file extension.
func extensionToMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if mime, ok := extMap[ext]; ok {
		return mime
	}
	return ""
}

// Verify that Detector implements extractor.Detector at compile time.
var _ extractor.Detector = (*Detector)(nil)

// Verify that DetectFunc implements extractor.Detector at compile time.
var _ extractor.Detector = DetectFunc(nil)

// Error alias for convenience.
var (
	ErrDetectionFailed = extractor.ErrDetectionFailed
	ErrNilReader       = extractor.ErrNilReader
)
