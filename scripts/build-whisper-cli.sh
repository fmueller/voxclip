#!/usr/bin/env bash
set -euo pipefail

WHISPER_REF="${WHISPER_REF:-v1.8.3}"

usage() {
  cat <<'HELP'
Build whisper-cli and stage it for Voxclip packaging.

Usage:
  scripts/build-whisper-cli.sh [--ref vX.Y.Z]

Environment:
  WHISPER_REF   whisper.cpp git tag/branch (default: v1.8.3)

Output:
  packaging/whisper/<os>_<arch>/whisper-cli

This script builds for the current host OS/arch only.
HELP
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ref)
      WHISPER_REF="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

for cmd in git cmake; do
  if ! need_cmd "$cmd"; then
    echo "Missing required command: $cmd" >&2
    exit 1
  fi
done

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_RAW="$(uname -m)"

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH_RAW" >&2
    exit 1
    ;;
esac

ROOT_DIR="$(git rev-parse --show-toplevel)"
OUT_DIR="${ROOT_DIR}/packaging/whisper/${OS}_${ARCH}"
OUT_FILE="${OUT_DIR}/whisper-cli"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "Building whisper-cli (${WHISPER_REF}) for ${OS}/${ARCH}"
git clone --depth 1 --branch "$WHISPER_REF" https://github.com/ggml-org/whisper.cpp.git "$TMP_DIR/whisper.cpp"

EXTRA_CMAKE_FLAGS=()
if [[ "$ARCH" == "amd64" ]]; then
  EXTRA_CMAKE_FLAGS=(
    -DGGML_NATIVE=OFF
    -DGGML_AVX=OFF
    -DGGML_AVX2=OFF
    -DGGML_AVX512=OFF
    -DGGML_AVX512_VBMI=OFF
    -DGGML_AVX512_VNNI=OFF
    -DGGML_FMA=OFF
    -DGGML_F16C=OFF
  )
fi

cmake -S "$TMP_DIR/whisper.cpp" -B "$TMP_DIR/whisper.cpp/build" \
  -DCMAKE_BUILD_TYPE=Release \
  -DBUILD_SHARED_LIBS=OFF \
  -DGGML_OPENMP=OFF \
  "${EXTRA_CMAKE_FLAGS[@]}"
cmake --build "$TMP_DIR/whisper.cpp/build" --config Release --target whisper-cli -j "$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"

SRC_FILE=""
if [[ -f "$TMP_DIR/whisper.cpp/build/bin/whisper-cli" ]]; then
  SRC_FILE="$TMP_DIR/whisper.cpp/build/bin/whisper-cli"
elif [[ -f "$TMP_DIR/whisper.cpp/build/src/whisper-cli" ]]; then
  SRC_FILE="$TMP_DIR/whisper.cpp/build/src/whisper-cli"
else
  echo "Could not locate whisper-cli build output" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
cp "$SRC_FILE" "$OUT_FILE"
chmod +x "$OUT_FILE"

if [[ "$OS" == "linux" ]]; then
  if ldd "$OUT_FILE" | grep -E 'libwhisper|libggml' >/dev/null 2>&1; then
    echo "whisper-cli still links against libwhisper/libggml; expected static linkage" >&2
    ldd "$OUT_FILE" >&2
    exit 1
  fi
fi

if [[ "$OS" == "darwin" ]]; then
  if otool -L "$OUT_FILE" | grep -E 'libwhisper|libggml' >/dev/null 2>&1; then
    echo "whisper-cli still links against libwhisper/libggml; expected static linkage" >&2
    otool -L "$OUT_FILE" >&2
    exit 1
  fi
fi

echo "Staged bundled engine at: $OUT_FILE"
