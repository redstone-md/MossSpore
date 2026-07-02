package spore

import (
	"encoding/json"
	"testing"
)

// The monitor /info handler passes node.MeshInfoJSON() through verbatim, so the
// relay counters the pool dashboard needs must be present in that JSON. This
// locks the three fields S3 promised; relay_bytes_total is intentionally absent
// (moss exposes no byte counter yet — deferred).
func TestInfoJSONCarriesRelayCounters(t *testing.T) {
	// A minimal MeshInfo JSON as moss emits it.
	sample := `{"relay_session_count":0,"relay_route_count":0,"supernode_ready":false}`
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(sample), &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"relay_session_count", "relay_route_count", "supernode_ready"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing relay counter %q in /info payload", k)
		}
	}
	if _, ok := m["relay_bytes_total"]; ok {
		t.Error("relay_bytes_total unexpectedly present — deferral note is stale")
	}
}
