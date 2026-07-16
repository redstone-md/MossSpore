package spore

import "path/filepath"

// RelayMeshID is retained for backward compatibility with old config files.
// Deprecated: relay-mesh mode no longer exists — every spore joins the one
// shared substrate and relays for every room regardless of mesh id.
const RelayMeshID = "moss-relay/1"

// Config defines the configuration for a MossSpore daemon instance.
type Config struct {
	// MeshID is the room this spore joins on the shared substrate. A spore
	// relays for EVERY room no matter what — the substrate is shared — so a
	// pure relay leaves this empty (the default). Set it only if you also want
	// the spore to be a pub/sub member of a specific room.
	MeshID string `json:"mesh_id,omitempty"`

	// PSK is an optional pre-shared key hex-encoded string (32 bytes).
	// If set, all handshakes require knowledge of this key.
	PSK string `json:"psk,omitempty"`

	// ListenPort is the port for peer connections. 0 means OS-assigned.
	ListenPort int `json:"listen_port,omitempty"`

	// MaxPeers caps the number of concurrent direct peer connections.
	MaxPeers int `json:"max_peers,omitempty"`

	// Trackers lists BitTorrent tracker URLs for peer discovery.
	Trackers []string `json:"trackers,omitempty"`

	// StaticPeers lists peer addresses to always try connecting to.
	StaticPeers []string `json:"static_peers,omitempty"`

	// AnnounceIntervalSec controls how often the node announces to trackers.
	AnnounceIntervalSec int `json:"announce_interval_sec,omitempty"`

	// BootstrapTimeoutSec is the timeout for initial tracker responses.
	BootstrapTimeoutSec int `json:"bootstrap_timeout_sec,omitempty"`

	// LANDiscovery enables automatic peer discovery on the local network.
	LANDiscovery bool `json:"lan_discovery,omitempty"`

	// Relay configures the relay/supernode behaviour of this spore.
	Relay RelayConfig `json:"relay"`

	// NAT configures NAT traversal strategies.
	NAT NATConfig `json:"nat"`

	// Transport configures low-level buffer sizes.
	Transport TransportConfig `json:"transport"`

	// IdentityPath is the filesystem path for the persistent node identity.
	// If empty, Normalize defaults it to a path under the state dir so the
	// spore keeps a stable peer-id across restarts.
	IdentityPath string `json:"identity_path,omitempty"`

	// RelayMesh is accepted but ignored, kept so existing config files still
	// parse. Deprecated: the shared substrate makes relay-mesh mode obsolete —
	// a spore already relays for every room.
	RelayMesh RelayMeshConfig `json:"relay_mesh,omitempty"`

	// Monitor configures the built-in HTTP monitoring endpoint.
	Monitor MonitorConfig `json:"monitor"`

	// LogDir is the directory for log output. Empty means stdout.
	LogDir string `json:"log_dir,omitempty"`

	// Verbose enables detailed debug-level logging.
	Verbose bool `json:"verbose,omitempty"`

	// AutoUpdate enables automatic binary updates from GitHub releases.
	AutoUpdate AutoUpdateConfig `json:"auto_update"`

	// Telemetry opts this spore into the privacy-preserving network
	// observability layer that gateways aggregate. On by default for relay
	// infrastructure; the contribution is DP-noised and carries no address or
	// stable identity.
	Telemetry TelemetryConfig `json:"telemetry"`

	// Axiom ships this spore's errors and periodic node-stats to an Axiom
	// dataset for fleet-wide observability. Disabled unless a token is set.
	Axiom AxiomConfig `json:"axiom"`

	// Veil configures the DPI-resistant "Reality" bearer. A public relay
	// sets role="listener" to front the mesh as ordinary HTTPS so a client
	// behind DPI (e.g. RU TSPU) can join through it. Disabled by default.
	Veil VeilConfig `json:"veil"`
}

// VeilConfig configures the Veil "Reality" DPI-mask bearer on a spore.
type VeilConfig struct {
	// Enabled turns the bearer on. Off by default.
	Enabled bool `json:"enabled"`
	// Role is "listener" for a relay that fronts the mesh, or "dialer"
	// (default) for a node that only bootstraps through veil relays.
	Role string `json:"role,omitempty"`
	// ListenAddr is the listener bind (host:port), e.g. ":8443".
	ListenAddr string `json:"listen_addr,omitempty"`
	// CoverSNI is the domain the masked TLS impersonates. Must match on both
	// legs and SHOULD be a real site the listener can reach, so spliced
	// probes see a genuine certificate (e.g. "www.wikipedia.org").
	CoverSNI string `json:"cover_sni,omitempty"`
	// TargetAddr is the real origin unauthenticated probes are spliced to
	// (default CoverSNI:443).
	TargetAddr string `json:"target_addr,omitempty"`
	// Relays lists veil-fronted relays a dialer bootstraps through.
	Relays []VeilRelay `json:"relays,omitempty"`
}

// VeilRelay is one veil-fronted relay a spore dials to survive DPI.
type VeilRelay struct {
	// Addr is the relay's veil listener, host:port.
	Addr string `json:"addr"`
	// CoverSNI must match the relay's cover domain.
	CoverSNI string `json:"cover_sni"`
	// PubKeyHex is the relay's 32-byte static Noise public key (hex).
	PubKeyHex string `json:"pubkey"`
}

