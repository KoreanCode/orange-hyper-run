#!/usr/bin/env sh
set -eu

repo="${HYPER_RUN_UPDATE_SOURCE:-KoreanCode/orange-hyper-run}"
install_dir="${HYPER_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *)
    echo "Unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  arm64|aarch64) arch="arm64" ;;
  x86_64|amd64) arch="amd64" ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

case "$repo" in
  http://*|https://*) url="$repo" ;;
  github:*) url="https://github.com/${repo#github:}/releases/latest/download/hyper-$os-$arch" ;;
  *) url="https://github.com/$repo/releases/latest/download/hyper-$os-$arch" ;;
esac

target="$install_dir/hyper"
tmp="${TMPDIR:-/tmp}/hyper-install-$$"

mkdir -p "$install_dir"
echo "Installing Hyper Run from $url"
curl -fsSL "$url" -o "$tmp"
chmod 0755 "$tmp"
mv "$tmp" "$target"

echo "Installed: $target"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    echo "Warning: $install_dir is not on PATH" >&2
    echo "Add this to your shell profile:" >&2
    echo "  export PATH=\"$install_dir:\$PATH\"" >&2
    ;;
esac

if "$target" version >/dev/null 2>&1; then
  "$target" version
else
  "$target"
fi
