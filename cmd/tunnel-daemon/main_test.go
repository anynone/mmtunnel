package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultProfilePathUsesMMTunnelNamespace(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path := defaultProfilePath()
	wantSuffix := filepath.Join("mmtunnel", "profiles.json")
	if !strings.HasSuffix(filepath.ToSlash(path), filepath.ToSlash(wantSuffix)) {
		t.Fatalf("defaultProfilePath() = %q, want suffix %q", path, wantSuffix)
	}
}
