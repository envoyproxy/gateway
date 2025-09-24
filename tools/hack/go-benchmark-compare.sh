#!/bin/bash

# go-benchmark-compare.sh - Compare benchmark results between PR and main branch
# Usage: go-benchmark-compare.sh [--help]

set -euo pipefail

# Environment variables with defaults
REGRESSION_THRESHOLD=${REGRESSION_THRESHOLD:-5}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Function to display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Compare benchmark performance between PR branch and main branch.

OPTIONS:
    --help          Show this help message

ENVIRONMENT VARIABLES:
    REGRESSION_THRESHOLD    Regression threshold percentage (default: 5)

EXIT CODES:
    0    Success (no significant regressions)
    1    Error in execution
    2    Significant performance regressions detected
EOF
}

# Function to cleanup temporary files
cleanup() {
    local exit_code=$?
    rm -f "$REPO_ROOT/pr-bench.txt" "$REPO_ROOT/main-bench.txt" "$REPO_ROOT/comparison.txt"
    exit $exit_code
}

# Function to log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# Function to run benchmarks with error handling
run_benchmark() {
    local output_file=$1
    local branch_name=$2

    log "Running benchmarks on $branch_name..."
    if ! make -C "$REPO_ROOT" go-benchmark > "$output_file" 2>&1; then
        log "ERROR: Failed to run benchmarks on $branch_name"
        cat "$output_file"
        return 1
    fi
    log "Benchmarks completed for $branch_name"
}

# Function to check for regressions
check_regressions() {
    local comparison_file=$1
    local threshold=$2
    local regression_found=false

    log "Analyzing for performance regressions (threshold: ${threshold}%)..."

    # Check for CPU time regressions
    if grep -E "^\s*Benchmark.*\+[${threshold}-9]\.[0-9]+%|^\s*Benchmark.*\+[1-9][0-9]\.[0-9]+%" "$comparison_file" >/dev/null 2>&1; then
        log "WARNING: CPU time regression detected (>${threshold}%)"
        regression_found=true
    fi

    # Check for memory allocation regressions
    if grep -E "B/op.*\+[${threshold}-9]\.[0-9]+%|B/op.*\+[1-9][0-9]\.[0-9]+%" "$comparison_file" >/dev/null 2>&1; then
        log "WARNING: Memory allocation regression detected (>${threshold}%)"
        regression_found=true
    fi

    if [[ "$regression_found" == "true" ]]; then
        log "RESULT: Performance regressions detected - consider reviewing the performance impact"
        return 2
    else
        log "RESULT: No significant performance regressions detected"
        return 0
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            usage
            exit 0
            ;;
        *)
            log "ERROR: Unknown option $1"
            usage
            exit 1
            ;;
    esac
done

# Validate threshold is a number
if ! [[ "$REGRESSION_THRESHOLD" =~ ^[0-9]+$ ]]; then
    log "ERROR: REGRESSION_THRESHOLD must be a positive integer, got: $REGRESSION_THRESHOLD"
    exit 1
fi

# Set up cleanup trap
trap cleanup EXIT INT TERM

# Change to repository root
cd "$REPO_ROOT"

# Verify we're in a git repository
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    log "ERROR: Not in a git repository"
    exit 1
fi

# Verify benchstat is available
if ! command -v benchstat >/dev/null 2>&1; then
    log "ERROR: benchstat not found. Install with: go install golang.org/x/perf/cmd/benchstat@latest"
    exit 1
fi

log "Starting benchmark comparison with ${REGRESSION_THRESHOLD}% regression threshold"

# Store current state
current_branch=$(git rev-parse --abbrev-ref HEAD)
current_commit=$(git rev-parse HEAD)
log "Current branch: $current_branch, commit: ${current_commit:0:8}"

# Run benchmarks on current PR branch
if ! run_benchmark "pr-bench.txt" "$current_branch"; then
    exit 1
fi

# Switch to main branch and run benchmarks
log "Switching to main branch..."
if ! git checkout origin/main --quiet; then
    log "ERROR: Failed to checkout main branch"
    exit 1
fi

if ! run_benchmark "main-bench.txt" "origin/main"; then
    # Return to original state before failing
    git checkout "$current_commit" --quiet || true
    exit 1
fi

# Return to original commit
log "Returning to original commit..."
if ! git checkout "$current_commit" --quiet; then
    log "WARNING: Failed to return to original commit"
fi

# Compare benchmarks using benchstat
log "Comparing benchmark results..."
if ! benchstat main-bench.txt pr-bench.txt > comparison.txt 2>&1; then
    log "WARNING: Benchstat comparison had issues, but continuing..."
fi

# Display results
echo
echo "Benchmark Comparison Results:"
echo "======================================"
cat comparison.txt
echo "======================================"
echo

# Check for regressions and exit with appropriate code
check_regressions "comparison.txt" "$REGRESSION_THRESHOLD"
exit_code=$?

log "Benchmark comparison completed"
exit $exit_code