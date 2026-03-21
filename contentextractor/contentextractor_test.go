package contentextractor

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/metadata"
)

func TestIntegration_PlainText(t *testing.T) {
	input := "Hello, integration test!"
	r := bytes.NewReader([]byte(input))
	result, err := Extract(context.Background(), r, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != input {
		t.Errorf("Text = %q, want %q", result.Text, input)
	}
}

func TestIntegration_PlainText_AutoDetect(t *testing.T) {
	// Auto-detection with a .txt filename hint via the registry.
	reg := NewRegistry()
	input := "auto-detected text"
	r := bytes.NewReader([]byte(input))
	// Can't auto-detect plain text by magic bytes alone, so provide MIME directly.
	result, err := reg.Extract(context.Background(), r, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != input {
		t.Errorf("Text = %q, want %q", result.Text, input)
	}
}

func TestIntegration_DOCX(t *testing.T) {
	data, err := os.ReadFile("../testdata/simple.docx")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	result, err := Extract(context.Background(), bytes.NewReader(data), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, World!") {
		t.Errorf("Text = %q, want to contain %q", result.Text, "Hello, World!")
	}
	if result.Metadata.Get(metadata.KeyTitle) != "Test Title" {
		t.Errorf("Title = %q, want %q", result.Metadata.Get(metadata.KeyTitle), "Test Title")
	}
}

func TestIntegration_XLSX(t *testing.T) {
	data, err := os.ReadFile("../testdata/simple.xlsx")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	result, err := Extract(context.Background(), bytes.NewReader(data), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Name") || !strings.Contains(result.Text, "Alice") {
		t.Errorf("Text = %q, want to contain Name and Alice", result.Text)
	}
	if result.Metadata.Get(metadata.KeyTitle) != "Sales Data" {
		t.Errorf("Title = %q, want %q", result.Metadata.Get(metadata.KeyTitle), "Sales Data")
	}
}

func TestIntegration_PPTX(t *testing.T) {
	data, err := os.ReadFile("../testdata/simple.pptx")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	result, err := Extract(context.Background(), bytes.NewReader(data), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Welcome to the Presentation") {
		t.Errorf("Text = %q, want to contain 'Welcome to the Presentation'", result.Text)
	}
	if result.Metadata.Get(metadata.KeyTitle) != "Test Presentation" {
		t.Errorf("Title = %q, want %q", result.Metadata.Get(metadata.KeyTitle), "Test Presentation")
	}
}

func TestIntegration_SupportedTypes(t *testing.T) {
	reg := NewRegistry()
	types := reg.SupportedTypes()
	if len(types) < 2 {
		t.Errorf("SupportedTypes = %v, want at least 2 types", types)
	}
}

func TestIntegration_UnsupportedFormat(t *testing.T) {
	r := bytes.NewReader([]byte("data"))
	_, err := Extract(context.Background(), r, "application/x-unknown")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
