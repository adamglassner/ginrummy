#!/usr/bin/env bash
# Builds Gin Rummy for Windows (64-bit).
# Output: dist/GinRummy/ with GinRummy.exe and its text files, ready for the
# Ironite Launcher (the release zip) or for installer/GinRummySetup.iss.
# Needs Go 1.24+ and bash (Git Bash on Windows is fine). No C compiler needed.
set -euo pipefail
cd "$(dirname "$0")"
ROOT=$PWD
BUNDLE=GinRummy
EXE=GinRummy
OUT=$ROOT/dist
export GOFLAGS=-mod=mod
rm -rf "$OUT" && mkdir -p "$OUT/$BUNDLE"

VERSION=$(grep -m1 -oE '^## \[[0-9.]+' CHANGELOG.md | tr -d '#[ ')

# Windows resource (icon + version info) for the .exe
( cd tools/mkres && GOFLAGS=-mod=vendor go build -o "$ROOT/mkres.tmp" . )
"$ROOT/mkres.tmp" -version "$VERSION" -icon icon-1024.png -out rsrc_windows_amd64.syso \
  -name "Gin Rummy" -exe "$EXE.exe" -internal "$EXE" -identity "Ironite.$BUNDLE" -desc "Gin Rummy - A Club Car Card Game"
rm -f "$ROOT/mkres.tmp"

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-H windowsgui -s -w -X main.version=$VERSION" -o "$OUT/$BUNDLE/$EXE.exe" .
rm -f rsrc_windows_amd64.syso

crlf() { sed 's/\r$//; s/$/\r/' "$1" > "$2"; }
crlf CHANGELOG.md "$OUT/$BUNDLE/CHANGELOG.txt"
crlf assets/fonts/OFL.txt "$OUT/$BUNDLE/OFL.txt"
sed "s/@VERSION@/$VERSION/g" README.txt  | sed 's/\r$//; s/$/\r/' > "$OUT/$BUNDLE/README.txt"
sed "s/@VERSION@/$VERSION/g" CREDITS.txt | sed 's/\r$//; s/$/\r/' > "$OUT/$BUNDLE/CREDITS.txt"
printf '%s\r\n' "$VERSION" > "$OUT/$BUNDLE/VERSION"
echo "$BUNDLE $VERSION -> $OUT/$BUNDLE/$EXE.exe"
