#!/bin/bash
echo "=== Processing Statistics ==="
echo ""

if [ -f stats.json ]; then
    cat stats.json
else
    echo "No statistics yet"
fi

echo ""
echo "=== Recent Files ==="
echo ""
echo "Inbox notes:"
ls -lht vault/Inbox/ 2>/dev/null | head -10 || echo "No notes yet"

echo ""
echo "Attachments:"
ls -lht attachments/ 2>/dev/null | head -10 || echo "No attachments yet"
