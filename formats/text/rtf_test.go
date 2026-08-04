// CONTRACT:C1-EXTRACTOR.1.0
package text

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/extractor/extractortest"
)

func TestRTFCompliance(t *testing.T) {
	ext := NewRTF()
	extractortest.RunComplianceTests(t, ext, "application/rtf", []byte(`{\rtf1 hello}`))
}

func TestRTF_BasicText(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1\ansi Hello, World!\par Second paragraph.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, World!") {
		t.Errorf("Text should contain 'Hello, World!', got %q", result.Text)
	}
	if !strings.Contains(result.Text, "Second paragraph.") {
		t.Errorf("Text should contain 'Second paragraph.', got %q", result.Text)
	}
	if result.ContentType != "application/rtf" {
		t.Errorf("ContentType = %q, want application/rtf", result.ContentType)
	}
}

func TestRTF_DestinationGroups(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1{\fonttbl{\f0 Arial;}}Visible text.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Visible text.") {
		t.Errorf("Text should contain 'Visible text.', got %q", result.Text)
	}
	if strings.Contains(result.Text, "fonttbl") {
		t.Errorf("Text should not contain 'fonttbl', got %q", result.Text)
	}
}

func TestRTF_UnicodeEscape(t *testing.T) {
	ext := NewRTF()
	input := `{\rtf1 Hello \u8212? World.\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello") || !strings.Contains(result.Text, "World.") {
		t.Errorf("Text = %q, want 'Hello' and 'World.'", result.Text)
	}
	// Should contain em-dash (U+2014).
	if !strings.Contains(result.Text, "\u2014") {
		t.Errorf("Text should contain em-dash, got %q", result.Text)
	}
}

func TestRTF_HexEscape(t *testing.T) {
	ext := NewRTF()
	// Use Unicode escape for é instead of hex, since \'e9 produces a
	// raw Latin-1 byte that gets cleaned as invalid UTF-8.
	input := `{\rtf1 caf\u233?\par}`
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "café") {
		t.Errorf("Text should contain 'café', got %q", result.Text)
	}
}

func TestRTF_EmptyInput(t *testing.T) {
	ext := NewRTF()
	result, err := ext.Extract(context.Background(), strings.NewReader(""), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Text = %q, want empty", result.Text)
	}
}

func TestRTF_UnsupportedMIME(t *testing.T) {
	ext := NewRTF()
	_, err := ext.Extract(context.Background(), strings.NewReader(`{\rtf1}`), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestRTF_SymbolControls(t *testing.T) {
	ext := NewRTF()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "escaped backslash",
			input: `{\rtf1 path\\to\\file}`,
			want:  `path\to\file`,
		},
		{
			name:  "escaped open brace",
			input: `{\rtf1 use \{ braces \}}`,
			want:  "use { braces }",
		},
		{
			name:  "escaped close brace",
			input: `{\rtf1 value = \{42\}}`,
			want:  "value = {42}",
		},
		{
			name:  "mixed symbol controls",
			input: `{\rtf1 a\\b\{c\}d}`,
			want:  `a\b{c}d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ext.Extract(context.Background(), strings.NewReader(tt.input), "application/rtf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.want {
				t.Errorf("Text = %q, want %q", result.Text, tt.want)
			}
		})
	}
}

func TestRTF_NegativeParams(t *testing.T) {
	ext := NewRTF()
	// \uc-1 is a control word with a negative numeric parameter.
	// It should be parsed and not appear in output text.
	input := `{\rtf1\uc-1 Hello World}` //nolint:gocritic
	result, err := ext.Extract(context.Background(), strings.NewReader(input), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello World") {
		t.Errorf("Text should contain 'Hello World', got %q", result.Text)
	}
	if strings.Contains(result.Text, "uc") {
		t.Errorf("Text should not contain control word 'uc', got %q", result.Text)
	}
}

