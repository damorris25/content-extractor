// CONTRACT:C1-EXTRACTOR.1.0
package extractor

import "errors"

var (
	// ErrNilReader is returned when a nil reader is passed to Extract.
	ErrNilReader = errors.New("extractor: nil reader")

	// ErrUnsupportedFormat is returned when the MIME type is not handled.
	ErrUnsupportedFormat = errors.New("extractor: unsupported format")

	// ErrCorruptInput is returned when the input is malformed or truncated.
	ErrCorruptInput = errors.New("extractor: corrupt input")

	// ErrPasswordProtected is returned when the file requires a password.
	ErrPasswordProtected = errors.New("extractor: password protected")

	// ErrFileTooLarge is returned when the input exceeds the configured size limit.
	ErrFileTooLarge = errors.New("extractor: file too large")

	// ErrDetectionFailed is returned when format auto-detection cannot determine the type.
	ErrDetectionFailed = errors.New("extractor: detection failed")
)
