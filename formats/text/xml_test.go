// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"bytes"
	"context"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
)

func TestXMLCompliance(t *testing.T) {
	ext := NewXML()
	extractortest.RunComplianceTests(t, ext, "application/xml", []byte("<root>hello</root>"))
}

func TestXMLExtract_BasicXML(t *testing.T) {
	ext := NewXML()
	input := `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <title>Hello World</title>
  <body>This is a test.</body>
</root>`
	result, err := ext.Extract(context.Background(), bytes.NewReader([]byte(input)), "application/xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ContentType != "application/xml" {
		t.Errorf("ContentType = %q, want application/xml", result.ContentType)
	}
	// Should contain the text content without tags.
	if result.Text == "" {
		t.Error("Text should not be empty")
	}
	if !contains(result.Text, "Hello World") {
		t.Errorf("Text should contain 'Hello World', got %q", result.Text)
	}
	if !contains(result.Text, "This is a test.") {
		t.Errorf("Text should contain 'This is a test.', got %q", result.Text)
	}
}

func TestXMLExtract_TextXML(t *testing.T) {
	ext := NewXML()
	input := `<note><to>User</to><body>Hello</body></note>`
	result, err := ext.Extract(context.Background(), bytes.NewReader([]byte(input)), "text/xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ContentType != "text/xml" {
		t.Errorf("ContentType = %q, want text/xml", result.ContentType)
	}
	if !contains(result.Text, "User") || !contains(result.Text, "Hello") {
		t.Errorf("Text should contain 'User' and 'Hello', got %q", result.Text)
	}
}

func TestXMLExtract_EmptyReader(t *testing.T) {
	ext := NewXML()
	result, err := ext.Extract(context.Background(), bytes.NewReader(nil), "application/xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestXMLExtract_NilReader(t *testing.T) {
	ext := NewXML()
	_, err := ext.Extract(context.Background(), nil, "application/xml")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestXMLExtract_UnsupportedMIME(t *testing.T) {
	ext := NewXML()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("<x/>")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestXMLExtract_ContextCancelled(t *testing.T) {
	ext := NewXML()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, bytes.NewReader([]byte("<x/>")), "application/xml")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
