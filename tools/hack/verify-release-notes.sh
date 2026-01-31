#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' 

errors=0

echo "Verifying release note filenames in 'release-notes'..."

if [ -d "release-notes" ]; then
    for file in release-notes/*.yaml; do
        [ -e "$file" ] || continue
        filename=$(basename "$file")

        if [ "$filename" = "current.yaml" ]; then
            echo " Skipping current.yaml (special file)"
            continue
        fi
        
        if ! [[ "$filename" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?\.yaml$ ]]; then
            echo -e "${RED}Error: Invalid filename format: $filename${NC}"
            echo "   Expected: vX.Y.Z.yaml or vX.Y.Z-rc.N.yaml"
            echo "   Example:  v1.7.0-rc.1.yaml"
            
            if [[ "$filename" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc[0-9]+\.yaml$ ]]; then
                suggested="${filename//-rc/-rc.}"
                echo -e "${YELLOW} Suggestion: Rename to $suggested${NC}"
            fi
            
            ((errors++))
        else
            echo " $filename"
        fi
    done
fi

SITE_NOTES_DIR="site/content/en/news/releases/notes"
if [ -d "$SITE_NOTES_DIR" ]; then
    echo ""
    echo "🔍 Verifying release note filenames in '$SITE_NOTES_DIR'..."
    
    for file in "$SITE_NOTES_DIR"/*.md; do
        [ -e "$file" ] || continue
        filename=$(basename "$file")
        
        if [ "$filename" = "_index.md" ]; then
            echo "Skipping _index.md (index file)"
            continue
        fi
        
        if ! [[ "$filename" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?\.md$ ]]; then
            echo -e "${RED}Error: Invalid filename format: $filename${NC}"
            echo "   Expected: vX.Y.Z.md or vX.Y.Z-rc.N.md"
            echo "   Example:  v1.7.0-rc.1.md"
            
            if [[ "$filename" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc[0-9]+\.md$ ]]; then
                suggested="${filename//-rc/-rc.}"
                echo -e "${YELLOW}  Suggestion: Rename to $suggested${NC}"
            fi
            
            ((errors++))
        else
            echo " $filename"
        fi
    done
fi

echo ""
if [ "$errors" -eq 0 ]; then
    echo -e "${GREEN} All release note filenames are valid!${NC}"
    exit 0
else
    echo -e "${RED} Verification failed with $errors error(s).${NC}"
    echo ""
    echo "To fix these issues:"
    echo "   1. Rename files with old RC format (rcN → rc.N)"
    echo "   2. Ensure all version files follow: vX.Y.Z.yaml or vX.Y.Z-rc.N.yaml"
    exit 1
fi