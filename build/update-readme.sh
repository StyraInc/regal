#!/usr/bin/env bash

README_SECTIONS_DIR="./docs/readme-sections"
README_PATH="./README.md"
MANIFEST="$README_SECTIONS_DIR/github-manifest"

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 {check|write}"
  exit 1
fi

MODE="$1"

# Create a temporary file to hold the new content
tmpfile=$(mktemp)

# Build new content into tmpfile
while IFS= read -r file; do
  section_path="$README_SECTIONS_DIR/$file"

  if [[ -f "$section_path" ]]; then
    cat "$section_path" >> "$tmpfile"
    echo -e "\n" >> "$tmpfile"
  else
    echo "Section file not found: $section_path" >&2
    exit 1
  fi
done < "$MANIFEST"

if [[ "$MODE" == "check" ]]; then
  if ! cmp -s "$tmpfile" "$README_PATH"; then
    echo "README.md is out of date. Please run '$0 write' to update it."
    rm "$tmpfile"
    exit 1
  else
    echo "README.md is up to date."
  fi
elif [[ "$MODE" == "write" ]]; then
  mv "$tmpfile" "$README_PATH"
  echo "README.md has been updated."
else
  echo "Unknown mode: $MODE"
  echo "Usage: $0 {check|write}"
  rm "$tmpfile"
  exit 1
fi
