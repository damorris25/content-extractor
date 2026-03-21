// CONTRACT:C1-EXTRACTOR.1.0
package cad

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

func TestDXF_Compliance(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.dxf")
	extractortest.RunComplianceTests(t, ext, "image/vnd.dxf", data)
}

func TestDXF_TextExtraction(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.dxf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/vnd.dxf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTexts := []string{
		"Floor Plan - Level 1",
		"Main Entrance",
		"Conference Room",
		"Emergency Exit",
		"DoorBlock",
		"12.5m",
	}
	for _, want := range expectedTexts {
		if !strings.Contains(result.Text, want) {
			t.Errorf("Text missing %q, got:\n%s", want, result.Text)
		}
	}
}

func TestDXF_MTEXTFormatting(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.dxf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/vnd.dxf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// \P should be converted to newline.
	if !strings.Contains(result.Text, "Capacity: 20") {
		t.Errorf("MTEXT formatting not cleaned, got:\n%s", result.Text)
	}
}

func TestDXF_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.dxf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/vnd.dxf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"cad:version", "AC1027"},
		{"cad:units", "Millimeters"},
		{metadata.KeyFormat, "image/vnd.dxf"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestDXF_Layers(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.dxf")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/vnd.dxf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	layers := result.Metadata.Get("cad:layers")
	for _, want := range []string{"Walls", "Doors", "Electrical"} {
		if !strings.Contains(layers, want) {
			t.Errorf("layers missing %q, got %q", want, layers)
		}
	}
}

func TestDXF_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "image/vnd.dxf")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestDXF_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestDXF_CleanMTEXT(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`Hello\PWorld`, "Hello\nWorld"},
		{`{\fArial;Bold text}`, "Bold text"},
		{`Normal\\Escaped`, `Normal\Escaped`},
		{`{\H1.5;Sized}`, "Sized"},
		{`Simple text`, "Simple text"},
	}
	for _, tt := range tests {
		got := cleanMTEXT(tt.input)
		if got != tt.want {
			t.Errorf("cleanMTEXT(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
