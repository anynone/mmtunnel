package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mmtunnel/internal/config"
	"mmtunnel/internal/desktop/profile"
	"mmtunnel/internal/desktop/runtime"
	"mmtunnel/internal/protocol"
)

func TestProfileCRUDAndRuntimeStatus(t *testing.T) {
	store := profile.Store{ActiveProfileID: "company", Profiles: []profile.Profile{{
		ID: "company", Name: "Company",
		Client:  profile.ClientIdentity{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"},
		Tunnels: []profile.Tunnel{{ID: "api", Name: "API", PublicPath: "/api", Target: "http://127.0.0.1:3000", Enabled: true}},
	}}}
	manager := runtime.NewManager(runtime.StaticProvider{ProfileID: "company", Config: config.ClientConfig{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"}}, func(ctx context.Context, cfg config.ClientConfig, emit func(runtime.Event)) error {
		<-ctx.Done()
		return ctx.Err()
	}, 10)
	api := New(store, manager)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("profiles status = %d", resp.StatusCode)
	}

	body, _ := json.Marshal(profile.Profile{ID: "local", Name: "Local", Client: profile.ClientIdentity{ID: "local", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"}})
	resp, err = http.Post(server.URL+"/api/profiles", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}

	resp, err = http.Post(server.URL+"/api/runtime/start", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("start status = %d", resp.StatusCode)
	}
	resp, err = http.Get(server.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestConnectionTestEndpoints(t *testing.T) {
	tunnelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := protocol.Upgrade(w, r)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		msg, _ := protocol.Decode(data)
		if msg.Type == protocol.TypeAuth && msg.ClientID == "dev" && msg.Token == "token" {
			encoded, _ := protocol.Encode(protocol.Message{Type: protocol.TypeAuthOK, ClientID: "dev"})
			_ = conn.WriteMessage(protocol.OpcodeText, encoded)
		}
	}))
	defer tunnelServer.Close()

	wsURL := "ws" + tunnelServer.URL[len("http"):] + "/_tunnel/connect"
	store := profile.Store{ActiveProfileID: "company", Profiles: []profile.Profile{{
		ID: "company", Name: "Company",
		Client: profile.ClientIdentity{ID: "dev", Token: "token", Server: wsURL},
	}}}
	manager := runtime.NewManager(runtime.StaticProvider{ProfileID: "company", Config: config.ClientConfig{ID: "dev", Token: "token", Server: wsURL}}, func(ctx context.Context, cfg config.ClientConfig, emit func(runtime.Event)) error {
		<-ctx.Done()
		return ctx.Err()
	}, 10)
	api := New(store, manager)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/profiles/company/test-server", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("test-server status = %d", resp.StatusCode)
	}

	resp, err = http.Post(server.URL+"/api/profiles/company/test-auth", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("test-auth status = %d", resp.StatusCode)
	}
}
