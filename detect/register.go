// CONTRACT:C2-DETECT.1.0
package detect

import "github.com/dmorris/content-extractor/extractor"

// NewDetectorFactory returns a factory function suitable for extractor.RegisterDetector.
func NewDetectorFactory() func() extractor.Detector {
	return func() extractor.Detector {
		return NewDetector()
	}
}
