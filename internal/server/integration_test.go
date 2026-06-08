package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tunnelclient "mmsocket/internal/client"
	"mmsocket/internal/config"
	"mmsocket/internal/protocol"
	tunnelserver "mmsocket/internal/server"
)

func TestHTTPForwardingAndPathStripping(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/users?id=1" {
			t.Fatalf("target path = %s", r.URL.String())
		}
		if r.Header.Get("X-Tunnel-Client-Id") != "dev" {
			t.Fatalf("missing forwarded client id")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	server := newTestServer(t, target.URL, true)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/users?id=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated || string(body) != "ok" {
		t.Fatalf("status/body = %d %q", resp.StatusCode, body)
	}
}

func TestStatusEndpointReadOnly(t *testing.T) {
	s := httptest.NewServer(tunnelserver.New(config.ServerConfig{Listen: ":0", Clients: []config.TrustedClient{{ID: "dev", Token: "token"}}, RequestTimeout: time.Second, MaxRequestBodySize: 1024, PingInterval: time.Second, PongTimeout: time.Second}, slog.Default()))
	defer s.Close()
	resp, err := http.Post(s.URL+"/_tunnel/status", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestKeepaliveDoesNotDisconnectHealthyClient(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	cfg := config.ServerConfig{
		Listen:             ":0",
		RequestTimeout:     time.Second,
		MaxRequestBodySize: 1024,
		PingInterval:       20 * time.Millisecond,
		PongTimeout:        20 * time.Millisecond,
		Clients:            []config.TrustedClient{{ID: "dev", Token: "token"}},
	}
	srv := httptest.NewServer(tunnelserver.New(cfg, slog.Default()))
	defer srv.Close()

	clientCfg := config.ClientConfig{
		ID:                "dev",
		Token:             "token",
		Server:            "ws" + strings.TrimPrefix(srv.URL, "http") + "/_tunnel/connect",
		ReconnectInterval: time.Second,
		Tunnels: []config.TunnelConfig{{
			Name:       "api",
			PublicPath: "/api",
			Target:     target.URL,
			StripPath:  true,
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = tunnelclient.New(clientCfg, slog.Default()).Run(ctx)
	}()

	waitForStatus(t, srv.URL)
	time.Sleep(120 * time.Millisecond)
	waitForStatus(t, srv.URL)
}

func TestWebSocketForwarding(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := protocol.Upgrade(w, r)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer conn.Close()
			op, body, err := conn.ReadMessage()
			if err == nil {
				_ = conn.WriteMessage(op, append([]byte("echo:"), body...))
			}
		}()
	}))
	defer target.Close()

	server := newTestServer(t, target.URL, true)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/socket"
	conn, _, err := protocol.Dial(wsURL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteMessage(protocol.OpcodeText, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	_, body, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "echo:hello" {
		t.Fatalf("body = %q", body)
	}
}

func newTestServer(t *testing.T, target string, stripPath bool) *httptest.Server {
	t.Helper()
	cfg := config.ServerConfig{
		Listen:             ":0",
		RequestTimeout:     5 * time.Second,
		MaxRequestBodySize: 1024 * 1024,
		PingInterval:       time.Second,
		PongTimeout:        time.Second,
		Clients:            []config.TrustedClient{{ID: "dev", Token: "token"}},
	}
	srv := httptest.NewServer(tunnelserver.New(cfg, slog.Default()))
	clientCfg := config.ClientConfig{
		ID:                "dev",
		Token:             "token",
		Server:            "ws" + strings.TrimPrefix(srv.URL, "http") + "/_tunnel/connect",
		ReconnectInterval: time.Second,
		Tunnels: []config.TunnelConfig{{
			Name:       "api",
			PublicPath: "/api",
			Target:     target,
			StripPath:  stripPath,
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = tunnelclient.New(clientCfg, slog.Default()).Run(ctx)
	}()
	waitForStatus(t, srv.URL)
	return srv
}

func waitForStatus(t *testing.T, base string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/_tunnel/status")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && strings.Contains(string(body), `"id":"dev"`) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("client did not register before deadline")
}
