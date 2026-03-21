// CONTRACT:I2-REGISTRY.1.0
package extractor

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
)

// stubExtractor is a minimal Extractor for registry tests.
type stubExtractor struct {
	mimeTypes []string
	text      string
}

func (s *stubExtractor) Extract(_ context.Context, r io.Reader, mime string) (*Result, error) {
	if r == nil {
		return nil, ErrNilReader
	}
	for _, mt := range s.mimeTypes {
		if mt == mime {
			return &Result{Text: s.text, ContentType: mime}, nil
		}
	}
	return nil, ErrUnsupportedFormat
}

func (s *stubExtractor) SupportedMIMETypes() []string {
	return s.mimeTypes
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	ext := &stubExtractor{mimeTypes: []string{"text/plain"}, text: "hello"}
	reg.Register(ext)

	got := reg.Lookup("text/plain")
	if got != ext {
		t.Error("Lookup did not return registered extractor")
	}
}

func TestRegistry_Lookup_Unregistered(t *testing.T) {
	reg := NewRegistry()
	if got := reg.Lookup("application/pdf"); got != nil {
		t.Errorf("Lookup unregistered = %v, want nil", got)
	}
}

func TestRegistry_Register_DuplicatePanics(t *testing.T) {
	reg := NewRegistry()
	ext := &stubExtractor{mimeTypes: []string{"text/plain"}}
	reg.Register(ext)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	reg.Register(ext)
}

func TestRegistry_Register_NilPanics(t *testing.T) {
	reg := NewRegistry()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil registration")
		}
	}()
	reg.Register(nil)
}

func TestRegistry_SupportedTypes_Sorted(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubExtractor{mimeTypes: []string{"text/plain"}})
	reg.Register(&stubExtractor{mimeTypes: []string{"application/pdf"}})
	reg.Register(&stubExtractor{mimeTypes: []string{"image/png"}})

	types := reg.SupportedTypes()
	if len(types) != 3 {
		t.Fatalf("SupportedTypes len = %d, want 3", len(types))
	}
	if types[0] != "application/pdf" || types[1] != "image/png" || types[2] != "text/plain" {
		t.Errorf("SupportedTypes = %v, want sorted", types)
	}
}

func TestRegistry_Extract_Dispatches(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubExtractor{mimeTypes: []string{"text/plain"}, text: "hello"})

	result, err := reg.Extract(context.Background(), bytes.NewReader([]byte("hello")), "text/plain")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}
	if result.Text != "hello" {
		t.Errorf("Extract text = %q, want %q", result.Text, "hello")
	}
}

func TestRegistry_Extract_UnsupportedMIME(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Extract(context.Background(), bytes.NewReader([]byte("data")), "application/x-unknown")
	if err != ErrUnsupportedFormat {
		t.Errorf("Extract unsupported = %v, want ErrUnsupportedFormat", err)
	}
}

func TestRegistry_Extract_NilReader(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Extract(context.Background(), nil, "text/plain")
	if err != ErrNilReader {
		t.Errorf("Extract nil reader = %v, want ErrNilReader", err)
	}
}

func TestRegistry_ConcurrentLookup(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubExtractor{mimeTypes: []string{"text/plain"}, text: "hello"})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ext := reg.Lookup("text/plain")
			if ext == nil {
				t.Error("concurrent Lookup returned nil")
			}
		}()
	}
	wg.Wait()
}
