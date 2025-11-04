#!/bin/bash

set -e

echo "=== Yoink Integration Tests ==="
echo

# Build the binary
echo "Building yoink..."
go build -o bin/yoink ./cmd/yoink
echo "✓ Build successful"
echo

# Test 1: Help command
echo "Test 1: Help command"
./bin/yoink --help > /dev/null
echo "✓ Help command works"
echo

# Test 2: Revoke command help
echo "Test 2: Revoke command help"
./bin/yoink revoke --help > /dev/null
echo "✓ Revoke help works"
echo

# Test 3: Dry-run with argument
echo "Test 3: Dry-run with argument"
OUTPUT=$(./bin/yoink revoke --dry-run ghp_1234567890abcdefghijklmnopqrstuvwxyz123456 2>&1)
if echo "$OUTPUT" | grep -q "DRY-RUN"; then
    echo "✓ Dry-run with argument works"
else
    echo "✗ Dry-run with argument failed"
    exit 1
fi
echo

# Test 4: Dry-run with stdin
echo "Test 4: Dry-run with stdin"
OUTPUT=$(echo "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456" | ./bin/yoink revoke --dry-run 2>&1)
if echo "$OUTPUT" | grep -q "DRY-RUN"; then
    echo "✓ Dry-run with stdin works"
else
    echo "✗ Dry-run with stdin failed"
    exit 1
fi
echo

# Test 5: Multiple credentials
echo "Test 5: Multiple credentials"
OUTPUT=$(echo -e "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456\ngho_1234567890abcdefghijklmnopqrstuvwxyz123456" | ./bin/yoink revoke --dry-run 2>&1)
if echo "$OUTPUT" | grep -q "Total processed: 2"; then
    echo "✓ Multiple credentials work"
else
    echo "✗ Multiple credentials failed"
    exit 1
fi
echo

# Test 6: Invalid credential format
echo "Test 6: Invalid credential format"
OUTPUT=$(echo "invalid_credential" | ./bin/yoink revoke --dry-run 2>&1 || true)
if echo "$OUTPUT" | grep -q "Failed"; then
    echo "✓ Invalid credential properly rejected"
else
    echo "✗ Invalid credential handling failed"
    exit 1
fi
echo

# Test 7: No credentials provided
echo "Test 7: No credentials provided"
OUTPUT=$(./bin/yoink revoke 2>&1 || true)
if echo "$OUTPUT" | grep -q "no credentials provided"; then
    echo "✓ Empty input properly handled"
else
    echo "✗ Empty input handling failed"
    exit 1
fi
echo

# Test 8: Version flag
echo "Test 8: Version flag"
OUTPUT=$(./bin/yoink --version 2>&1)
if echo "$OUTPUT" | grep -q "1.0.0"; then
    echo "✓ Version flag works"
else
    echo "✗ Version flag failed"
    exit 1
fi
echo

# Test 9: File watcher (basic test - just verify it starts)
echo "Test 9: File watcher help"
./bin/yoink watch --help > /dev/null
echo "✓ Watch command help works"
echo

# Test 10: Server help
echo "Test 10: Server help"
./bin/yoink serve --help > /dev/null
echo "✓ Serve command help works"
echo

echo "=== All Integration Tests Passed! ==="
