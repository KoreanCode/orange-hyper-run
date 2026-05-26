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
    signature_url="${HYPER_RUN_SIGNATURE_URL:-}"
    identity_regexp="${HYPER_RUN_COSIGN_IDENTITY_REGEXP:-}"
    ;;
  github:*)
    release_repo="${repo#github:}"
    asset="hyper-$os-$arch"
    url="https://github.com/$release_repo/releases/latest/download/$asset"
    checksum_url="https://github.com/$release_repo/releases/latest/download/checksums.txt"
    signature_url="https://github.com/$release_repo/releases/latest/download/$asset.sigstore.json"
    identity_regexp="${HYPER_RUN_COSIGN_IDENTITY_REGEXP:-https://github.com/$release_repo/.github/workflows/release.yml@refs/tags/v.*}"
    ;;
  *)
    release_repo="$repo"
    asset="hyper-$os-$arch"
    url="https://github.com/$release_repo/releases/latest/download/$asset"
    checksum_url="https://github.com/$release_repo/releases/latest/download/checksums.txt"
    signature_url="https://github.com/$release_repo/releases/latest/download/$asset.sigstore.json"
    identity_regexp="${HYPER_RUN_COSIGN_IDENTITY_REGEXP:-https://github.com/$release_repo/.github/workflows/release.yml@refs/tags/v.*}"
    ;;
esac

asset="${asset:-$(basename "$url")}"

target="$install_dir/hyper"
tmp="${TMPDIR:-/tmp}/hyper-install-$$"
checksums_tmp="$tmp.checksums"
signature_tmp="$tmp.sigstore.json"

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

verify_signature="${HYPER_RUN_VERIFY_SIGNATURE:-auto}"
if [ -n "${signature_url:-}" ]; then
  if command -v cosign >/dev/null 2>&1; then
    if [ -z "${identity_regexp:-}" ]; then
      echo "Signature verification requires HYPER_RUN_COSIGN_IDENTITY_REGEXP for custom URLs" >&2
      rm -f "$tmp" "$checksums_tmp" "$signature_tmp"
      exit 1
    fi
    echo "Verifying signature from $signature_url"
    if curl -fsSL "$signature_url" -o "$signature_tmp"; then
      cosign verify-blob \
        --bundle "$signature_tmp" \
        --certificate-identity-regexp "$identity_regexp" \
        --certificate-oidc-issuer "${HYPER_RUN_COSIGN_OIDC_ISSUER:-https://token.actions.githubusercontent.com}" \
        "$tmp"
      rm -f "$signature_tmp"
    else
      case "$verify_signature" in
        1|true|required|always)
          echo "Signature bundle download failed for $asset" >&2
          rm -f "$tmp" "$checksums_tmp" "$signature_tmp"
          exit 1
          ;;
        *)
          echo "Signature verification skipped: signature bundle not found; checksum still verified"
          rm -f "$signature_tmp"
          ;;
      esac
    fi
  else
    case "$verify_signature" in
      1|true|required|always)
        echo "Signature verification requires cosign. Install cosign or unset HYPER_RUN_VERIFY_SIGNATURE." >&2
        rm -f "$tmp" "$checksums_tmp" "$signature_tmp"
        exit 1
        ;;
      *)
        echo "Signature verification skipped: cosign not found; checksum still verified"
        ;;
    esac
  fi
else
  case "$verify_signature" in
    1|true|required|always)
      echo "Signature verification requires HYPER_RUN_SIGNATURE_URL for custom URLs." >&2
      rm -f "$tmp" "$checksums_tmp" "$signature_tmp"
      exit 1
      ;;
  esac
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