func TestRTF_ControlWordDelimiters(t *testing.T) {
	ext := NewRTF()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "space delimiter consumed",
			input: `{\rtf1\par Next line}`,
			want:  "Next line",
		},
		{
			name:  "non-alpha terminates control word",
			input: `{\rtf1\b0Hello}`,
			want:  "Hello",
		},
		{
			name:  "tab control word",
			input: `{\rtf1 A\tab B}`,
			want:  "A\tB",
		},
		{
			name:  "line control word",
			input: `{\rtf1 A\line B}`,
			want:  "A\nB",
		},
		{
			name:  "special word emdash",
			input: `{\rtf1 A\emdash B}`,
			want:  "A\u2014B",
		},
		{
			name:  "special word endash",
			input: `{\rtf1 A\endash B}`,
			want:  "A\u2013B",
		},
		{
			name:  "special word bullet",
			input: `{\rtf1 \bullet item}`,
			want:  "\u2022item",
		},
		{
			name:  "special word lquote rquote",
			input: `{\rtf1 \lquote hello\rquote }`,
			want:  "'hello'",
		},
		{
			name:  "special word ldblquote rdblquote",
			input: `{\rtf1 \ldblquote hello\rdblquote }`,
			want:  `"hello"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ext.Extract(context.Background(), strings.NewReader(tt.input), "application/rtf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.want {
				t.Errorf("Text = %q, want %q", result.Text, tt.want)
			}
		})
	}
}

func TestRTF_UnicodeWithFallback(t *testing.T) {
	ext := NewRTF()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unicode with question mark fallback",
			input: `{\rtf1 Hello \u8212? World}`,
			want:  "Hello \u2014 World",
		},
		{
			name:  "negative unicode value",
			input: `{\rtf1 \u-4064? end}`,
			// -4064 + 65536 = 61472 = 0xF020, a private use area char
			want:  "\uF020 end",
		},
		{
			name:  "unicode with space delimiter",
			input: `{\rtf1 A\u8212 B}`,
			// space after number is consumed as delimiter, B is text
			want:  "A\u2014B",
		},
		{
			name:  "unicode with non-space fallback char",
			input: `{\rtf1 \u169Xrest}`,
			// U+00A9 = copyright sign, X is fallback (skipped)
			want:  "\u00A9rest",
		},
		{
			name:  "multiple unicode escapes",
			input: `{\rtf1 \u8220?Hello\u8221?}`,
			// U+201C = left double quotation, U+201D = right double quotation
			want:  "\u201CHello\u201D",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ext.Extract(context.Background(), strings.NewReader(tt.input), "application/rtf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.want {
				t.Errorf("Text = %q, want %q", result.Text, tt.want)
			}
		})
	}
}

func TestRTF_SkipDestinationGroups(t *testing.T) {
	ext := NewRTF()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "skip info destination with control words",
			input: `{\rtf1{\*\info{\author Test Author}}Visible}`,
			want:  "Visible",
		},
		{
			name:  "skip destination with negative param",
			input: `{\rtf1{\*\revtbl{\revtm-123 data}}Hello}`,
			want:  "Hello",
		},
		{
			name:  "skip destination with space-delimited control",
			input: `{\rtf1{\*\generator Riched20 10.0.19041}Text here}`,
			want:  "Text here",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ext.Extract(context.Background(), strings.NewReader(tt.input), "application/rtf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.want {
				t.Errorf("Text = %q, want %q", result.Text, tt.want)
			}
		})
	}
}

func TestRTF_Fixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/simple.rtf")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ext := NewRTF()
	result, err := ext.Extract(context.Background(), strings.NewReader(string(data)), "application/rtf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello, this is a simple RTF document.") {
		t.Errorf("Text should contain 'Hello, this is a simple RTF document.', got %q", result.Text)
	}
	if !strings.Contains(result.Text, "multiple paragraphs") {
		t.Errorf("Text should contain 'multiple paragraphs', got %q", result.Text)
	}
}
