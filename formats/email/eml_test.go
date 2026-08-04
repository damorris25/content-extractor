// CONTRACT:C1-EXTRACTOR.1.0
package email

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
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

func TestExtract_DefaultMaxBytes(t *testing.T) {
	ext := &Extractor{MaxBytes: 0}
	got := ext.maxBytes()
	if got != DefaultMaxBytes {
		t.Errorf("maxBytes() = %d, want %d", got, DefaultMaxBytes)
	}
}

func TestExtract_CustomMaxBytes(t *testing.T) {
	ext := &Extractor{MaxBytes: 1024}
	got := ext.maxBytes()
	if got != 1024 {
		t.Errorf("maxBytes() = %d, want 1024", got)
	}
}

func TestExtract_HTMLOnlyMultipart(t *testing.T) {
	// Multipart with only HTML part (no plain text) — should strip HTML.
	input := "From: test@example.com\r\n" +
		"Subject: HTML Only\r\n" +
		"Content-Type: multipart/alternative; boundary=\"HTMLBND\"\r\n" +
		"\r\n" +
		"--HTMLBND\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		"<html><body><p>Hello <b>World</b></p><script>alert('x')</script></body></html>\r\n" +
		"--HTMLBND--\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain visible text, not HTML tags or script content.
	if !strings.Contains(result.Text, "Hello") {
		t.Errorf("Text should contain 'Hello', got %q", result.Text)
	}
	if !strings.Contains(result.Text, "World") {
		t.Errorf("Text should contain 'World', got %q", result.Text)
	}
	if strings.Contains(result.Text, "<p>") {
		t.Errorf("Text should not contain HTML tags, got %q", result.Text)
	}
	if strings.Contains(result.Text, "alert") {
		t.Errorf("Text should not contain script content, got %q", result.Text)
	}
}

func TestExtract_SinglePartHTML(t *testing.T) {
	// Single-part HTML email — should strip tags.
	input := "From: test@example.com\r\n" +
		"Subject: Single HTML\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		"<html><body><style>body{color:red}</style><p>Visible text</p></body></html>\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Visible text") {
		t.Errorf("Text should contain 'Visible text', got %q", result.Text)
	}
	if strings.Contains(result.Text, "color:red") {
		t.Errorf("Text should not contain style content, got %q", result.Text)
	}
}

func TestExtract_MultipartMissingBoundary(t *testing.T) {
	// Multipart content type without boundary — should return result without body.
	input := "From: test@example.com\r\n" +
		"Subject: No Boundary\r\n" +
		"Content-Type: multipart/mixed\r\n" +
		"\r\n" +
		"This is the body\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// extractBody error → best-effort empty text.
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestExtract_NestedMultipart(t *testing.T) {
	// Nested multipart message (multipart/mixed containing multipart/alternative).
	input := "From: test@example.com\r\n" +
		"Subject: Nested\r\n" +
		"Content-Type: multipart/mixed; boundary=\"OUTER\"\r\n" +
		"\r\n" +
		"--OUTER\r\n" +
		"Content-Type: multipart/alternative; boundary=\"INNER\"\r\n" +
		"\r\n" +
		"--INNER\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Nested plain text\r\n" +
		"--INNER--\r\n" +
		"--OUTER--\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Nested plain text") {
		t.Errorf("Text should contain 'Nested plain text', got %q", result.Text)
	}
}

func TestExtract_DepthLimit(t *testing.T) {
	// Build deeply nested multipart (11 levels) to hit maxMultipartDepth.
	// This is synthetic — real email never nests this deep.
	var builder strings.Builder
	builder.WriteString("From: test@example.com\r\nSubject: Deep\r\n")
	builder.WriteString("Content-Type: multipart/mixed; boundary=\"L0\"\r\n\r\n")

	for i := 0; i < 12; i++ {
		builder.WriteString(fmt.Sprintf("--%s\r\n", fmt.Sprintf("L%d", i)))
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"L%d\"\r\n\r\n", i+1))
	}
	// Innermost part.
	builder.WriteString("--L12\r\nContent-Type: text/plain\r\n\r\nDeep text\r\n--L12--\r\n")

	// Close all levels.
	for i := 11; i >= 0; i-- {
		builder.WriteString(fmt.Sprintf("\r\n--%s--\r\n", fmt.Sprintf("L%d", i)))
	}

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(builder.String()), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Due to depth limit, innermost text may not be reached — just ensure no panic.
	_ = result
}

func TestExtract_CorruptInput(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), strings.NewReader("not a valid email at all"), "message/rfc822")
	// mail.ReadMessage may parse this as a header-less message — shouldn't error fatally.
	// The key is it doesn't panic.
	_ = err
}

func TestExtract_BinaryAttachmentSkipped(t *testing.T) {
	// Multipart with a non-text part — should be skipped.
	input := "From: test@example.com\r\n" +
		"Subject: Attachment\r\n" +
		"Content-Type: multipart/mixed; boundary=\"ATT\"\r\n" +
		"\r\n" +
		"--ATT\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Body text\r\n" +
		"--ATT\r\n" +
		"Content-Type: application/octet-stream\r\n" +
		"\r\n" +
		"binary data here\r\n" +
		"--ATT--\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Body text") {
		t.Errorf("Text should contain 'Body text', got %q", result.Text)
	}
	if strings.Contains(result.Text, "binary data") {
		t.Errorf("Text should not contain binary attachment content")
	}
}

func TestExtract_InvalidContentType(t *testing.T) {
	// Content-Type that fails mime.ParseMediaType — should fall back to plain text.
	input := "From: test@example.com\r\n" +
		"Subject: Bad CT\r\n" +
		"Content-Type: ;;;invalid\r\n" +
		"\r\n" +
		"Fallback body text\r\n"

	ext := New()
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "message/rfc822")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Fallback body text") {
		t.Errorf("Text should contain 'Fallback body text', got %q", result.Text)
	}
}
