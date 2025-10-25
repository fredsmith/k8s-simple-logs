#!/bin/bash
set -e

# Generate CalVer version: YYYY.M.PATCH
# Format: 2025.1.0, 2025.1.1, 2025.10.0, etc.

YEAR=$(date +%Y)
MONTH=$(date +%-m)  # No leading zero

# Get the latest tag matching current year/month pattern (with 'v' prefix)
CURRENT_PATTERN="v${YEAR}.${MONTH}."
LATEST_TAG=$(git tag -l "${CURRENT_PATTERN}*" | sort -V | tail -n 1)

if [ -z "$LATEST_TAG" ]; then
  # No tags for this month yet, start at 0
  PATCH=0
else
  # Extract patch version from tag (strip 'v' prefix first)
  TAG_WITHOUT_V="${LATEST_TAG#v}"
  PATCH=$(echo "$TAG_WITHOUT_V" | cut -d'.' -f3)
  PATCH=$((PATCH + 1))
fi

VERSION="${YEAR}.${MONTH}.${PATCH}"
echo "$VERSION"
