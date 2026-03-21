# Repository Guidelines

## Read Before Coding

**The Cold Start Quad (every session, every agent, no exceptions):**
1. `README.md` — universal orientation (ALWAYS first)
2. `QUICKCONTEXT.md` — current state (verify against `git log --oneline -10`)
3. `TODO.md` — tasks + known issues + blockers
4. `AGENTS.md` (this file) — norms, contracts, collaboration

**Reference (read as needed):**
- `DESIGN.md` — the philosophy (contracts are the operating system)
- `architecture/CONTRACT-REGISTRY.md` — contract index
- `practices/` — specialized guidance (orchestration, E2E, deployment, worktrees)
- **ASK CLI** — query role-based agents: `ask architect "question"`, `ask product "question"`, `ask steward summary`. See `bin/ask help`.

## Core Tenets

1. **Pure Go, Minimal Dependencies** — Prefer stdlib. When third-party deps are necessary, only MIT/BSD/Apache 2.0 licensed. No CGO unless absolutely unavoidable.
2. **Contract-First** — Write the contract, then the code. Every extractor implements a defined behavioral contract.
3. **Memory-Efficient** — Stream large files. Never buffer an entire file in memory. Target environments include IoT devices with limited RAM.
4. **Accuracy Over Speed** — Correct extraction is more important than fast extraction. But do both when possible.
5. **Commercially Safe** — No GPL/LGPL/AGPL dependencies. Every dependency must be auditable for license compliance.

---

## Agent Autonomy

**Maximum autonomy granted.** Act decisively. Ship code. Don't ask permission for routine work.

### Full Authority (no approval needed)
- Write, edit, refactor, delete code
- Run, write, fix tests
- Git: commit, push, branch, merge, rebase
- Add/remove/upgrade dependencies (respecting license constraints)
- Create, update, reorganize, archive documentation
- Fix bugs, improve error handling, optimize performance
- Implement features that follow existing patterns

### Requires Discussion (enter plan mode)
Only **fundamental architectural decisions** that are hard to reverse:
- New third-party dependencies (must evaluate license + necessity)
- Changes to the core `Extractor` interface or `Result` type
- Adding CGO dependencies (violates pure-Go preference)
- New format families beyond the defined scope
- Changes to the metadata model (`I1-METADATA` contract)
- Breaking changes to the public API

### Never Without Explicit Request
- `git push --force` to shared branches
- `git reset --hard` on commits others have
- Adding any GPL/LGPL/AGPL licensed dependency
- Modifying production secrets

**Rule of thumb:** If it follows existing patterns and is reversible — just do it. If it establishes new patterns or is hard to undo — plan mode.

---

## Cold Start Methodology (MANDATORY for New Agent Sessions)

**When starting a new session, always perform this sanity check before acting:**

### Step 1: Verify Document Freshness (5 min)
Don't trust docs blindly. Cross-reference against actual state:

```bash
# 1. Check current branch (docs may reference wrong branch)
git branch --show-current
git log --oneline -10

# 2. Compare QUICKCONTEXT.md branch claim against reality
grep -i "branch" QUICKCONTEXT.md

# 3. Check TODO.md "Last synced" date
head -10 TODO.md

# 4. Verify recent commits match documented priorities
git log --oneline -20 | head -10
```

### Step 1b: Verify Ground Truth Metrics

```bash
# Compare documented metrics against codebase reality
[ -x scripts/check-ground-truth.sh ] && ./scripts/check-ground-truth.sh
```

If metrics have drifted, update `METRICS` BEFORE proceeding with your task.

### Step 2: Identify Discrepancies
Look for these common drift patterns:
- **Branch mismatch**: Docs say one branch, you're on another
- **Stale dates**: "Last Updated" > 2 weeks old warrants scrutiny
- **Missing features**: Grep for features in code vs docs

### Step 3: Update Before Acting
If you find discrepancies:
1. **Minor drift**: Update the doc inline while working
2. **Major drift**: Update docs FIRST, then proceed with task
3. **Conflicting signals**: Ask user for clarification

### Step 4: Strategic Assessment
Before diving into code, ask:
- What's the **actual** current state? (git log, file structure)
- What's the **documented** next step? (TODO.md priorities)
- Do they align? If not, which is authoritative?
- Are there **blocked** items I should avoid?

---

## Multi-Agent Orchestration

> For subagent templates, fan-out patterns, pre-launch audit, and feature
> inventory protocol, see **`practices/multi-agent-orchestration.md`**.
>
> **Key rules (always apply):** Subagents writing code MUST use worktree
> isolation. Subagents MUST commit before completing. Run a pre-launch audit
> before any parallel agent campaign.

---

## Testing Cascade (MANDATORY)

**Fast inner loops, rigorous outer gates.** Iterate at the speed of a single test file.

### The Tiers

| Tier | Name | Target | Speed | When to Run | Command |
|------|------|--------|-------|-------------|---------|
| **T0** | Vet/Build | Changed package | <5s | Every meaningful edit | `go vet ./pkg/...` |
| **T1** | Targeted | Single test file | <10s | Every change cycle | `go test -run TestName ./pkg/path/` |
| **T2** | Package | One package's suite | <30s | Before committing | `go test ./pkg/path/...` |
| **T3** | All tests | All unit/integration | <60s | Before pushing | `go test ./...` |
| **T4** | Race + Bench | Race detection + benchmarks | <2min | Concurrency changes | `go test -race ./... && go test -bench=. ./...` |
| **T5** | Full suite | Everything + lint | <10min | Release prep | `go test -race ./... && golangci-lint run` |

### Rules for Agents

