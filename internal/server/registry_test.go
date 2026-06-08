package server

import (
	"testing"
	"time"

	"log/slog"

	"mmsocket/internal/protocol"
)

func TestRegistryRejectsConflictingPath(t *testing.T) {
	reg := NewRegistry()
	session := &Session{ClientID: "a", ConnectedAt: time.Now(), log: slog.Default()}
	if err := reg.RegisterClient(session, []protocol.Tunnel{{Name: "api", PublicPath: "/api", Target: "http://127.0.0.1:3000", StripPath: true}}); err != nil {
		t.Fatal(err)
	}
	err := reg.RegisterClient(&Session{ClientID: "b", ConnectedAt: time.Now(), log: slog.Default()}, []protocol.Tunnel{{Name: "v1", PublicPath: "/api/v1", Target: "http://127.0.0.1:3001", StripPath: true}})
	if err == nil {
		t.Fatal("expected conflict")
	}
	if err.Error() != "路径冲突，已经存在当前路径 /api" {
		t.Fatalf("unexpected conflict error: %v", err)
	}
}

func TestRegistryLongestPrefixMatch(t *testing.T) {
	reg := NewRegistry()
	session := &Session{ClientID: "a", ConnectedAt: time.Now(), log: slog.Default()}
	reg.routes["/api"] = Route{ClientID: "a", PublicPath: "/api", session: session}
	reg.routes["/api/v1"] = Route{ClientID: "a", PublicPath: "/api/v1", session: session}
	route, ok := reg.Match("/api/v1/users")
	if !ok {
		t.Fatal("expected match")
	}
	if route.PublicPath != "/api/v1" {
		t.Fatalf("got %s", route.PublicPath)
	}
}
