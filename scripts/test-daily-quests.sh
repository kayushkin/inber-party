#!/bin/bash

# Test script for Daily Quest system
# This script starts the server, generates daily quests, and shows the results

echo "🎮 Testing Daily Quest System..."
echo "================================="

# Check if build exists
if [ ! -f "bin/inber-party" ]; then
    echo "❌ Build not found. Run 'make build' first."
    exit 1
fi

# Set test database URL (using SQLite for testing)
export DATABASE_URL="postgres://localhost/test_inber_party?sslmode=disable"
export PORT="8081"

echo "🚀 Starting test server on port 8081..."
./bin/inber-party &
SERVER_PID=$!

# Give server time to start
sleep 3

# Test daily quest generation
echo "📋 Testing daily quest generation..."
response=$(curl -s -X POST http://localhost:8081/api/daily-quests/generate)

if echo "$response" | grep -q "success"; then
    echo "✅ Daily quest generation successful"
else
    echo "❌ Daily quest generation failed: $response"
fi

# Get daily quest stats
echo "📊 Daily quest stats:"
curl -s http://localhost:8081/api/daily-quests/stats | python3 -m json.tool

# Get active daily quests
echo "📜 Active daily quests:"
curl -s http://localhost:8081/api/daily-quests | python3 -m json.tool

# Cleanup
echo "🧹 Stopping test server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "✅ Daily Quest test completed!"