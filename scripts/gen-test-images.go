//go:build ignore

// gen-test-images generates minimal image test fixtures.
package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createJPEG(filepath.Join(dir, "simple.jpg"))
	createPNG(filepath.Join(dir, "simple.png"))
	createPNGWithText(filepath.Join(dir, "text.png"))
	createGIF(filepath.Join(dir, "simple.gif"))
}

func createJPEG(path string) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for y := range 50 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	jpeg.Encode(f, img, &jpeg.Options{Quality: 80})
}

func createPNG(path string) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 150))
	for y := range 150 {
		for x := range 200 {
			img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

func createPNGWithText(path string) {
	// Create a PNG with tEXt chunks manually.
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := range 64 {
		for x := range 64 {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	// Encode to standard PNG first.
	var buf bytes.Buffer
	png.Encode(&buf, img)
	data := buf.Bytes()

	// Insert tEXt chunks before IEND.
	// PNG structure: signature(8) + chunks...
	// IEND is the last chunk.
	iendPos := bytes.LastIndex(data, []byte("IEND"))
	if iendPos < 4 {
		// Fallback: just write as-is.
		os.WriteFile(path, data, 0o644)
		return
	}
	iendStart := iendPos - 4 // 4 bytes for length field

	var out bytes.Buffer
	out.Write(data[:iendStart])

	// Add tEXt chunks.
	writeTextChunk(&out, "Title", "Green Square")
	writeTextChunk(&out, "Author", "Test Generator")
	writeTextChunk(&out, "Description", "A 64x64 green square for testing")
	writeTextChunk(&out, "Software", "gen-test-images")

	out.Write(data[iendStart:])

	os.WriteFile(path, out.Bytes(), 0o644)
}

func writeTextChunk(buf *bytes.Buffer, key, value string) {
	// tEXt chunk: keyword\0text
	payload := append([]byte(key), 0)
	payload = append(payload, []byte(value)...)

	// Write length (4 bytes big-endian).
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(payload)))
	buf.Write(lenBytes)

	// Write chunk type.
	chunkType := []byte("tEXt")
	buf.Write(chunkType)

	// Write payload.
	buf.Write(payload)

	// Write CRC (over type + payload).
	crc := crc32.NewIEEE()
	crc.Write(chunkType)
	crc.Write(payload)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc.Sum32())
	buf.Write(crcBytes)
}

func createGIF(path string) {
	img := image.NewPaletted(image.Rect(0, 0, 32, 32), color.Palette{
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
	})
	f, _ := os.Create(path)
	defer f.Close()
	gif.Encode(f, img, nil)
}
