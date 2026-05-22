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
  http://*|https://*)
    url="$repo"
    checksum_url="${HYPER_RUN_CHECKSUM_URL:-}"
    ;;
  github:*)
    release_repo="${repo#github:}"
    asset="hyper-$os-$arch"
    url="https://github.com/$release_repo/releases/latest/download/$asset"
    checksum_url="https://github.com/$release_repo/releases/latest/download/checksums.txt"
    ;;
  *)
    release_repo="$repo"
    asset="hyper-$os-$arch"
    url="https://github.com/$release_repo/releases/latest/download/$asset"
    checksum_url="https://github.com/$release_repo/releases/latest/download/checksums.txt"
    ;;
esac

asset="${asset:-$(basename "$url")}"

target="$install_dir/hyper"
tmp="${TMPDIR:-/tmp}/hyper-install-$$"
checksums_tmp="$tmp.checksums"

mkdir -p "$install_dir"
echo "Installing Hyper Run from $url"
curl -fsSL "$url" -o "$tmp"

if [ -n "${checksum_url:-}" ]; then
  echo "Verifying checksum from $checksum_url"
  curl -fsSL "$checksum_url" -o "$checksums_tmp"
  expected="$(awk -v name="$asset" '$2 == name { print $1 }' "$checksums_tmp")"
  if [ -z "$expected" ]; then
    echo "Checksum not found for $asset in checksums.txt" >&2
    rm -f "$tmp" "$checksums_tmp"
    exit 1
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$tmp" | awk '{ print $1 }')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$tmp" | awk '{ print $1 }')"
  else
    echo "No SHA256 tool found. Install sha256sum or shasum, or set HYPER_RUN_CHECKSUM_URL= to skip custom URL verification." >&2
    rm -f "$tmp" "$checksums_tmp"
    exit 1
  fi
  if [ "$actual" != "$expected" ]; then
    echo "Checksum mismatch for $asset" >&2
    echo "Expected: $expected" >&2
    echo "Actual:   $actual" >&2
    rm -f "$tmp" "$checksums_tmp"
    exit 1
  fi
  rm -f "$checksums_tmp"
fi

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
