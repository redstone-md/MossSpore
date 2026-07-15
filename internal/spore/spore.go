package spore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/redstone-md/moss"

	"github.com/moss/mossspore/internal/update"
	"github.com/moss/mossspore/internal/version"
)

// Event type labels for human-readable logging.
var eventLabels = map[int32]string{
	1: "PeerJoined",
	2: "PeerLeft",
	3: "SupernodePromoted",
	4: "SupernodeRevoked",
	5: "TrackerAnnounce",
	6: "TrackerFailure",
	7: "RelayMigrated",
}

// Spore is the core daemon that wraps a Moss node with lifecycle management,
// event logging, and the monitoring endpoint.
type Spore struct {
	cfg     Config
	node    *moss.Node
	monitor *Monitor
	updater *update.Updater
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopped chan struct{}
}

// New creates a new Spore daemon from the given configuration.
func New(cfg Config) (*Spore, error) {
	dataDir := defaultDataDir()
	s := &Spore{
		cfg:     cfg,
		stopped: make(chan struct{}),
	}
	if cfg.AutoUpdate.Enabled {
		s.updater = update.NewUpdater(version.Version, dataDir)
	}
	return s, nil
}

// Start initialises the Moss node and begins operation.
func (s *Spore) Start() error {
	if err := s.resolveUpdate(); err != nil {
		log.Printf("[spore] update cleanup: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.node != nil {
		return fmt.Errorf("spore: already started")
	}

	mossCfg := s.toMossConfig()
	psk, err := s.parsePSK()
	if err != nil {
		return fmt.Errorf("spore: %w", err)
	}

	node, err := moss.NewNode(s.cfg.MeshID, psk, mossCfg)
	if err != nil {
		return fmt.Errorf("spore: node init failed: %w", err)
	}

	node.SetEventCallback(s.onEvent)

	if err := node.Start(); err != nil {
		return fmt.Errorf("spore: node start failed: %w", err)
	}

	s.node = node

	pubKey := node.PublicKey()
	room := s.cfg.MeshID
	if room == "" {
		room = "(relay-only)"
	}
	log.Printf("[spore] node started — room=%s key=%s nat=%s port=%d",
		room,
		hex.EncodeToString(pubKey[:8]),
		node.NATType(),
		node.ListenPort(),
	)

	if s.cfg.Monitor.Enabled {
		m := NewMonitor(node, s.cfg.Monitor.Listen)
		ln, err := net.Listen("tcp", s.cfg.Monitor.Listen)
		if err != nil {
			return fmt.Errorf("spore: monitor listen: %w", err)
		}
		s.monitor = m
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			m.ServeOn(ln)
		}()
		log.Printf("[spore] monitor listening on %s", ln.Addr().String())
	}

	if s.updater != nil {
		s.startUpdateLoop()
	}

	return nil
}

// Stop gracefully shuts down the spore daemon.
func (s *Spore) Stop() error {
	s.mu.Lock()
	node := s.node
	mon := s.monitor
	s.node = nil
	s.monitor = nil
	s.mu.Unlock()

	if mon != nil {
		if err := mon.Close(); err != nil {
			log.Printf("[spore] monitor close: %v", err)
		}
	}
	if node != nil {
		if err := node.Stop(); err != nil {
			return fmt.Errorf("spore: node stop: %w", err)
		}
		log.Printf("[spore] node stopped")
	}
	close(s.stopped)
	return nil
}

// Wait blocks until the spore is stopped (e.g. by SIGINT/SIGTERM).
func (s *Spore) Wait() {
	<-s.stopped
	s.wg.Wait()
}

// WaitSignal blocks until a termination signal is received, then calls Stop.
func (s *Spore) WaitSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	sigReceived := <-sig
	log.Printf("[spore] signal %v received, shutting down", sigReceived)
	_ = s.Stop()
}

// Node returns the underlying Moss node. May be nil before Start.
func (s *Spore) Node() *moss.Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.node
}

func (s *Spore) onEvent(eventType int32, detailJSON string) {
	label, ok := eventLabels[eventType]
	if !ok {
		label = fmt.Sprintf("Event%d", eventType)
	}
	detail := detailJSON
	if detail == "" {
		detail = "{}"
	}

	// Structured log output: event type + detail.
	// Machine-parseable: each event is a single JSON line.
	type eventLog struct {
		Type   string `json:"type"`
		Code   int32  `json:"code"`
		Detail string `json:"detail"`
	}
	raw, _ := json.Marshal(eventLog{Type: label, Code: eventType, Detail: detail})
	log.Printf("[event] %s", string(raw))
}

