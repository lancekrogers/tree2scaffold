#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
commit="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
date="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
out_dir="${OUT_DIR:-dist/release}"
main_path="./cmd/tree2scaffold"

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
)

rm -rf "$out_dir"
mkdir -p "$out_dir"
stage_root="$(mktemp -d)"
trap 'rm -rf "$stage_root"' EXIT

for target in "${targets[@]}"; do
  read -r goos goarch <<<"$target"
  archive="tree2scaffold_${version}_${goos}_${goarch}.tar.gz"
  stage="$stage_root/${goos}_${goarch}"
  mkdir -p "$stage"

  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${version} -X main.commit=${commit} -X main.date=${date}" \
    -o "$stage/tree2scaffold" \
    "$main_path"

  ln -sf tree2scaffold "$stage/t2s"
  cp LICENSE "$stage/LICENSE"
  {
    echo "# tree2scaffold ${version}"
    echo
    echo "Install one of these files on PATH:"
    echo
    echo "- tree2scaffold"
    echo "- t2s"
  } >"$stage/README.md"

  tar -C "$stage" -czf "$out_dir/$archive" .
  (cd "$out_dir" && shasum -a 256 "$archive" >"$archive.sha256")
done

echo "Release assets written to $out_dir"
