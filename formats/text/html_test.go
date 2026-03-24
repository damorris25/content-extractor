// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"bytes"
	"context"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
)

func TestHTMLCompliance(t *testing.T) {
	ext := NewHTML()
	extractortest.RunComplianceTests(t, ext, "text/html", []byte("<html><body>hello</body></html>"))
}

func TestHTMLExtract_BasicHTML(t *testing.T) {
	ext := NewHTML()
	input := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
  <h1>Hello World</h1>
  <p>This is a test paragraph.</p>
</body>
</html>`
	result, err := ext.Extract(context.Background(), bytes.NewReader([]byte(input)), "text/html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ContentType != "text/html" {
		t.Errorf("ContentType = %q, want text/html", result.ContentType)
	}
	if !contains(result.Text, "Hello World") {
		t.Errorf("Text should contain 'Hello World', got %q", result.Text)
	}
	if !contains(result.Text, "This is a test paragraph.") {
		t.Errorf("Text should contain paragraph text, got %q", result.Text)
	}
}

func TestHTMLExtract_TitleExtraction(t *testing.T) {
	ext := NewHTML()
	input := `<html><head><title>My Title</title></head><body><p>Content</p></body></html>`
	result, err := ext.Extract(context.Background(), bytes.NewReader([]byte(input)), "text/html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := result.Metadata.Get("dc:title")
	if title != "My Title" {
		t.Errorf("title = %q, want 'My Title'", title)
	}
}

func TestHTMLExtract_SkipsScriptAndStyle(t *testing.T) {
	ext := NewHTML()
	input := `<html><body>
<script>var x = 1;</script>
<style>.foo { color: red; }</style>
<p>Visible text</p>
</body></html>`
	result, err := ext.Extract(context.Background(), bytes.NewReader([]byte(input)), "text/html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(result.Text, "var x") {
		t.Error("Text should not contain script content")
	}
	if contains(result.Text, "color: red") {
		t.Error("Text should not contain style content")
	}
	if !contains(result.Text, "Visible text") {
		t.Errorf("Text should contain 'Visible text', got %q", result.Text)
	}
}

func TestHTMLExtract_EmptyReader(t *testing.T) {
	ext := NewHTML()
	result, err := ext.Extract(context.Background(), bytes.NewReader(nil), "text/html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestHTMLExtract_NilReader(t *testing.T) {
	ext := NewHTML()
	_, err := ext.Extract(context.Background(), nil, "text/html")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestHTMLExtract_UnsupportedMIME(t *testing.T) {
	ext := NewHTML()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("<html/>")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestHTMLExtract_ContextCancelled(t *testing.T) {
	ext := NewHTML()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, bytes.NewReader([]byte("<html/>")), "text/html")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
