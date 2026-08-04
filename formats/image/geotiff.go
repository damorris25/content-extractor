// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/damorris25/content-extractor/metadata"
)

// GeoTIFF tag IDs.
const (
	tagModelPixelScale   = 33550
	tagModelTiepoint     = 33922
	tagGeoKeyDirectory   = 34735
	tagGeoAsciiParams    = 34737
)

// GeoKey IDs.
const (
	geoKeyGeogCitation = 2049
)

// extractGeoTIFF parses GeoTIFF metadata from TIFF IFD tags.
// It extracts CRS, bounds, and pixel scale from ModelTiepointTag,
// ModelPixelScaleTag, GeoKeyDirectoryTag, and GeoAsciiParamsTag.
// Non-GeoTIFF files silently produce no geo metadata.
func extractGeoTIFF(data []byte, md metadata.Metadata) {
	if len(data) < 8 {
		return
	}

	// Determine byte order.
	var bo binary.ByteOrder
	if data[0] == 'I' && data[1] == 'I' {
		bo = binary.LittleEndian
	} else if data[0] == 'M' && data[1] == 'M' {
		bo = binary.BigEndian
	} else {
		return
	}

	// Verify TIFF magic (42).
	if bo.Uint16(data[2:]) != 42 {
		return
	}

	// Read IFD0 offset.
	ifdOffset := bo.Uint32(data[4:])
	if int(ifdOffset)+2 > len(data) {
		return
	}

	numEntries := int(bo.Uint16(data[ifdOffset:]))
	entryStart := int(ifdOffset) + 2

	var pixelScale [3]float64
	var hasPixelScale bool
	var tiepoints []float64
	var geoAscii string
	var geoKeys []uint16
	var tiffWidth, tiffHeight int

	for i := 0; i < numEntries; i++ {
		off := entryStart + i*12
		if off+12 > len(data) {
			break
		}
		tag := bo.Uint16(data[off:])
		typ := bo.Uint16(data[off+2:])
		count := bo.Uint32(data[off+4:])
		valueOff := off + 8

		switch tag {
		case 256: // ImageWidth
			tiffWidth = readShortOrLongValue(data, bo, typ, valueOff)
		case 257: // ImageLength
			tiffHeight = readShortOrLongValue(data, bo, typ, valueOff)

		case tagModelPixelScale:
			// 3 DOUBLEs: ScaleX, ScaleY, ScaleZ
			doubles := readDoubles(data, bo, typ, count, valueOff)
			if len(doubles) >= 3 {
				pixelScale = [3]float64{doubles[0], doubles[1], doubles[2]}
				hasPixelScale = true
			}

		case tagModelTiepoint:
			// 6*N DOUBLEs: (I, J, K, X, Y, Z) groups
			tiepoints = readDoubles(data, bo, typ, count, valueOff)

		case tagGeoKeyDirectory:
			// Array of SHORTs
			geoKeys = readShorts(data, bo, typ, count, valueOff)

		case tagGeoAsciiParams:
			// ASCII string
			geoAscii = readASCII(data, bo, typ, count, valueOff)
		}
	}

	// Set TIFF dimensions if not already present (stdlib may not decode TIFF).
	if tiffWidth > 0 && !md.Has("exif:width") {
		md.Set("exif:width", fmt.Sprintf("%d", tiffWidth))
	}
	if tiffHeight > 0 && !md.Has("exif:height") {
		md.Set("exif:height", fmt.Sprintf("%d", tiffHeight))
	}

	// Extract CRS from GeoAsciiParams via GeoKeyDirectory.
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "" {
		md.Set("geo:crs", crs)
	}

	// Extract pixel scale.
	if hasPixelScale {
		md.Set("geo:pixel-scale", fmt.Sprintf("%.6f,%.6f", pixelScale[0], pixelScale[1]))
	}

	// Compute bounds from first tiepoint + pixel scale + image dimensions.
	if len(tiepoints) >= 6 && hasPixelScale && tiffWidth > 0 && tiffHeight > 0 {
		originX := tiepoints[3]
		originY := tiepoints[4]
		minX := originX
		maxY := originY
		maxX := originX + float64(tiffWidth)*pixelScale[0]
		minY := originY - float64(tiffHeight)*pixelScale[1]
		md.Set("geo:bounds", fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", minX, minY, maxX, maxY))
	}
}

// readShortOrLongValue reads a single SHORT (type 3) or LONG (type 4) value from a TIFF IFD entry.
func readShortOrLongValue(data []byte, bo binary.ByteOrder, typ uint16, valueOff int) int {
	if valueOff+4 > len(data) {
		return 0
	}
	switch typ {
	case 3: // SHORT
		return int(bo.Uint16(data[valueOff:]))
	case 4: // LONG
		return int(bo.Uint32(data[valueOff:]))
	}
	return 0
}

// readDoubles reads DOUBLE (type 12) values from a TIFF IFD entry.
func readDoubles(data []byte, bo binary.ByteOrder, typ uint16, count uint32, valueOff int) []float64 {
	if typ != 12 { // DOUBLE
		return nil
	}
	// Cap count to prevent excessive allocation.
	if count > 10000 {
		return nil
	}
	totalBytes := count * 8
	// Data is at the offset stored in the value field (never inline for doubles).
	if valueOff+4 > len(data) {
		return nil
	}
	dataOff := int(bo.Uint32(data[valueOff:]))
	if dataOff+int(totalBytes) > len(data) {
		return nil
	}
	result := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		bits := bo.Uint64(data[dataOff+int(i)*8:])
		result[i] = math.Float64frombits(bits)
	}
	return result
}

