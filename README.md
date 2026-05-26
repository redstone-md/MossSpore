# MossSpore

[![CI](https://github.com/redstone-md/MossSpore/actions/workflows/ci.yml/badge.svg)](https://github.com/redstone-md/MossSpore/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-brightgreen)](LICENSE)
[![Moss](https://img.shields.io/badge/powered%20by-Moss-2ea44f)](https://github.com/redstone-md/moss)

> **Run a Spore, grow the Moss.**
>
> MossSpore is a lightweight, headless relay daemon for the Moss P2P mesh protocol.
> Each instance is a new point of presence — it relays traffic, coordinates routes,
> and fills coverage gaps. Launch-and-forget infrastructure for a decentralised world.

---

## One-line Install

```bash
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
```

Detects your OS and architecture, downloads the latest release binary to `/usr/local/bin`,
creates a default config, and optionally sets up a systemd service.

Works on **Linux** (amd64 / arm64) and **macOS** (Intel / Apple Silicon).

---

## Quick Start

```bash
# 1. One-command install (Linux / macOS)
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh

# 2. Start the relay
sudo systemctl start mossspore

# 3. Check that it's alive
curl http://127.0.0.1:9800/health
# → {"status":"ok","nat_type":"port_restricted_cone"}

# 4. Or run directly (no service)
mossspore --mesh-id my-mesh --identity-path ~/.mossspore/identity.key
```

---

## The Spore

In biology, a spore is the smallest self-sufficient unit of a fungus or moss. It
drifts on the wind, lands on bare ground, and grows into a new patch — expanding
the organism across the landscape.

MossSpore is the digital equivalent.

A spore does not generate user traffic. It **relays**, **routes**, and **extends**
the mesh. Each running MossSpore instance is another node on the network that helps
peers behind NATs reach each other, propagates gossip, and makes the whole system
more resilient.

| Biological spore | MossSpore |
|---|---|
| Carried by wind to new terrain | Discovered via trackers / LAN / static peers |
| Roots and becomes part of a colony | Connects and becomes a mesh relay node |
| Requires no external care | Zero configuration, runs headless |
| Expands the organism's range | Extends network coverage and redundancy |

---

## Quick Start

```bash
# 1. Build
git clone https://github.com/moss/MossSpore
cd MossSpore
make build

# 2. Run (zero-config, ephemeral identity)
./build/mossspore --mesh-id my-mesh

# 3. Check that it's alive
curl http://127.0.0.1:9800/health
# → {"status":"ok","nat_type":"public"}
```

**Run a Spore, grow the Moss.** That single command turns your machine into a
mesh relay node. No account, no registration, no central server.

---

## Installation

### One-line install (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
```

This downloads the latest pre-built binary for your platform, installs it to `/usr/local/bin`,
and interactively asks whether you want a systemd service.

### From source

```bash
go build -o /usr/local/bin/mossspore ./cmd/mossspore
```

### Cross-compile

```bash
make build-linux    # → build/mossspore-linux-amd64
make build-windows  # → build/mossspore-windows-amd64.exe
make build-darwin   # → build/mossspore-darwin-arm64
```

Binaries are statically linked and require no runtime dependencies.

---

## Configuration

MossSpore works out of the box with sensible defaults tuned for relay operation.
Override anything via JSON config file:

```json
{
  "mesh_id": "moss-spore-production",
  "listen_port": 4001,
  "identity_path": "/var/lib/mossspore/identity.key",

  "relay": {
    "enabled": true,
    "max_bandwidth_kbps": 2048,
    "max_sessions": 200
  },

  "monitor": {
    "enabled": true,
    "listen": ":9800"
  }
}
```

```bash
mossspore --config /etc/mossspore/config.json
```

### Options

| Field | Type | Default | Description |
|---|---|---|---|
| `mesh_id` | `string` | `global` | Mesh network identifier **(required)** |
| `psk` | `string` | `""` | Pre-shared key (64 hex chars, 32 bytes) |
| `listen_port` | `int` | `0` (OS-assigned) | Peer connection port |
| `max_peers` | `int` | `200` | Maximum concurrent direct peers |
| `identity_path` | `string` | `""` (ephemeral) | Path to persist Ed25519 + Noise identity |
| `lan_discovery` | `bool` | `true` | LAN multicast peer discovery |
| `trackers` | `[]string` | Moss defaults | BitTorrent tracker announce URLs |
| `static_peers` | `[]string` | `[]` | Always-connect peer addresses |
| `announce_interval_sec` | `int` | `120` | Tracker re-announce interval |
| `bootstrap_timeout_sec` | `int` | `5` | Initial bootstrap timeout |
| `relay.enabled` | `bool` | `true` | Volunteer as relay / supernode |
| `relay.max_bandwidth_kbps` | `int` | `1024` | Relay throughput cap |
| `relay.max_sessions` | `int` | `100` | Max concurrent relay sessions |
| `relay.session_ttl_sec` | `int` | `1800` | Inactive session TTL |
| `relay.min_uptime_sec` | `int` | `60` | Grace period before supernode promotion |
| `nat.upnp` | `bool` | `true` | UPnP IGD port mapping |
| `nat.natpmp` | `bool` | `true` | NAT-PMP port mapping |
| `nat.pcp` | `bool` | `true` | PCP port mapping |
| `nat.hole_punch_attempts` | `int` | `3` | UDP hole-punch retries |
| `transport.high_throughput` | `bool` | `true` | Large buffers for relay traffic |
| `monitor.enabled` | `bool` | `true` | Enable monitoring HTTP endpoint |
| `monitor.listen` | `string` | `:9800` | Monitoring endpoint address |
| `identity_path` | `string` | `""` | Identity persistence file |
| `verbose` | `bool` | `false` | Debug-level logging |

---

## Monitoring API

MossSpore exposes a lightweight HTTP endpoint for health checks and operational
insight. Enabled by default on port `9800`.

### `GET /health`

Liveness check. Returns the detected NAT type.

```json
{
  "status": "ok",
  "nat_type": "port_restricted_cone"
}
```

### `GET /info`

Full mesh node state — peers, relay sessions, subscription channels, NAT profile,
public key fingerprint.

```json
{
  "mesh_id": "my-mesh",
  "listen_port": 47001,
  "advertised_addr": "203.0.113.42:47001",
  "nat_type": "public",
  "supernode_ready": true,
  "peer_count": 14,
  "relay_session_count": 3,
  "relay_route_count": 3,
  "public_key": "a1b2c3d4e5f6..."
}
```

### `GET /version`

Build metadata for inventory and debugging.

```json
{
  "version": "0.1.0-dev",
  "commit": "a1b2c3d4",
  "build_date": "2026-05-26T18:00:00Z",
  "go_version": "go1.26.1",
  "os": "linux",
  "arch": "amd64"
}
```

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  MossSpore                      │
│  ┌───────────────────────────────────────────┐  │
│  │            Moss Node (moss)               │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐  │  │
│  │  │  P2P     │ │  Relay   │ │ GossipSub│  │  │
│  │  │  Mesh    │ │  Engine  │ │ Pub/Sub  │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘  │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐  │  │
│  │  │  NAT     │ │  Noise   │ │ Tracker  │  │  │
│  │  │  Travs.  │ │  Crypt   │ │ Discovery│  │  │
│  │  └──────────┘ └──────────┘ └──────────┘  │  │
│  └───────────────────────────────────────────┘  │
│                                                   │
│  ┌───────────────────────────────────────────┐  │
│  │         Monitor (HTTP :9800)              │  │
│  │   /health    /info    /version            │  │
│  └───────────────────────────────────────────┘  │
│                                                   │
│  ┌───────────────────────────────────────────┐  │
│  │         Event Logger (structured JSON)    │  │
│  │   PeerJoined  SupernodePromoted  Tracker  │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### Layers

| Layer | Responsibility |
|---|---|
| **Moss Node** | Core P2P runtime — encrypted sessions, NAT traversal, gossip protocol, relay sessions |
| **Monitor** | HTTP health endpoint — `GET /health`, `/info`, `/version` |
| **Event Logger** | Structured JSON logging of lifecycle events |

### Events

MossSpore logs all significant network events as structured JSON, one per line:

```
{"type":"PeerJoined",       "code":1, "detail":"{\"peer\":\"a1b2...\",\"addr\":\"10.0.0.1:47001\"}"}
{"type":"SupernodePromoted","code":3, "detail":"{\"nat_type\":\"public\"}"}
{"type":"TrackerAnnounce",  "code":5, "detail":"{\"candidate_peers\":4,\"connected_peers\":1}"}
```

Consume these with `jq`, `journald`, or any log aggregator for real-time
visibility into mesh health.

---

## Identity Persistence

Each node needs a cryptographic identity (Ed25519 + Noise static keypair).
MossSpore stores it in the file specified by `identity_path`:

```json
{
  "identity_path": "/var/lib/mossspore/identity.key"
}
```

If the file exists, the identity is loaded. Otherwise a new one is generated
and written. Without `identity_path`, a new ephemeral identity is created on
every start — fine for testing, but running a public spore should use a
persistent identity so peers recognise you across restarts.

The identity file is 97 bytes: version byte (1) + ed25519 private key (64) +
Noise DH private key (32). Keep it secret, keep it safe.

---

## Running as a Service

The [install.sh](install.sh) script handles all of this automatically on Linux:

```bash
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
```

### Manual systemd (Linux)

```ini
[Unit]
Description=MossSpore P2P Relay Daemon
Documentation=https://github.com/redstone-md/MossSpore
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mossspore --config /etc/mossspore/config.json
Restart=always
RestartSec=30
User=mossspore
Group=mossspore
StateDirectory=mossspore
MemoryMax=256M

[Install]
WantedBy=multi-user.target
```

### Windows (NSSM)

```batch
nssm install MossSpore "C:\Program Files\MossSpore\mossspore.exe" "--config C:\ProgramData\MossSpore\config.json"
nssm set MossSpore AppStdout C:\ProgramData\MossSpore\stdout.log
nssm set MossSpore AppStderr C:\ProgramData\MossSpore\stderr.log
nssm set MossSpore AppRestartDelay 30000
nssm start MossSpore
```

### Docker

```dockerfile
FROM alpine:latest
COPY mossspore /usr/local/bin/
COPY config.json /etc/mossspore/
EXPOSE 9800
ENTRYPOINT ["mossspore", "--config", "/etc/mossspore/config.json"]
```

```bash
docker build -t mossspore .
docker run -d --restart=always -p 9800:9800 mossspore
```

---

## Project Map

```
cmd/mossspore/           CLI entry point — flags, config loading, lifecycle
internal/spore/          Core daemon — node management, event handling, monitor
internal/version/        Build version injection
.github/workflows/       CI + Release GitHub Actions workflows
install.sh               One-line installer (curl pipe)
mossspore.example.json   Documented configuration template
Makefile                 Build targets for Linux / Windows / macOS
```

---

## Relationship to Moss

```
┌─────────────────────────────────────────────────────┐
│                   Ecosystem                          │
│                                                       │
│  ┌──────────────────┐   ┌────────────────────────┐  │
│  │   Moss (core)     │   │   MossSpore (daemon)   │  │
│  │   github.com/     │   │   github.com/         │  │
│  │   redstone-md/moss│   │   redstone-md/MossSpore│  │
│  │                   │   │                        │  │
│  │  • P2P protocol   │   │  • Standalone binary   │  │
│  │  • Mesh routing   │   │  • Relay-optimised     │  │
│  │  • NAT traversal  │   │  • Zero-config headless│  │
│  │  • C FFI library  │   │  • Monitor endpoint    │  │
│  │  • Go API (moss)  │   │  • Uses moss Go API    │  │
│  └──────────────────┘   └────────────────────────┘  │
│                                                       │
│  ┌──────────────────┐                                │
│  │   MOSH (client)   │                                │
│  │   github.com/     │                                │
│  │   redstone-md/mosh│                                │
│  │                   │                                │
│  │  • Desktop chat   │                                │
│  │  • FFI consumer   │                                │
│  └──────────────────┘                                │
└─────────────────────────────────────────────────────┘
```

- **Moss** — the core runtime, protocol, and C-shared library
- **MossSpore** — a standalone relay daemon built on top of the Moss Go API
- **MOSH** — a desktop chat client consuming Moss via the C FFI

All three are MIT-licensed open source.

---

## Development

```bash
git clone https://github.com/redstone-md/MossSpore
cd MossSpore

go test ./...
go vet ./...

make build
./build/mossspore --version
```

PRs and issues welcome. See the contribution guide in each repository for details.

---

## License

MIT — see [LICENSE](LICENSE) for the full text.

---

<p align="center">
  <strong>Run a Spore, grow the Moss.</strong><br>
  <sub>Decentralised infrastructure, one spore at a time.</sub>
</p>
