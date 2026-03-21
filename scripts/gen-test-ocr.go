//go:build ignore

// gen-test-ocr generates test images containing known text for OCR testing.
// Run: go run scripts/gen-test-ocr.go
//
// Generates:
//   testdata/ocr-text.png — 400x60 white image with "Hello World 12345" in black
//   testdata/ocr-text.jpg — JPEG version of the same image
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {
	// Create a 400x60 white image with black text.
	// We use a simple bitmap font approach — draw each character as a pattern of black pixels.
	width, height := 400, 60
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white.
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, white)
		}
	}

	// Draw "Hello World 12345" using simple block letters.
	// Each character is drawn at a fixed size for OCR readability.
	text := "Hello World 12345"
	drawText(img, text, 20, 15, 3) // x=20, y=15, scale=3

	// Write PNG.
	if err := writePNG("testdata/ocr-text.png", img); err != nil {
		fmt.Fprintf(os.Stderr, "write PNG: %v\n", err)
		os.Exit(1)
	}

	// Write JPEG.
	if err := writeJPEG("testdata/ocr-text.jpg", img); err != nil {
		fmt.Fprintf(os.Stderr, "write JPEG: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated testdata/ocr-text.png and testdata/ocr-text.jpg")
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func writeJPEG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
}

// drawText renders text using a simple 5x7 bitmap font at the given scale.
func drawText(img *image.RGBA, text string, startX, startY, scale int) {
	black := color.RGBA{0, 0, 0, 255}
	x := startX

	for _, ch := range text {
		glyph, ok := font5x7[ch]
		if !ok {
			// Unknown character — skip space.
			x += 6 * scale
			continue
		}
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				if glyph[row]&(1<<(4-col)) != 0 {
					// Draw a scale×scale block for each pixel.
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							px := x + col*scale + dx
							py := startY + row*scale + dy
							if px < img.Bounds().Max.X && py < img.Bounds().Max.Y {
								img.Set(px, py, black)
							}
						}
					}
				}
			}
		}
		x += 6 * scale // 5 pixels + 1 pixel gap
	}
}

// font5x7 is a minimal 5×7 bitmap font for uppercase + digits.
// Each character is 7 rows of 5-bit patterns (MSB = leftmost pixel).
var font5x7 = map[rune][7]byte{
	'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'e': {0x00, 0x00, 0x0E, 0x11, 0x1F, 0x10, 0x0E},
	'l': {0x0C, 0x04, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'o': {0x00, 0x00, 0x0E, 0x11, 0x11, 0x11, 0x0E},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x1B, 0x11},
	'r': {0x00, 0x00, 0x16, 0x19, 0x10, 0x10, 0x10},
	'd': {0x01, 0x01, 0x0D, 0x13, 0x11, 0x13, 0x0D},
	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F},
	'3': {0x1F, 0x02, 0x04, 0x02, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
}
