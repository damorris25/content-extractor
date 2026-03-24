// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func loadTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("failed to load test file %s: %v", name, err)
	}
	return data
}

func TestImage_Compliance_JPEG(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.jpg")
	extractortest.RunComplianceTests(t, ext, "image/jpeg", data)
}

func TestImage_Compliance_PNG(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.png")
	extractortest.RunComplianceTests(t, ext, "image/png", data)
}

func TestImage_Compliance_GIF(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.gif")
	extractortest.RunComplianceTests(t, ext, "image/gif", data)
}

func TestImage_JPEG_Dimensions(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.jpg")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get("exif:width"); got != "100" {
		t.Errorf("width = %q, want %q", got, "100")
	}
	if got := result.Metadata.Get("exif:height"); got != "50" {
		t.Errorf("height = %q, want %q", got, "50")
	}
	if got := result.Metadata.Get(metadata.KeyFormat); got != "image/jpeg" {
		t.Errorf("format = %q, want %q", got, "image/jpeg")
	}
}

func TestImage_PNG_Dimensions(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.png")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get("exif:width"); got != "200" {
		t.Errorf("width = %q, want %q", got, "200")
	}
	if got := result.Metadata.Get("exif:height"); got != "150" {
		t.Errorf("height = %q, want %q", got, "150")
	}
}

func TestImage_PNG_TextChunks(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "text.png")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "Green Square"},
		{metadata.KeyCreator, "Test Generator"},
		{metadata.KeyDescription, "A 64x64 green square for testing"},
		{metadata.KeyApplication, "gen-test-images"},
		{"png:title", "Green Square"},
		{"png:author", "Test Generator"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestImage_GIF_Dimensions(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.gif")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/gif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get("exif:width"); got != "32" {
		t.Errorf("width = %q, want %q", got, "32")
	}
	if got := result.Metadata.Get("exif:height"); got != "32" {
		t.Errorf("height = %q, want %q", got, "32")
	}
}

func TestImage_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "image/jpeg")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestImage_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestImage_EmptyReader(t *testing.T) {
	ext := New()
	result, err := ext.Extract(context.Background(), bytes.NewReader(nil), "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestImage_TextSummary(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "text.png")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Text should contain a summary of key metadata.
	if result.Text == "" {
		t.Error("Text summary is empty, expected metadata summary")
	}
}

// --- Contract error path tests ---

func TestImage_DefaultMaxSize(t *testing.T) {
	ext := &Extractor{MaxSize: 0}
	if got := ext.maxSize(); got != DefaultMaxImageSize {
		t.Errorf("maxSize() = %d, want %d", got, DefaultMaxImageSize)
	}
}

func TestImage_CustomMaxSize(t *testing.T) {
	ext := &Extractor{MaxSize: 1024}
	if got := ext.maxSize(); got != 1024 {
		t.Errorf("maxSize() = %d, want 1024", got)
	}
}

func TestImage_FileTooLarge(t *testing.T) {
	ext := &Extractor{MaxSize: 10}
	data := make([]byte, 20) // larger than MaxSize
	// Give it a valid JPEG header so it doesn't fail on MIME check.
	copy(data, []byte{0xFF, 0xD8, 0xFF, 0xE0})
	_, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/jpeg")
	if err != extractor.ErrFileTooLarge {
		t.Errorf("error = %v, want ErrFileTooLarge", err)
	}
}

func TestImage_ContextCancelled(t *testing.T) {
	ext := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, bytes.NewReader([]byte{0xFF, 0xD8}), "image/jpeg")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

// --- cleanTag tests ---

func TestCleanTag(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"quoted"`, "quoted"},
		{"  spaces  ", "spaces"},
		{`"with quotes"`, "with quotes"},
		{"unquoted", "unquoted"},
		{`""`, ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := cleanTag(tt.input)
		if got != tt.want {
			t.Errorf("cleanTag(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- PNG text chunk edge cases ---

func TestImage_PNG_ChunkOverflow(t *testing.T) {
	// Build a minimal PNG with a tEXt chunk that claims more bytes than available.
	var buf []byte
	// PNG signature
	buf = append(buf, 0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n')
	// tEXt chunk with size 0xFFFF (larger than remaining data)
	buf = append(buf, 0x00, 0x00, 0xFF, 0xFF) // chunk length (65535)
	buf = append(buf, 't', 'E', 'X', 't')     // chunk type
	buf = append(buf, make([]byte, 20)...)     // only 20 bytes of payload (not 65535)
	buf = append(buf, 0, 0, 0, 0)             // CRC

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractPNGText(buf, result)
	// Should not panic — overflow chunk should be skipped.
}

func TestImage_PNG_WellKnownKeys(t *testing.T) {
	// Build a PNG with tEXt chunks for all well-known keys.
	var buf []byte
	// PNG signature
	buf = append(buf, 0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n')

	// Helper: append a tEXt chunk.
	addText := func(key, val string) {
		payload := append([]byte(key), 0)
		payload = append(payload, []byte(val)...)
		size := len(payload)
		buf = append(buf, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))
		buf = append(buf, 't', 'E', 'X', 't')
		buf = append(buf, payload...)
		buf = append(buf, 0, 0, 0, 0) // CRC (ignored for extraction)
	}

	addText("Title", "My Image")
	addText("Author", "Test Author")
	addText("Description", "A test image")
	addText("Software", "TestGen")

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractPNGText(buf, result)

	checks := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "My Image"},
		{metadata.KeyCreator, "Test Author"},
		{metadata.KeyDescription, "A test image"},
		{metadata.KeyApplication, "TestGen"},
		{"png:title", "My Image"},
		{"png:author", "Test Author"},
		{"png:description", "A test image"},
		{"png:software", "TestGen"},
	}
	for _, c := range checks {
		if got := result.Metadata.Get(c.key); got != c.want {
			t.Errorf("Metadata[%q] = %q, want %q", c.key, got, c.want)
		}
	}
}

// --- buildTextSummary ---

func TestBuildTextSummary_Empty(t *testing.T) {
	m := make(metadata.Metadata)
	got := buildTextSummary(m)
	if got != "" {
		t.Errorf("buildTextSummary(empty) = %q, want empty", got)
	}
}

func TestBuildTextSummary_WithData(t *testing.T) {
	m := make(metadata.Metadata)
	m.Set(metadata.KeyTitle, "Test Image")
	m.Set("exif:make", "Canon")
	got := buildTextSummary(m)
	if got == "" {
		t.Error("buildTextSummary should not be empty")
	}
}
