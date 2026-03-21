//go:build cgo && ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestOCR_EngineCreation(t *testing.T) {
	engine, err := NewOCREngine("eng")
	if err != nil {
		t.Fatalf("NewOCREngine failed: %v", err)
	}
	defer engine.Close()
}

func TestOCR_RecognizeTextImage(t *testing.T) {
	engine, err := NewOCREngine("eng")
	if err != nil {
		t.Fatalf("NewOCREngine failed: %v", err)
	}
	defer engine.Close()

	data := loadTestFile(t, "ocr-text.png")
	text, confidence, err := engine.RecognizeBytes(context.Background(), data)
	if err != nil {
		t.Fatalf("RecognizeBytes failed: %v", err)
	}

	// Should find "Hello World 12345" (or close).
	if !strings.Contains(text, "Hello") {
		t.Errorf("OCR text = %q, expected to contain 'Hello'", text)
	}
	if confidence < 50 {
		t.Errorf("confidence = %.1f, expected > 50", confidence)
	}
}

func TestOCR_EmptyInput(t *testing.T) {
	engine, err := NewOCREngine("eng")
	if err != nil {
		t.Fatalf("NewOCREngine failed: %v", err)
	}
	defer engine.Close()

	text, _, err := engine.RecognizeBytes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if text != "" {
		t.Errorf("expected empty text for nil input, got %q", text)
	}
}

func TestOCR_ContextCancellation(t *testing.T) {
	engine, err := NewOCREngine("eng")
	if err != nil {
		t.Fatalf("NewOCREngine failed: %v", err)
	}
	defer engine.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, _, err = engine.RecognizeBytes(ctx, []byte("data"))
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}

func TestOCR_IntegrationWithExtractor(t *testing.T) {
	engine, err := NewOCREngine("eng")
	if err != nil {
		t.Fatalf("NewOCREngine failed: %v", err)
	}
	defer engine.Close()

	ext := &Extractor{OCR: engine, OCRLanguage: "eng"}
	data := loadTestFile(t, "ocr-text.png")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := ext.Extract(ctx, bytes.NewReader(data), "image/png")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// OCR text should be in result.Text.
	if !strings.Contains(result.Text, "Hello") {
		t.Errorf("result.Text = %q, expected to contain 'Hello'", result.Text)
	}

	// OCR metadata should be present.
	if result.Metadata.Get("ocr:confidence") == "" {
		t.Error("expected ocr:confidence metadata")
	}
	if result.Metadata.Get("ocr:language") != "eng" {
		t.Errorf("ocr:language = %q, want 'eng'", result.Metadata.Get("ocr:language"))
	}
}
