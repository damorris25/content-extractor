// CONTRACT:C1-EXTRACTOR.1.0
package odf

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxODFSize is the maximum uncompressed size for an ODF file.
	DefaultMaxODFSize = 100 * 1024 * 1024 // 100MB

	// maxDecompressedEntrySize limits the decompressed size of individual ZIP entries
	// to prevent zip bomb attacks.
	maxDecompressedEntrySize = 100 * 1024 * 1024 // 100MB

	textNS  = "urn:oasis:names:tc:opendocument:xmlns:text:1.0"
	tableNS = "urn:oasis:names:tc:opendocument:xmlns:table:1.0"
	drawNS  = "urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
	metaNS  = "urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
	dcNS    = "http://purl.org/dc/elements/1.1/"
)

// openODF reads an ODF file from r, validates it, and returns the zip.Reader.
// Returns an empty Result for empty input.
func openODF(r io.Reader, maxSize int64, mime string) (*zip.Reader, []byte, *extractor.Result, error) {
	limited := io.LimitReader(r, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	if int64(len(data)) > maxSize {
		return nil, nil, nil, extractor.ErrFileTooLarge
	}
	if len(data) == 0 {
		return nil, nil, &extractor.Result{ContentType: mime, Metadata: make(metadata.Metadata)}, nil
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid ZIP: %v", extractor.ErrCorruptInput, err)
	}

	return zr, data, nil, nil
}

// findZipFile finds a file by name in a ZIP archive.
func findZipFile(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// extractODFMeta reads meta.xml and populates the Result metadata.
func extractODFMeta(zr *zip.Reader, result *extractor.Result) {
	f := findZipFile(zr, "meta.xml")
	if f == nil {
		return
	}
	rc, err := f.Open()
	if err != nil {
		return
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 64*1024))
	if err != nil {
		return
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	var currentTag string
	var currentNS string

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			currentTag = t.Name.Local
			currentNS = t.Name.Space
		case xml.EndElement:
			currentTag = ""
			currentNS = ""
		case xml.CharData:
			val := strings.TrimSpace(string(t))
			if val == "" {
				continue
			}
			switch {
			case currentNS == dcNS && currentTag == "title":
				result.Metadata.Set(metadata.KeyTitle, val)
			case currentNS == dcNS && currentTag == "creator":
				result.Metadata.Set(metadata.KeyCreator, val)
			case currentNS == dcNS && currentTag == "subject":
				result.Metadata.Set(metadata.KeySubject, val)
			case currentNS == dcNS && currentTag == "description":
				result.Metadata.Set(metadata.KeyDescription, val)
			case currentNS == dcNS && currentTag == "date":
				result.Metadata.Set(metadata.KeyModified, metadata.NormalizeDate(val))
			case currentNS == dcNS && currentTag == "language":
				result.Metadata.Set(metadata.KeyLanguage, val)
				result.Language = val
			case currentNS == metaNS && currentTag == "creation-date":
				result.Metadata.Set(metadata.KeyDate, metadata.NormalizeDate(val))
			case currentNS == metaNS && currentTag == "generator":
				result.Metadata.Set(metadata.KeyApplication, val)
			}
		}
	}
}
