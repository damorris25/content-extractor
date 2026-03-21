//go:build ignore

// gen-test-geotiff generates a minimal GeoTIFF test fixture.
// Run: go run scripts/gen-test-geotiff.go
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func main() {
	outPath := "testdata/geotiff-sample.tif"

	// Build a minimal little-endian TIFF with GeoTIFF tags.
	// Image: 4x4 grayscale (8-bit), single strip.
	// GeoTIFF: EPSG:4326, tie point at origin (-122.0, 37.0), pixel scale (0.001, 0.001).
	var buf []byte

	// TIFF Header (8 bytes)
	buf = appendU16LE(buf, 0x4949) // Byte order: little-endian
	buf = appendU16LE(buf, 42)     // TIFF magic
	buf = appendU32LE(buf, 8)      // Offset to first IFD

	// IFD0 starts at offset 8.
	// We'll have these tags:
	// 256 (ImageWidth), 257 (ImageLength), 258 (BitsPerSample),
	// 259 (Compression), 262 (PhotometricInterpretation),
	// 273 (StripOffsets), 278 (RowsPerStrip), 279 (StripByteCounts),
	// 33550 (ModelPixelScaleTag), 33922 (ModelTiepointTag),
	// 34735 (GeoKeyDirectoryTag), 34737 (GeoAsciiParamsTag)
	numTags := uint16(12)
	buf = appendU16LE(buf, numTags)

	// Tag entries (12 bytes each), starting at offset 10.
	// After IFD: 10 + 12*12 + 4(next IFD ptr) = 10 + 144 + 4 = 158
	// Data area starts at offset 158.
	dataAreaStart := uint32(10 + 12*numTags + 4)
	dataOffset := dataAreaStart
	var dataArea []byte

	// Tag 256: ImageWidth = 4 (SHORT)
	buf = appendIFDEntry(buf, 256, 3, 1, 4) // type=SHORT, count=1, value=4

	// Tag 257: ImageLength = 4 (SHORT)
	buf = appendIFDEntry(buf, 257, 3, 1, 4)

	// Tag 258: BitsPerSample = 8 (SHORT)
	buf = appendIFDEntry(buf, 258, 3, 1, 8)

	// Tag 259: Compression = 1 (no compression) (SHORT)
	buf = appendIFDEntry(buf, 259, 3, 1, 1)

	// Tag 262: PhotometricInterpretation = 1 (BlackIsZero) (SHORT)
	buf = appendIFDEntry(buf, 262, 3, 1, 1)

	// Tag 273: StripOffsets — will point to image data (LONG)
	// Image data will go at the end; we'll fix this up.
	stripOffsetPos := len(buf) + 8 // position of the value field in this entry
	buf = appendIFDEntry(buf, 273, 4, 1, 0) // placeholder

	// Tag 278: RowsPerStrip = 4 (SHORT)
	buf = appendIFDEntry(buf, 278, 3, 1, 4)

	// Tag 279: StripByteCounts = 16 (4*4 pixels) (LONG)
	buf = appendIFDEntry(buf, 279, 4, 1, 16)

	// Tag 33550: ModelPixelScaleTag — 3 DOUBLEs (ScaleX, ScaleY, ScaleZ)
	// Type 12 (DOUBLE), count 3, data at offset
	pixelScaleData := make([]byte, 24)
	binary.LittleEndian.PutUint64(pixelScaleData[0:], math.Float64bits(0.001))  // ScaleX
	binary.LittleEndian.PutUint64(pixelScaleData[8:], math.Float64bits(0.001))  // ScaleY
	binary.LittleEndian.PutUint64(pixelScaleData[16:], math.Float64bits(0.0))   // ScaleZ
	buf = appendIFDEntry(buf, 33550, 12, 3, dataOffset)
	dataArea = append(dataArea, pixelScaleData...)
	dataOffset += 24

	// Tag 33922: ModelTiepointTag — 6 DOUBLEs (I, J, K, X, Y, Z)
	// Tie point: pixel (0,0,0) -> geo (-122.0, 37.0, 0.0)
	tiepointData := make([]byte, 48)
	binary.LittleEndian.PutUint64(tiepointData[0:], math.Float64bits(0.0))     // I
	binary.LittleEndian.PutUint64(tiepointData[8:], math.Float64bits(0.0))     // J
	binary.LittleEndian.PutUint64(tiepointData[16:], math.Float64bits(0.0))    // K
	binary.LittleEndian.PutUint64(tiepointData[24:], math.Float64bits(-122.0)) // X
	binary.LittleEndian.PutUint64(tiepointData[32:], math.Float64bits(37.0))   // Y
	binary.LittleEndian.PutUint64(tiepointData[40:], math.Float64bits(0.0))    // Z
	buf = appendIFDEntry(buf, 33922, 12, 6, dataOffset)
	dataArea = append(dataArea, tiepointData...)
	dataOffset += 48

	// Tag 34735: GeoKeyDirectoryTag — array of SHORTs
	// Format: [KeyDirectoryVersion, KeyRevision, MinorRevision, NumberOfKeys, ...keys]
	// Key entry: [KeyID, TIFFTagLocation, Count, Value_Offset]
	// GTModelTypeGeoKey (1024) = 2 (Geographic)
	// GTRasterTypeGeoKey (1025) = 1 (RasterPixelIsArea)
	// GeographicTypeGeoKey (2048) = 4326 (EPSG:4326)
	// GeogCitationGeoKey (2049) -> GeoAsciiParamsTag
	geoKeys := []uint16{
		1, 1, 0, 4, // version, revision, minor, numKeys
		1024, 0, 1, 2,       // GTModelTypeGeoKey = Geographic
		1025, 0, 1, 1,       // GTRasterTypeGeoKey = RasterPixelIsArea
		2048, 0, 1, 4326,    // GeographicTypeGeoKey = EPSG:4326
		2049, 34737, 7, 0,   // GeogCitationGeoKey -> GeoAsciiParamsTag, 7 chars at offset 0
	}
	geoKeyData := make([]byte, len(geoKeys)*2)
	for i, v := range geoKeys {
		binary.LittleEndian.PutUint16(geoKeyData[i*2:], v)
	}
	buf = appendIFDEntry(buf, 34735, 3, uint32(len(geoKeys)), dataOffset)
	dataArea = append(dataArea, geoKeyData...)
	dataOffset += uint32(len(geoKeyData))

	// Tag 34737: GeoAsciiParamsTag — ASCII string "WGS 84|"
	geoAscii := []byte("WGS 84|")
	buf = appendIFDEntry(buf, 34737, 2, uint32(len(geoAscii)), dataOffset)
	dataArea = append(dataArea, geoAscii...)
	dataOffset += uint32(len(geoAscii))

	// Next IFD offset = 0 (no more IFDs)
	buf = appendU32LE(buf, 0)

	// Append data area
	buf = append(buf, dataArea...)

	// Fix up strip offset to point to image data at current end
	imageDataOffset := uint32(len(buf))
	binary.LittleEndian.PutUint32(buf[stripOffsetPos:], imageDataOffset)

	// Append 4x4 grayscale image data (16 bytes of gray values)
	imageData := make([]byte, 16)
	for i := range imageData {
		imageData[i] = byte(i * 16)
	}
	buf = append(buf, imageData...)

	if err := os.WriteFile(outPath, buf, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s (%d bytes)\n", outPath, len(buf))
}

func appendU16LE(buf []byte, v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return append(buf, b...)
}

func appendU32LE(buf []byte, v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return append(buf, b...)
}

func appendIFDEntry(buf []byte, tag uint16, typ uint16, count uint32, value uint32) []byte {
	buf = appendU16LE(buf, tag)
	buf = appendU16LE(buf, typ)
	buf = appendU32LE(buf, count)
	buf = appendU32LE(buf, value)
	return buf
}
