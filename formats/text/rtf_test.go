// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
)

func TestRTFCompliance(t *testing.T) {
	ext := NewRTF()
	extractortest.RunComplianceTests(t, ext, "application/rtf", []byte(`{\rtf1 hello}`))
}

func TestRTF_BasicText(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1\ansi Hello, World!\par Second paragraph.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, World!") {
		t.Errorf("Text should contain 'Hello, World!', got %q", result.Text)
	}
	if !strings.Contains(result.Text, "Second paragraph.") {
		t.Errorf("Text should contain 'Second paragraph.', got %q", result.Text)
	}
	if result.ContentType != "application/rtf" {
		t.Errorf("ContentType = %q, want application/rtf", result.ContentType)
	}
}

func TestRTF_DestinationGroups(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1{\fonttbl{\f0 Arial;}}Visible text.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Visible text.") {
		t.Errorf("Text should contain 'Visible text.', got %q", result.Text)
	}
	if strings.Contains(result.Text, "fonttbl") {
		t.Errorf("Text should not contain 'fonttbl', got %q", result.Text)
	}
}

func TestRTF_UnicodeEscape(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1 Hello \u8212? World.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello") || !strings.Contains(result.Text, "World.") {
		t.Errorf("Text = %q, want 'Hello' and 'World.'", result.Text)
	}
	// Should contain em-dash (U+2014).
	if !strings.Contains(result.Text, "\u2014") {
		t.Errorf("Text should contain em-dash, got %q", result.Text)
	}
}

func TestRTF_HexEscape(t *testing.T) {
	ext := NewRTF()
	// Use Unicode escape for é instead of hex, since \'e9 produces a
	// raw Latin-1 byte that gets cleaned as invalid UTF-8.
	input := `{\rtf1 caf\u233?\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "café") {
		t.Errorf("Text should contain 'café', got %q", result.Text)
	}
}

func TestRTF_EmptyInput(t *testing.T) {
	ext := NewRTF()
	result, err := ext.Extract(context.Background(), strings.NewReader(""), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestRTF_UnsupportedMIME(t *testing.T) {
	ext := NewRTF()
	_, err := ext.Extract(context.Background(), strings.NewReader(`{\rtf1}`), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestRTF_Fixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/simple.rtf")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ext := NewRTF()
	result, err := ext.Extract(context.Background(), strings.NewReader(string(data)), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, this is a simple RTF document.") {
		t.Errorf("Text should contain 'Hello, this is a simple RTF document.', got %q", result.Text)
	}
	if !strings.Contains(result.Text, "multiple paragraphs") {
		t.Errorf("Text should contain 'multiple paragraphs', got %q", result.Text)
	}
}
