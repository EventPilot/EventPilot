#!/bin/bash

echo "🦋 Bluesky Connection Test"
echo "=========================================="
echo ""

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found"
    echo ""
    echo "Create one first:"
    echo "  cp .env.example .env"
    echo "  nano .env"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Check credentials
if [ -z "$BSKY_HANDLE" ]; then
    echo "❌ BSKY_HANDLE not set in .env"
    exit 1
fi

if [ -z "$BSKY_PASSWORD" ]; then
    echo "❌ BSKY_PASSWORD not set in .env"
    exit 1
fi

echo "✅ Credentials loaded"
echo ""

# Run the test
go run test-bluesky.go