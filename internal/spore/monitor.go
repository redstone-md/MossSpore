package spore

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"time"

	"moss"

	"github.com/moss/mossspore/internal/version"
)

// Monitor provides an HTTP health-check and metrics endpoint for a Moss node.
type Monitor struct {
	node *moss.Node
	srv  *http.Server
}

// NewMonitor creates a Monitor bound to the given listen address.
func NewMonitor(node *moss.Node, addr string) *Monitor {
	m := &Monitor{node: node}
	mux := http.NewServeMux()
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
