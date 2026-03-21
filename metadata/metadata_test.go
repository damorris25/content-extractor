// CONTRACT:I1-METADATA.1.0
package metadata

import (
	"testing"
)

func TestMetadata_Get_NilSafe(t *testing.T) {
	var m Metadata
	if got := m.Get("dc:title"); got != "" {
		t.Errorf("Get on nil Metadata = %q, want empty", got)
	}
}

func TestMetadata_Has_NilSafe(t *testing.T) {
	var m Metadata
	if m.Has("dc:title") {
		t.Error("Has on nil Metadata = true, want false")
	}
}

func TestMetadata_GetAll_NilSafe(t *testing.T) {
	var m Metadata
	if got := m.GetAll("dc:title"); got != nil {
		t.Errorf("GetAll on nil Metadata = %v, want nil", got)
	}
}

func TestMetadata_SetAndGet(t *testing.T) {
	m := make(Metadata)
	m.Set("dc:title", "Hello")
	if got := m.Get("dc:title"); got != "Hello" {
		t.Errorf("Get = %q, want %q", got, "Hello")
	}
}

func TestMetadata_Set_EmptyOmitted(t *testing.T) {
	m := make(Metadata)
	m.Set("dc:title", "Hello")
	m.Set("dc:title", "")
	if m.Has("dc:title") {
		t.Error("Set with empty string should remove the key")
	}
}

func TestMetadata_Add_MultiValue(t *testing.T) {
	m := make(Metadata)
	m.Add("dc:creator", "Alice")
	m.Add("dc:creator", "Bob")
	got := m.GetAll("dc:creator")
	if len(got) != 2 || got[0] != "Alice" || got[1] != "Bob" {
		t.Errorf("GetAll = %v, want [Alice Bob]", got)
	}
}

func TestMetadata_Add_EmptyOmitted(t *testing.T) {
	m := make(Metadata)
	m.Add("dc:creator", "")
	if m.Has("dc:creator") {
		t.Error("Add with empty value should not create key")
	}
}

func TestNormalizeDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-03-21T10:30:00Z", "2026-03-21T10:30:00Z"},
		{"2026-03-21", "2026-03-21T00:00:00Z"},
		{"2026-03-21 14:30:00", "2026-03-21T14:30:00Z"},
		{"", ""},
		{"not a date", "not a date"}, // unparseable returns original
	}
	for _, tt := range tests {
		got := NormalizeDate(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCleanUTF8(t *testing.T) {
	// Valid UTF-8 passes through unchanged.
	if got := CleanUTF8("hello"); got != "hello" {
		t.Errorf("CleanUTF8(valid) = %q, want %q", got, "hello")
	}
	// Invalid bytes get replaced with U+FFFD.
	invalid := "hello\x80world"
	got := CleanUTF8(invalid)
	if got != "hello\uFFFDworld" {
		t.Errorf("CleanUTF8(invalid) = %q, want %q", got, "hello\uFFFDworld")
	}
}
