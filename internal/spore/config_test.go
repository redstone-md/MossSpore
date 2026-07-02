package spore

import "testing"

func TestNormalizeRelayMeshModeForcesMeshID(t *testing.T) {
	c := DefaultConfig() // RelayMesh.Enabled defaults true
	c.MeshID = "global"
	c.Normalize()
	if c.MeshID != RelayMeshID {
		t.Fatalf("relay-mesh mode should force mesh_id=%q, got %q", RelayMeshID, c.MeshID)
	}
}

func TestNormalizeSingleMeshOptOut(t *testing.T) {
	c := DefaultConfig()
	c.RelayMesh.Enabled = false
	c.MeshID = "my-mesh"
	c.Normalize()
	if c.MeshID != "my-mesh" {
		t.Fatalf("opt-out should keep operator mesh_id, got %q", c.MeshID)
	}
}

func TestNormalizeDefaultsPersistentIdentity(t *testing.T) {
	c := DefaultConfig()
	c.IdentityPath = ""
	c.Normalize()
	if c.IdentityPath == "" {
		t.Fatal("empty identity_path should default to a persistent state-dir path")
	}
}

func TestNormalizeKeepsExplicitIdentity(t *testing.T) {
	c := DefaultConfig()
	c.IdentityPath = "/custom/id.key"
	c.Normalize()
	if c.IdentityPath != "/custom/id.key" {
		t.Fatalf("explicit identity_path must be preserved, got %q", c.IdentityPath)
	}
}
