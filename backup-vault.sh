#!/bin/bash
cd vault

if [[ -n $(git status -s 2>/dev/null) ]]; then
    git add .
    git commit -m "Auto-backup: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "✅ Vault backed up"
    
    if git remote | grep -q origin; then
        git push origin main 2>/dev/null && echo "✅ Pushed to GitHub"
    fi
else
    echo "No changes to backup"
fi
