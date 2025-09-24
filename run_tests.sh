#!/bin/bash

# Comprehensive test runner for yfinance-go
# Implements the testing strategy from Step 11

set -e

echo "🧪 Running yfinance-go test suite..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Parse command line arguments
RUN_UNIT=true
RUN_INTEGRATION=true
RUN_MAPPING=true
RUN_CROSSLANG=false
RUN_SOAK=false
RUN_RACE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            RUN_INTEGRATION=false
            RUN_MAPPING=false
            RUN_CROSSLANG=false
            RUN_SOAK=false
            shift
            ;;
        --integration-only)
            RUN_UNIT=false
            RUN_MAPPING=false
            RUN_CROSSLANG=false
            RUN_SOAK=false
            shift
            ;;
        --mapping-only)
            RUN_UNIT=false
            RUN_INTEGRATION=false
            RUN_CROSSLANG=false
            RUN_SOAK=false
            shift
            ;;
        --crosslang)
            RUN_CROSSLANG=true
            shift
            ;;
        --soak)
            RUN_SOAK=true
            shift
            ;;
        --race)
            RUN_RACE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --unit-only        Run only unit tests"
            echo "  --integration-only Run only integration tests"
            echo "  --mapping-only     Run only mapping regression tests"
            echo "  --crosslang        Run cross-language round-trip tests"
            echo "  --soak             Run soak tests (requires build tag)"
            echo "  --race             Run tests with race detector"
            echo "  --verbose          Verbose output"
            echo "  --help             Show this help"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Set test flags
TEST_FLAGS="-count=1"
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [ "$RUN_RACE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -race"
    print_warning "Running with race detector (slower)"
fi

# 1. Unit Tests
if [ "$RUN_UNIT" = true ]; then
    print_status "Running unit tests..."
    if go test $TEST_FLAGS ./tests/unit/...; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        exit 1
    fi
fi

# 2. Integration Tests
if [ "$RUN_INTEGRATION" = true ]; then
    print_status "Running integration tests..."
    if go test $TEST_FLAGS ./tests/integration/...; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        exit 1
    fi
fi

# 3. Mapping Regression Tests
if [ "$RUN_MAPPING" = true ]; then
    print_status "Running mapping regression tests..."
    if go test $TEST_FLAGS ./tests/mapping/...; then
        print_success "Mapping regression tests passed"
    else
        print_error "Mapping regression tests failed"
        exit 1
    fi
fi

# 4. Cross-Language Round-Trip Tests
if [ "$RUN_CROSSLANG" = true ]; then
    print_status "Running cross-language round-trip tests..."
    
    # Check if Python is available
    if ! command -v python3 &> /dev/null; then
        print_warning "Python3 not found, skipping cross-language tests"
    else
        # Check if virtual environment exists
        if [ ! -d ".venv" ]; then
            print_status "Creating Python virtual environment..."
            python3 -m venv .venv
        fi
        
        # Activate virtual environment and install dependencies
        print_status "Installing Python dependencies..."
        source .venv/bin/activate
        pip install -r tests/python/requirements.txt
        
        # Run cross-language tests
        if go test $TEST_FLAGS ./tests/crosslang/...; then
            print_success "Cross-language round-trip tests passed"
        else
            print_error "Cross-language round-trip tests failed"
            exit 1
        fi
    fi
fi

# 5. Soak Tests
if [ "$RUN_SOAK" = true ]; then
    print_status "Running soak tests..."
    print_warning "Soak tests can take 30+ minutes to complete"
    
    if go test -tags=soak $TEST_FLAGS -timeout=40m ./tests/soak/...; then
        print_success "Soak tests passed"
    else
        print_error "Soak tests failed"
        exit 1
    fi
fi

# 6. Run existing tests in internal packages
print_status "Running existing internal package tests..."
if go test $TEST_FLAGS ./internal/...; then
    print_success "Internal package tests passed"
else
    print_error "Internal package tests failed"
    exit 1
fi

# 7. Run CLI tests
print_status "Running CLI tests..."
if go test $TEST_FLAGS ./cmd/...; then
    print_success "CLI tests passed"
else
    print_error "CLI tests failed"
    exit 1
fi

print_success "All tests completed successfully! 🎉"

# Summary
echo ""
echo "=== Test Summary ==="
if [ "$RUN_UNIT" = true ]; then
    echo "✓ Unit tests"
fi
if [ "$RUN_INTEGRATION" = true ]; then
    echo "✓ Integration tests"
fi
if [ "$RUN_MAPPING" = true ]; then
    echo "✓ Mapping regression tests"
fi
if [ "$RUN_CROSSLANG" = true ]; then
    echo "✓ Cross-language round-trip tests"
fi
if [ "$RUN_SOAK" = true ]; then
    echo "✓ Soak tests"
fi
echo "✓ Internal package tests"
echo "✓ CLI tests"

echo ""
echo "To run specific test suites:"
echo "  $0 --unit-only"
echo "  $0 --integration-only"
echo "  $0 --mapping-only"
echo "  $0 --crosslang"
echo "  $0 --soak"
echo "  $0 --race"
