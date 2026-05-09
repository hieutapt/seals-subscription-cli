#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
MUTED='\033[0;2m'
NC='\033[0m'

REPO="hieutapt/seals-subscription-cli"
BINARY_NAME="seal-cli"

die() {
  echo -e "${RED}Error: $1${NC}" >&2
  exit 1
}

usage() {
  cat <<EOF
seal-cli Installer

Usage: install.sh [options]

Options:
    -h, --help              Display this help message
    -v, --version <version> Install a specific version (e.g., 0.2.0)
        --install-dir <dir> Override the installation directory
        --no-modify-path    Don't add install dir to shell config files

Environment Variables:
    SEAL_INSTALL_DIR        Override the installation directory
    SEAL_VERSION            Install a specific version (e.g., 0.2.0)

Examples:
    curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash
    curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash -s -- --version 0.2.0
    SEAL_VERSION=0.2.0 curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash
    SEAL_INSTALL_DIR=~/.local/bin curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash
EOF
}

# --- Parse args ---
requested_version="${SEAL_VERSION:-}"
install_dir="${SEAL_INSTALL_DIR:-}"
no_modify_path=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    -v|--version)
      [[ -n "${2:-}" ]] || die "--version requires a version argument"
      requested_version="$2"; shift 2 ;;
    --install-dir)
      [[ -n "${2:-}" ]] || die "--install-dir requires a path argument"
      install_dir="$2"; shift 2 ;;
    --no-modify-path)
      no_modify_path=true; shift ;;
    *) die "Unknown option: $1" ;;
  esac
done

# --- Detect OS ---
case "$(uname -s)" in
  Darwin*) os="darwin" ;;
  Linux*)  os="linux"  ;;
  *) die "Unsupported OS: $(uname -s)" ;;
esac

# --- Detect architecture ---
case "$(uname -m)" in
  x86_64)        arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) die "Unsupported architecture: $(uname -m)" ;;
esac

# --- Resolve version ---
if [[ -z "$requested_version" ]]; then
  echo -e "${MUTED}Fetching latest seal-cli version...${NC}"
  requested_version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
  [[ -n "$requested_version" ]] || die "Failed to fetch latest version from GitHub"
fi

# Strip leading 'v' for asset name construction; keep it for the tag URL
version="${requested_version#v}"
tag="v${version}"

# GoReleaser asset name format: seal-cli_<version>_<os>_<arch>.tar.gz
asset="seal-cli_${version}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${tag}/${asset}"

echo -e "${MUTED}Downloading seal-cli v${version} (${os}/${arch})...${NC}"

# --- Download to temp dir ---
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

curl -fsSL "$url" -o "${tmpdir}/${asset}" \
  || die "Download failed. Check that release ${tag} exists at:\n  https://github.com/${REPO}/releases"

tar -xzf "${tmpdir}/${asset}" -C "$tmpdir"

tmp_binary="${tmpdir}/${BINARY_NAME}"
[[ -f "$tmp_binary" ]] || die "Binary '${BINARY_NAME}' not found in archive"
chmod +x "$tmp_binary"

# --- Choose install directory ---
if [[ -z "$install_dir" ]]; then
  if [[ -w "/usr/local/bin" ]]; then
    install_dir="/usr/local/bin"
  else
    install_dir="$HOME/.local/bin"
  fi
fi

mkdir -p "$install_dir"
mv "$tmp_binary" "${install_dir}/${BINARY_NAME}"

echo -e "${GREEN}✓ seal-cli v${version} installed to ${install_dir}/${BINARY_NAME}${NC}"

# --- Add to PATH if needed ---
if ! command -v "$BINARY_NAME" >/dev/null 2>&1 && [[ "$no_modify_path" == "false" ]]; then
  shell_config=""
  case "${SHELL:-}" in
    */zsh)  shell_config="$HOME/.zshrc" ;;
    */bash) shell_config="$HOME/.bashrc" ;;
  esac

  if [[ -n "$shell_config" ]]; then
    export_line="export PATH=\"${install_dir}:\$PATH\""
    if ! grep -qF "$install_dir" "$shell_config" 2>/dev/null; then
      echo "" >> "$shell_config"
      echo "# Added by seal-cli installer" >> "$shell_config"
      echo "$export_line" >> "$shell_config"
      echo -e "${MUTED}Added ${install_dir} to PATH in ${shell_config}${NC}"
      echo -e "${MUTED}Restart your shell or run: source ${shell_config}${NC}"
    fi
  else
    echo -e "${MUTED}Add ${install_dir} to your PATH to use seal-cli from anywhere.${NC}"
  fi
fi

echo ""
echo -e "Run ${GREEN}seal-cli --help${NC} to get started."
echo -e "Set your token: ${GREEN}seal-cli profile set default --token <your-token>${NC}"
