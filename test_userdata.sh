#!/bin/bash
export AGENT_BROWSER_USER_DATA_DIR="$PWD/test-data"
echo "UserDataDir set to: $AGENT_BROWSER_USER_DATA_DIR"
./agent-browser-go --backend chromedp --head open https://example.com
sleep 2
echo ""
echo "Checking for files in test-data:"
ls -la test-data/ 2>&1 || echo "Directory not created"
