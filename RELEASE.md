# Release Checklist

This checklist is for maintainers publishing `tree2scaffold` release artifacts,
npm packages, and Homebrew tap updates.

## Credential Boundary

The repository must not contain npm tokens, Homebrew credentials, signing keys,
or private release credentials. The following steps require maintainer-owned
accounts or configured local credentials:

- `npm login`
- `npm publish`
- pushing to the Homebrew tap repository
- creating or editing GitHub releases outside the tag-triggered workflow

## Pre-Release Checks

Run these from the repository root:

```bash
go test ./...
go vet ./...
just build release-native
just build release-native-verify
npm run package:check
just build homebrew-render
```

The Homebrew render command writes generated files under `dist/homebrew/`.
Generated `dist/` contents are build outputs and should not be committed to this
repository.

## Version And Tag

Choose a semver release tag with a leading `v`, for example:

```bash
export TAG=v1.2.3
export VERSION=1.2.3
```

The native release workflow attaches these assets to GitHub releases on `v*`
tags:

```text
tree2scaffold_${TAG}_darwin_amd64.tar.gz
tree2scaffold_${TAG}_darwin_amd64.tar.gz.sha256
tree2scaffold_${TAG}_darwin_arm64.tar.gz
tree2scaffold_${TAG}_darwin_arm64.tar.gz.sha256
tree2scaffold_${TAG}_linux_amd64.tar.gz
tree2scaffold_${TAG}_linux_amd64.tar.gz.sha256
```

Create and push the tag after pre-release checks pass:

```bash
git tag "$TAG"
git push origin "$TAG"
```

## npm Publication

Confirm package ownership before publishing. If `tree2scaffold` is unavailable
on npm, use the scoped fallback package name decided by the maintainer.

Prepare the package from release assets:

```bash
just build release-native
npm run package:stage-native
npm run package:check
npm pack --dry-run
```

Publish only with maintainer credentials:

```bash
npm login
npm publish
```

Post-publish verification:

```bash
npm install -g tree2scaffold
tree2scaffold --version
t2s --version
npm uninstall -g tree2scaffold
```

## Homebrew Tap Handoff

Render tap-ready files after native release assets exist:

```bash
TAG=v1.2.3 VERSION=1.2.3 just build release-native
TAG=v1.2.3 VERSION=1.2.3 just build homebrew-render
```

Generated files:

```text
dist/homebrew/Casks/tree2scaffold.rb
dist/homebrew/Formula/tree2scaffold.rb
```

Copy the generated files into the maintainer-controlled tap repository. The
formula is the recommended Homebrew path for this CLI; the cask is provided for
cask-specific distribution needs.

Post-tap verification:

```bash
brew install lancekrogers/tap/tree2scaffold
tree2scaffold --version
t2s --version
brew uninstall tree2scaffold

brew install --cask lancekrogers/tap/tree2scaffold
tree2scaffold --version
t2s --version
brew uninstall --cask tree2scaffold
```

## Correction Notes

- Bad GitHub release asset: delete the bad release asset, rebuild from the
  intended tag, and re-upload with matching checksum.
- Bad npm package metadata: publish a corrected patch version. Avoid unpublish
  unless the maintainer has confirmed npm policy allows it for the case.
- Bad Homebrew tap file: push a corrective tap commit with updated checksums or
  URLs and rerun install verification.
