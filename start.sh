#!/bin/sh

echo "Architecture: $(uname -m)"
/usr/local/bin/ollama serve &
./telegram-bot