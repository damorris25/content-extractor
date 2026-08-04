// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
)

func TestCompliance(t *testing.T) {
	ext := New()
	extractortest.RunComplianceTests(t, ext, "text/plain", []byte("hello world"))
}

func TestExtract_BasicText(t *testing.T) {
	ext := New()
	input := "Hello, World!\nLine two."
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != input {
		t.Errorf("Text = %q, want %q", result.Text, input)
	}
	if result.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", result.ContentType)
	}
}

func TestExtract_EmptyReader(t *testing.T) {
	ext := New()
	result, err := ext.Extract(context.Background(), bytes.NewReader(nil), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestExtract_InvalidUTF8(t *testing.T) {
	ext := New()
	input := []byte("hello\x80world")
	result, err := ext.Extract(context.Background(), bytes.NewReader(input), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "hello\uFFFDworld"
	if result.Text != want {
		t.Errorf("Text = %q, want %q", result.Text, want)
	}
}

func TestExtract_MaxBytesLimit(t *testing.T) {
	ext := &Extractor{MaxBytes: 10}
	input := strings.Repeat("a", 100)
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Text) != 10 {
		t.Errorf("Text length = %d, want 10", len(result.Text))
	}
}

func TestExtract_ContextCancelled(t *testing.T) {
	ext := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, strings.NewReader("hello"), "text/plain")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestExtract_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), strings.NewReader("hello"), "application/pdf")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestExtract_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "text/plain")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestTextLike_MIMETypes(t *testing.T) {
	tests := []struct {
		mime  string
		input string
	}{
		{"application/json", `{"key": "value"}`},
		{"text/csv", "a,b,c\n1,2,3"},
		{"text/tab-separated-values", "a\tb\tc\n1\t2\t3"},
		{"application/x-yaml", "key: value"},
		{"text/markdown", "# Hello\n\nWorld"},
	}

	ext := NewTextLike(
		"application/json",
		"text/csv",
		"text/tab-separated-values",
		"application/x-yaml",
		"text/markdown",
	)

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			result, err := ext.Extract(context.Background(), strings.NewReader(tt.input), tt.mime)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.input {
				t.Errorf("Text = %q, want %q", result.Text, tt.input)
			}
			if result.ContentType != tt.mime {
				t.Errorf("ContentType = %q, want %q", result.ContentType, tt.mime)
			}
		})
	}
}

func TestTextLike_RejectsPlainText(t *testing.T) {
	ext := NewTextLike("application/json")
	_, err := ext.Extract(context.Background(), strings.NewReader("hello"), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestTextLike_Fixtures(t *testing.T) {
	tests := []struct {
		mime     string
		file     string
		contains string
	}{
		{"application/json", "../../testdata/simple.json", `"content-extractor"`},
		{"text/csv", "../../testdata/simple.csv", "Alice,30,New York"},
		{"application/x-yaml", "../../testdata/simple.yaml", "name: content-extractor"},
		{"text/markdown", "../../testdata/simple.md", "# Content Extractor"},
	}

	ext := NewTextLike(
		"application/json",
		"text/csv",
		"text/tab-separated-values",
		"application/x-yaml",
		"text/markdown",
	)

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			result, err := ext.Extract(context.Background(), strings.NewReader(string(data)), tt.mime)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result.Text, tt.contains) {
				t.Errorf("Text should contain %q", tt.contains)
			}
		})
	}
}
