// CONTRACT:C2-DETECT.1.0
package detect

import (
	"bytes"
	"io"
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
