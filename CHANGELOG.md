# Changelog

All notable changes to MossSpore are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
semantic versioning. Most releases track a moss runtime bump; the moss changelog
has the transport/protocol detail.

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
