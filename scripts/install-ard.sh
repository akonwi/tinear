#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-v0.40.0}"
DESTINATION="${2:-${HOME}/.local/bin}"

case "$(uname -s)" in
  Darwin) PLATFORM=darwin ;;
  Linux) PLATFORM=linux ;;
  *) echo "Unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

ARCHIVE="ard_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
BASE_URL="https://github.com/akonwi/ard/releases/download/${VERSION}"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${TEMP_DIR}/${ARCHIVE}"
curl -fsSL "${BASE_URL}/checksums.txt" -o "${TEMP_DIR}/checksums.txt"

EXPECTED="$(awk -v path="dist/${ARCHIVE}" '$2 == path { print $1 }' "${TEMP_DIR}/checksums.txt")"
if [[ -z "$EXPECTED" ]]; then
  echo "No checksum published for ${ARCHIVE}" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "${TEMP_DIR}/${ARCHIVE}" | awk '{ print $1 }')"
else
  ACTUAL="$(shasum -a 256 "${TEMP_DIR}/${ARCHIVE}" | awk '{ print $1 }')"
fi
if [[ "$ACTUAL" != "$EXPECTED" ]]; then
  echo "Checksum mismatch for ${ARCHIVE}" >&2
  exit 1
fi

tar -xzf "${TEMP_DIR}/${ARCHIVE}" -C "$TEMP_DIR"
mkdir -p "$DESTINATION"
install -m 0755 "${TEMP_DIR}/ard" "${DESTINATION}/ard"

if [[ "$("${DESTINATION}/ard" version)" != "$VERSION" ]]; then
  echo "Installed Ard version did not match ${VERSION}" >&2
  exit 1
fi
