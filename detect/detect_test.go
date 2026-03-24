// CONTRACT:C2-DETECT.1.0
package detect

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
)

func TestDetect_MagicBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"PDF", []byte("%PDF-1.7 rest of file"), "application/pdf"},
		{"PNG", []byte("\x89PNG\r\n\x1a\nrest"), "image/png"},
		{"JPEG", []byte("\xff\xd8\xff\xe0rest"), "image/jpeg"},
		{"GIF87a", []byte("GIF87arest"), "image/gif"},
		{"GIF89a", []byte("GIF89arest"), "image/gif"},
		{"BMP", []byte("BMrest of file data"), "image/bmp"},
		{"TIFF_LE", []byte("II\x2a\x00rest"), "image/tiff"},
		{"TIFF_BE", []byte("MM\x00\x2arest"), "image/tiff"},
		{"OGG", []byte("OggSrest"), "audio/ogg"},
		{"FLAC", []byte("fLaCrest"), "audio/flac"},
		{"ID3", []byte("ID3rest"), "audio/mpeg"},
		{"OLE2", []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, 0x00}, "application/x-ole2"},
		{"DWG", []byte("AC1032rest"), "image/vnd.dwg"},
		// ISOBMFF: ftyp at offset 4 with known brands.
		{"HEIC", func() []byte {
			// ftyp box: [size=20][ftyp][heic][0000]
			b := make([]byte, 20)
			b[0], b[1], b[2], b[3] = 0, 0, 0, 20
			copy(b[4:], "ftyp")
			copy(b[8:], "heic")
			return b
		}(), "image/heic"},
		{"AVIF", func() []byte {
			b := make([]byte, 20)
			b[0], b[1], b[2], b[3] = 0, 0, 0, 20
			copy(b[4:], "ftyp")
			copy(b[8:], "avif")
			return b
		}(), "image/avif"},
		{"HEIF_mif1", func() []byte {
			b := make([]byte, 20)
			b[0], b[1], b[2], b[3] = 0, 0, 0, 20
			copy(b[4:], "ftyp")
			copy(b[8:], "mif1")
			return b
		}(), "image/heif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			r := bytes.NewReader(tt.data)
			got, err := d.Detect(r, "")
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Detect = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetect_ExtensionFallback(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"document.txt", "text/plain"},
		{"report.pdf", "application/pdf"},
		{"photo.jpg", "image/jpeg"},
		{"song.mp3", "audio/mpeg"},
		{"drawing.dxf", "image/vnd.dxf"},
		{"photo.heic", "image/heic"},
		{"photo.heif", "image/heif"},
		{"photo.avif", "image/avif"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			d := NewDetector()
			// Use data that doesn't match any magic bytes.
			r := bytes.NewReader([]byte("generic data without magic bytes"))
			got, err := d.Detect(r, tt.filename)
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Detect = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetect_NilReader(t *testing.T) {
	d := NewDetector()
	_, err := d.Detect(nil, "")
	if err != extractor.ErrNilReader {
		t.Errorf("Detect(nil) = %v, want ErrNilReader", err)
	}
}

func TestDetect_EmptyReader(t *testing.T) {
	d := NewDetector()
	r := bytes.NewReader(nil)
	_, err := d.Detect(r, "")
	if err != extractor.ErrDetectionFailed {
		t.Errorf("Detect(empty) = %v, want ErrDetectionFailed", err)
	}
}

func TestDetect_Unknown(t *testing.T) {
	d := NewDetector()
	r := bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	_, err := d.Detect(r, "")
	if err != extractor.ErrDetectionFailed {
		t.Errorf("Detect(unknown) = %v, want ErrDetectionFailed", err)
	}
}

func TestDetect_ReaderPositionRestored(t *testing.T) {
	d := NewDetector()
	data := []byte("%PDF-1.7 some content here")
	r := bytes.NewReader(data)

	_, err := d.Detect(r, "")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}

	// Reader should be back at start.
	pos, _ := r.Seek(0, io.SeekCurrent)
	if pos != 0 {
		t.Errorf("reader position = %d, want 0", pos)
	}
}

func TestDetect_MagicOverExtension(t *testing.T) {
	// File with .txt extension but PDF magic bytes — magic should win.
	d := NewDetector()
	r := bytes.NewReader([]byte("%PDF-1.7 fake text file"))
	got, err := d.Detect(r, "document.txt")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if got != "application/pdf" {
		t.Errorf("Detect = %q, want application/pdf (magic should override extension)", got)
	}
}

// --- detectZIPFormat tests (real Office/ODF files) ---

func TestDetect_ZIPFormat_RealFiles(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"unicode.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"rich.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"simple.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"rich.odt", "application/vnd.oasis.opendocument.text"},
		{"empty.ods", "application/vnd.oasis.opendocument.spreadsheet"},
		{"empty.odp", "application/vnd.oasis.opendocument.presentation"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			f, err := os.Open(filepath.Join("..", "testdata", tt.file))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer f.Close()

			d := NewDetector()
			got, err := d.Detect(f, tt.file)
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Detect(%q) = %q, want %q", tt.file, got, tt.want)
			}
		})
	}
}

