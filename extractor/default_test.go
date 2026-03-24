// CONTRACT:I2-REGISTRY.1.0
package extractor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

// stubDetector is a minimal Detector for testing.
type stubDetector struct {
	mime string
	err  error
}

func (d *stubDetector) Detect(r io.ReadSeeker, filename string) (string, error) {
	return d.mime, d.err
}

func TestRegistry_Extract_WithDetector(t *testing.T) {
	reg := NewRegistry()
	reg.SetDetector(&stubDetector{mime: "text/plain"})
	reg.Register(&stubExtractor{mimeTypes: []string{"text/plain"}, text: "hello"})

	result, err := reg.Extract(context.Background(), bytes.NewReader([]byte("hello")), "")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}
	if result.Text != "hello" {
		t.Errorf("Extract text = %q, want %q", result.Text, "hello")
	}
	if result.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", result.ContentType)
	}
}

func TestRegistry_Extract_DetectorError(t *testing.T) {
	reg := NewRegistry()
	reg.SetDetector(&stubDetector{err: errors.New("detection failed")})

	_, err := reg.Extract(context.Background(), bytes.NewReader([]byte("data")), "")
	if err == nil {
		t.Fatal("expected error from detector")
	}
	if !errors.Is(err, ErrDetectionFailed) {
		t.Errorf("error = %v, want wrapped ErrDetectionFailed", err)
	}
}

func TestRegistry_Extract_NoDetector(t *testing.T) {
	reg := NewRegistry()
	// No detector set, empty mime — should return ErrDetectionFailed.
	_, err := reg.Extract(context.Background(), bytes.NewReader([]byte("data")), "")
	if err != ErrDetectionFailed {
		t.Errorf("error = %v, want ErrDetectionFailed", err)
	}
}

func TestDefaultRegistry_WithBuilders(t *testing.T) {
	// Save and restore global state.
	savedBuilders := builders
	savedFactory := detectorFactory
	defer func() {
		builders = savedBuilders
		detectorFactory = savedFactory
	}()

	builders = nil
	detectorFactory = nil

	builderCalled := false
	RegisterBuilder(func(r *Registry) {
		builderCalled = true
		r.Register(&stubExtractor{mimeTypes: []string{"test/format"}, text: "test"})
	})

	detectorCalled := false
	RegisterDetector(func() Detector {
		detectorCalled = true
		return &stubDetector{mime: "test/format"}
	})

	reg := DefaultRegistry()
	if !builderCalled {
		t.Error("builder was not called")
	}
	if !detectorCalled {
		t.Error("detector factory was not called")
	}

	// Verify the registry works.
	ext := reg.Lookup("test/format")
	if ext == nil {
		t.Fatal("Lookup returned nil for registered format")
	}
}

func TestExtract_ConvenienceFunction(t *testing.T) {
	// Save and restore global state.
	savedBuilders := builders
	savedFactory := detectorFactory
	defer func() {
		builders = savedBuilders
		detectorFactory = savedFactory
	}()

	builders = nil
	detectorFactory = nil

	RegisterBuilder(func(r *Registry) {
		r.Register(&stubExtractor{mimeTypes: []string{"text/plain"}, text: "hello"})
	})

	result, err := Extract(bytes.NewReader([]byte("hello")), "text/plain")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}
	if result.Text != "hello" {
		t.Errorf("Extract text = %q, want %q", result.Text, "hello")
	}
}

func TestDefaultRegistry_NoBuilders(t *testing.T) {
	savedBuilders := builders
	savedFactory := detectorFactory
	defer func() {
		builders = savedBuilders
		detectorFactory = savedFactory
	}()

	builders = nil
	detectorFactory = nil

	reg := DefaultRegistry()
	types := reg.SupportedTypes()
	if len(types) != 0 {
		t.Errorf("SupportedTypes = %v, want empty", types)
	}
}
