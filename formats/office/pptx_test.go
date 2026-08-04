// CONTRACT:C1-EXTRACTOR.1.0
package office

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
	"github.com/damorris25/content-extractor/metadata"
)

func TestPptx_Compliance(t *testing.T) {
	ext := NewPptx()
	data := loadTestFile(t, "simple.pptx")
	extractortest.RunComplianceTests(t, ext, mimePptx, data)
}

func TestPptx_ExtractText(t *testing.T) {
	ext := NewPptx()
	data := loadTestFile(t, "simple.pptx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePptx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Welcome to the Presentation", "By Test Author"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
}

func TestPptx_MultiSlide(t *testing.T) {
	ext := NewPptx()
	data := loadTestFile(t, "multislide.pptx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePptx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Introduction", "Main Content", "Bullet point one", "Conclusion", "Thank you!"} {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got: %q", want, result.Text)
		}
	}
	// Slides should be separated by double newlines.
	if !strings.Contains(result.Text, "\n\n") {
		t.Error("slides not separated by blank lines")
	}
}

func TestPptx_EmptySlide(t *testing.T) {
	ext := NewPptx()
	data := loadTestFile(t, "empty.pptx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePptx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestPptx_Metadata(t *testing.T) {
	ext := NewPptx()
	data := loadTestFile(t, "simple.pptx")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), mimePptx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Test Presentation" {
		t.Errorf("Title = %q, want %q", got, "Test Presentation")
	}
	if got := result.Metadata.Get(metadata.KeyCreator); got != "Slide Author" {
		t.Errorf("Creator = %q, want %q", got, "Slide Author")
	}
	if got := result.Metadata.Get(metadata.KeyFormat); got != mimePptx {
		t.Errorf("Format = %q, want %q", got, mimePptx)
	}
}

func TestPptx_CorruptInput(t *testing.T) {
	ext := NewPptx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("not a zip")), mimePptx)
	if err == nil {
		t.Fatal("expected error for corrupt input")
	}
}

func TestPptx_NilReader(t *testing.T) {
	ext := NewPptx()
	_, err := ext.Extract(context.Background(), nil, mimePptx)
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestPptx_UnsupportedMIME(t *testing.T) {
	ext := NewPptx()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}
