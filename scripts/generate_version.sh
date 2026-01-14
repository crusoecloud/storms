#!/usr/bin/env bash
set -e

# 1. Read Arguments (Defaults to 0 if missing)
MAJOR="${1:-0}"
MINOR="${2:-0}"
TAG_PREFIX="${3:-}" # e.g., "storms-" or empty

# Define the base version pattern. 
# Your logic implies: [PREFIX]v[MAJOR].[MINOR].[PATCH]
# Example: v1.2.x or storms-v1.2.x
BASE_PATTERN="${TAG_PREFIX}v${MAJOR}.${MINOR}"

echo "🔍 Searching for latest tag matching: ${BASE_PATTERN}.*"

# 2. Fetch Tags
# We prune to ensure we don't calculate based on a tag that was deleted remotely.
git fetch -q --tags --prune --prune-tags || echo "⚠️ Warning: git fetch failed, relying on local tags."

# 3. Find the Latest Tag
# --sort=-v:refname sorts SemVer correctly (e.g., v1.2.10 comes after v1.2.9)
# head -n 1 picks the highest one.
LATEST_TAG=$(git tag -l "${BASE_PATTERN}.*" --sort=-v:refname | head -n 1)

if [[ -z "$LATEST_TAG" ]]; then
  echo "ℹ️ No existing tag found. Starting at patch 0."
  NEW_PATCH="0"
else
  # Extract the patch version by stripping the BASE_PATTERN and the trailing dot
  # Example: "v1.2.15" remove "v1.2." leaves "15"
  LAST_PATCH=$(echo "${LATEST_TAG}" | sed "s/^${BASE_PATTERN}.//")
  
  # specific check to ensure we actually got a number
  if [[ ! "$LAST_PATCH" =~ ^[0-9]+$ ]]; then
    echo "❌ Error: Could not parse patch number from tag '${LATEST_TAG}'."
    exit 1
  fi
  
  NEW_PATCH=$((LAST_PATCH + 1))
  echo "✅ Found latest tag: ${LATEST_TAG}. Incrementing patch to ${NEW_PATCH}."
fi

# 4. Construct New Version
NEW_VERSION="${BASE_PATTERN}.${NEW_PATCH}"

echo "🚀 New Version: ${NEW_VERSION}"

# 5. Export to variables.env
# We use '>' to OVERWRITE the file, ensuring no duplicate variables exist.
echo "RELEASE_VERSION=${NEW_VERSION}" > variables.env