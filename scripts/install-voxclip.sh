#!/bin/sh

usage() {
  cat <<'HELP'
Install voxclip from GitHub Releases.

Usage:
  scripts/install-voxclip.sh [--version vX.Y.Z] [--prefix PATH] [--bin-dir PATH]

Options:
  --version   Release tag to install (default: latest release)
  --prefix    Installation prefix (default: ~/.local)
  --bin-dir   Binary directory (default: <prefix>/bin)
  -h, --help  Show this help

Examples:
  curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
  curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh -s -- --version v1.2.3
HELP
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux|darwin) echo "$os" ;;
    *)
      echo "Unsupported OS: $os" >&2
      return 1
      ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "Unsupported architecture: $arch" >&2
      return 1
      ;;
  esac
}

latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n 1
}

sha256_file() {
  local file="$1"
  if need_cmd sha256sum; then
    sha256sum "$file" | awk '{print $1}'
  elif need_cmd shasum; then
    shasum -a 256 "$file" | awk '{print $1}'
  else
    echo "Need sha256sum or shasum for checksum verification" >&2
    return 1
  fi
}

install_file() {
  local src="$1" dst="$2" mode="$3"
  local dst_dir
  dst_dir="$(dirname "$dst")"

  if mkdir -p "$dst_dir" 2>/dev/null; then
    install -m "$mode" "$src" "$dst"
    return
  fi

  if need_cmd sudo; then
    sudo mkdir -p "$dst_dir"
    sudo install -m "$mode" "$src" "$dst"
    return
  fi

  echo "Cannot write to ${dst_dir} and sudo is unavailable" >&2
  return 1
}

main() {
  set -eu

  local REPO="fmueller/voxclip"
  local VERSION=""
  local PREFIX="${HOME}/.local"
  local BIN_DIR=""

  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        VERSION="$2"
        shift 2
        ;;
      --prefix)
        PREFIX="$2"
        shift 2
        ;;
      --bin-dir)
        BIN_DIR="$2"
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

  if [ -z "$BIN_DIR" ]; then
    BIN_DIR="${PREFIX}/bin"
  fi

  if ! need_cmd curl; then
    echo "curl is required" >&2
    exit 1
  fi
  if ! need_cmd tar; then
    echo "tar is required" >&2
    exit 1
  fi
  if ! need_cmd install; then
    echo "install command is required" >&2
    exit 1
  fi

  local OS ARCH
  OS="$(detect_os)"
  ARCH="$(detect_arch)"

  if [ -z "$VERSION" ]; then
    VERSION="$(latest_tag)"
  fi

  if [ -z "$VERSION" ]; then
    echo "Could not determine release version" >&2
    exit 1
  fi

  local VERSION_STRIPPED="${VERSION#v}"
  local ARTIFACT="voxclip_${VERSION_STRIPPED}_${OS}_${ARCH}.tar.gz"
  local CHECKSUMS="checksums.txt"
  local BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

  TMP_DIR="$(mktemp -d)"
  cleanup() {
    rm -rf "$TMP_DIR"
  }
  trap cleanup EXIT

  echo "Installing Voxclip ${VERSION} (${OS}/${ARCH})"
  echo "Downloading ${ARTIFACT}"
  curl -fsSL -o "${TMP_DIR}/${ARTIFACT}" "${BASE_URL}/${ARTIFACT}"
  curl -fsSL -o "${TMP_DIR}/${CHECKSUMS}" "${BASE_URL}/${CHECKSUMS}"

  local expected_sha
  expected_sha="$(grep " ${ARTIFACT}$" "${TMP_DIR}/${CHECKSUMS}" | awk '{print $1}')"
  if [ -z "$expected_sha" ]; then
    echo "Checksum for ${ARTIFACT} not found in ${CHECKSUMS}" >&2
    exit 1
  fi

  local actual_sha
  actual_sha="$(sha256_file "${TMP_DIR}/${ARTIFACT}")"
  if [ "$expected_sha" != "$actual_sha" ]; then
    echo "Checksum mismatch for ${ARTIFACT}" >&2
    echo "Expected: ${expected_sha}" >&2
    echo "Actual:   ${actual_sha}" >&2
    exit 1
  fi

  mkdir -p "${TMP_DIR}/extract"
  tar -xzf "${TMP_DIR}/${ARTIFACT}" -C "${TMP_DIR}/extract"

  if [ ! -f "${TMP_DIR}/extract/voxclip" ]; then
    echo "Archive is missing voxclip binary" >&2
    exit 1
  fi
  if [ ! -f "${TMP_DIR}/extract/libexec/whisper/whisper-cli" ]; then
    echo "Archive is missing bundled whisper engine at libexec/whisper/whisper-cli" >&2
    exit 1
  fi

  local LIBEXEC_DIR="${PREFIX}/libexec/whisper"

  install_file "${TMP_DIR}/extract/voxclip" "${BIN_DIR}/voxclip" 0755
  install_file "${TMP_DIR}/extract/libexec/whisper/whisper-cli" "${LIBEXEC_DIR}/whisper-cli" 0755

  echo ""
  echo "Installed: ${BIN_DIR}/voxclip"
  echo "Bundled engine: ${LIBEXEC_DIR}/whisper-cli"
  echo ""
  echo "Next steps:"
  echo "  voxclip setup"
  echo "  voxclip devices"
  echo "  voxclip"

  case ":$PATH:" in
    *":${BIN_DIR}:"*) ;;
    *)
      echo ""
      echo "Add this to your shell profile if needed:"
      echo "  export PATH=\"${BIN_DIR}:\$PATH\""
      ;;
  esac
}

if [ "${_VOXCLIP_TESTING:-}" != "1" ]; then
  main "$@"
fi
