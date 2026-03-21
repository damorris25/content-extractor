# CLAUDE.md

Project instructions for Claude Code. These override defaults—follow them exactly.

## Project: content-extractor

Go library for extracting text content and metadata from unstructured file formats
(Office, ODF, PDF, images, audio, CAD). Inspired by Apache Tika. Commercial-friendly,
minimal dependencies, low memory footprint. Pure Go.

## Core Tenets

1. **Pure Go, Minimal Dependencies** — Prefer stdlib. When third-party deps are necessary, only MIT/BSD/Apache 2.0 licensed. No CGO unless absolutely unavoidable.
2. **Contract-First** — Write the contract, then the code. Every extractor implements a defined behavioral contract.
3. **Memory-Efficient** — Stream large files. Never buffer an entire file in memory. Target environments include IoT devices with limited RAM.
4. **Accuracy Over Speed** — Correct extraction is more important than fast extraction. But do both when possible.
5. **Commercially Safe** — No GPL/LGPL/AGPL dependencies. Every dependency must be auditable for license compliance.

## Cold Start (New Agent?)

**Read in order (5 min total):**
1. `README.md` → universal orientation (ALWAYS read first, every session)
2. `QUICKCONTEXT.md` → current state (verify against `git log --oneline -10`)
3. `TODO.md` → tasks + known issues + blockers
4. `AGENTS.md` → norms, testing cascade, contracts, collaboration

**Reference:**
- `DESIGN.md` → the philosophy (contracts, BDD, autonomy model)
- `architecture/` → contract documents
- `agents/` → subagent templates

## Commands

```bash
go mod download              # install dependencies
go build ./...               # build all packages
go test ./...                # run all tests
go test -race ./...          # run all tests with race detector
go test -run TestName ./pkg/ # run a single test
go test -bench=. ./...       # run benchmarks
go vet ./...                 # static analysis
golangci-lint run            # lint (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
gofmt -w .                   # format all Go files
```

## Structure

- `extractor/` — core extraction interface, registry, and result types
- `formats/office/` — Microsoft Office extractors (DOCX, XLSX, PPTX, legacy)
- `formats/odf/` — Open Document Format extractors (ODT, ODS, ODP)
- `formats/pdf/` — PDF text extraction
- `formats/image/` — image metadata extraction (EXIF, PNG metadata)
- `formats/audio/` — audio metadata extraction (ID3, Vorbis)
- `formats/cad/` — CAD text/metadata extraction (DXF, DWG)
- `metadata/` — metadata model and normalization
- `detect/` — MIME type / format detection (magic bytes + extension)
- `testdata/` — test fixture files
- `architecture/` — contract documents
- `scripts/` — enforcement and quality scanning

## Coding Style

Go standard conventions. `gofmt` formatting. No tabs-vs-spaces debate — Go settles it.

**Naming:**
- Packages: short, lowercase, no underscores (`detect`, `metadata`, `extractor`)
- Interfaces: behavior-named (`Extractor`, `Detector`), not `IExtractor`
- Errors: `Err` prefix as package-level vars (`ErrUnsupportedFormat`, `ErrCorruptFile`)
- Test files: co-located (`foo.go` → `foo_test.go`)

**Patterns:**
- Accept `io.Reader` or `io.ReadSeeker`, return concrete types
- Use `context.Context` as first parameter for cancellation
- Prefer table-driven tests
- Use `testdata/` directories for fixture files
- No `init()` functions — use explicit registration

**Memory discipline:**
- Never read an entire file into `[]byte` unless the file is known-small (<1MB)
- Use `io.LimitReader` for safety bounds
- Pool buffers with `sync.Pool` for hot paths
- Document memory expectations in contract behavioral tables

### Contract Linking (MANDATORY)

Every source file must have a header comment declaring which contract it
implements or belongs to. See `architecture/README.md` for the full system.

```go
// CONTRACT:C1-EXTRACTOR.1.0
package extractor
```

```go
// Architecture: CONTRACT:I1-METADATA.1.0
package metadata
```

When editing a file: **read its contract first.** When updating a contract:
**search for all implementing code** (`grep -rn "CONTRACT:{id}"`).
See `DESIGN.md` for the full philosophy.

## Testing

Co-locate unit tests beside source code (`extractor.go` → `extractor_test.go`).
Use `testdata/` directories for fixture files. Table-driven tests for behavioral
contracts. Always run with `-race` before pushing.

### Testing Cascade

**Iterate at the speed of a single test, not the full pipeline.** See `AGENTS.md` "Testing Cascade" for full rules.

| Tier | Speed | What | When |
|------|-------|------|------|
| T0 Vet/Build | <5s | Changed package | Every edit |
| T1 Targeted | <10s | Single test | Every change |
| T2 Package | <30s | One package suite | Before commit |
| T3 All tests | <60s | All unit tests | Before push |
| T4 Race+Bench | <2min | Race detector + benchmarks | Concurrency changes |
| T5 Full suite | <10min | Everything + lint | Release prep |

---

## Allowed Commands

The following command patterns are pre-approved for autonomous execution:

### Go Build & Test
- `go build ./...`, `go install ./...`
- `go test ./...`, `go test -race ./...`, `go test -run <name> ./path/`
- `go test -bench=. ./...`, `go test -benchmem ./...`
- `go vet ./...`
- `go mod download`, `go mod tidy`, `go mod verify`
- `go get <package>`, `go get -u <package>`
- `gofmt -w .`, `goimports -w .`
- `golangci-lint run`

