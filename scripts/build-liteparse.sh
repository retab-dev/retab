#!/usr/bin/env bash
#
# Build the LiteParse `lit` CLI and stage it next to a matching prebuilt
# libpdfium for one (os, arch) target.
#
# Why this exists
# ---------------
# The Retab CLI shells out to `lit` (LiteParse, Apache-2.0) for the local-first
# `retab files parse|grep|inspect` commands. Upstream LiteParse only publishes
# release tarballs for 3 of the 6 platforms Retab supports, and even those
# tarballs ship ONLY the `lit` binary — not libpdfium. But `lit` dlopen()s
# `libpdfium.{dylib,so,dll}` at runtime and panics if it can't find one, so a
# bare `lit` binary cannot parse PDFs. This script builds `lit` from a pinned
# source tag and stages it together with the matching libpdfium so the pair can
# be archived and shipped as one self-contained bundle.
#
# What it does
# ------------
#   1. Clones LiteParse at $LITEPARSE_REF (shallow) if a source tree isn't
#      already supplied via $LITEPARSE_SRC.
#   2. Builds `lit` in release mode with cargo. OCR (Tesseract, statically
#      linked via the `tesseract` feature) is on by default; set LIT_OCR=0 to
#      build without it (used as the windows/arm64 fallback).
#   3. Downloads the prebuilt libpdfium matching $PDFIUM_TAG for this target
#      from run-llama/pdfium-binaries (the same fork + tag LiteParse pins in
#      crates/pdfium-sys/build.rs).
#   4. Stages `lit` + the pdfium shared library (+ LICENSE) into $STAGE_DIR.
#
# Archiving + checksums are intentionally left to the caller (the workflow),
# because portable archive creation differs per-OS (tar.gz on unix,
# Compress-Archive on Windows).
#
# Consumers must set PDFIUM_LIB_PATH to the directory containing the staged
# libpdfium before exec'ing `lit`.
#
# Inputs (env vars)
# -----------------
#   LIT_OS        linux | darwin | windows         (required)
#   LIT_ARCH      amd64 | arm64                     (required)
#   LIT_OCR       1 (default) | 0                   (OCR on/off)
#   LITEPARSE_REF git tag/branch/sha to build       (default: crates-v2.0.3)
#   LITEPARSE_SRC pre-existing source tree          (optional; skips clone)
#   PDFIUM_TAG    pdfium-binaries release tag        (default: chromium/7847)
#   STAGE_DIR     output staging dir                 (default: ./dist/lit-<os>-<arch>)
#
set -euo pipefail

die() { echo "::error::$*" >&2; exit 1; }
log() { echo ">> $*" >&2; }

LIT_OS="${LIT_OS:?LIT_OS is required (linux|darwin|windows)}"
LIT_ARCH="${LIT_ARCH:?LIT_ARCH is required (amd64|arm64)}"
LIT_OCR="${LIT_OCR:-1}"
LITEPARSE_REF="${LITEPARSE_REF:-crates-v2.0.3}"
PDFIUM_TAG="${PDFIUM_TAG:-chromium/7847}"
LITEPARSE_REPO="${LITEPARSE_REPO:-https://github.com/run-llama/liteparse.git}"
PDFIUM_REPO="${PDFIUM_REPO:-run-llama/pdfium-binaries}"
STAGE_DIR="${STAGE_DIR:-dist/lit-${LIT_OS}-${LIT_ARCH}}"

# --- Map target -> rust triple, pdfium asset, and lib layout ----------------
case "${LIT_OS}/${LIT_ARCH}" in
  linux/amd64)   PDFIUM_ASSET=pdfium-linux-x64.tgz;   PDFIUM_MEMBER=lib/libpdfium.so;    LIB_NAME=libpdfium.so    ;;
  linux/arm64)   PDFIUM_ASSET=pdfium-linux-arm64.tgz; PDFIUM_MEMBER=lib/libpdfium.so;    LIB_NAME=libpdfium.so    ;;
  darwin/amd64)  PDFIUM_ASSET=pdfium-mac-x64.tgz;     PDFIUM_MEMBER=lib/libpdfium.dylib; LIB_NAME=libpdfium.dylib ;;
  darwin/arm64)  PDFIUM_ASSET=pdfium-mac-arm64.tgz;   PDFIUM_MEMBER=lib/libpdfium.dylib; LIB_NAME=libpdfium.dylib ;;
  windows/amd64) PDFIUM_ASSET=pdfium-win-x64.tgz;     PDFIUM_MEMBER=bin/pdfium.dll;      LIB_NAME=pdfium.dll      ;;
  windows/arm64) PDFIUM_ASSET=pdfium-win-arm64.tgz;   PDFIUM_MEMBER=bin/pdfium.dll;      LIB_NAME=pdfium.dll      ;;
  *) die "unsupported target ${LIT_OS}/${LIT_ARCH}" ;;
esac

