#!/usr/bin/env bash
#
# Download the 6 prebuilt LiteParse `lit` bundles for a pinned release tag into
# cli/cmd/assets/, so a `-tags retab_embed_lit` build can go:embed them into the
# fat `retab` binary.
#
# Why this exists
# ---------------
# The release build of `retab` ships as a single self-contained binary with
# `lit` + libpdfium + eng.traineddata baked in (see cli/cmd/litembed_*.go). The
# bundles themselves are ~25MB x 6 = ~150MB of binary assets — far too large to
# commit to the repo, so cli/cmd/assets/*.tar.gz is gitignored. This script
# stages them at release time from the `lit-v*` GitHub release that
# build-liteparse.yml publishes.
#
# GoReleaser cross-compiles all 6 targets in one run, and each target's
# litembed_<os>_<arch>.go has a `//go:embed assets/lit-<os>-<arch>.tar.gz`
# directive that must resolve at compile time. So ALL 6 assets must be present
# before the build, even though any single binary only embeds its own.
#
# Inputs (env vars)
# -----------------
#   LIT_BUNDLE_TAG       release tag to fetch (default: lit-v2.0.3-r2 — keep in
#                        sync with litBundleTag in cli/cmd/liteparse_managed.go)
#   LIT_BUNDLE_BASE_URL  download root (default: the retab-dev/retab releases)
#   ASSETS_DIR           output dir (default: <repo>/cli/cmd/assets)
#   VERIFY_CHECKSUMS     1 (default) | 0 — verify each asset against the
#                        release checksums.txt when it is fetchable
#
set -euo pipefail

die() { echo "::error::$*" >&2; exit 1; }
log() { echo ">> $*" >&2; }

LIT_BUNDLE_TAG="${LIT_BUNDLE_TAG:-lit-v2.0.3-r2}"
LIT_BUNDLE_BASE_URL="${LIT_BUNDLE_BASE_URL:-https://github.com/retab-dev/retab/releases/download}"
VERIFY_CHECKSUMS="${VERIFY_CHECKSUMS:-1}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ASSETS_DIR="${ASSETS_DIR:-${REPO_ROOT}/cli/cmd/assets}"

PLATFORMS=(
  linux-amd64
  linux-arm64
  darwin-amd64
  darwin-arm64
  windows-amd64
  windows-arm64
)

BASE="${LIT_BUNDLE_BASE_URL%/}/${LIT_BUNDLE_TAG}"
mkdir -p "${ASSETS_DIR}"

# Best-effort fetch of the release-wide checksums.txt for verification.
CHECKSUMS=""
if [ "${VERIFY_CHECKSUMS}" = "1" ]; then
  CHECKSUMS="$(mktemp)"
  if curl -fsSL -o "${CHECKSUMS}" "${BASE}/checksums.txt"; then
    log "fetched checksums.txt for ${LIT_BUNDLE_TAG}"
  else
    log "WARNING: no checksums.txt at ${BASE}; skipping verification"
    CHECKSUMS=""
  fi
fi

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

for plat in "${PLATFORMS[@]}"; do
  asset="lit-${plat}.tar.gz"
  url="${BASE}/${asset}"
  out="${ASSETS_DIR}/${asset}"
  log "downloading ${asset}"
  curl -fsSL -o "${out}" "${url}" || die "failed to download ${url}"
  [ -s "${out}" ] || die "downloaded ${asset} is empty"

  if [ -n "${CHECKSUMS}" ]; then
    want="$(grep -E "  ${asset}\$|\\*?${asset}\$" "${CHECKSUMS}" | awk '{print $1}' | head -n1 || true)"
    if [ -n "${want}" ]; then
      got="$(sha256_of "${out}")"
      [ "${got}" = "${want}" ] || die "checksum mismatch for ${asset}: got ${got}, want ${want}"
      log "  verified ${asset} (${got})"
    else
      log "  WARNING: ${asset} not listed in checksums.txt; not verified"
    fi
  fi
done

log "staged $(printf '%s ' "${PLATFORMS[@]}")bundles into ${ASSETS_DIR}"
ls -lh "${ASSETS_DIR}"/*.tar.gz >&2
