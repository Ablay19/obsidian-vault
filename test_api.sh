#!/bin/bash

echo "Testing /modelinfo endpoint..."
curl -s http://localhost:8080/info | jq .

echo ""
echo "Testing /status endpoint..."
curl -s http://localhost:8080/status | jq .
