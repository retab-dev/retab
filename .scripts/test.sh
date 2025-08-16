#!/bin/bash

# Main test script for open-source SDK
# Runs tests for both Node.js and Python clients

set -e  # Exit on any error

echo "🚀 Running SDK tests..."
echo "========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to run a test script and capture results
run_test() {
    local test_name="$1"
    local test_path="$2"
    local client_dir="$3"
    
    echo -e "\n${YELLOW}📋 Running $test_name tests...${NC}"
    echo "Path: $test_path"
    echo "Client directory: $client_dir"
    echo "----------------------------------------"
    
    if [ -f "$test_path" ]; then
        cd "$client_dir"
        if bash "$test_path"; then
            echo -e "${GREEN}✅ $test_name tests passed${NC}"
            return 0
        else
            echo -e "${RED}❌ $test_name tests failed${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Test script not found: $test_path${NC}"
        return 1
    fi
}

# Store the original directory
ORIGINAL_DIR=$(pwd)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_ROOT="$(dirname "$SCRIPT_DIR")"

# Initialize test results
FAILED_TESTS=()
PASSED_TESTS=()

# Run Node.js client tests
if run_test "Node.js Client" "$SDK_ROOT/clients/node/.scripts/test.sh" "$SDK_ROOT/clients/node"; then
    PASSED_TESTS+=("Node.js Client")
else
    FAILED_TESTS+=("Node.js Client")
fi

# Return to original directory
cd "$ORIGINAL_DIR"

# Run Python client tests
if run_test "Python Client" "$SDK_ROOT/clients/python/.scripts/test.sh" "$SDK_ROOT/clients/python"; then
    PASSED_TESTS+=("Python Client")
else
    FAILED_TESTS+=("Python Client")
fi

# Return to original directory
cd "$ORIGINAL_DIR"

# Print summary
echo -e "\n${YELLOW}📊 Test Summary${NC}"
echo "========================="

if [ ${#PASSED_TESTS[@]} -gt 0 ]; then
    echo -e "${GREEN}✅ Passed tests:${NC}"
    for test in "${PASSED_TESTS[@]}"; do
        echo "   - $test"
    done
fi

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "   - $test"
    done
    echo -e "\n${RED}Some tests failed. Please check the output above for details.${NC}"
    exit 1
else
    echo -e "\n${GREEN}🎉 All tests passed successfully!${NC}"
    exit 0
fi
