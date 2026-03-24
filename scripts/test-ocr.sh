#!/bin/bash
# test-ocr.sh — Run OCR tests for Go and TypeScript.
# Requires Tesseract to be installed for Go tests.
# TypeScript tests auto-detect tesseract.js availability.
#
# Usage: bash scripts/test-ocr.sh

set -euo pipefail

echo "=== OCR Test Runner ==="
echo

# --- Go OCR tests ---
if command -v tesseract >/dev/null 2>&1; then
    echo "Tesseract found: $(tesseract --version 2>&1 | head -1)"
    echo "Running Go OCR tests (CGO_ENABLED=1 -tags ocr)..."
    CGO_ENABLED=1 go test -tags ocr -race -v ./formats/image/ -run "OCR|Ocr" 2>&1
    echo
    echo "Go OCR tests: PASSED"
else
    echo "Tesseract not installed — skipping Go OCR tests."
    echo "  Install: brew install tesseract (macOS) or apt install libtesseract-dev (Linux)"
fi

echo

# --- TypeScript OCR tests ---
echo "Running TypeScript OCR tests..."
(cd ts && npx vitest run __tests__/ocr.test.ts 2>&1)
echo
echo "TypeScript OCR tests: PASSED"

echo
echo "=== OCR tests complete ==="
