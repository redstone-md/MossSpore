#!/bin/sh
#
# MossSpore — Run a Spore, grow the Moss.
# Quick installer: detects OS/arch, downloads the latest release binary,
# and optionally sets up a systemd service.
#
# Usage: curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
#        curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh -s -- --version 0.1.0

set -eu

# ── Config ───────────────────────────────────────────────────────────
REPO="redstone-md/MossSpore"
BINDIR="${BINDIR:-/usr/local/bin}"
CONFIGDIR="${CONFIGDIR:-/etc/mossspore}"
IDENTITYDIR="${IDENTITYDIR:-/var/lib/mossspore}"
VERSION="${VERSION:-latest}"

# ── Color helpers ────────────────────────────────────────────────────
info()  { printf "\033[0;34m•\033[0m %s\n" "$*" >&2; }
ok()    { printf "\033[0;32m✓\033[0m %s\n" "$*" >&2; }
warn()  { printf "\033[0;33m⚠\033[0m %s\n" "$*" >&2; }
err()   { printf "\033[0;31m✗\033[0m %s\n" "$*" >&2; exit 1; }

# ── Detect platform ──────────────────────────────────────────────────
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)   PLATFORM="linux"   ;;
    darwin)  PLATFORM="darwin"  ;;
    *)       err "Unsupported OS: $OS (only linux and darwin are supported)" ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)            err "Unsupported architecture: $ARCH" ;;
  esac

  SUFFIX=""
  [ "$PLATFORM" = "darwin" ] && SUFFIX=""
  BINARY="mossspore-${PLATFORM}-${ARCH}${SUFFIX}"
}

# ── Download ──────────────────────────────────────────────────────────
download() {
  if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
  else
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY}"
  fi

  info "Downloading ${BINARY}..."
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  if command -v curl >/dev/null 2>&1; then
    curl -sSL --fail "$URL" -o "${tmp}/${BINARY}"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$URL" -O "${tmp}/${BINARY}"
  else
    err "Neither curl nor wget found. Install one of them and retry."
  fi

  chmod +x "${tmp}/${BINARY}"
  eval "BINARY_PATH=\"${tmp}/${BINARY}\""
}

# ── Install ───────────────────────────────────────────────────────────
install_binary() {
  info "Installing to ${BINDIR}/mossspore..."

  if [ ! -d "$BINDIR" ]; then
    mkdir -p "$BINDIR" 2>/dev/null || true
  fi

  if [ -f "${BINDIR}/mossspore" ]; then
    warn "Overwriting existing binary at ${BINDIR}/mossspore"
  fi

  if cp "$BINARY_PATH" "${BINDIR}/mossspore" 2>/dev/null; then
    chmod +x "${BINDIR}/mossspore"
    ok "Installed to ${BINDIR}/mossspore"
  else
    # Fall back to sudo
    if command -v sudo >/dev/null 2>&1; then
      sudo cp "$BINARY_PATH" "${BINDIR}/mossspore"
      sudo chmod +x "${BINDIR}/mossspore"
      ok "Installed to ${BINDIR}/mossspore (via sudo)"
    else
      err "Cannot write to ${BINDIR}. Try: sudo curl ... | sh"
    fi
  fi
}

