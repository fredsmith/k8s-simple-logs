#!/bin/bash
set -e

# Generate Go-compatible CalVer version: v0.YYMM.PATCH
# Format: v0.2510.0, v0.2510.1, v0.2511.0, etc.
# Uses v0 major version for Go modules compatibility

YEAR=$(date +%y)    # 2-digit year (25 for 2025)
MONTH=$(date +%m)   # 2-digit month with leading zero (01-12)

# Get the latest tag matching current year/month pattern
CURRENT_PATTERN="v0.${YEAR}${MONTH}."
LATEST_TAG=$(git tag -l "${CURRENT_PATTERN}*" | sort -V | tail -n 1)

if [ -z "$LATEST_TAG" ]; then
  # No tags for this month yet, start at 0
  PATCH=0
else
  # Extract patch version from tag
  PATCH=$(echo "$LATEST_TAG" | cut -d'.' -f3)
  PATCH=$((PATCH + 1))
fi

VERSION="0.${YEAR}${MONTH}.${PATCH}"
echo "$VERSION"
