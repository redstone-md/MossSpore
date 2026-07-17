# Changelog

All notable changes to MossSpore are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
semantic versioning. Most releases track a moss runtime bump; the moss changelog
has the transport/protocol detail.

## [0.8.1] - 2026-07-17

### Changed
- **Bundled moss → v0.8.14: relays stopped drowning each other.** A spore with
  seven peers was taking 21,808 supernode announcements in two minutes — against
  29 pings — and silently discarding 142,125 packets a minute at 2% capacity.
  Disagreeing nodes forwarded their own unverifiable view of a peer back and
  forth forever, each correction going to every peer they had; the flood stalled
  the synchronous read loop, the stream buffer filled, and everything behind it
  was dropped without a trace. The pings among them are why sessions died at six
  misses with the connection healthy. Fixed in three layers: do not re-tell what
  we cannot vouch for (v0.8.8), survive a peer that floods us anyway (v0.8.9),
  and stop redialling a peer that connects and then goes silent (v0.8.10). Links
  between spores on this build show zero six-miss deaths and a median session
  life of 632s, against 37s before.
- moss v0.8.3-v0.8.5 are retracted and must not be bundled: they carry a session
  goodbye that tore down live links. See the moss changelog.

### Added
- Fleet telemetry a spore can be judged by: `stream_drops` (packets thrown away
  for want of buffer — an overflow hook had sat uninstalled in the tree, so these
  had never once been counted), `in_<envelope_type>` (what is arriving, which is
  how the flood got a name), `peer_capacity_pct` / `relay_capacity_pct` with
  their denominators (8 peers is half-idle at MaxPeers=16 and wedged at
  MaxPeers=8 — a spore at 90% warns), and `peer` on `session_end` (so both halves
  of one link can be joined; from inside one node "zero packets arrived" cannot
  be told from "they never sent").

## [0.7.2] - [0.8.0] - 2026-07-16 — 2026-07-17

These shipped without changelog entries; recorded here after the fact. Each is a
moss runtime bump, and the moss changelog carries the detail.

### Changed
- moss v0.7.2–v0.7.6: NAT classification stopped stealing a public node's road to
  reachable; the relay became a fallback rather than a terminus; the overlay
  learned to bootstrap itself, to name the path that opens every session, and to
  stop draining its own routing table.
- moss v0.7.7–v0.7.9: the overlay left the hot path — a Kademlia lookup per peer
  per tick was adding ~12s to every connect, which players felt as the stall; the
  node stopped redialling peers it already held (a storm of ~1.7 sessions/second,
  95% of them duplicates closed on arrival); and sessions began counting what
  actually arrives, since a UDP write succeeds locally whether or not anyone
  receives it.
- moss v0.8.0: duplicate sessions are decided by transport, the one fact both ends
  agree on. A bootstrap race leaves each side holding two sessions in the same
  direction, and diverging choices left the loser's half a ghost.

## [0.7.1] - 2026-07-16

### Changed
- Bundled moss → v0.7.1: **a node can finally classify its own NAT.** Fleet
  telemetry showed every event carrying `observations=1` and ~89% of connect
  attempts failing at ~10s each — the same bug. A node never had the two vantage
  points symmetric NAT requires (the gossip binding reply echoed back the
  asker's own port; STUN stopped at the first server), so `nat_type` stayed
  `unknown` forever and the relay-preference gate — which needs both ends known
  symmetric — never fired. Relay selection also now asks the overlay where a
  peer is attached instead of guessing which neighbour might reach it.

## [0.7.0] - 2026-07-16

### Changed
- **Bundled moss → v0.7.0: a spore is now the overlay's core.** A publicly
  reachable spore holds Kademlia buckets and answers the lookups that let two
  peers on a sparse channel find each other at all — the gap that made game
  lobbies invisible once the shared substrate diluted every topic among
  hundreds of unrelated peers. Leaves cannot serve this: a lookup can only be
  delivered to a node that can be dialed, so **the relay fleet has to be
  upgraded for clients to benefit**. Records are keyed by the opaque room topic,
  so a spore holding one still cannot tell which room or game it belongs to.
- moss v0.7.0 also ships per-attempt connectivity telemetry (`nat_attempt`,
  `topic_rendezvous`, `session_end`) — no addresses — and fixes the FFI dropping
  the Axiom config, which is why no desktop client had ever reported.

## [0.6.16] - 2026-07-16

### Fixed
- **Bundled moss → v0.6.24, which reverts a NAT hole-punch regression.** moss
  v0.6.22/v0.6.23 — shipped in MossSpore **v0.6.14 and v0.6.15** — dropped the
  UDP session a node keeps with its bootstrap peers, breaking direct P2P
  between NAT'd peers (every moss node answers an `observe` request over that
  session — a built-in STUN — and those observations drive NAT classification
  and hole-punch port prediction). **Anyone on v0.6.14/v0.6.15 should upgrade.**

