#!/usr/bin/env sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
cd "$repo_dir"

image="dpokidov/imagemagick:7.1.1-47"
source_dir="${1:-reference/exercises}"
output_dir="${2:-assets/static/images/exercises}"

if [ ! -d "$source_dir" ]; then
  echo "source directory not found: $source_dir" >&2
  exit 1
fi

mkdir -p "$output_dir"
source_dir=$(cd "$source_dir" && pwd)
output_dir=$(cd "$output_dir" && pwd)
source_mount=$source_dir
output_mount=$output_dir
if command -v cygpath >/dev/null 2>&1; then
  source_mount=$(cygpath -w "$source_dir")
  output_mount=$(cygpath -w "$output_dir")
fi

MSYS_NO_PATHCONV=1 docker run --rm \
  --entrypoint sh \
  -v "$source_mount:/input:ro" \
  -v "$output_mount:/output" \
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
      case "$base" in
        *-female|*-male) ;;
        *)
          echo "unexpected source filename: $file" >&2
          exit 1
          ;;
      esac
      magick "$file" -strip -define webp:method=6 -quality 88 "/output/${base}.webp"
    done
    if [ "$found" -eq 0 ]; then
      echo "no PNG source images found" >&2
      exit 1
    fi
  '
