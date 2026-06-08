package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActiveClientConfigFiltersDisabledTunnels(t *testing.T) {
	store := Store{
		ActiveProfileID: "company",
		Profiles: []Profile{{
			ID:   "company",
			Name: "Company",
			Client: ClientIdentity{
				ID:                "dev",
				Token:             "token",
				Server:            "wss://callback.example.com/_tunnel/connect",
				ReconnectInterval: "5s",
			},
			Tunnels: []Tunnel{
				{ID: "payment", Name: "Payment", PublicPath: "/payment", Target: "http://127.0.0.1:9098", StripPath: true, Enabled: true},
				{ID: "docs", Name: "Docs", PublicPath: "/docs", Target: "https://docs.example.com", StripPath: true, Enabled: false},
			},
		}},
	}

	cfg, statuses, err := store.ActiveClientConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tunnels) != 1 {
		t.Fatalf("runtime tunnels = %d, want 1", len(cfg.Tunnels))
	}
	if cfg.Tunnels[0].Name != "Payment" {
		t.Fatalf("runtime tunnel = %s", cfg.Tunnels[0].Name)
	}
	if statuses["payment"] != TunnelStatusPending || statuses["docs"] != TunnelStatusDisabled {
		t.Fatalf("unexpected statuses: %#v", statuses)
	}
}

func TestValidateRejectsEnabledPathConflict(t *testing.T) {
	store := Store{
		ActiveProfileID: "company",
		Profiles: []Profile{{
			ID:     "company",
			Client: ClientIdentity{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"},
			Tunnels: []Tunnel{
				{ID: "a", Name: "A", PublicPath: "/api", Target: "http://127.0.0.1:3000", Enabled: true},
				{ID: "b", Name: "B", PublicPath: "/api/v1", Target: "http://127.0.0.1:3001", Enabled: true},
			},
		}},
	}
	if err := store.Validate(); err == nil {
		t.Fatal("expected enabled tunnel conflict")
	}
}

func TestSaveLoadImportExportCLIYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.json")
	store := Store{
		ActiveProfileID: "company",
		Profiles: []Profile{{
			ID:      "company",
			Name:    "Company",
			Client:  ClientIdentity{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect", ReconnectInterval: "5s"},
			Tunnels: []Tunnel{{ID: "api", Name: "api", PublicPath: "/api", Target: "http://127.0.0.1:3000", StripPath: true, Enabled: true}},
		}},
	}
	if err := store.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ActiveProfileID != "company" || len(loaded.Profiles) != 1 {
		t.Fatalf("loaded store mismatch: %#v", loaded)
	}

	yamlPath := filepath.Join(dir, "client.yaml")
	if err := loaded.ExportCLIYAML("company", yamlPath); err != nil {
		t.Fatal(err)
	}
	imported, err := ImportCLIYAML("imported", "Imported", yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Client.ID != "dev" || len(imported.Tunnels) != 1 || !imported.Tunnels[0].Enabled {
		t.Fatalf("imported profile mismatch: %#v", imported)
	}
	if data, err := os.ReadFile(yamlPath); err != nil || len(data) == 0 {
		t.Fatalf("expected exported yaml: %v", err)
	}
}
