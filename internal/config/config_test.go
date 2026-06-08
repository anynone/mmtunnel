package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClientDefaultsStripPath(t *testing.T) {
	path := writeTemp(t, `
client:
  id: dev
  token: token
  server: ws://127.0.0.1:8080/_tunnel/connect

tunnels:
  - name: api
    publicPath: /api
    target: http://127.0.0.1:3000
`)
	cfg, err := LoadClient(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Tunnels[0].StripPath {
		t.Fatal("stripPath should default to true")
	}
}

func TestLoadClientRejectsReservedPath(t *testing.T) {
	path := writeTemp(t, `
client:
  id: dev
  token: token
  server: ws://127.0.0.1:8080/_tunnel/connect

tunnels:
  - name: bad
    publicPath: /_tunnel/app
    target: http://127.0.0.1:3000
`)
	if _, err := LoadClient(path); err == nil {
		t.Fatal("expected reserved path validation error")
	}
}

func TestLoadServerRejectsInvalidLogLevel(t *testing.T) {
	path := writeTemp(t, `
server:
  logLevel: noisy

clients:
  - id: dev
    token: token
`)
	if _, err := LoadServer(path); err == nil {
		t.Fatal("expected invalid log level")
	}
}

func writeTemp(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
