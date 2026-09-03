#!/bin/sh
# herdr-bridge installer — downloads a prebuilt release binary.
#
#   curl -fsSL https://raw.githubusercontent.com/sarathsp06/herdrweb/main/install.sh | sh
#
# Environment overrides:
#   HERDR_VERSION   version to install, e.g. v0.1.0 or 0.1.0 (default: latest release)
#   BINDIR          install directory (default: /usr/local/bin if writable, else ~/.local/bin)
set -eu

REPO="sarathsp06/herdrweb"
PROJECT="herdrweb"
BIN="herdr-bridge"

info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[33mwarn:\033[0m %s\n' "$*" >&2; }
err()  { printf '\033[31merror:\033[0m %s\n' "$*" >&2; exit 1; }

# ---- fetch helper: curl or wget ----
if command -v curl >/dev/null 2>&1; then
  fetch()   { curl -fsSL "$1"; }
  fetchto() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch()   { wget -qO- "$1"; }
  fetchto() { wget -qO "$2" "$1"; }
else
  err "need curl or wget on PATH"
fi

# ---- platform ----
os=$(uname -s)
case "$os" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *) err "unsupported OS: $os (linux and darwin only)" ;;
esac
arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) err "unsupported arch: $arch (amd64 and arm64 only)" ;;
esac

# ---- resolve version ----
ver="${HERDR_VERSION:-}"
if [ -z "$ver" ]; then
  info "resolving latest release"
  ver=$(fetch "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4) || true
  [ -n "$ver" ] || err "could not resolve latest release — set HERDR_VERSION=vX.Y.Z"
fi
case "$ver" in v*) tag="$ver" ;; *) tag="v$ver" ;; esac
num="${tag#v}"
asset="${PROJECT}_${num}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$tag/$asset"

# ---- install dir ----
bindir="${BINDIR:-}"
if [ -z "$bindir" ]; then
  if [ -w /usr/local/bin ]; then bindir=/usr/local/bin; else bindir="$HOME/.local/bin"; fi
fi
mkdir -p "$bindir" || err "cannot create $bindir"

# ---- download, verify, install ----
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

info "downloading $asset ($tag)"
fetchto "$url" "$tmp/$asset" || err "download failed: $url"

if fetchto "https://github.com/$REPO/releases/download/$tag/checksums.txt" "$tmp/checksums.txt" 2>/dev/null; then
  want=$(grep " $asset\$" "$tmp/checksums.txt" | awk '{print $1}' || true)
  if [ -n "$want" ]; then
    if command -v sha256sum >/dev/null 2>&1; then got=$(sha256sum "$tmp/$asset" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then got=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')
    else got=""; warn "no sha256 tool found; skipping checksum verification"; fi
    if [ -n "$got" ]; then
      [ "$got" = "$want" ] || err "checksum mismatch for $asset"
      info "checksum ok"
    fi
  else
    warn "no checksum entry for $asset; skipping verification"
  fi
else
  warn "checksums.txt not available; skipping verification"
fi

tar -xzf "$tmp/$asset" -C "$tmp"
[ -f "$tmp/$BIN" ] || err "archive did not contain $BIN"
chmod +x "$tmp/$BIN"
mv -f "$tmp/$BIN" "$bindir/$BIN"

info "installed $BIN $num -> $bindir/$BIN"
case ":$PATH:" in
  *":$bindir:"*) ;;
  *) warn "$bindir is not on your PATH — add:  export PATH=\"$bindir:\$PATH\"" ;;
esac

cat <<EOF

Run the bridge (loopback only by default):
  $BIN

Reach it from your phone over Tailscale (HTTPS = installable PWA + push):
  tailscale serve --bg --https=443 127.0.0.1:7331
  # then open https://<machine>.<tailnet>.ts.net on the phone
EOF
