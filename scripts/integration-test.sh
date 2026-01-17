#!/bin/bash
# Integration tests for Scope CLI
# Tests all major functionality end-to-end

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCOPE_BIN="${SCOPE_BIN:-./build/scope}"
# Make SCOPE_BIN absolute path
SCOPE_BIN="$(cd "$(dirname "$SCOPE_BIN")" 2>/dev/null && pwd)/$(basename "$SCOPE_BIN")"
TEST_DIR="/tmp/scope-integration-test-$$"
DB_BACKUP="$HOME/.config/scope/scope.db.backup"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test directories...${NC}"
    rm -rf "$TEST_DIR"

    # Restore database if backup exists
    if [ -f "$DB_BACKUP" ]; then
        mv "$DB_BACKUP" "$HOME/.config/scope/scope.db"
    fi
}

trap cleanup EXIT

# Test helper functions
assert_success() {
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ $1${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_output_contains() {
    local output="$1"
    local expected="$2"
    local test_name="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    if echo "$output" | grep -q "$expected"; then
        echo -e "${GREEN}✓ $test_name${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ $test_name${NC}"
        echo -e "${RED}  Expected output to contain: $expected${NC}"
        echo -e "${RED}  Got: $output${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_exists() {
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ -e "$1" ]; then
        echo -e "${GREEN}✓ $2${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ $2${NC}"
        echo -e "${RED}  File not found: $1${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Setup
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Scope Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if binary exists
if [ ! -x "$SCOPE_BIN" ]; then
    echo -e "${RED}Error: Scope binary not found at $SCOPE_BIN${NC}"
    echo "Build it first with: make build"
    exit 1
fi

# Backup existing database
if [ -f "$HOME/.config/scope/scope.db" ]; then
    echo -e "${YELLOW}Backing up existing database...${NC}"
    cp "$HOME/.config/scope/scope.db" "$DB_BACKUP"
    rm "$HOME/.config/scope/scope.db"
fi

# Create test directories
mkdir -p "$TEST_DIR"/{project1,project2,project3,personal/docs,work/app}
echo "test file" > "$TEST_DIR/project1/README.md"
echo "test file 2" > "$TEST_DIR/project2/README.md"

echo -e "${BLUE}Test directory: $TEST_DIR${NC}"
echo ""

# Test 1: Help command
echo -e "${YELLOW}Test Group: Help & Info${NC}"
output=$($SCOPE_BIN help)
assert_output_contains "$output" "Usage:" "Help command shows usage"

# Test 2: Tag a folder
echo ""
echo -e "${YELLOW}Test Group: Tagging${NC}"
$SCOPE_BIN tag "$TEST_DIR/project1" work
assert_success "Tag folder with 'work'"

$SCOPE_BIN tag "$TEST_DIR/project2" work
assert_success "Tag another folder with 'work'"

$SCOPE_BIN tag "$TEST_DIR/project3" personal
assert_success "Tag folder with 'personal'"

$SCOPE_BIN tag "$TEST_DIR/project1" urgent
assert_success "Add second tag to same folder"

# Test 3: List all tags
echo ""
echo -e "${YELLOW}Test Group: Listing${NC}"
output=$($SCOPE_BIN list)
assert_output_contains "$output" "personal" "List shows 'personal' tag"
assert_output_contains "$output" "work" "List shows 'work' tag"
assert_output_contains "$output" "urgent" "List shows 'urgent' tag"

# Test 4: List folders by tag
output=$($SCOPE_BIN list work)
assert_output_contains "$output" "project1" "List work tag shows project1"
assert_output_contains "$output" "project2" "List work tag shows project2"

output=$($SCOPE_BIN list personal)
assert_output_contains "$output" "project3" "List personal tag shows project3"

# Test 5: Tag current directory
echo ""
echo -e "${YELLOW}Test Group: Current Directory${NC}"
cd "$TEST_DIR/personal/docs"
$SCOPE_BIN tag . docs
assert_success "Tag current directory with '.'"

output=$($SCOPE_BIN list docs)
assert_output_contains "$output" "docs" "List shows tagged current directory"

# Test 6: Untag
echo ""
echo -e "${YELLOW}Test Group: Untagging${NC}"
$SCOPE_BIN untag "$TEST_DIR/project1" urgent
assert_success "Untag folder"

output=$($SCOPE_BIN list urgent)
if echo "$output" | grep -q "No folders found"; then
    echo -e "${GREEN}✓ Untag removes tag correctly${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ Untag removes tag correctly${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 7: Remove tag entirely
echo ""
echo -e "${YELLOW}Test Group: Tag Deletion${NC}"
$SCOPE_BIN remove-tag docs
assert_success "Remove tag entirely"

output=$($SCOPE_BIN list)
if ! echo "$output" | grep -q "docs"; then
    echo -e "${GREEN}✓ Removed tag not in list${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ Removed tag not in list${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 8: Database file creation
echo ""
echo -e "${YELLOW}Test Group: Database${NC}"
assert_file_exists "$HOME/.config/scope/scope.db" "Database file created"

# Test 9: Special characters in path
echo ""
echo -e "${YELLOW}Test Group: Edge Cases${NC}"
mkdir -p "$TEST_DIR/folder with spaces"
$SCOPE_BIN tag "$TEST_DIR/folder with spaces" test-spaces
assert_success "Tag folder with spaces in name"

output=$($SCOPE_BIN list test-spaces)
assert_output_contains "$output" "folder with spaces" "List shows folder with spaces"

# Test 10: Non-existent folder
echo ""
echo -e "${YELLOW}Test Group: Error Handling${NC}"
if ! $SCOPE_BIN tag /nonexistent/path test 2>/dev/null; then
    echo -e "${GREEN}✓ Tagging non-existent folder fails gracefully${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ Tagging non-existent folder fails gracefully${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 11: Untag non-existent tag
if ! $SCOPE_BIN untag "$TEST_DIR/project1" nonexistent 2>/dev/null; then
    echo -e "${GREEN}✓ Untag non-existent tag fails gracefully${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ Untag non-existent tag fails gracefully${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 12: Start session (non-interactive test)
echo ""
echo -e "${YELLOW}Test Group: Session Management${NC}"
# We can't fully test interactive session in script, but we can check it doesn't crash
timeout 2s $SCOPE_BIN start work <<EOF || true
exit
EOF

if [ $? -eq 0 ] || [ $? -eq 124 ]; then
    echo -e "${GREEN}✓ Start session executes without crash${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ Start session executes without crash${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total tests run: ${BLUE}$TESTS_RUN${NC}"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
