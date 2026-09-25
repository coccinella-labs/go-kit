#!/usr/bin/env bash
set -euo pipefail

REPO="coccinella-labs/go-kit"
INSTALL_DIR="${KIT_INSTALL_DIR:-$HOME/.local/bin}"
BIN_NAME="kit"
TMPDIR=""

info() {
  printf "[INFO] %s\n" "$1"
}

warn() {
  printf "[WARN] %s\n" "$1"
}

error() {
  printf "[ERROR] %s\n" "$1" >&2
}

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       error "Unsupported OS: $(uname -s)"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)  echo "amd64" ;;
    arm64)   echo "arm64" ;;
    aarch64) echo "arm64" ;;
    *)       error "Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
}

get_latest_version() {
  local version
  version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [[ -z "$version" ]]; then
    error "Could not determine latest version"
    exit 1
  fi
  echo "$version"
}

get_expected_checksum() {
  local checksum_url="$1"
  local filename="$2"
  curl -fsSL "$checksum_url" | awk -v f="$filename" '$2 == f {print $1}'
}

cleanup() {
  if [[ -n "${TMPDIR:-}" && -d "$TMPDIR" ]]; then
    rm -rf "$TMPDIR"
  fi
}

main() {
  info "Installing Kit CLI"

  local os arch version
  os=$(detect_os)
  arch=$(detect_arch)
  version=$(get_latest_version)

  info "Detected OS: ${os}, Architecture: ${arch}"
  info "Latest version: ${version}"

  local asset_name="${BIN_NAME}-${os}-${arch}"
  if [[ "$os" == "windows" ]]; then
    asset_name="${asset_name}.exe"
  fi

  local download_url="https://github.com/${REPO}/releases/download/${version}/${asset_name}"
  local checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

  TMPDIR=$(mktemp -d)
  trap cleanup EXIT

  local binary_path="${TMPDIR}/${BIN_NAME}"

  info "Downloading ${asset_name}..."
  curl -fsSL "$download_url" -o "$binary_path"

  if [[ ! -f "$binary_path" ]]; then
    error "Download failed"
    exit 1
  fi

  info "Verifying checksum..."
  local expected_checksum
  expected_checksum=$(get_expected_checksum "$checksum_url" "$asset_name")

  if [[ -z "$expected_checksum" ]]; then
    warn "Could not fetch checksum, skipping verification"
  else
    local actual_checksum
    actual_checksum=$(sha256sum "$binary_path" | awk '{print $1}')

    if [[ "$actual_checksum" != "$expected_checksum" ]]; then
      error "Checksum verification failed"
      error "Expected: $expected_checksum"
      error "Actual:   $actual_checksum"
      rm -f "$binary_path"
      exit 1
    fi

    info "Checksum verified"
  fi

  info "Installing ${BIN_NAME} to ${INSTALL_DIR}"

  if [[ ! -d "$INSTALL_DIR" ]]; then
    mkdir -p "$INSTALL_DIR"
  fi

  cp "$binary_path" "${INSTALL_DIR}/${BIN_NAME}"
  chmod +x "${INSTALL_DIR}/${BIN_NAME}"

  if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    warn "${INSTALL_DIR} is not in your PATH"
    warn "Add this to your shell config:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi

  info "Successfully installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"
}

main "$@"
