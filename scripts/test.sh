#!/bin/bash

# =============================================================================
# Test Script for Pipeline Architecture
# Runs unit tests, integration tests, and generates reports
# =============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COVERAGE_FILE="coverage.txt"
REPORT_DIR="test-reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_test() { echo -e "${BLUE}[TEST]${NC} $1"; }

# Ensure we're in the project root
cd "$SCRIPT_DIR/.."

# Create reports directory
mkdir -p "$REPORT_DIR"

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."

    go test -v -race \
        -coverprofile="${COVERAGE_FILE}" \
        -covermode=atomic \
        ./internal/... ./pkg/... 2>&1 | tee "${REPORT_DIR}/unit-tests-${TIMESTAMP}.log"

    local result=$?
    if [[ $result -eq 0 ]]; then
        TESTS_PASSED=$((TESTS_PASSED + $(grep -c "PASS:" "${REPORT_DIR}/unit-tests-${TIMESTAMP}.log" || 0)))
        log_info "Unit tests completed"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "Unit tests failed"
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."

    if [[ ! -d "tests/integration" ]]; then
        log_warn "No integration tests found"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        return 0
    fi

    # Integration tests might require external services
    # You can skip them if DB_REDIS_URL is not set
    if [[ -z "${DATABASE_URL:-}" ]] && [[ -z "${SKIP_INTEGRATION:-}" ]]; then
        log_warn "DATABASE_URL not set. Skipping integration tests."
        log_warn "Set DATABASE_URL to run integration tests."
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        return 0
    fi

    go test -v \
        -coverprofile="coverage-integration.txt" \
        -covermode=atomic \
        ./tests/integration/... 2>&1 | tee "${REPORT_DIR}/integration-tests-${TIMESTAMP}.log"

    local result=$?
    if [[ $result -eq 0 ]]; then
        log_info "Integration tests completed"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "Integration tests failed"
    fi
}

# Run end-to-end tests
run_e2e_tests() {
    log_info "Running end-to-end tests..."

    if [[ ! -d "tests/e2e" ]]; then
        log_warn "No e2e tests found"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        return 0
    fi

    # E2E tests require running application
    # They should be run in a separate environment
    log_warn "E2E tests require a running application cluster"
    log_warn "Skipping E2E tests in this run"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
}

# Generate coverage report
generate_coverage_report() {
    log_info "Generating coverage report..."

    if [[ -f "$COVERAGE_FILE" ]]; then
        # Generate coverage breakdown
        echo "## Coverage Report" > "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        echo "" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        echo "Generated: $(date)" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        echo "" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"

        echo "### Overall Coverage" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        go tool cover -func="${COVERAGE_FILE}" | tail -5 >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        echo "" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"

        echo "### Package Coverage" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        go tool cover -func="${COVERAGE_FILE}" | grep -E "^(internal|pkg|cmd)" >> "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"

        log_info "Coverage report generated: ${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
        cat "${REPORT_DIR}/coverage-summary-${TIMESTAMP}.md"
    else
        log_warn "Coverage file not found"
    fi
}

# Generate HTML coverage report
generate_html_coverage() {
    log_info "Generating HTML coverage report..."

    if command -v go &> /dev/null; then
        go tool cover -html="${COVERAGE_FILE}" -o "${REPORT_DIR}/coverage-${TIMESTAMP}.html" 2>/dev/null || true

        if [[ -f "${REPORT_DIR}/coverage-${TIMESTAMP}.html" ]]; then
            log_info "HTML coverage report: ${REPORT_DIR}/coverage-${TIMESTAMP}.html"
        fi
    fi
}

# Run linter checks
run_linters() {
    log_info "Running linter checks..."

    # Check for golangci-lint
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./... 2>&1 | tee "${REPORT_DIR}/lint-${TIMESTAMP}.log" || true
    else
        # Fall back to basic checks
        log_info "golangci-lint not found, running basic checks..."

        # Check formatting
        if ! gofmt -d . 2>&1 | grep -q .; then
            log_info "Code formatting is correct"
        else
            log_warn "Code formatting issues found. Run 'gofmt -w .' to fix."
            gofmt -d .
        fi

        # Check for common issues
        go vet ./... 2>&1 | tee "${REPORT_DIR}/vet-${TIMESTAMP}.log" || true
    fi
}

# Run benchmark tests
run_benchmarks() {
    log_info "Running benchmark tests..."

    go test -bench=. -benchmem -run=^$ ./... 2>&1 | tee "${REPORT_DIR}/benchmarks-${TIMESTAMP}.log" || true

    log_info "Benchmark results saved to: ${REPORT_DIR}/benchmarks-${TIMESTAMP}.log"
}

# Generate JUnit XML report (for CI/CD)
generate_junit_report() {
    log_info "Generating JUnit XML report..."

    # Run tests with JUnit output format
    go test -v -coverprofile="coverage-junit.xml" \
        -covermode=atomic \
        ./... 2>&1 | go-junit-report > "${REPORT_DIR}/junit-${TIMESTAMP}.xml" || true

    if [[ -f "${REPORT_DIR}/junit-${TIMESTAMP}.xml" ]]; then
        log_info "JUnit report: ${REPORT_DIR}/junit-${TIMESTAMP}.xml"
    fi
}

# Print test summary
print_summary() {
    echo ""
    echo "============================================"
    echo -e "${GREEN}Test Summary${NC}"
    echo "============================================"
    echo -e "Tests Passed:  ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests Failed:  ${RED}${TESTS_FAILED}${NC}"
    echo -e "Tests Skipped: ${YELLOW}${TESTS_SKIPPED}${NC}"
    echo "============================================"
    echo "Reports Directory: ${REPORT_DIR}"
    echo "Coverage File: ${COVERAGE_FILE}"
    echo ""

    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Some tests failed!"
        exit 1
    fi

    log_info "All tests passed!"
}

# Main function
main() {
    log_info "Starting test suite..."
    log_info "Timestamp: ${TIMESTAMP}"
    echo ""

    # Parse command line arguments
    case "${1:-all}" in
        all)
            run_unit_tests
            run_integration_tests
            run_linters
            generate_coverage_report
            generate_html_coverage
            ;;
        unit)
            run_unit_tests
            generate_coverage_report
            generate_html_coverage
            ;;
        integration)
            run_integration_tests
            ;;
        e2e)
            run_e2e_tests
            ;;
        lint)
            run_linters
            ;;
        coverage)
            generate_coverage_report
            generate_html_coverage
            ;;
        benchmark)
            run_benchmarks
            ;;
        junit)
            generate_junit_report
            ;;
        ci)
            run_unit_tests
            run_integration_tests
            run_linters
            generate_coverage_report
            generate_junit_report
            ;;
        *)
            echo "Usage: $0 [all|unit|integration|e2e|lint|coverage|benchmark|junit|ci]"
            echo ""
            echo "Commands:"
            echo "  all         - Run all tests and generate reports (default)"
            echo "  unit        - Run unit tests only"
            echo "  integration - Run integration tests only"
            echo "  e2e         - Run end-to-end tests only"
            echo "  lint        - Run linters only"
            echo "  coverage    - Generate coverage reports"
            echo "  benchmark   - Run benchmark tests"
            echo "  junit       - Generate JUnit XML report"
            echo "  ci          - Full CI test suite"
            exit 1
            ;;
    esac

    print_summary
}

main "$@"