1. **Iterate at T1.** Your inner loop is: edit — run the specific test — edit. This should take <10 seconds.
2. **Promote on success.** Only escalate to the next tier when the current one passes. Never skip tiers.
3. **Background the expensive tiers.** T3+ should run in background sub-agents while you keep coding.
4. **Use `-run` to target.** `go test -run TestExtractDOCX ./formats/office/`
5. **Never run T5 in your inner loop.** T5 is a release gate, not a development tool.
6. **Always run with `-race` before pushing.** Data races in concurrent extractors are critical bugs.

### Anti-Patterns

- **Run T5 after every change** — You'll spend more time waiting than coding.
- **Skip T1 and go straight to T3** — T3 runs all tests. T1 runs one. The feedback delay kills iteration speed.
- **Ignore `-race`** — This library will be used concurrently. Race conditions are silent killers.
- **Run `go test ./...` to check one function** — Find the exact test. Run that one.

---

## Testing Expectations

### Go Testing Conventions
- Test files are co-located with source: `extractor.go` → `extractor_test.go`
- Table-driven tests for behavioral contracts
- Use `testdata/` directories for fixture files
- Use `testing.Short()` to skip slow tests in T1/T2 tiers
- Always run with `-race` flag in T4+ tiers
- Benchmark critical paths: `func BenchmarkExtractDOCX(b *testing.B)`

### The Scout Rule: Zero Tolerance for Broken Tests

**"You're a scout. You leave the camp cleaner than when you came."**

| Situation | Action |
|-----------|--------|
| Skipped test | Fix the skip. Scope it properly or delete it. Never leave a `t.Skip`. |
| Failing test | P0. Fix it before continuing your task. No exceptions. |
| Flaky test | Stabilize it or delete it. Flaky = lying about coverage. |
| Obsolete test (OBE) | Remove it carefully. Verify the behavior is gone or covered elsewhere. |

### Contract-Driven Development

**Contracts are the operating system.** See `DESIGN.md` for the full
philosophy and `architecture/README.md` for the naming/linking system.

**The rules:**
1. **Don't implement without a contract.** Write the contract first.
2. **Don't modify code without checking its contract.** Read the `CONTRACT:`
   header comment, then read the contract document.
3. **Don't update a contract without searching implementations.**
   `grep -rn "CONTRACT:{id}" .` finds all implementing code.
4. **Contract changes that break interfaces — plan mode.**

**Every source file** has a header declaring its contract:
```go
// CONTRACT:C1-EXTRACTOR.1.0
package extractor
```

**Contract tests are king.** If a contract test fails, nothing ships.

---

## Dependency Review Protocol

Before adding ANY third-party dependency:

1. **Check the license.** Only MIT, BSD-2-Clause, BSD-3-Clause, Apache-2.0, ISC, and Unlicense are acceptable. Run `go-licenses check` or manually inspect.
2. **Check transitive dependencies.** A BSD library that depends on a GPL library is effectively GPL.
3. **Evaluate necessity.** Can this be implemented in <200 lines of Go? If yes, prefer the stdlib approach.
4. **Check maintenance status.** Abandoned libraries become security liabilities.
5. **Document the decision.** Add a note to the contract's Dependencies section.

**Banned licenses:** GPL-2.0, GPL-3.0, LGPL-2.1, LGPL-3.0, AGPL-3.0, SSPL, BSL, CPAL, OSL

---

## Documentation Maintenance Policy

**Principle**: Code and docs must stay in sync.

**After every code change or task completion**, update affected files:

| Change Type | Docs to Update |
|-------------|----------------|
| **New extractor** | Architecture contract, CONTRACT-REGISTRY.md, README format list |
| **API change** | Contract first (contract-first!), then implementation |
| **Bug fix** | Relevant contract if it clarifies behavior |
| **New dependency** | Contract Dependencies section, license audit |
| **Test change** | METRICS if coverage shifts |

### Single Source of Truth for Metrics

| Metric | Computed From | Verified By | Referenced In |
|--------|--------------|-------------|---------------|
| Test count | `go test -count=1 ./...` | `check-ground-truth.sh` | QUICKCONTEXT.md |
| Contract count | `architecture/CONTRACT-*.md` | `check-ground-truth.sh` | CONTRACT-REGISTRY.md |
| Format count | Registered extractors | `check-ground-truth.sh` | README.md |

---

## Quality Gates (run before every push)

```bash
go vet ./...
go test -race ./...
golangci-lint run
scripts/ci-check.sh
```

**Skip Policy**
- No skipping core contract tests. If they flake, fix or revert.

## Commit & PR Guidelines

Use conventional prefixes (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `build:`). PRs must describe the user-facing impact, list touched packages, link to contracts, and note new tests. Never commit secrets.

---

## TODO Tracking (MANDATORY PRE-COMMIT)

**This is a hard requirement for all agents.**

### Two-Tag System

| Tag | Meaning | Commit Allowed? |
|-----|---------|-----------------|
| `TODO:` | Untracked work | No - must track first |
| `TRACKED-TASK:` | In TODO.md/docs | Yes |

### Before Every Commit

```bash
# 1. Find untracked TODOs (should be 0 before commit)
grep -rn "TODO:" --include="*.go" .

# 2. If untracked TODOs found:
#    - Add to TODO.md
#    - Convert TODO: -> TRACKED-TASK: in code
#    - Re-run check

# 3. Only commit when untracked TODOs = 0
```

### Periodic Scrub

Weekly, audit `TRACKED-TASK:` comments:
1. Verify each is still documented in TODO.md
2. Remove completed items from both code and docs
3. Update stale references

**See:** `CLAUDE.md` "TODO Tracking Methodology" for full details.
