#!/bin/bash
set -e

# Generate CalVer version: YYYY.MM.PATCH
# Format: 2025.01.0, 2025.01.1, etc.

YEAR=$(date +%Y)
MONTH=$(date +%-m)  # No leading zero

# Get the latest tag matching current year/month pattern
CURRENT_PATTERN="${YEAR}.${MONTH}."
LATEST_TAG=$(git tag -l "${CURRENT_PATTERN}*" | sort -V | tail -n 1)

if [ -z "$LATEST_TAG" ]; then
  # No tags for this month yet, start at 0
  PATCH=0
else
  # Extract patch version and increment
  PATCH=$(echo "$LATEST_TAG" | cut -d'.' -f3)
  PATCH=$((PATCH + 1))
fi

VERSION="${YEAR}.${MONTH}.${PATCH}"
echo "$VERSION"
