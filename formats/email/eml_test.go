// CONTRACT:C1-EXTRACTOR.1.0
package email

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
)

func TestCompliance(t *testing.T) {
	ext := New()
	sample := []byte("From: test@example.com\r\nSubject: Test\r\n\r\nHello world")
	extractortest.RunComplianceTests(t, ext, "message/rfc822", sample)
}

func TestExtract_SimpleEML(t *testing.T) {
	ext := New()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "simple.eml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ContentType != "message/rfc822" {
		t.Errorf("ContentType = %q, want message/rfc822", result.ContentType)
	}

	// Check metadata.
	if got := result.Metadata.Get("dc:title"); got != "Test Email" {
		t.Errorf("dc:title = %q, want 'Test Email'", got)
	}
	if got := result.Metadata.Get("dc:creator"); got != "sender@example.com" {
		t.Errorf("dc:creator = %q, want 'sender@example.com'", got)
	}
	if got := result.Metadata.Get("email:to"); got != "recipient@example.com" {
		t.Errorf("email:to = %q, want 'recipient@example.com'", got)
	}
	if got := result.Metadata.Get("dc:date"); got != "2024-01-01T12:00:00Z" {
		t.Errorf("dc:date = %q, want '2024-01-01T12:00:00Z'", got)
	}

	// Check body text.
	if !strings.Contains(result.Text, "simple test email") {
		t.Errorf("Text should contain 'simple test email', got %q", result.Text)
	}
}

func TestExtract_MultipartEML(t *testing.T) {
	ext := New()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "multipart.eml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should prefer plain text part.
	if !strings.Contains(result.Text, "plain text version") {
		t.Errorf("Text should contain 'plain text version', got %q", result.Text)
	}

	// Should have Cc metadata.
	if got := result.Metadata.Get("email:cc"); got != "carol@example.com" {
		t.Errorf("email:cc = %q, want 'carol@example.com'", got)
	}

	// Should have subject.
	if got := result.Metadata.Get("dc:title"); got != "Multipart Test" {
		t.Errorf("dc:title = %q, want 'Multipart Test'", got)
	}
}

func TestExtract_EmptyBody(t *testing.T) {
	ext := New()
	input := "From: test@example.com\r\nSubject: Empty\r\n\r\n"
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestExtract_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "message/rfc822")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestExtract_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), strings.NewReader("test"), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestExtract_ContextCancelled(t *testing.T) {
	ext := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, strings.NewReader("From: x\r\n\r\ntest"), "message/rfc822")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
