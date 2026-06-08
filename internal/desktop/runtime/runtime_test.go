package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"mmtunnel/internal/config"
)

func TestManagerStartStopRestartAndEvents(t *testing.T) {
	provider := StaticProvider{ProfileID: "company", Config: config.ClientConfig{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"}}
	starts := 0
	manager := NewManager(provider, func(ctx context.Context, cfg config.ClientConfig, emit func(Event)) error {
		starts++
		emit(Event{Type: EventLog, Message: "runner started"})
		<-ctx.Done()
		return ctx.Err()
	}, 10)

	if err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitState(t, manager, StateConnected)
	if err := manager.Restart(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return starts >= 2 })
	if err := manager.Stop(); err != nil {
		t.Fatal(err)
	}
	waitState(t, manager, StateStopped)
	if len(manager.Events()) == 0 {
		t.Fatal("expected events")
	}
}

func TestManagerRecordsRunnerError(t *testing.T) {
	provider := StaticProvider{ProfileID: "company", Config: config.ClientConfig{ID: "dev", Token: "token", Server: "ws://127.0.0.1:8080/_tunnel/connect"}}
	manager := NewManager(provider, func(ctx context.Context, cfg config.ClientConfig, emit func(Event)) error {
		return errors.New("boom")
	}, 10)
	if err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitState(t, manager, StateError)
	if manager.Status().LastError != "boom" {
		t.Fatalf("last error = %q", manager.Status().LastError)
	}
}

func waitState(t *testing.T, manager *Manager, state State) {
	t.Helper()
	waitFor(t, func() bool { return manager.Status().State == state })
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met")
}
