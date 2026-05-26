## MossSpore v0.1.4

**Run a Spore, grow the Moss.**

### What's New

- **NAT detection finally works on bare metal** — After three rounds of fixes in Moss, the root cause was `appendObservation` deduplicating identical STUN results, preventing `WithBindingObservations` from ever seeing 2+ entries. Final fix: fast-path in `applyExternalObservation` upgrades `TypeUnknown` → `TypePublic` immediately when the external address is a global unicast, without waiting for a second observation. Servers with all ports open now detect as `public` within seconds and promote to supernode after min uptime.

- **Auto-update mechanism** — New config option `auto_update.enabled`. When enabled, the spore periodically checks GitHub releases, downloads the new binary, verifies SHA256 checksum, creates a backup, applies the update atomically with sentinel-based crash recovery. On Unix, re-execs into the new binary in-place via `syscall.Exec`. If the new binary fails to start, the sentinel triggers a rollback to the previous version on next boot.

  ```json
  {
    "auto_update": {
      "enabled": true,
      "interval": "24h"
    }
  }
  ```

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