# ── Config ────────────────────────────────────────────────────────────
setup_config() {
  if [ -f "${CONFIGDIR}/config.json" ]; then
    ok "Config already exists at ${CONFIGDIR}/config.json"
    return
  fi

  info "Creating default config at ${CONFIGDIR}/config.json..."

  if mkdir -p "$CONFIGDIR" 2>/dev/null; then
    true
  else
    sudo mkdir -p "$CONFIGDIR"
  fi
  if mkdir -p "$IDENTITYDIR" 2>/dev/null; then
    true
  else
    sudo mkdir -p "$IDENTITYDIR"
  fi
  sudo chmod 0700 "$IDENTITYDIR" 2>/dev/null || true

  cat <<-EOF | sudo tee "${CONFIGDIR}/config.json" >/dev/null
{
  "mesh_id": "moss-relay/1",
  "listen_port": 0,
  "identity_path": "${IDENTITYDIR}/identity.key",
  "lan_discovery": true,
  "relay_mesh": {
    "enabled": true
  },
  "relay": {
    "enabled": true,
    "max_bandwidth_kbps": 1024,
    "max_sessions": 100,
    "session_ttl_sec": 1800,
    "min_uptime_sec": 60
  },
  "monitor": {
    "enabled": true,
    "listen": ":9800"
  }
}
EOF
  ok "Config written to ${CONFIGDIR}/config.json"

  info "This spore joins the shared relay mesh (moss-relay/1) by default."
  warn "It only becomes a viable SuperNode if its peer port is inbound-reachable (public IP or forwarded)."
  info "Check reachability once running: curl localhost:9800/health, look at 'nat_type' —"
  info "  public / cone: reachable, good SuperNode candidate."
  warn "  symmetric / cgnat: not reachable, will relay through the mesh but cannot serve as a SuperNode."
}

# ── systemd ───────────────────────────────────────────────────────────
setup_systemd() {
  if [ "$PLATFORM" != "linux" ]; then
    info "systemd service is only available on Linux — skipping"
    return
  fi

  if [ -f /etc/systemd/system/mossspore.service ]; then
    ok "systemd service already exists at /etc/systemd/system/mossspore.service"
    return
  fi

  info "Installing systemd service..."

  cat <<-EOF | sudo tee /etc/systemd/system/mossspore.service >/dev/null
[Unit]
Description=MossSpore P2P Relay Daemon
Documentation=https://github.com/${REPO}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${BINDIR}/mossspore --config ${CONFIGDIR}/config.json
Restart=always
RestartSec=30
User=mossspore
Group=mossspore
StateDirectory=mossspore
MemoryMax=256M

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl daemon-reload 2>/dev/null || true
  sudo systemctl enable mossspore 2>/dev/null || warn "systemctl enable failed — run manually: systemctl enable mossspore"

  ok "systemd service installed: /etc/systemd/system/mossspore.service"
}

# ── system user ───────────────────────────────────────────────────────
setup_user() {
  if [ "$PLATFORM" != "linux" ]; then
    return
  fi
  if id mossspore >/dev/null 2>&1; then
    return
  fi
  info "Creating system user 'mossspore'..."
  sudo useradd --system --no-create-home --shell /usr/sbin/nologin mossspore 2>/dev/null || true
}

# ── Main ──────────────────────────────────────────────────────────────
main() {
  detect_platform

  printf "\033[0;32m"
  printf "  ╔══════════════════════════════════════╗\n"
  printf "  ║         MossSpore Installer          ║\n"
  printf "  ║    Run a Spore, grow the Moss.       ║\n"
  printf "  ╚══════════════════════════════════════╝\n"
  printf "\033[0m\n"

  download
  install_binary

  # Ask about systemd
  if [ "${PLATFORM}" = "linux" ]; then
    printf "\n"
    info "Would you like to install MossSpore as a systemd service?"
    printf "  [Y]es / [n]o / [s]kip config: "
    read -r answer < /dev/tty 2>/dev/null || answer="y"

    case "$answer" in
      n|N|no|NO)
        info "Skipping service setup. Run manually: mossspore --config ${CONFIGDIR}/config.json"
        info "Creating config anyway..."
        setup_config
        ;;
      s|S|skip|SKIP)
        info "Skipping config and service setup entirely."
        ;;
      *)
        setup_user
        setup_config
        setup_systemd
        ;;
    esac
  fi

  printf "\n"
  ok "MossSpore installed successfully!"
  echo ""
  echo "  Binary:  ${BINDIR}/mossspore"
  echo "  Config:  ${CONFIGDIR}/config.json"
  echo "  Service: mossspore.service"
  echo ""
  echo "  Start now:  sudo systemctl start mossspore"
  echo "  Check logs: journalctl -u mossspore -f"
  echo "  Check health: curl http://127.0.0.1:9800/health"
  echo ""
}

main
