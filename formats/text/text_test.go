// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
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
