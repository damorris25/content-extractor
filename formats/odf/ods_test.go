// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
	"github.com/damorris25/content-extractor/metadata"
)

func TestODS_Compliance(t *testing.T) {
	ext := NewODS()
	data := loadTestFile(t, "simple.ods")
	extractortest.RunComplianceTests(t, ext, mimeODS, data)
}

func TestODS_ExtractText(t *testing.T) {
	ext := NewODS()
	data := loadTestFile(t, "simple.ods")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Name", "Age", "City", "Alice", "30", "Springfield"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
	// Cells should be tab-separated.
	if !strings.Contains(result.Text, "\t") {
		t.Error("cells not tab-separated")
	}
}

func TestODS_MultiSheet(t *testing.T) {
	ext := NewODS()
	data := loadTestFile(t, "multisheet.ods")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Product", "Widget", "9.99", "Total"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
}

func TestODS_Empty(t *testing.T) {
	ext := NewODS()
	data := loadTestFile(t, "empty.ods")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestODS_Metadata(t *testing.T) {
	ext := NewODS()
	data := loadTestFile(t, "simple.ods")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Test Document" {
		t.Errorf("Title = %q, want %q", got, "Test Document")
	}
}

func TestODS_NilReader(t *testing.T) {
	ext := NewODS()
	_, err := ext.Extract(context.Background(), nil, mimeODS)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestODS_UnsupportedMIME(t *testing.T) {
	ext := NewODS()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
