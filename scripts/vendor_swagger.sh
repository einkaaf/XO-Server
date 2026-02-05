#!/usr/bin/env sh
set -e

DEST="$(dirname "$0")/../docs/swagger-ui"
mkdir -p "$DEST"

BASE="https://unpkg.com/swagger-ui-dist@5"
FILES="swagger-ui.css swagger-ui-bundle.js swagger-ui-standalone-preset.js favicon-16x16.png favicon-32x32.png"

for f in $FILES; do
  url="$BASE/$f"
  out="$DEST/$f"
  echo "Downloading $url -> $out"
  curl -fsSL "$url" -o "$out"
done

echo "Swagger UI assets downloaded to $DEST"
