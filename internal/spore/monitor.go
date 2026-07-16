package spore

import (
	"context"
	_ "embed"
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/redstone-md/moss"

	"github.com/moss/mossspore/internal/version"
)

// buildInfo caches the module + VCS build metadata so the dashboard can show
// exactly which moss version and commit are running — the release ldflags are
// not set on a from-source build (e.g. Flux), so version.Version is unreliable,
// but the build info embedded by the Go toolchain always is.
var mossModuleVersion, buildRevision = func() (string, string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	var mossVer, rev string
	for _, d := range bi.Deps {
		if d.Path == "github.com/redstone-md/moss" {
			mossVer = d.Version
			break
		}
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			rev = s.Value
			if len(rev) > 12 {
				rev = rev[:12]
			}
		}
	}
	return mossVer, rev
}()

// dashboardHTML is the self-contained status page served at "/". It polls
// /api/summary client-side; nothing sensitive is inlined.
//
//go:embed web/dashboard.html
var dashboardHTML string

// Monitor provides an HTTP health-check and metrics endpoint for a Moss node.
type Monitor struct {
	node      *moss.Node
	srv       *http.Server
	startedAt time.Time
}

// NewMonitor creates a Monitor bound to the given listen address.
func NewMonitor(node *moss.Node, addr string) *Monitor {
	m := &Monitor{node: node, startedAt: time.Now()}
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handleRoot)
	mux.HandleFunc("/api/summary", m.handleSummary)
	mux.HandleFunc("/health", m.handleHealth)
	mux.HandleFunc("/info", m.handleInfo)
	mux.HandleFunc("/version", m.handleVersion)
	m.srv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return m
}

// ListenAndServe starts the HTTP listener and begins serving.
func (m *Monitor) ListenAndServe() error {
	ln, err := net.Listen("tcp", m.srv.Addr)
	if err != nil {
		return err
	}
	return m.srv.Serve(ln)
}

// ServeOn starts the server on the given listener and returns immediately.
// The caller must WaitGroup or otherwise coordinate lifecycle.
func (m *Monitor) ServeOn(ln net.Listener) {
	_ = m.srv.Serve(ln)
}

// Close stops the HTTP server.
func (m *Monitor) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return m.srv.Shutdown(ctx)
}

// handleRoot serves the status dashboard at exactly "/" and 404s everything
// else (the "/" mux pattern is a catch-all).
func (m *Monitor) handleRoot(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

// handleSummary returns a sanitized view for the public dashboard: aggregate
// counts and health only. It deliberately DROPS peer/known-peer addresses,
// peer_details, advertised_addr and the full public key — the monitor port is
// mapped to a public URL on Flux, and a peer IP list would deanonymize the
// mesh. node_id is a short prefix, enough to identify without exposing the key.
func (m *Monitor) handleSummary(w http.ResponseWriter, r *http.Request) {
	node := m.node
	if node == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "node not initialised"})
		return
	}
	_ = r.Body.Close()
	var info map[string]any
	_ = json.Unmarshal([]byte(node.MeshInfoJSON()), &info)
	out := map[string]any{
		"node_id":       shortID(info["public_key"]),
		"mesh_id":       info["mesh_id"],
		"nat_type":      info["nat_type"],
		"supernode_ready":          info["supernode_ready"],
		"peer_count":               info["peer_count"],
		"direct_peer_count":        info["direct_peer_count"],
		"relayed_peer_count":       info["relayed_peer_count"],
		"known_peer_count":         info["known_peer_count"],
		"relay_capable_peer_count": info["relay_capable_peer_count"],
		"relay_session_count":      info["relay_session_count"],
		"relay_route_count":        info["relay_route_count"],
		"channel_count":            arrayLen(info["channels"]),
		"uptime_sec":               int(time.Since(m.startedAt).Seconds()),
		"version":                  version.Version,
		"commit":                   version.Commit,
		"moss_version":             mossModuleVersion,
		"build_rev":                buildRevision,
	}
	writeJSON(w, http.StatusOK, out)
}

// shortID returns the first 16 hex chars of a public key for display.
func shortID(v any) string {
	s, _ := v.(string)
	if len(s) > 16 {
		return s[:16]
	}
	return s
}

// arrayLen reports the length of a JSON array decoded into any, else 0.
func arrayLen(v any) int {
	if a, ok := v.([]any); ok {
		return len(a)
	}
	return 0
}

func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	node := m.node
	if node == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
			"error":  "node not initialised",
		})
		return
	}
	_ = r.Body.Close()
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"nat_type": node.NATType(),
	})
}

// handleInfo returns node.MeshInfoJSON() verbatim — it already carries
// relay_session_count, relay_route_count and supernode_ready for pool
// dashboards. relay_bytes_total is deferred: moss exposes no relay byte
// counter yet; add it here once moss's MeshInfo does.
func (m *Monitor) handleInfo(w http.ResponseWriter, r *http.Request) {
	node := m.node
	if node == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "node not initialised",
		})
		return
	}
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(node.MeshInfoJSON()))
}

func (m *Monitor) handleVersion(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	writeJSON(w, http.StatusOK, map[string]string{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.Date,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
