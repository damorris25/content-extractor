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
		// ISOBMFF-based formats need brand inspection.
		if mime == "application/x-isobmff" {
			if bmffMIME := detectISOBMFF(buf); bmffMIME != "" {
				return bmffMIME, nil
			}
			// Unknown ISOBMFF brand — fall through to extension/text detection.
		} else if mime == "application/x-riff" {
			// RIFF container — disambiguate by subtype at bytes 8-11.
			if riffMIME := detectRIFF(buf); riffMIME != "" {
				return riffMIME, nil
			}
			// Unknown RIFF subtype — fall through to extension/text detection.
		} else {
			return mime, nil
		}
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

// isobmffBrands maps ISOBMFF major/compatible brands to MIME types.
var isobmffBrands = map[string]string{
	"heic": "image/heic",
	"heix": "image/heic",
	"heim": "image/heic",
	"heis": "image/heic",
	"mif1": "image/heif",
	"msf1": "image/heif",
	"avif": "image/avif",
	"avis": "image/avif",
}

// detectISOBMFF reads the ftyp box to determine the specific ISOBMFF format.
// The ftyp box layout: [4B size][4B "ftyp"][4B major_brand][4B minor_version][4B compatible_brands...]
func detectISOBMFF(buf []byte) string {
	if len(buf) < 12 {
		return ""
	}

	// Read box size (big-endian uint32).
	boxSize := int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])
	if boxSize < 12 || boxSize > len(buf) {
		boxSize = len(buf) // Use available data if box extends beyond buffer.
	}

	// Major brand at bytes 8-11.
	major := string(buf[8:12])
	if mime, ok := isobmffBrands[major]; ok {
		return mime
	}

	// Scan compatible brands starting at byte 16, in 4-byte chunks.
	for i := 16; i+4 <= boxSize; i += 4 {
		brand := string(buf[i : i+4])
		if mime, ok := isobmffBrands[brand]; ok {
			return mime
		}
	}

	return ""
}

// detectRIFF disambiguates RIFF-based formats by checking the subtype at bytes 8-11.
func detectRIFF(buf []byte) string {
	if len(buf) < 12 {
		return ""
	}
	subtype := string(buf[8:12])
	switch subtype {
	case "WEBP":
		return "image/webp"
	case "WAVE":
		return "audio/wav"
	default:
		return ""
	}
}

// detectTextFormat checks if the buffer looks like a known text-based format.
func detectTextFormat(buf []byte) string {
	s := strings.TrimSpace(string(buf))

	// DXF files start with a section header.
	if strings.HasPrefix(s, "0\nSECTION") ||
		strings.HasPrefix(s, "0\r\nSECTION") {
		return "image/vnd.dxf"
	}

	// RTF: starts with {\rtf
	if strings.HasPrefix(s, "{\\rtf") {
		return "application/rtf"
	}

	// XML: starts with <?xml. Check for SVG before generic XML.
	if strings.HasPrefix(s, "<?xml") {
		if strings.Contains(s, "<svg") {
			return "image/svg+xml"
		}
		return "application/xml"
	}

	// SVG without XML declaration.
	if strings.HasPrefix(s, "<svg") {
		return "image/svg+xml"
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

	// JSON: starts with { or [
	if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
		return "application/json"
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
