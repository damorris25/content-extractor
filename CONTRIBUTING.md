# Contributing to content-extractor

Thanks for your interest in contributing! This project is a dual-language
(Go + TypeScript) content-extraction library with strict output parity
between implementations.

## Getting Started

```bash
# Go
go mod download
go build ./...
go test ./...

# TypeScript
cd ts && npm install
npx vitest run
```

Requirements: Go 1.26+, Node.js 20+. Tesseract is optional (OCR support only).

## Ground Rules

- **Licenses:** Only MIT/BSD/Apache 2.0 dependencies. No GPL/LGPL/AGPL — this
  library must stay commercially safe.
- **Minimal dependencies:** Go prefers stdlib; TypeScript has one runtime
  dependency (`fflate`). New dependencies need strong justification.
- **Output parity:** Go is the reference implementation. Any change to
  extraction behavior in either language must keep both in sync — regenerate
  golden files and run both test suites:

  ```bash
  go run scripts/gen-golden-json.go && cd ts && npx vitest run
  ```

- **Contract linking:** Every source file declares the contract it implements
  in a header comment (e.g. `// CONTRACT:C1-EXTRACTOR.1.0`). Read the
  contract in `architecture/` before editing a file, and keep contracts and
  code in agreement.
- **Memory discipline:** Stream large files in Go (`io.Reader`, `io.LimitReader`);
  respect size limits in TypeScript. Target environments include browsers and
  IoT devices.

## Before Submitting a PR

1. `gofmt -w .` and `go vet ./...` pass
2. `go test -race ./...` passes
3. `cd ts && npx tsc --noEmit && npx vitest run` passes
4. Golden files regenerated if extraction output changed
5. New behavior covered by tests (table-driven in Go, parity tests in TS)

See `AGENTS.md` for the full testing cascade and development norms.

## Reporting Bugs

Open an issue with the file format, a minimal reproduction (attach a sample
file if possible — make sure it contains no sensitive data), expected output,
and actual output.