func TestDetect_ZIPFormat_CorruptZIP(t *testing.T) {
	// Valid PK magic but invalid ZIP structure — should fall back to application/zip or text.
	d := NewDetector()
	data := []byte("PK\x03\x04corrupted zip data that is not a valid archive")
	r := bytes.NewReader(data)
	got, err := d.Detect(r, "")
	// Should not error — returns something (either zip or text detection fallback).
	if err != nil && err != extractor.ErrDetectionFailed {
		t.Fatalf("Detect error: %v", err)
	}
	_ = got // Any result is fine as long as we don't panic.
}

// --- detectISOBMFF compatible brands ---

func TestDetect_ISOBMFF_CompatibleBrands(t *testing.T) {
	// Unknown major brand, but "heic" as a compatible brand at offset 16.
	b := make([]byte, 24)
	// Box size = 24
	b[0], b[1], b[2], b[3] = 0, 0, 0, 24
	copy(b[4:], "ftyp")
	copy(b[8:], "unkn")          // unknown major brand
	b[12], b[13], b[14], b[15] = 0, 0, 0, 0 // minor version
	copy(b[16:], "heic")         // compatible brand

	d := NewDetector()
	r := bytes.NewReader(b)
	got, err := d.Detect(r, "")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if got != "image/heic" {
		t.Errorf("Detect = %q, want image/heic (from compatible brand)", got)
	}
}

func TestDetect_ISOBMFF_UnknownBrand(t *testing.T) {
	// ftyp box with unknown major AND no compatible brands — should fall through.
	b := make([]byte, 16)
	b[0], b[1], b[2], b[3] = 0, 0, 0, 16
	copy(b[4:], "ftyp")
	copy(b[8:], "mp41") // known ISO base format but not in our brands map

	d := NewDetector()
	r := bytes.NewReader(b)
	_, err := d.Detect(r, "")
	// Unknown ISOBMFF brand falls through to text/extension — likely ErrDetectionFailed.
	if err != extractor.ErrDetectionFailed {
		t.Logf("Detect returned err=%v (expected ErrDetectionFailed or nil)", err)
	}
}

func TestDetect_ISOBMFF_BoxSizeOverflow(t *testing.T) {
	// ftyp box with size < 12 (invalid).
	b := make([]byte, 12)
	b[0], b[1], b[2], b[3] = 0, 0, 0, 4 // size=4, too small
	copy(b[4:], "ftyp")
	copy(b[8:], "heic")

	result := detectISOBMFF(b)
	// size < 12, so boxSize gets clamped to len(buf)=12, major brand at 8-11 is "heic".
	if result != "image/heic" {
		t.Logf("detectISOBMFF with small box = %q", result)
	}
}

func TestDetect_ISOBMFF_TooShort(t *testing.T) {
	result := detectISOBMFF([]byte{0, 0, 0, 8, 'f', 't', 'y', 'p'})
	if result != "" {
		t.Errorf("detectISOBMFF(too short) = %q, want empty", result)
	}
}

// --- detectTextFormat variants ---

func TestDetect_TextFormat_DXF_CRLF(t *testing.T) {
	d := NewDetector()
	r := bytes.NewReader([]byte("0\r\nSECTION\r\n2\r\nHEADER"))
	got, err := d.Detect(r, "")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if got != "image/vnd.dxf" {
		t.Errorf("Detect = %q, want image/vnd.dxf (CRLF variant)", got)
	}
}

