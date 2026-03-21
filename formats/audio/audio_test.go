// CONTRACT:C1-EXTRACTOR.1.0
package audio

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func loadTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("failed to load test file %s: %v", name, err)
	}
	return data
}

func TestAudio_Compliance_MP3(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	extractortest.RunComplianceTests(t, ext, "audio/mpeg", data)
}

func TestAudio_Compliance_OGG(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.ogg")
	extractortest.RunComplianceTests(t, ext, "audio/ogg", data)
}

func TestAudio_Compliance_FLAC(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.flac")
	extractortest.RunComplianceTests(t, ext, "audio/flac", data)
}

func TestAudio_Compliance_WAV(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.wav")
	extractortest.RunComplianceTests(t, ext, "audio/wav", data)
}

func TestAudio_MP3_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "Test Song"},
		{metadata.KeyCreator, "Test Artist"},
		{"audio:album", "Test Album"},
		{"audio:track", "3"},
		{"audio:genre", "Electronic"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_OGG_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.ogg")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/ogg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "OGG Test Song"},
		{metadata.KeyCreator, "OGG Test Artist"},
		{"audio:album", "OGG Test Album"},
		{"audio:track", "5"},
		{"audio:genre", "Ambient"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_FLAC_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.flac")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/flac")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "FLAC Test Song"},
		{metadata.KeyCreator, "FLAC Test Artist"},
		{"audio:album", "FLAC Test Album"},
		{"audio:track", "7"},
		{"audio:genre", "Classical"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_WAV_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.wav")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/wav")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "WAV Test Song"},
		{metadata.KeyCreator, "WAV Test Artist"},
		{"audio:album", "WAV Test Album"},
		{"audio:genre", "Jazz"},
		{"audio:channels", "2"},
		{"audio:sample-rate", "44100"},
		{"audio:bits-per-sample", "16"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "audio/mpeg")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestAudio_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestAudio_TextSummary(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text == "" {
		t.Error("Text summary is empty")
	}
}
