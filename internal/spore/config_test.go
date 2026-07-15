package spore

import "testing"

func TestDefaultConfigIsRoomlessRelay(t *testing.T) {
	c := DefaultConfig()
	if c.MeshID != "" {
		t.Fatalf("default spore should be a roomless relay (empty mesh_id), got %q", c.MeshID)
	}
}

func TestNormalizePreservesMeshID(t *testing.T) {
	c := DefaultConfig()
	c.MeshID = "my-room"
	c.Normalize()
	if c.MeshID != "my-room" {
		t.Fatalf("Normalize must preserve an explicit mesh_id, got %q", c.MeshID)
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
