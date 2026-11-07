#!/usr/bin/env sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$repo_dir"

image="dpokidov/imagemagick:7.1.1-47"
source_dir="${1:-reference/exercises}"
output_dir="${2:-internal/app/static/images/exercises}"
processing_file="assets/exercise-images/processing.tsv"

go run ./tools/exercise-images -source "$source_dir"

if [ ! -d "$source_dir" ]; then
  echo "source directory not found: $source_dir" >&2
  exit 1
fi

mkdir -p "$output_dir"
rm -f "$output_dir"/*.webp
source_dir=$(cd "$source_dir" && pwd)
output_dir=$(cd "$output_dir" && pwd)
processing_file=$(cd "$(dirname "$processing_file")" && pwd)/$(basename "$processing_file")
source_mount=$source_dir
output_mount=$output_dir
processing_mount=$processing_file
if command -v cygpath >/dev/null 2>&1; then
  source_mount=$(cygpath -w "$source_dir")
  output_mount=$(cygpath -w "$output_dir")
  processing_mount=$(cygpath -w "$processing_file")
fi

MSYS_NO_PATHCONV=1 docker run --rm \
  --entrypoint sh \
  -v "$source_mount:/input:ro" \
  -v "$output_mount:/output" \
  -v "$processing_mount:/processing.tsv:ro" \
  "$image" -ceu '
    found=0
    for file in /input/*.png; do
      [ -f "$file" ] || continue
      found=1
      dimensions=$(magick identify -format "%wx%h" "$file")
      if [ "$dimensions" != "1536x1024" ]; then
        echo "unexpected dimensions for $file: $dimensions" >&2
        exit 1
      fi
      base=$(basename "$file" .png)
      trim=$(awk -v name="$base" "\$1 == name { print \$2 }" /processing.tsv)
      trim=${trim:-0}
      if [ "$trim" -gt 0 ]; then
        half=$(((1536 - trim) / 2))
        right=$(((1536 + trim) / 2))
        magick \
          \( "$file" -crop "${half}x1024+0+0" +repage \) \
          \( "$file" -crop "${half}x1024+${right}+0" +repage \) \
          +append -resize 1536x1024\! -strip -define webp:method=6 -quality 88 "/output/${base}.webp"
      else
        magick "$file" -strip -define webp:method=6 -quality 88 "/output/${base}.webp"
      fi
    done
    if [ "$found" -eq 0 ]; then
      echo "no PNG source images found" >&2
      exit 1
    fi
  '
