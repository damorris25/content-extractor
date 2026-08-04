// CONTRACT:C1-EXTRACTOR.1.0
// Package extractortest provides contract compliance test helpers for Extractor implementations.
package extractortest

import (
	"bytes"
	"context"
	"testing"
	"unicode/utf8"

	"github.com/damorris25/content-extractor/extractor"
)

// RunComplianceTests runs the standard contract compliance tests against
// any Extractor implementation. Call this from format-specific test files.
func RunComplianceTests(t *testing.T, ext extractor.Extractor, validMIME string, validInput []byte) {
	t.Helper()

	t.Run("NilReader", func(t *testing.T) {
		_, err := ext.Extract(context.Background(), nil, validMIME)
		if err == nil {
			t.Fatal("Extract(nil reader) should return error")
		}
		if err != extractor.ErrNilReader {
			t.Errorf("Extract(nil reader) = %v, want ErrNilReader", err)
		}
	})

	t.Run("EmptyReader", func(t *testing.T) {
		result, err := ext.Extract(context.Background(), bytes.NewReader(nil), validMIME)
		if err != nil {
			t.Fatalf("Extract(empty reader) unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Extract(empty reader) returned nil result")
		}
	})

	t.Run("UnsupportedMIME", func(t *testing.T) {
		_, err := ext.Extract(context.Background(), bytes.NewReader(validInput), "application/x-never-supported-12345")
		if err == nil {
			t.Fatal("Extract(unsupported MIME) should return error")
		}
		if err != extractor.ErrUnsupportedFormat {
			t.Errorf("Extract(unsupported MIME) = %v, want ErrUnsupportedFormat", err)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := ext.Extract(ctx, bytes.NewReader(validInput), validMIME)
		if err == nil {
			// Small inputs may complete before checking context; that's OK.
			return
		}
		if err != context.Canceled {
			t.Errorf("Extract(cancelled ctx) = %v, want context.Canceled", err)
		}
	})

	t.Run("SupportedMIMETypes", func(t *testing.T) {
		types := ext.SupportedMIMETypes()
		if len(types) == 0 {
			t.Fatal("SupportedMIMETypes() returned empty slice")
		}
		found := false
		for _, mt := range types {
			if mt == validMIME {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedMIMETypes() does not contain %q", validMIME)
		}
	})

	t.Run("ValidInput", func(t *testing.T) {
		result, err := ext.Extract(context.Background(), bytes.NewReader(validInput), validMIME)
		if err != nil {
			t.Fatalf("Extract(valid input) unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Extract(valid input) returned nil result")
		}
		if !utf8.ValidString(result.Text) {
			t.Error("Result.Text is not valid UTF-8")
		}
	})

	t.Run("ConcurrentSafety", func(t *testing.T) {
		const goroutines = 10
		errs := make(chan error, goroutines)
		for range goroutines {
			go func() {
				_, err := ext.Extract(context.Background(), bytes.NewReader(validInput), validMIME)
				errs <- err
			}()
		}
		for range goroutines {
			if err := <-errs; err != nil {
				t.Errorf("concurrent Extract error: %v", err)
			}
		}
	})
}
