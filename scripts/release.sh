#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
  shift
fi

if [[ $# -ne 1 || ! "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Usage: $0 [--dry-run] v<major>.<minor>.<patch>" >&2
  exit 1
fi

VERSION="$1"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -n "$(git status --porcelain --untracked-files=all)" ]]; then
  echo "Error: working tree has tracked or untracked changes" >&2
  exit 1
fi
if git rev-parse "refs/tags/$VERSION" >/dev/null 2>&1; then
  echo "Error: tag $VERSION already exists" >&2
  exit 1
fi

printf '==> Validating %s on %s/%s\n' "$VERSION" "$(go env GOOS)" "$(go env GOARCH)"
ard check main.ard
ard build main.ard --out "${RUNNER_TEMP:-/tmp}/tinear-release-check"
ard test
go test ./...
python3 test/pty/test_tinear.py

if [[ "$DRY_RUN" == true ]]; then
  echo "==> Dry run complete; no tag created"
else
  git tag -a "$VERSION" -m "$VERSION"
  echo "==> Created $VERSION"
  echo "Push with: git push origin $VERSION"
  echo "The Release workflow will build native archives and publish the GitHub release."
fi
