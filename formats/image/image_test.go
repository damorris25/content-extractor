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
