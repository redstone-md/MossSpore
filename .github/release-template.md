## MossSpore v0.3.2

**Run a Spore, grow the Moss.**

### Fixed

- **Clients that join after promotion now see the spore as a relay.** A SuperNode
  advertised its relay capability only once, at promotion — so a client that
  connected afterwards (the normal case) never learned it could relay through the
  spore, and relay selection found nothing. The spore now periodically
  re-announces its SuperNode status, so every connected peer converges (~10s)
  regardless of when it joined. Bundled Moss core bumped to `f3bb2fb`.

- **NAT detection on IPv4-only hosts.** The bundled Moss core now resolves STUN
  and peer addresses as IPv4 to match its IPv4-only transport. Previously, on a
  host with no IPv6 route (common on VPS/cloud), STUN could resolve to an IPv6
  address and stall — leaving a genuinely public spore stuck at
  `nat_type: unknown`, never promoting to SuperNode. Such spores now detect
  `public` and promote correctly once a peer connects.

### Relay mode (since v0.3.0)

- **Relay-mesh mode, on by default.** A spore now joins the shared relay mesh
  `moss-relay/1` out of the box and, once it detects a public/full-cone NAT and
  meets the uptime bar, promotes itself to a SuperNode that relays direct
  messages for Mosh peers stuck behind hard/carrier-grade NAT. The relay only
  ever forwards ciphertext — it cannot read the messages it carries. Opt out
  with `"relay_mesh": { "enabled": false }` to run a private mesh instead.

- **Stable identity by default.** The node key is persisted under the state
  directory (`/var/lib/mossspore` via systemd `StateDirectory`, or the platform
  data dir), so a spore keeps the same peer-id across restarts and upgrades
  instead of re-rolling on every boot.

- **Relay observability on `/info`.** The status endpoint now reports
  `relay_session_count`, `relay_route_count`, and `supernode_ready` so you can
  confirm a spore is actually relaying.

- **One-command deploy.**
  - `install.sh` writes a ready-to-run relay-mesh config, provisions the state
    directory, and prints a reachability notice (open UDP/TCP `4001`).
  - **Docker**: `docker build -f MossSpore/Dockerfile ..` (parent context so the
    bundled `../moss` resolves); baked relay-mesh config, identity on a volume.
  - **Fly.io**: `fly.toml` with UDP+TCP `4001`, a persistent volume, and a
    `/health` check.
  - New guide: `docs/running-a-spore.md`.

### Binaries

| Platform | Arch | File |
|---|---|---|
| Linux | amd64 | `mossspore-linux-amd64` |
| Linux | arm64 | `mossspore-linux-arm64` |
| macOS | Intel | `mossspore-darwin-amd64` |
| macOS | Apple Silicon | `mossspore-darwin-arm64` |
| Windows | amd64 | `mossspore-windows-amd64.exe` |

### Checksums

SHA-256 checksums for all binaries are in `checksums.txt`.

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
```

### Usage

See the [README](https://github.com/redstone-md/MossSpore) for configuration and deployment documentation.
