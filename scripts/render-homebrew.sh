#!/usr/bin/env bash
set -euo pipefail

tag="${TAG:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
version="${VERSION:-${tag#v}}"
release_dir="${RELEASE_DIR:-dist/release}"
out_dir="${OUT_DIR:-dist/homebrew}"

checksum_for() {
  local artifact="$1"
  local checksum_file="$release_dir/$artifact.sha256"
  test -f "$checksum_file"
  awk '{print $1}' "$checksum_file"
}

render() {
  local template="$1"
  local destination="$2"
  local sha_darwin_amd64="$3"
  local sha_darwin_arm64="$4"

  mkdir -p "$(dirname "$destination")"
  sed \
    -e "s/{{VERSION}}/$version/g" \
    -e "s/{{TAG}}/$tag/g" \
    -e "s/{{SHA256_DARWIN_AMD64}}/$sha_darwin_amd64/g" \
    -e "s/{{SHA256_DARWIN_ARM64}}/$sha_darwin_arm64/g" \
    "$template" >"$destination"
}

sha_darwin_amd64="$(checksum_for "tree2scaffold_${tag}_darwin_amd64.tar.gz")"
sha_darwin_arm64="$(checksum_for "tree2scaffold_${tag}_darwin_arm64.tar.gz")"

rm -rf "$out_dir"
render "packaging/homebrew/Casks/tree2scaffold.rb.template" \
  "$out_dir/Casks/tree2scaffold.rb" \
  "$sha_darwin_amd64" \
  "$sha_darwin_arm64"
render "packaging/homebrew/Formula/tree2scaffold.rb.template" \
  "$out_dir/Formula/tree2scaffold.rb" \
  "$sha_darwin_amd64" \
  "$sha_darwin_arm64"

ruby -c "$out_dir/Casks/tree2scaffold.rb"
ruby -c "$out_dir/Formula/tree2scaffold.rb"

if command -v brew >/dev/null 2>&1; then
  brew style "$out_dir/Casks/tree2scaffold.rb" "$out_dir/Formula/tree2scaffold.rb"

  audit_output="$(brew audit --cask --strict "$out_dir/Casks/tree2scaffold.rb" 2>&1)" || {
    if grep -q "Calling .*brew audit \\[path" <<<"$audit_output"; then
      echo "brew audit skipped for local cask path: $audit_output"
    else
      echo "$audit_output"
      exit 1
    fi
  }

  audit_output="$(brew audit --formula --strict "$out_dir/Formula/tree2scaffold.rb" 2>&1)" || {
    if grep -q "Calling .*brew audit \\[path" <<<"$audit_output"; then
      echo "brew audit skipped for local formula path: $audit_output"
    else
      echo "$audit_output"
      exit 1
    fi
  }
else
  echo "brew not found; skipped Homebrew audit"
fi

echo "Rendered Homebrew files in $out_dir"
