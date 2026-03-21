// CONTRACT:C1-EXTRACTOR.1.0
package office

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func TestXlsx_Compliance(t *testing.T) {
	ext := NewXlsx()
	data := loadTestFile(t, "simple.xlsx")
	extractortest.RunComplianceTests(t, ext, mimeXlsx, data)
}

func TestXlsx_ExtractText(t *testing.T) {
	ext := NewXlsx()
	data := loadTestFile(t, "simple.xlsx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeXlsx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain headers and data.
	for _, want := range []string{"Name", "Age", "City", "Alice", "30", "Springfield"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
	// Cells in a row should be tab-separated.
	if !strings.Contains(result.Text, "Name\tAge\tCity") {
		t.Errorf("expected tab-separated headers, got: %q", result.Text)
	}
}

func TestXlsx_MultiSheet(t *testing.T) {
	ext := NewXlsx()
	data := loadTestFile(t, "multisheet.xlsx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeXlsx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Product", "Widget", "9.99", "Total"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
}

func TestXlsx_EmptySheet(t *testing.T) {
	ext := NewXlsx()
	data := loadTestFile(t, "empty.xlsx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeXlsx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestXlsx_Metadata(t *testing.T) {
	ext := NewXlsx()
	data := loadTestFile(t, "simple.xlsx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeXlsx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Sales Data" {
		t.Errorf("Title = %q, want %q", got, "Sales Data")
	}
	if got := result.Metadata.Get(metadata.KeyCreator); got != "Spreadsheet Author" {
		t.Errorf("Creator = %q, want %q", got, "Spreadsheet Author")
	}
	if got := result.Metadata.Get(metadata.KeyFormat); got != mimeXlsx {
		t.Errorf("Format = %q, want %q", got, mimeXlsx)
	}
}

func TestXlsx_CorruptInput(t *testing.T) {
	ext := NewXlsx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("not a zip")), mimeXlsx)
	if err == nil {
		t.Fatal("expected error for corrupt input")
	}
}

func TestXlsx_NilReader(t *testing.T) {
	ext := NewXlsx()
	_, err := ext.Extract(context.Background(), nil, mimeXlsx)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestXlsx_UnsupportedMIME(t *testing.T) {
	ext := NewXlsx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