// readShorts reads SHORT (type 3) values from a TIFF IFD entry.
func readShorts(data []byte, bo binary.ByteOrder, typ uint16, count uint32, valueOff int) []uint16 {
	if typ != 3 { // SHORT
		return nil
	}
	if count > 100000 {
		return nil
	}
	totalBytes := count * 2
	var dataOff int
	if totalBytes <= 4 {
		dataOff = valueOff
	} else {
		if valueOff+4 > len(data) {
			return nil
		}
		dataOff = int(bo.Uint32(data[valueOff:]))
	}
	if dataOff+int(totalBytes) > len(data) {
		return nil
	}
	result := make([]uint16, count)
	for i := uint32(0); i < count; i++ {
		result[i] = bo.Uint16(data[dataOff+int(i)*2:])
	}
	return result
}

// readASCII reads ASCII (type 2) values from a TIFF IFD entry.
func readASCII(data []byte, bo binary.ByteOrder, typ uint16, count uint32, valueOff int) string {
	if typ != 2 { // ASCII
		return ""
	}
	if count > 100000 {
		return ""
	}
	var dataOff int
	if count <= 4 {
		dataOff = valueOff
	} else {
		if valueOff+4 > len(data) {
			return ""
		}
		dataOff = int(bo.Uint32(data[valueOff:]))
	}
	if dataOff+int(count) > len(data) {
		return ""
	}
	s := string(data[dataOff : dataOff+int(count)])
	// Strip trailing NUL and pipe characters (GeoASCII uses | as separator).
	for len(s) > 0 && (s[len(s)-1] == 0 || s[len(s)-1] == '|') {
		s = s[:len(s)-1]
	}
	return s
}

// extractCRS extracts the CRS string from GeoKeyDirectory + GeoAsciiParams.
func extractCRS(geoKeys []uint16, geoAscii string) string {
	// GeoKeyDirectory: [version, revision, minor, numKeys, key1ID, key1Loc, key1Count, key1ValueOff, ...]
	if len(geoKeys) < 4 {
		return ""
	}
	numKeys := int(geoKeys[3])
	for i := 0; i < numKeys; i++ {
		base := 4 + i*4
		if base+3 >= len(geoKeys) {
			break
		}
		keyID := geoKeys[base]
		tiffTagLoc := geoKeys[base+1]
		count := int(geoKeys[base+2])
		valueOff := int(geoKeys[base+3])

		if keyID == geoKeyGeogCitation && tiffTagLoc == tagGeoAsciiParams {
			if valueOff+count <= len(geoAscii) {
				s := geoAscii[valueOff : valueOff+count]
				// Strip trailing | and NUL.
				for len(s) > 0 && (s[len(s)-1] == '|' || s[len(s)-1] == 0) {
					s = s[:len(s)-1]
				}
				return s
			}
			// Fall back to entire geoAscii.
			return geoAscii
		}
	}

	// No citation key; if we have geoAscii, return it.
	if geoAscii != "" {
		return geoAscii
	}
	return ""
}

