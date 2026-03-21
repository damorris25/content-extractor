// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func TestODP_Compliance(t *testing.T) {
	ext := NewODP()
	data := loadTestFile(t, "simple.odp")
	extractortest.RunComplianceTests(t, ext, mimeODP, data)
}

func TestODP_ExtractText(t *testing.T) {
	ext := NewODP()
	data := loadTestFile(t, "simple.odp")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Welcome to the Presentation", "By Test Author"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
}

func TestODP_MultiSlide(t *testing.T) {
	ext := NewODP()
	data := loadTestFile(t, "multislide.odp")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Introduction", "Main Content", "Details here", "Conclusion"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
	// Slides should be separated by double newlines.
	if !strings.Contains(result.Text, "\n\n") {
		t.Error("slides not separated by blank lines")
	}
}

func TestODP_Empty(t *testing.T) {
	ext := NewODP()
	data := loadTestFile(t, "empty.odp")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestODP_Metadata(t *testing.T) {
	ext := NewODP()
	data := loadTestFile(t, "simple.odp")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimeODP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Test Document" {
		t.Errorf("Title = %q, want %q", got, "Test Document")
	}
	if got := result.Metadata.Get(metadata.KeyCreator); got != "ODF Author" {
		t.Errorf("Creator = %q, want %q", got, "ODF Author")
	}
}

func TestODP_NilReader(t *testing.T) {
	ext := NewODP()
	_, err := ext.Extract(context.Background(), nil, mimeODP)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestODP_UnsupportedMIME(t *testing.T) {
	ext := NewODP()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