func (s *Spore) toMossConfig() moss.Config {
	var cfg moss.Config
	if s.cfg.ListenPort > 0 {
		cfg.ListenPort = s.cfg.ListenPort
	}
	if s.cfg.MaxPeers > 0 {
		cfg.MaxPeers = s.cfg.MaxPeers
	}
	if len(s.cfg.Trackers) > 0 {
		cfg.Trackers = s.cfg.Trackers
	}
	if len(s.cfg.StaticPeers) > 0 {
		cfg.StaticPeers = s.cfg.StaticPeers
	}
	if s.cfg.AnnounceIntervalSec > 0 {
		cfg.AnnounceIntervalSec = s.cfg.AnnounceIntervalSec
	}
	if s.cfg.BootstrapTimeoutSec > 0 {
		cfg.BootstrapTimeoutSec = s.cfg.BootstrapTimeoutSec
	}

	lan := s.cfg.LANDiscovery
	cfg.LANDiscoveryEnabled = &lan

	if s.cfg.Relay.Enabled {
		if s.cfg.Relay.MaxBandwidthKBPS > 0 {
			cfg.RelayMaxBandwidthKBPS = s.cfg.Relay.MaxBandwidthKBPS
		}
		if s.cfg.Relay.MaxSessions > 0 {
			cfg.RelayMaxSessions = s.cfg.Relay.MaxSessions
		}
		if s.cfg.Relay.SessionTTLSec > 0 {
			cfg.RelaySessionTTLSec = s.cfg.Relay.SessionTTLSec
		}
		if s.cfg.Relay.MinUptimeSec > 0 {
			cfg.SuperNodeMinUptimeSec = s.cfg.Relay.MinUptimeSec
		}
	}

	upnp := s.cfg.NAT.UPnP
	natpmp := s.cfg.NAT.NATPMP
	pcp := s.cfg.NAT.PCP
	cfg.UPnPEnabled = &upnp
	cfg.NATPMPEnabled = &natpmp
	cfg.PCPEnabled = &pcp

	if s.cfg.NAT.HolePunchAttempts > 0 {
		cfg.HolePunchAttempts = s.cfg.NAT.HolePunchAttempts
	}

	ht := s.cfg.Transport.HighThroughput
	cfg.HighThroughput = &ht

	if s.cfg.Telemetry.Enabled {
		en := true
		cfg.TelemetryEnabled = &en
		if s.cfg.Telemetry.EpochSec > 0 {
			cfg.TelemetryEpochSec = s.cfg.Telemetry.EpochSec
		}
		if s.cfg.Telemetry.KAnon > 0 {
			cfg.TelemetryKAnon = s.cfg.Telemetry.KAnon
		}
	}

	cfg.IdentityPath = s.cfg.IdentityPath

	return cfg
}

// resolveUpdate handles update sentinel cleanup/rollback at startup.
func (s *Spore) resolveUpdate() error {
	if s.updater == nil {
		return nil
	}
	return s.updater.CleanupOnStart()
}

// startUpdateLoop begins the periodic update check in a background goroutine.
func (s *Spore) startUpdateLoop() {
	interval := 24 * time.Hour
	if s.cfg.AutoUpdate.Interval != "" {
		if d, err := time.ParseDuration(s.cfg.AutoUpdate.Interval); err == nil && d > 0 {
			interval = d
		}
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Check immediately on start, then on ticker.
		s.checkUpdateNow()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkUpdateNow()
			case <-s.stopped:
				return
			}
		}
	}()
}

func (s *Spore) checkUpdateNow() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := s.updater.DoFullUpdate(ctx); err != nil {
		log.Printf("[update] %v", err)
	}
}

// defaultDataDir returns the directory for update state and other spore data.
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "mossspore")
	}
	dir := filepath.Join(home, ".local", "share", "mossspore")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

func (s *Spore) parsePSK() ([]byte, error) {
	if s.cfg.PSK == "" {
		return nil, nil
	}
	raw, err := hex.DecodeString(s.cfg.PSK)
	if err != nil {
		return nil, fmt.Errorf("psk is not valid hex: %w", err)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("psk must be 32 bytes (64 hex chars), got %d", len(raw))
	}
	return raw, nil
}
