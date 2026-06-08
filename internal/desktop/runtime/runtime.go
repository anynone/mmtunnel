package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mmtunnel/internal/client"
	"mmtunnel/internal/config"
)

type State string

const (
	StateStopped      State = "stopped"
	StateConnecting   State = "connecting"
	StateConnected    State = "connected"
	StateReconnecting State = "reconnecting"
	StateStopping     State = "stopping"
	StateError        State = "error"
)

type EventType string

const (
	EventState   EventType = "state"
	EventLog     EventType = "log"
	EventRequest EventType = "request"
	EventError   EventType = "error"
)

type Event struct {
	Time      time.Time `json:"time"`
	Type      EventType `json:"type"`
	Level     string    `json:"level,omitempty"`
	Message   string    `json:"message,omitempty"`
	State     State     `json:"state,omitempty"`
	ProfileID string    `json:"profileId,omitempty"`
	Tunnel    string    `json:"tunnel,omitempty"`
	Method    string    `json:"method,omitempty"`
	Path      string    `json:"path,omitempty"`
	Status    int       `json:"status,omitempty"`
	Duration  int64     `json:"durationMs,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type Status struct {
	State     State     `json:"state"`
	ProfileID string    `json:"profileId,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
	LastError string    `json:"lastError,omitempty"`
}

type Provider interface {
	ActiveClientConfig() (config.ClientConfig, error)
	ActiveProfileID() string
}

type StaticProvider struct {
	ProfileID string
	Config    config.ClientConfig
}

func (p StaticProvider) ActiveClientConfig() (config.ClientConfig, error) {
	return p.Config, nil
}

func (p StaticProvider) ActiveProfileID() string {
	return p.ProfileID
}

type Runner func(context.Context, config.ClientConfig, func(Event)) error

type Manager struct {
	provider Provider
	runner   Runner
	limit    int

	mu      sync.Mutex
	state   State
	started time.Time
	err     string
	cancel  context.CancelFunc
	events  []Event
	subs    map[chan Event]struct{}
}

func NewManager(provider Provider, runner Runner, eventLimit int) *Manager {
	if eventLimit <= 0 {
		eventLimit = 200
	}
	return &Manager{provider: provider, runner: runner, limit: eventLimit, state: StateStopped, subs: map[chan Event]struct{}{}}
}

func NewClientRunner() Runner {
	return func(ctx context.Context, cfg config.ClientConfig, emit func(Event)) error {
		return client.New(cfg, nil, client.WithHook(func(event client.Event) {
			switch event.Type {
			case "request_completed":
				emit(Event{Type: EventRequest, Method: event.Method, Path: event.Path, Status: event.Status, Duration: event.Duration.Milliseconds()})
			case "request_failed":
				emit(Event{Type: EventError, Method: event.Method, Path: event.Path, Status: event.Status, Duration: event.Duration.Milliseconds(), Error: event.Error, Message: "request failed"})
			case "connected":
				emit(Event{Type: EventLog, Level: "info", Message: "client connected"})
			}
		})).Run(ctx)
	}
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.cancel != nil {
		m.mu.Unlock()
		return nil
	}
	cfg, err := m.provider.ActiveClientConfig()
	if err != nil {
		m.setLocked(StateError, err.Error())
		m.mu.Unlock()
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.started = time.Now()
	m.setLocked(StateConnecting, "")
	m.mu.Unlock()

	go func() {
		m.set(StateConnected, "")
		err := m.runner(runCtx, cfg, m.Emit)
		m.mu.Lock()
		defer m.mu.Unlock()
		m.cancel = nil
		if err == nil || err == context.Canceled || runCtx.Err() != nil {
			m.setLocked(StateStopped, "")
			return
		}
		m.setLocked(StateError, err.Error())
	}()
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	cancel := m.cancel
	if cancel == nil {
		m.setLocked(StateStopped, "")
		m.mu.Unlock()
		return nil
	}
	m.setLocked(StateStopping, "")
	m.mu.Unlock()
	cancel()
	return nil
}

func (m *Manager) Restart(ctx context.Context) error {
	m.mu.Lock()
	running := m.cancel != nil
	m.mu.Unlock()
	if running {
		if err := m.Stop(); err != nil {
			return err
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if m.Status().State == StateStopped {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	m.set(StateReconnecting, "")
	return m.Start(ctx)
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Status{State: m.state, ProfileID: m.provider.ActiveProfileID(), StartedAt: m.started, LastError: m.err}
}

func (m *Manager) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

func (m *Manager) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 32)
	m.mu.Lock()
	m.subs[ch] = struct{}{}
	m.mu.Unlock()
	return ch, func() {
		m.mu.Lock()
		delete(m.subs, ch)
		close(ch)
		m.mu.Unlock()
	}
}

func (m *Manager) Emit(event Event) {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	m.mu.Lock()
	m.events = append(m.events, event)
	if len(m.events) > m.limit {
		m.events = m.events[len(m.events)-m.limit:]
	}
	for ch := range m.subs {
		select {
		case ch <- event:
		default:
		}
	}
	m.mu.Unlock()
}

func (m *Manager) set(state State, err string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setLocked(state, err)
}

func (m *Manager) setLocked(state State, err string) {
	m.state = state
	m.err = err
	event := Event{Time: time.Now(), Type: EventState, State: state, ProfileID: m.provider.ActiveProfileID(), Error: err}
	if err != "" {
		event.Type = EventError
		event.Message = fmt.Sprintf("runtime error: %s", err)
	}
	m.events = append(m.events, event)
	if len(m.events) > m.limit {
		m.events = m.events[len(m.events)-m.limit:]
	}
	for ch := range m.subs {
		select {
		case ch <- event:
		default:
		}
	}
}
