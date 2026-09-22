#!/usr/bin/env bash

set -euo pipefail

repo="${OUTPIPE_REPO:-outpipe/outpipe}"
download_base_url="${OUTPIPE_DOWNLOAD_BASE_URL:-https://cli.outpipe.dev/releases/cli}"
version="${OUTPIPE_VERSION:-latest}"
install_dir="${OUTPIPE_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "unsupported operating system" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported architecture" >&2; exit 1 ;;
esac

asset="outpipe-cli_${os}_${arch}.tar.gz"
if [ "$version" = "latest" ]; then
  download_url="${download_base_url}/latest/${asset}"
  fallback_url="https://github.com/${repo}/releases/latest/download/${asset}"
else
  download_url="${download_base_url}/${version}/${asset}"
  fallback_url="https://github.com/${repo}/releases/download/${version}/${asset}"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

if ! curl --fail --location --silent --show-error "$download_url" -o "$tmp_dir/$asset"; then
  curl --fail --location --silent --show-error "$fallback_url" -o "$tmp_dir/$asset"
fi
tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"

mkdir -p "$install_dir"
install "$tmp_dir/outpipe-cli" "$install_dir/outpipe"
echo "installed outpipe to $install_dir/outpipe"