// AxiomConfig configures error/log shipping to Axiom.
type AxiomConfig struct {
	// Token is an ingest-only Axiom API token. Empty disables shipping.
	Token string `json:"token,omitempty"`
	// Dataset is the target Axiom dataset (e.g. "moss-events").
	Dataset string `json:"dataset,omitempty"`
	// Endpoint is the Axiom ingest base URL. Empty → cloud default; set the
	// region edge for non-default orgs (e.g. https://eu-central-1.aws.edge.axiom.co).
	Endpoint string `json:"endpoint,omitempty"`
	// Service overrides the host identifier attached to events (default
	// "mossspore").
	Service string `json:"service,omitempty"`
}

// TelemetryConfig controls the observability contribution.
type TelemetryConfig struct {
	// Enabled contributes DP-noised per-epoch aggregate metrics to the network.
	Enabled bool `json:"enabled"`

	// EpochSec is the telemetry epoch length. Default 300 (moss default).
	EpochSec int `json:"epoch_sec,omitempty"`

	// KAnon suppresses detailed metrics below this many contributors.
	KAnon int `json:"k_anon,omitempty"`
}

// AutoUpdateConfig controls the automatic update behaviour.
type AutoUpdateConfig struct {
	// Enabled periodically checks GitHub for new releases and applies them.
	Enabled bool `json:"enabled"`

	// Interval between update checks. Default "24h".
	Interval string `json:"interval,omitempty"`
}

// RelayConfig controls relay (supernode) behaviour.
type RelayConfig struct {
	// Enabled allows this spore to act as a relay for other peers.
	// When true, the node will volunteer as a supernode if it has a
	// public or full-cone NAT.
	Enabled bool `json:"enabled"`

	// MaxBandwidthKBPS limits relay throughput in kilobytes per second.
	MaxBandwidthKBPS int `json:"max_bandwidth_kbps,omitempty"`

	// MaxSessions caps the number of concurrent relay sessions.
	MaxSessions int `json:"max_sessions,omitempty"`

	// SessionTTLSec is the time-to-live for inactive relay sessions.
	SessionTTLSec int `json:"session_ttl_sec,omitempty"`

	// MinUptimeSec is the minimum uptime before being promoted to supernode.
	MinUptimeSec int `json:"min_uptime_sec,omitempty"`
}

// RelayMeshConfig enables shared-relay-mesh mode: the spore joins RelayMeshID
// and volunteers as a SuperNode for the whole pool. Default on.
type RelayMeshConfig struct {
	// Enabled points this spore at the shared relay mesh. When true it
	// overrides MeshID with RelayMeshID. Set false for a single-mesh spore.
	Enabled bool `json:"enabled"`
}

// NATConfig controls NAT traversal strategies.
type NATConfig struct {
	// UPnP enables UPnP IGD port mapping.
	UPnP bool `json:"upnp,omitempty"`

	// NATPMP enables NAT-PMP port mapping.
	NATPMP bool `json:"natpmp,omitempty"`

	// PCP enables PCP port mapping.
	PCP bool `json:"pcp,omitempty"`

	// HolePunchAttempts is the number of hole-punch attempts per peer.
	HolePunchAttempts int `json:"hole_punch_attempts,omitempty"`
}

// TransportConfig controls low-level networking buffers.
type TransportConfig struct {
	// HighThroughput enables larger buffers for high-bandwidth use cases.
	HighThroughput bool `json:"high_throughput,omitempty"`
}

// MonitorConfig controls the monitoring HTTP endpoint.
type MonitorConfig struct {
	// Enabled starts the monitoring HTTP server.
	Enabled bool `json:"enabled"`

	// Listen is the address:port for the monitoring endpoint.
	// Default ":9800".
	Listen string `json:"listen,omitempty"`
}

// DefaultConfig returns a sensible configuration for a relay-optimised spore.
func DefaultConfig() Config {
	return Config{
		// Empty room: a pure substrate relay that serves every room.
		MeshID: "",
		Relay: RelayConfig{
			Enabled:          true,
			MaxBandwidthKBPS: 1024,
			MaxSessions:      100,
			SessionTTLSec:    1800,
			MinUptimeSec:     60,
		},
		NAT: NATConfig{
			UPnP:              true,
			NATPMP:            true,
			PCP:               true,
			HolePunchAttempts: 3,
		},
		Transport: TransportConfig{
			HighThroughput: true,
		},
		Monitor: MonitorConfig{
			Enabled: true,
			Listen:  ":9800",
		},
		LANDiscovery:        true,
		AnnounceIntervalSec: 120,
		BootstrapTimeoutSec: 5,
		MaxPeers:            200,
		Verbose:             false,
		AutoUpdate: AutoUpdateConfig{
			Enabled:  false,
			Interval: "24h",
		},
		Telemetry: TelemetryConfig{
			Enabled: true,
		},
	}
}

// defaultIdentityPath returns the persistent identity file under the state dir,
// so a spore keeps a stable peer-id / SuperNode identity across restarts.
func defaultIdentityPath() string {
	return filepath.Join(defaultDataDir(), "identity.key")
}

// Normalize applies the persistent-identity default. Call after loading a
// config file and BEFORE applying explicit CLI flag overrides, so an explicit
// --mesh-id still wins.
func (c *Config) Normalize() {
	if c.IdentityPath == "" {
		c.IdentityPath = defaultIdentityPath()
	}
}
