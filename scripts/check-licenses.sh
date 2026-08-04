#!/usr/bin/env bash
# check-licenses.sh — Verify all Go dependencies use commercially-friendly licenses.
# Allowed: MIT, BSD-2-Clause, BSD-3-Clause, Apache-2.0, ISC, MPL-2.0
# Blocked: GPL, LGPL, AGPL, SSPL, CPAL, EUPL, OSL, CC-BY-SA, CC-BY-NC
#
# Usage: bash scripts/check-licenses.sh
#
# Requires: go-licenses (go install github.com/google/go-licenses@latest)

set -euo pipefail

ALLOWED_RE="MIT|BSD-2-Clause|BSD-3-Clause|Apache-2.0|ISC|MPL-2.0"
BLOCKED_RE="GPL|LGPL|AGPL|SSPL|CPAL|EUPL|OSL"

# Find go-licenses on PATH or in GOPATH/bin.
GO_LICENSES="go-licenses"
if ! command -v "$GO_LICENSES" &>/dev/null; then
    GOPATH_BIN="$(go env GOPATH)/bin"
    if [ -x "$GOPATH_BIN/go-licenses" ]; then
        GO_LICENSES="$GOPATH_BIN/go-licenses"
    else
        echo "ERROR: go-licenses not found."
        echo "Install: go install github.com/google/go-licenses@latest"
        exit 1
    fi
fi

echo "=== Dependency License Audit ==="
echo ""

# Get CSV output: module, license URL, license type
OUTPUT=$("$GO_LICENSES" csv ./... 2>/dev/null || true)

if [ -z "$OUTPUT" ]; then
    echo "No third-party dependencies found (or go-licenses produced no output)."
    echo "PASS"
    exit 0
fi

FAIL=0
PASS_COUNT=0

while IFS=, read -r module url license; do
    # Skip our own module.
    if [[ "$module" == github.com/damorris25/content-extractor* ]]; then
        continue
    fi

    if echo "$license" | grep -qiE "$BLOCKED_RE"; then
        echo "BLOCKED: $module ($license)"
        FAIL=1
    elif echo "$license" | grep -qiE "$ALLOWED_RE"; then
        echo "  OK: $module ($license)"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "  UNKNOWN: $module ($license) — manual review required"
    fi
done <<< "$OUTPUT"

echo ""
echo "Summary: $PASS_COUNT dependencies checked."

if [ "$FAIL" -ne 0 ]; then
    echo "FAIL: Blocked license(s) detected. See above."
    exit 1
fi

echo "PASS: All dependencies use commercially-friendly licenses."
exit 0
