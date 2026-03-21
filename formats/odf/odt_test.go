// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"bytes"
	"context"
	"os"
	"strings"
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

func TestODT_Compliance(t *testing.T) {
	ext := NewODT()
	data := loadTestFile(t, "simple.odt")
	extractortest.RunComplianceTests(t, ext, mimeODT, data)
}

func TestODT_ExtractText(t *testing.T) {
	ext := NewODT()
	data := loadTestFile(t, "simple.odt")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, ODT World!") {
		t.Errorf("Text = %q, want to contain %q", result.Text, "Hello, ODT World!")
	}
}

func TestODT_MultiParagraph(t *testing.T) {
	ext := NewODT()
	data := loadTestFile(t, "multipar.odt")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"First paragraph.", "Second paragraph.", "A heading", "Third paragraph."} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
	if !strings.Contains(result.Text, "\n") {
		t.Error("paragraphs not separated by newlines")
	}
}

func TestODT_Empty(t *testing.T) {
	ext := NewODT()
	data := loadTestFile(t, "empty.odt")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestODT_Metadata(t *testing.T) {
	ext := NewODT()
	data := loadTestFile(t, "simple.odt")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Test Document" {
		t.Errorf("Title = %q, want %q", got, "Test Document")
	}
	if got := result.Metadata.Get(metadata.KeyCreator); got != "ODF Author" {
		t.Errorf("Creator = %q, want %q", got, "ODF Author")
	}
	if got := result.Metadata.Get(metadata.KeyApplication); got != "content-extractor test" {
		t.Errorf("Application = %q, want %q", got, "content-extractor test")
	}
}

func TestODT_NilReader(t *testing.T) {
	ext := NewODT()
	_, err := ext.Extract(context.Background(), nil, mimeODT)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestODT_UnsupportedMIME(t *testing.T) {
	ext := NewODT()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
