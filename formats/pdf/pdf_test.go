// CONTRACT:C1-EXTRACTOR.1.0
package pdf

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

func TestPDF_Compliance(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.pdf")
	extractortest.RunComplianceTests(t, ext, mimePDF, data)
}

func TestPDF_ExtractText(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.pdf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello") {
		t.Errorf("Text = %q, want to contain 'Hello'", result.Text)
	}
	if result.ContentType != mimePDF {
		t.Errorf("ContentType = %q, want %q", result.ContentType, mimePDF)
	}
}

func TestPDF_MultiPage(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "multipage.pdf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Page one") {
		t.Errorf("Text missing 'Page one', got: %q", result.Text)
	}
	if !strings.Contains(result.Text, "Page two") {
		t.Errorf("Text missing 'Page two', got: %q", result.Text)
	}
}

func TestPDF_PageCount(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "multipage.pdf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyPageCount); got != "2" {
		t.Errorf("PageCount = %q, want %q", got, "2")
	}
}

func TestPDF_EmptyPage(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "empty.pdf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestPDF_CorruptInput(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("not a pdf")), mimePDF)
	if err == nil {
		t.Fatal("expected error for corrupt input")
	}
}

func TestPDF_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, mimePDF)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestPDF_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
