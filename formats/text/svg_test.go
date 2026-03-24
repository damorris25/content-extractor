// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func TestSVGCompliance(t *testing.T) {
	ext := NewSVG()
	extractortest.RunComplianceTests(t, ext, "image/svg+xml", []byte("<svg><text>hello</text></svg>"))
}

func TestSVG_BasicText(t *testing.T) {
	ext := NewSVG()
	input := `<svg><text x="10" y="20">Hello SVG</text></svg>`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "image/svg+xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "Hello SVG" {
		t.Errorf("Text = %q, want %q", result.Text, "Hello SVG")
	}
	if result.ContentType != "image/svg+xml" {
		t.Errorf("ContentType = %q, want image/svg+xml", result.ContentType)
	}
}

func TestSVG_TitleAndDesc(t *testing.T) {
	ext := NewSVG()
	input := `<svg><title>My Title</title><desc>My Description</desc><text>content</text></svg>`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "image/svg+xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "content" {
		t.Errorf("Text = %q, want %q", result.Text, "content")
	}
	title := result.Metadata.Get(metadata.KeyTitle)
	if title != "My Title" {
		t.Errorf("Title = %q, want %q", title, "My Title")
	}
	desc := result.Metadata.Get(metadata.KeyDescription)
	if desc != "My Description" {
		t.Errorf("Description = %q, want %q", desc, "My Description")
	}
}

func TestSVG_TspanAndTextPath(t *testing.T) {
	ext := NewSVG()
	input := `<svg><text><tspan>First</tspan> and <tspan>Second</tspan></text></svg>`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "image/svg+xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "First") || !strings.Contains(result.Text, "Second") {
		t.Errorf("Text = %q, want both 'First' and 'Second'", result.Text)
	}
}

func TestSVG_EmptyInput(t *testing.T) {
	ext := NewSVG()
	result, err := ext.Extract(context.Background(), strings.NewReader(""), "image/svg+xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestSVG_UnsupportedMIME(t *testing.T) {
	ext := NewSVG()
	_, err := ext.Extract(context.Background(), strings.NewReader("<svg/>"), "text/html")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestSVG_Fixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/simple.svg")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ext := NewSVG()
	result, err := ext.Extract(context.Background(), strings.NewReader(string(data)), "image/svg+xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello SVG") {
		t.Errorf("Text should contain 'Hello SVG', got %q", result.Text)
	}
	title := result.Metadata.Get(metadata.KeyTitle)
	if title != "Test SVG" {
		t.Errorf("Title = %q, want %q", title, "Test SVG")
	}
	desc := result.Metadata.Get(metadata.KeyDescription)
	if desc != "A simple test SVG with text elements" {
		t.Errorf("Description = %q, want %q", desc, "A simple test SVG with text elements")
	}
}
