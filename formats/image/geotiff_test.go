// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"testing"
)

func TestImage_GeoTIFF_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "geotiff-sample.tif")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/tiff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"geo:crs", "WGS 84"},
		{"geo:pixel-scale", "0.001000,0.001000"},
		{"geo:bounds", "-122.000000,36.996000,-121.996000,37.000000"},
		{"exif:width", "4"},
		{"exif:height", "4"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Text summary should include geo info.
	if result.Text == "" {
		t.Error("Text summary is empty, expected geo metadata in summary")
	}
}

func TestImage_RegularTIFF_NoGeoMetadata(t *testing.T) {
	// Non-GeoTIFF data should produce no geo: keys.
	md := make(map[string][]string)
	// Call extractGeoTIFF with non-GeoTIFF data.
	extractGeoTIFF([]byte{}, md)
	if len(md) != 0 {
		t.Errorf("expected no geo metadata from empty data, got %v", md)
	}
}

func TestGeoTIFF_TruncatedData(t *testing.T) {
	data := loadTestFile(t, "geotiff-sample.tif")
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too_short", []byte{0x49, 0x49}},
		{"truncated_header", data[:7]},
		{"truncated_ifd", data[:12]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := make(map[string][]string)
			extractGeoTIFF(tt.data, md)
			// Should not panic or produce geo metadata.
			if md["geo:crs"] != nil || md["geo:bounds"] != nil || md["geo:pixel-scale"] != nil {
				t.Errorf("expected no geo metadata from truncated data, got %v", md)
			}
		})
	}
}

func TestGeoTIFF_BigEndian(t *testing.T) {
	// Verify that a big-endian TIFF header is handled (even if no geo tags).
	md := make(map[string][]string)
	// Minimal big-endian TIFF: MM, 42, IFD offset=8, 0 entries
	data := []byte{
		'M', 'M', 0x00, 0x2A, // Byte order + magic
		0x00, 0x00, 0x00, 0x08, // IFD0 offset = 8
		0x00, 0x00, // 0 entries
	}
	extractGeoTIFF(data, md)
	// Should not panic or produce metadata.
	if len(md) != 0 {
		t.Errorf("expected no metadata from empty IFD, got %v", md)
	}
}