### Added
- `/api/summary` and the dashboard now report the bundled **moss version** and
  **build revision** (from the Go build info, which is reliable even on a
  from-source build with no release ldflags). Makes it possible to confirm
  exactly which moss a deployment is running — e.g. whether a Flux build
  actually picked up the latest commit.

## [0.6.15] - 2026-07-16

### Changed
- Bundled moss → v0.6.23 (**symmetric-NAT flap, real fix**: bootstrap now dials
  a public supernode TCP-first instead of racing a UDP session against it, so a
  node behind symmetric NAT — e.g. a Flux container — forms one stable TCP link
  per supernode instead of being pruned every ~38s. v0.6.14's dialer-side
  preference alone was insufficient because the supernode latched onto the
  raced UDP session on its own).

## [0.6.14] - 2026-07-16

### Added
- **Status dashboard at `/`.** The monitor port now serves a self-contained
  (no external assets) status page that polls a new sanitized `/api/summary`
  endpoint every few seconds — peer/relay counts, NAT type, supernode state,
  uptime, version. When the port is mapped to a public URL (e.g. a Flux
  deployment), the summary deliberately omits peer/known-peer addresses,
  `peer_details`, `advertised_addr` and the full public key, so the mesh is
  not deanonymized; `node_id` is a short prefix. `/info` is unchanged for
  local tooling.

### Changed
- Bundled moss → v0.6.22 (**symmetric-NAT flap fix**: a node behind symmetric
  NAT — e.g. a container on Flux — no longer drops its supernodes every ~38s;
  duplicate-session dedup now keeps the stable TCP link instead of letting a
  dead-return-path UDP session replace it).

## [0.6.13] - 2026-07-16

### Changed
- **Axiom shipping is now on by default**, with a baked ingest-only fleet
  token. A spore reports to the shared `moss-events` dataset even when it runs
  on `DefaultConfig` with no config file — the case for a Flux build that just
  executes the binary, which previously shipped nothing and was invisible.
  An explicit `[axiom]` block still overrides the token/dataset/endpoint/
  service; opt out entirely with `"axiom": {"disabled": true}`. The token is
  write-only and already present in the desktop clients, so this adds no
  exposure.

## [0.6.12] - 2026-07-16

### Added
- **Veil DPI-mask bearer configurable on a spore** (`[veil]` block). A public
  relay sets `role="listener"` with a `cover_sni` to front the mesh as
  ordinary HTTPS; a spore behind DPI lists veil-fronted relays under
  `relays` (`{addr, cover_sni, pubkey}`) to bootstrap through the masked
  tunnel when its plain UDP/TCP paths are throttled (e.g. RU TSPU). Off by
  default.

### Changed
- Bundled moss → v0.6.21 (Veil dialer integration: the mesh now bootstraps
  through the DPI-mask bearer, not just listens on it).

## [0.6.11] - 2026-07-15

### Added
- **Ships errors and node-stats to Axiom.** New `[axiom]` config block (`token`,
  `dataset`, `endpoint`, `service`) mapped into moss, so each spore reports its
  own failures and periodic peer/supernode/relay counts to a central dataset for
  fleet observability. Disabled unless a token is set.

### Changed
- Bundled moss → v0.6.18 (public Axiom config + `node_stats` emitter; telemetry
  on by default).

## [0.6.6] - [0.6.9] - 2026-07-15

### Changed
- Rolling moss runtime bumps across the shared-substrate reliability line:
  - v0.6.9 → moss v0.6.15 (explicit peer targets, neutral dial ranking).
  - v0.6.8 → moss v0.6.13 (relay-route leak fix).
  - v0.6.7 → moss v0.6.12 (supernode connection-glare fix).
  - v0.6.6 → moss v0.6.11 (client-flapping fix, `peer_details`).

## [0.6.0] - 2026-07-15

### BREAKING
- **Shared-substrate migration.** A spore now supports the whole mesh network
  regardless of `mesh_id`: `mesh_id` became a room (pub/sub namespace) layered on
  a shared substrate, so a spore relays and discovers for every room, not one.
  The default config is roomless. `relay_mesh` is deprecated (accepted, ignored).
  Wire-incompatible with the pre-substrate line — deploy the fleet together.

### Added
- Network-wide telemetry aggregation and geo-region mapping surfaced by the
  gateway; embedded binary; Flux-Orbit-friendly single-repo build.
