## MossSpore v0.1.1

**Run a Spore, grow the Moss.**

### What's New

- **Fixed NAT detection on bare metal** — `profiler.go` in Moss no longer downgrades `TypeUnknown` to `port_restricted_cone` when a public STUN address is obtained. Servers with all ports open now correctly get `PublicReachable`, enabling supernode promotion.

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
