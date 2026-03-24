//go:build cgo && ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"os/exec"
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

// TestOCR_MultiLanguage_German tests OCR with German text.
// Skipped if deu.traineddata is not installed.
func TestOCR_MultiLanguage_German(t *testing.T) {
	// Check if German language pack is available.
	out, err := exec.Command("tesseract", "--list-langs").CombinedOutput()
	if err != nil {
		t.Skip("tesseract not available")
	}
	if !strings.Contains(string(out), "deu") {
		t.Skip("deu.traineddata not installed — install with: brew install tesseract-lang")
	}

	engine, err := NewOCREngine("deu")
	if err != nil {
		t.Fatalf("NewOCREngine(deu) failed: %v", err)
	}
	defer engine.Close()

	data := loadTestFile(t, "ocr-german.png")
	text, confidence, err := engine.RecognizeBytes(context.Background(), data)
	if err != nil {
		t.Fatalf("RecognizeBytes failed: %v", err)
	}

	t.Logf("German OCR: text=%q, confidence=%.1f", text, confidence)

	// The image contains "Hallo Welt" — should find at least part of it.
	normalized := strings.ToLower(text)
	if !strings.Contains(normalized, "hallo") && !strings.Contains(normalized, "welt") {
		t.Errorf("OCR text = %q, expected to contain 'Hallo' or 'Welt'", text)
	}
}