### Git Operations
- `git status`, `git diff`, `git log`, `git show`
- `git add`, `git commit`, `git push`, `git pull`
- `git checkout`, `git branch`, `git fetch`
- `git stash`, `git reset`, `git restore`
- `git cherry-pick`, `git merge`, `git rebase`
- `git rm`, `git mv`, `git ls-tree`, `git show-ref`
- `git worktree`, `git remote`

### File & System Utilities
- `ls`, `tree`, `find`, `cat`, `head`, `tail`
- `echo`, `tee`, `stat`, `chmod`
- `strings`, `awk`, `xargs`, `test`
- `timeout`, `gtimeout`, `lsof`, `pkill`, `killall`, `jobs`
- `curl`

### Scripts
- `bash <script>.sh`
- Any script under `scripts/`
- Temp scripts in `/tmp/`

### Diagnostics
- Any script under `scripts/diagnostics/`
- `go-licenses check ./...`, `go-licenses csv ./...`

## Web Fetch Domains

Pre-approved for fetching:
- `pkg.go.dev`
- `github.com` (public repos for dependency evaluation)
- `tika.apache.org` (reference implementation docs)

## Agent Autonomy

**This project grants MAXIMUM autonomy.** Act decisively. Ship code. Don't ask permission for routine work.

### DO WITHOUT ASKING

1. **All coding tasks** - Write, edit, refactor, delete code freely
2. **All testing** - Run, write, fix, skip tests as needed
3. **All builds** - Build, rebuild, clean as needed
4. **All git operations** - Add, commit, push, branch, merge, rebase
5. **All dependency changes** - Add, remove, upgrade packages (respecting license constraints)
6. **All documentation** - Create, update, reorganize, archive docs
7. **All file operations** - Create, move, rename, delete files
8. **Bug fixes** - Fix bugs immediately without discussion
9. **Refactoring** - Improve code quality, reduce duplication
10. **Test fixes** - Update broken tests, add missing coverage
11. **Config changes** - Update configs, env vars, build settings
12. **Minor features** - Small enhancements that follow existing patterns
13. **Error handling** - Add/improve error handling and logging
14. **Performance fixes** - Optimize slow code paths

### CHECK TEMPLATES FIRST

Before doing specialized tasks (security audit, code review, etc.),
check `agents/subagent-prompts-index.md` for a matching template.

### ASK ONLY FOR

**Fundamental architectural decisions** that would be hard to reverse:

1. **New third-party dependencies** - Must evaluate license compatibility first
2. **Core interface changes** - Modifications to `Extractor`, `Result`, or metadata model
3. **CGO introduction** - Adding any C dependency (violates pure-Go preference)
4. **New format families** - Adding format categories beyond the defined scope
5. **Breaking changes** - Changes that break the public API
6. **GPL-licensed anything** - Never add without explicit authorization (answer is almost always no)

**When in doubt:** If a change follows existing patterns and is reversible, just do it. If it establishes a new pattern or is hard to undo, enter plan mode.

### Force Operations (require explicit user request)

- `git push --force` to shared branches
- `git reset --hard` on commits others might have
- Adding any dependency with a copyleft license

## Active Workstreams

- **P0** — Architecture contracts + project scaffolding (in progress)
- **P1** — Core extraction interface + first format extractors

## Environment Variables

```bash
# No environment variables required — this is a library.
# Test configuration uses Go build tags and test flags only.
```

---

## TODO Tracking Methodology (MANDATORY)

**This is a hard requirement. Follow exactly.**

### The Two-Tag System

| Tag | Meaning | Action Required |
|-----|---------|-----------------|
| `TODO:` | Untracked work item | Must be tracked before commit |
| `TRACKED-TASK:` | Already in TODO.md/docs | Periodically verify still documented |

### Workflow

**When you add a TODO in code:**
```go
// TODO: Handle edge case for X
```

**Before committing, you MUST either:**
1. Fix it immediately (remove the TODO), OR
2. Track it in `TODO.md` and convert to:
```go
// TRACKED-TASK: Handle edge case for X - see TODO.md "Code Debt"
```

### Pre-Commit Checklist

**Run before every commit:**
```bash
# Find untracked TODOs (should be 0 before commit)
grep -rn "TODO:" --include="*.go" .

# Find tracked tasks (audit these periodically)
grep -rn "TRACKED-TASK:" --include="*.go" .
```

**If untracked TODOs exist, you must:**
1. Add each to `TODO.md` under appropriate section
2. Convert `TODO:` → `TRACKED-TASK:` in source
3. Re-run check to confirm zero untracked TODOs

### Periodic Scrub (Weekly or Per-Sprint)

For `TRACKED-TASK:` items:
1. Verify each is still in `TODO.md` or relevant doc
2. If completed, remove from both code and docs
3. If stale (no longer relevant), remove from both
4. If doc reference is wrong, update the comment

### Why This Matters

- **TODOs get lost.** Scattered comments with no tracking = technical debt amnesia.
- **Agents work async.** Next agent needs to know what's pending without reading all code.
- **State consistency.** Code comments and TODO.md must agree.

---