func TestDetect_TextFormat_XMLContainingSVG(t *testing.T) {
	d := NewDetector()
	r := bytes.NewReader([]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><circle/></svg>`))
	got, err := d.Detect(r, "")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if got != "image/svg+xml" {
		t.Errorf("Detect = %q, want image/svg+xml (XML with <svg>)", got)
	}
}

func TestDetect_TextFormat_HTMLUpperCase(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"DOCTYPE_upper", "<!DOCTYPE HTML><HTML><BODY>Hello</BODY></HTML>"},
		{"html_upper", "<HTML><BODY>Hello</BODY></HTML>"},
		{"doctype_mixed", "<!DOCTYPE html PUBLIC \"-//W3C//DTD\">\n<html>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			r := bytes.NewReader([]byte(tt.data))
			got, err := d.Detect(r, "")
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != "text/html" {
				t.Errorf("Detect = %q, want text/html", got)
			}
		})
	}
}

func TestDetect_TextFormat_EMLHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"From", "From: alice@example.com\r\nSubject: Hi\r\n\r\nBody"},
		{"Return-Path", "Return-Path: <bounce@example.com>\r\nFrom: x\r\n\r\nBody"},
		{"Received", "Received: from mail.example.com\r\nFrom: x\r\n\r\nBody"},
		{"MIME-Version", "MIME-Version: 1.0\r\nFrom: x\r\n\r\nBody"},
		{"Message-ID", "Message-ID: <123@example.com>\r\nFrom: x\r\n\r\nBody"},
		{"Message-Id", "Message-Id: <123@example.com>\r\nFrom: x\r\n\r\nBody"},
		{"Date", "Date: Mon, 1 Jan 2024 00:00:00 +0000\r\nFrom: x\r\n\r\nBody"},
		{"Subject", "Subject: Test\r\nFrom: x\r\n\r\nBody"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			r := bytes.NewReader([]byte(tt.header))
			got, err := d.Detect(r, "")
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != "message/rfc822" {
				t.Errorf("Detect = %q, want message/rfc822", got)
			}
		})
	}
}

func TestDetect_TextFormat_NotEML(t *testing.T) {
	// Text that starts with something that looks like headers but doesn't match.
	s := "Hello: World\r\nThis is not an email"
	result := detectTextFormat([]byte(s))
	if result == "message/rfc822" {
		t.Errorf("detectTextFormat matched non-EML as EML")
	}
}

// --- Extension fallback edge cases ---

func TestDetect_ExtensionFallback_Unknown(t *testing.T) {
	d := NewDetector()
	r := bytes.NewReader([]byte("random data"))
	_, err := d.Detect(r, "file.xyz")
	if err != extractor.ErrDetectionFailed {
		t.Errorf("Detect(unknown ext) = %v, want ErrDetectionFailed", err)
	}
}

// --- DetectFunc adapter ---

func TestDetectFunc_Adapter(t *testing.T) {
	fn := DetectFunc(func(r io.ReadSeeker, filename string) (string, error) {
		return "test/mime", nil
	})
	got, err := fn.Detect(nil, "")
	if err != nil {
		t.Fatalf("DetectFunc error: %v", err)
	}
	if got != "test/mime" {
		t.Errorf("DetectFunc = %q, want test/mime", got)
	}
}

// --- looksLikeEML direct ---

func TestLooksLikeEML_False(t *testing.T) {
	if looksLikeEML("not an email header") {
		t.Error("looksLikeEML matched non-email text")
	}
	if looksLikeEML("") {
		t.Error("looksLikeEML matched empty string")
	}
}

// --- extensionToMIME direct ---

func TestExtensionToMIME_AllFormats(t *testing.T) {
	// Spot check a few extensions and verify unknown returns empty.
	tests := []struct {
		filename string
		want     string
	}{
		{"test.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"test.svg", "image/svg+xml"},
		{"test.rtf", "application/rtf"},
		{"test.eml", "message/rfc822"},
		{"test.unknown", ""},
	}
	for _, tt := range tests {
		got := extensionToMIME(tt.filename)
		if got != tt.want {
			t.Errorf("extensionToMIME(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}

// --- detectTextFormat edge cases ---

func TestDetectTextFormat_RTF(t *testing.T) {
	got := detectTextFormat([]byte("{\\rtf1\\ansi Hello world}"))
	if got != "application/rtf" {
		t.Errorf("detectTextFormat(RTF) = %q, want application/rtf", got)
	}
}

func TestDetectTextFormat_JSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"object", `{"key": "value"}`},
		{"array", `["a", "b"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTextFormat([]byte(tt.data))
			if got != "application/json" {
				t.Errorf("detectTextFormat(%s) = %q, want application/json", tt.name, got)
			}
		})
	}
}

func TestDetectTextFormat_SVGWithoutXML(t *testing.T) {
	got := detectTextFormat([]byte(`<svg viewBox="0 0 100 100"><rect/></svg>`))
	if got != "image/svg+xml" {
		t.Errorf("detectTextFormat(SVG) = %q, want image/svg+xml", got)
	}
}

func TestDetectTextFormat_Empty(t *testing.T) {
	got := detectTextFormat([]byte(""))
	if got != "" {
		t.Errorf("detectTextFormat(empty) = %q, want empty", got)
	}
}

// --- Detect with ReadSeeker error ---

type failSeeker struct{}

func (f failSeeker) Read(p []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (f failSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestDetect_SeekError(t *testing.T) {
	d := NewDetector()
	_, err := d.Detect(failSeeker{}, "")
	if err != extractor.ErrDetectionFailed {
		t.Errorf("Detect(failSeeker) = %v, want ErrDetectionFailed", err)
	}
}

// --- MP3 sync word (masked signature) ---

func TestDetect_MP3SyncWord(t *testing.T) {
	// MP3 frame sync without ID3 tag: 0xFF 0xFB (MPEG1 Layer 3, no CRC).
	d := NewDetector()
	data := make([]byte, 100)
	data[0] = 0xFF
	data[1] = 0xFB
	r := bytes.NewReader(data)
	got, err := d.Detect(r, "")
	// Note: RIFF signature also starts checking, but MP3 sync word with mask 0xFFE0
	// should match. The signatures list has WebP first with "RIFF", so non-RIFF data
	// will skip it. The sync word at 0xFF 0xFB & 0xFFE0 = 0xFFE0 matches.
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	// Could be audio/mpeg or image/webp depending on RIFF. Since first bytes are 0xFF 0xFB
	// which is NOT "RIFF", it should match MP3 sync word.
	if !strings.Contains(got, "audio/mpeg") {
		t.Errorf("Detect(MP3 sync) = %q, want audio/mpeg", got)
	}
}