if [ "${LIT_OS}" = "windows" ]; then
  LIT_BIN_NAME=lit.exe
else
  LIT_BIN_NAME=lit
fi

# --- 1. Obtain the LiteParse source tree ------------------------------------
WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

if [ -n "${LITEPARSE_SRC:-}" ]; then
  SRC="${LITEPARSE_SRC}"
  log "using existing LiteParse source at ${SRC}"
else
  SRC="${WORKDIR}/liteparse"
  log "cloning ${LITEPARSE_REPO} @ ${LITEPARSE_REF}"
  git clone --depth 1 --branch "${LITEPARSE_REF}" "${LITEPARSE_REPO}" "${SRC}"
fi

# --- 2. Build lit -----------------------------------------------------------
# Tesseract's vendored build (tesseract-rs build-tesseract feature) drives CMake;
# Ninja is the generator LiteParse CI uses and the one that reliably works on
# Windows. Harmless on unix when Ninja is installed.
export CMAKE_GENERATOR="${CMAKE_GENERATOR:-Ninja}"

CARGO_FLAGS=(--release -p liteparse --bin lit)
if [ "${LIT_OCR}" = "0" ]; then
  log "building lit WITHOUT OCR (--no-default-features)"
  CARGO_FLAGS+=(--no-default-features)
else
  log "building lit WITH OCR (Tesseract, statically linked)"
fi

( cd "${SRC}" && cargo build "${CARGO_FLAGS[@]}" )

LIT_BUILT="${SRC}/target/release/${LIT_BIN_NAME}"
[ -f "${LIT_BUILT}" ] || die "expected built binary not found at ${LIT_BUILT}"

# --- 3. Fetch the matching libpdfium ----------------------------------------
PDFIUM_URL="https://github.com/${PDFIUM_REPO}/releases/download/${PDFIUM_TAG//\//%2F}/${PDFIUM_ASSET}"
log "downloading ${PDFIUM_ASSET} from ${PDFIUM_REPO}@${PDFIUM_TAG}"
PDFIUM_TGZ="${WORKDIR}/pdfium.tgz"
curl -fsSL -o "${PDFIUM_TGZ}" "${PDFIUM_URL}"

PDFIUM_EXTRACT="${WORKDIR}/pdfium"
mkdir -p "${PDFIUM_EXTRACT}"
tar xzf "${PDFIUM_TGZ}" -C "${PDFIUM_EXTRACT}"
PDFIUM_LIB="${PDFIUM_EXTRACT}/${PDFIUM_MEMBER}"
[ -f "${PDFIUM_LIB}" ] || die "pdfium member ${PDFIUM_MEMBER} missing from ${PDFIUM_ASSET}"

# --- 4. Stage the bundle ----------------------------------------------------
rm -rf "${STAGE_DIR}"
mkdir -p "${STAGE_DIR}"
cp "${LIT_BUILT}" "${STAGE_DIR}/${LIT_BIN_NAME}"
cp "${PDFIUM_LIB}" "${STAGE_DIR}/${LIB_NAME}"
chmod +x "${STAGE_DIR}/${LIT_BIN_NAME}" 2>/dev/null || true

# Carry licenses so the redistributable bundle is compliant.
for lic in "${SRC}/LICENSE" "${SRC}/LICENSE-APACHE" "${SRC}/LICENSE.md"; do
  [ -f "${lic}" ] && cp "${lic}" "${STAGE_DIR}/LICENSE-liteparse.$(basename "${lic}")" || true
done
for lic in "${PDFIUM_EXTRACT}/LICENSE" "${PDFIUM_EXTRACT}/licenses"/*; do
  [ -e "${lic}" ] && cp -R "${lic}" "${STAGE_DIR}/" 2>/dev/null || true
done

cat > "${STAGE_DIR}/BUNDLE.txt" <<EOF
LiteParse \`lit\` bundle for ${LIT_OS}/${LIT_ARCH}
  liteparse ref : ${LITEPARSE_REF}
  ocr           : $( [ "${LIT_OCR}" = "0" ] && echo "disabled" || echo "enabled (tesseract)" )
  pdfium tag    : ${PDFIUM_TAG} (${PDFIUM_REPO})

Contents:
  ${LIT_BIN_NAME}   - LiteParse CLI
  ${LIB_NAME}       - PDFium shared library (dlopen'd by lit at runtime)

The consumer MUST set PDFIUM_LIB_PATH to this directory before running lit:
  PDFIUM_LIB_PATH=<this dir> ${LIT_BIN_NAME} parse <file>
EOF

log "staged bundle at ${STAGE_DIR}:"
ls -l "${STAGE_DIR}" >&2

# Expose key facts to the workflow if running under GitHub Actions.
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "stage_dir=${STAGE_DIR}"
    echo "lit_bin=${LIT_BIN_NAME}"
    echo "pdfium_lib=${LIB_NAME}"
  } >> "${GITHUB_OUTPUT}"
fi
