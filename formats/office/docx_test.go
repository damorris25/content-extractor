// CONTRACT:C1-EXTRACTOR.1.0
package office

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

func TestDocx_Compliance(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "simple.docx")
	extractortest.RunComplianceTests(t, ext, mimeDocx, data)
}

func TestDocx_ExtractText(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "simple.docx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, World!") {
		t.Errorf("Text = %q, want to contain %q", result.Text, "Hello, World!")
	}
}

func TestDocx_MultiParagraph(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "multipar.docx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "First paragraph.") {
		t.Errorf("missing first paragraph in %q", result.Text)
	}
	if !strings.Contains(result.Text, "Second paragraph.") {
		t.Errorf("missing second paragraph in %q", result.Text)
	}
	if !strings.Contains(result.Text, "Third paragraph.") {
		t.Errorf("missing third paragraph in %q", result.Text)
	}
	// Paragraphs should be separated by newlines.
	if !strings.Contains(result.Text, "\n") {
		t.Error("paragraphs not separated by newlines")
	}
}

func TestDocx_EmptyDocument(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "empty.docx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestDocx_Metadata(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "simple.docx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "Test Title"},
		{metadata.KeyCreator, "Test Author"},
		{metadata.KeyApplication, "content-extractor test"},
		{metadata.KeyPageCount, "1"},
		{metadata.KeyWordCount, "2"},
		{metadata.KeyCharCount, "13"},
		{metadata.KeyFormat, mimeDocx},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Date should be normalized to ISO 8601.
	if date := result.Metadata.Get(metadata.KeyDate); date == "" {
		t.Error("Metadata[dc:date] is empty")
	}
}

func TestDocx_CorruptInput(t *testing.T) {
	ext := NewDocx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("not a zip file")), mimeDocx)
	if err == nil {
		t.Fatal("expected error for corrupt input")
	}
}

func TestDocx_NilReader(t *testing.T) {
	ext := NewDocx()
	_, err := ext.Extract(context.Background(), nil, mimeDocx)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestDocx_UnsupportedMIME(t *testing.T) {
	ext := NewDocx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestDocx_ContextCancelled(t *testing.T) {
	ext := NewDocx()
	data := loadTestFile(t, "simple.docx")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, bytes.NewReader(data), mimeDocx)
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
