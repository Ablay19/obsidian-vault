#!/bin/bash
# Create Cloudflare KV namespace for sessions

set -e

echo "Creating Cloudflare KV namespace for sessions..."

# Create KV namespace
KV_RESPONSE=$(npx wrangler kv:namespace create "SESSIONS" 2>&1)

echo "$KV_RESPONSE"

# Extract the ID from the response
KV_ID=$(echo "$KV_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$KV_ID" ]; then
    echo "Could not extract KV namespace ID. Please manually update wrangler.toml"
    echo "KV Response: $KV_RESPONSE"
else
    echo "KV Namespace ID: $KV_ID"
    
    # Update wrangler.toml with the new ID
    sed -i "s/YOUR_KV_NAMESPACE_ID/$KV_ID/g" wrangler.toml
    
    # Generate preview ID
    echo "Creating preview namespace..."
    npx wrangler kv:namespace create "SESSIONS" --preview 2>&1 | head -20
    
    echo "Done! wrangler.toml has been updated with KV namespace configuration."
fi
