#!/bin/bash
# Test script to run UI with debug output visible

echo "=========================================="
echo "Running UI in mock mode with debug output"
echo "Debug messages will appear below the UI"
echo "=========================================="
echo ""
echo "Instructions:"
echo "1. Press Enter on 'Build Environment'"
echo "2. Type a name like 'test-env'"
echo "3. Press Enter"
echo "4. Look for green message at top of Actions panel"
echo ""
echo "OR:"
echo ""
echo "1. Press Tab to go to Observe tab"
echo "2. Use j/k to select an environment"
echo "3. Press 'x' to delete"
echo "4. Look for green message at top"
echo ""
echo "Press 'q' to quit when done"
echo "=========================================="
echo ""

cd /Users/roryhedderman/GolandProjects/imperm
./bin/imperm-ui --mock 2>&1
