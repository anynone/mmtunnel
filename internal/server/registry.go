package server

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"mmtunnel/internal/protocol"
	"mmtunnel/internal/validate"
)

type Route struct {
	ClientID   string    `json:"clientId"`
	Name       string    `json:"name"`
	PublicPath string    `json:"publicPath"`
	Target     string    `json:"target"`
	StripPath  bool      `json:"stripPath"`
	Connected  time.Time `json:"connectedAt"`
	session    *Session
}

type ClientStatus struct {
	ID        string    `json:"id"`
	Online    bool      `json:"online"`
	Connected time.Time `json:"connectedAt"`
	Tunnels   []Route   `json:"tunnels"`
}

type Registry struct {
	mu       sync.RWMutex
	routes   map[string]Route
	sessions map[string]*Session
}

func NewRegistry() *Registry {
	return &Registry{routes: map[string]Route{}, sessions: map[string]*Session{}}
}

func (r *Registry) RegisterClient(session *Session, tunnels []protocol.Tunnel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tunnel := range tunnels {
		if err := validate.PublicPath(tunnel.PublicPath); err != nil {
			return err
		}
		for existing := range r.routes {
			if existing == tunnel.PublicPath || validate.PathsConflict(existing, tunnel.PublicPath) {
				return fmt.Errorf("路径冲突，已经存在当前路径 %s", existing)
			}
		}
	}

	r.sessions[session.ClientID] = session
	for _, tunnel := range tunnels {
		r.routes[tunnel.PublicPath] = Route{
			ClientID:   session.ClientID,
			Name:       tunnel.Name,
			PublicPath: tunnel.PublicPath,
			Target:     tunnel.Target,
			StripPath:  tunnel.StripPath,
			Connected:  session.ConnectedAt,
			session:    session,
		}
	}
	return nil
}

func (r *Registry) UnregisterClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, clientID)
	for path, route := range r.routes {
		if route.ClientID == clientID {
			delete(r.routes, path)
		}
	}
}

func (r *Registry) Match(path string) (Route, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var best Route
	var found bool
	for publicPath, route := range r.routes {
		if path == publicPath || strings.HasPrefix(path, publicPath+"/") {
			if !found || len(publicPath) > len(best.PublicPath) {
				best = route
				found = true
			}
		}
	}
	return best, found
}

func (r *Registry) Status() []ClientStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	status := make([]ClientStatus, 0, len(r.sessions))
	for clientID, session := range r.sessions {
		item := ClientStatus{ID: clientID, Online: true, Connected: session.ConnectedAt}
		for _, route := range r.routes {
			if route.ClientID == clientID {
				route.session = nil
				item.Tunnels = append(item.Tunnels, route)
			}
		}
		sort.Slice(item.Tunnels, func(i, j int) bool {
			return item.Tunnels[i].PublicPath < item.Tunnels[j].PublicPath
		})
		status = append(status, item)
	}
	sort.Slice(status, func(i, j int) bool {
		return status[i].ID < status[j].ID
	})
	return status
}
