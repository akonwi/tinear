---
description: Validate, tag, and publish a native Tinear release
---
Create a Tinear release through the native GitHub Actions release pipeline.

1. Determine the next semantic version from commits since the latest tag. Propose the version and ask for confirmation.
2. Draft concise user-facing release notes grouped into features, improvements, and fixes.
3. Verify that the Tinear and pinned Cooper revisions are both remotely available.
4. Run `scripts/release.sh --dry-run vX.Y.Z`. Resolve every validation failure before continuing.
5. Ask for confirmation before creating or pushing a tag.
6. Run `scripts/release.sh vX.Y.Z`, then `git push origin vX.Y.Z`.
7. Monitor `.github/workflows/release.yml`. It must build on native Linux amd64/arm64 and macOS amd64/arm64 runners; do not substitute CGO cross-compilation.
8. Verify the resulting GitHub release contains four `tinear-<os>-<arch>.tar.gz` archives and `SHA256SUMS.txt`.
9. Update the Homebrew tap from the published archive URLs and checksums only after the release workflow succeeds.
