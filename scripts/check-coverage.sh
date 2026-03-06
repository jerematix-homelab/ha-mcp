#!/usr/bin/env bash
# Per-file coverage enforcement for ha-mcp
# Usage: bash scripts/check-coverage.sh coverage.out [threshold]
# Default threshold: 80%
#
# Files containing "// coverage-exempt: <reason>" are skipped.

set -euo pipefail

COVERAGE_FILE="${1:-coverage.out}"
THRESHOLD="${2:-80}"
MODULE="github.com/zorak1103/ha-mcp"

if [ ! -f "$COVERAGE_FILE" ]; then
  echo "Error: coverage file '$COVERAGE_FILE' not found"
  echo "Run: go test -coverprofile=coverage.out -covermode=atomic ./..."
  exit 1
fi

echo "Per-file coverage check (threshold: ${THRESHOLD}%)"
echo "Coverage file: $COVERAGE_FILE"
echo "──────────────────────────────────────────────────────────────"

FAILED=0
PASSED=0
SKIPPED=0

# Parse per-file coverage from 'go tool cover -func' output.
# Output format: "github.com/zorak1103/ha-mcp/internal/foo/bar.go:42: FuncName   100.0%"
# We collect the last line per file (which is the file total when using -func).
# Actually go tool cover -func doesn't emit per-file totals directly — we compute them.

# Build a map of file -> covered/total statements by parsing the raw coverage.out profile.
# coverage.out format: "mode: atomic" then lines of "file.go:line.col,line.col stmts count"
# stmts = number of statements in block, count = execution count (0=not covered, >0=covered)

declare -A FILE_COVERED
declare -A FILE_TOTAL

while IFS= read -r line; do
  # Skip mode line
  [[ "$line" == mode:* ]] && continue

  # Parse: pkg/file.go:start,end stmts count
  # e.g.: github.com/zorak1103/ha-mcp/internal/mcp/server.go:42.16,44.2 1 1
  if [[ "$line" =~ ^([^:]+):([0-9]+\.[0-9]+),([0-9]+\.[0-9]+)[[:space:]]+([0-9]+)[[:space:]]+([0-9]+)$ ]]; then
    file="${BASH_REMATCH[1]}"
    stmts="${BASH_REMATCH[4]}"
    count="${BASH_REMATCH[5]}"

    FILE_TOTAL["$file"]=$(( ${FILE_TOTAL["$file"]:-0} + stmts ))
    if [ "$count" -gt 0 ]; then
      FILE_COVERED["$file"]=$(( ${FILE_COVERED["$file"]:-0} + stmts ))
    else
      FILE_COVERED["$file"]=${FILE_COVERED["$file"]:-0}
    fi
  fi
done < "$COVERAGE_FILE"

for file in $(echo "${!FILE_TOTAL[@]}" | tr ' ' '\n' | sort); do
  total="${FILE_TOTAL[$file]}"
  covered="${FILE_COVERED[$file]:-0}"

  # Convert module path to filesystem path for exemption check
  rel_path="${file#${MODULE}/}"
  fs_path="$rel_path"

  # Check for coverage-exempt comment in the source file
  if [ -f "$fs_path" ] && grep -q "coverage-exempt:" "$fs_path" 2>/dev/null; then
    reason=$(grep -m1 "coverage-exempt:" "$fs_path" | sed 's/.*coverage-exempt:[[:space:]]*//')
    echo "SKIP  $rel_path (exempt: $reason)"
    SKIPPED=$(( SKIPPED + 1 ))
    continue
  fi

  if [ "$total" -eq 0 ]; then
    echo "SKIP  $rel_path (no statements)"
    SKIPPED=$(( SKIPPED + 1 ))
    continue
  fi

  # Calculate percentage (bash integer arithmetic: multiply by 100 first)
  pct=$(( covered * 100 / total ))

  if [ "$pct" -ge "$THRESHOLD" ]; then
    printf "PASS  %-70s %3d%%\n" "$rel_path" "$pct"
    PASSED=$(( PASSED + 1 ))
  else
    printf "FAIL  %-70s %3d%% (need %d%%)\n" "$rel_path" "$pct" "$THRESHOLD"
    FAILED=$(( FAILED + 1 ))
  fi
done

echo "──────────────────────────────────────────────────────────────"
echo "Results: ${PASSED} passed, ${FAILED} failed, ${SKIPPED} skipped"

if [ "$FAILED" -gt 0 ]; then
  echo ""
  echo "Coverage below ${THRESHOLD}% in ${FAILED} file(s)."
  echo "Add '// coverage-exempt: <reason>' to a file to exclude it from enforcement."
  exit 1
fi

echo "All files meet the ${THRESHOLD}% coverage threshold."
