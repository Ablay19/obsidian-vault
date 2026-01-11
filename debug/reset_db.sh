#!/bin/bash

# Reset local database for testing
# This removes the local SQLite database so it can be recreated with proper schema

set -e

echo "🔄 Resetting local database for testing..."

# Remove existing local database
if [ -f ".data/local/test.db" ]; then
    echo "🗑️  Removing existing database..."
    rm -f .data/local/test.db
fi

# Ensure directory exists
mkdir -p .data/local

echo "✅ Database reset complete"
echo "📝 Next time you start the bot, it will create a fresh local SQLite database"
echo "🔧 Run: GIN_MODE=release ./bot"