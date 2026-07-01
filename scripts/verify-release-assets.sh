#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
out_dir="${OUT_DIR:-dist/release}"

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<<"$target"
  archive="tree2scaffold_${version}_${goos}_${goarch}.tar.gz"

  test -f "$out_dir/$archive"
  test -f "$out_dir/$archive.sha256"

  (cd "$out_dir" && shasum -a 256 -c "$archive.sha256")

  listing="$(tar -tzf "$out_dir/$archive")"
  grep -Eq '(^|/)tree2scaffold$' <<<"$listing"
  grep -Eq '(^|/)t2s$' <<<"$listing"
  grep -Eq '(^|/)LICENSE$' <<<"$listing"
  grep -Eq '(^|/)README.md$' <<<"$listing"
done

echo "Release assets verified in $out_dir"
