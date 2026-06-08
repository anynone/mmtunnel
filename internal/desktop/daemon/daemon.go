package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"mmtunnel/internal/config"
	"mmtunnel/internal/desktop/profile"
	"mmtunnel/internal/desktop/runtime"
	"mmtunnel/internal/protocol"
)

type API struct {
	store   profile.Store
	path    string
	manager *runtime.Manager
}

func New(store profile.Store, manager *runtime.Manager) *API {
	return &API{store: store, manager: manager}
}

func NewPersistent(path string, managerFactory func(runtime.Provider) *runtime.Manager) (*API, error) {
	store, err := profile.Load(path)
	if err != nil {
		return nil, err
	}
	if store.ActiveProfileID == "" && len(store.Profiles) > 0 {
		store.ActiveProfileID = store.Profiles[0].ID
	}
	api := &API{store: store, path: path}
	api.manager = managerFactory(api)
	return api, nil
}

func (a *API) ActiveClientConfig() (config.ClientConfig, error) {
	return a.store.ActiveClientRuntimeConfig()
}

func (a *API) ActiveProfileID() string {
	return a.store.ActiveProfileID
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/profiles", a.handleProfiles)
	mux.HandleFunc("/api/profiles/", a.handleProfileItem)
	mux.HandleFunc("/api/runtime/start", a.handleStart)
	mux.HandleFunc("/api/runtime/stop", a.handleStop)
	mux.HandleFunc("/api/runtime/restart", a.handleRestart)
	mux.HandleFunc("/api/events", a.handleEvents)
	mux.HandleFunc("/api/logs", a.handleLogs)
	return loopbackOnly(mux)
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"runtime":         a.manager.Status(),
		"activeProfileId": a.store.ActiveProfileID,
		"profiles":        a.store.Profiles,
		"tunnels":         a.tunnelStatuses(),
	})
}

func (a *API) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/profiles/import" {
		a.handleImport(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store)
	case http.MethodPost:
		var p profile.Profile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.store.Profiles = append(a.store.Profiles, p)
		if a.store.ActiveProfileID == "" {
			a.store.ActiveProfileID = p.ID
		}
		if err := a.store.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = a.save()
		writeJSON(w, http.StatusCreated, p)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	if id == "" {
		id = "imported"
	}
	if name == "" {
		name = id
	}
	tmp, err := os.CreateTemp("", "mmtunnel-import-*.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := io.Copy(tmp, r.Body); err != nil {
		_ = tmp.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = tmp.Close()
	p, err := profile.ImportCLIYAML(id, name, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.store.Profiles = append(a.store.Profiles, p)
	if a.store.ActiveProfileID == "" {
		a.store.ActiveProfileID = p.ID
	}
	if err := a.store.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = a.save()
	writeJSON(w, http.StatusCreated, p)
}

func (a *API) handleProfileItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	if strings.HasSuffix(path, "/test-server") {
		a.handleTestServer(w, r, strings.TrimSuffix(path, "/test-server"))
		return
	}
	if strings.HasSuffix(path, "/test-auth") {
		a.handleTestAuth(w, r, strings.TrimSuffix(path, "/test-auth"))
		return
	}
	if strings.HasSuffix(path, "/export") {
		a.handleExport(w, r, strings.TrimSuffix(path, "/export"))
		return
	}
	if strings.HasSuffix(path, "/active") {
		a.handleSetActive(w, r, strings.TrimSuffix(path, "/active"))
		return
	}
	id := strings.Trim(path, "/")
	idx := a.profileIndex(id)
	if idx < 0 {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.Profiles[idx])
	case http.MethodPut:
		var p profile.Profile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p.ID = id
		a.store.Profiles[idx] = p
		if err := a.store.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = a.save()
		writeJSON(w, http.StatusOK, p)
	case http.MethodDelete:
		if a.store.ActiveProfileID == id {
			http.Error(w, "cannot delete active profile", http.StatusConflict)
			return
		}
		a.store.Profiles = append(a.store.Profiles[:idx], a.store.Profiles[idx+1:]...)
		_ = a.save()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) handleExport(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p, ok := a.profileByID(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	store := profile.Store{ActiveProfileID: p.ID, Profiles: []profile.Profile{p}}
	tmp, err := os.CreateTemp("", "mmtunnel-cli-*.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(path)
	if err := store.ExportCLIYAML(p.ID, path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	_, _ = w.Write(data)
}

func (a *API) handleSetActive(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.profileIndex(id) < 0 {
		http.NotFound(w, r)
		return
	}
	a.store.ActiveProfileID = id
	_ = a.save()
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := a.manager.Start(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = a.manager.Stop()
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := a.manager.Restart(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	flusher, _ := w.(http.Flusher)
	ch, unsubscribe := a.manager.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

func (a *API) handleLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.manager.Events())
}

func (a *API) handleTestServer(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p, ok := a.profileByID(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	conn, _, err := protocol.Dial(p.Client.Server, 5_000_000_000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_ = conn.Close()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) handleTestAuth(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p, ok := a.profileByID(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	conn, _, err := protocol.Dial(p.Client.Server, 5_000_000_000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()
	data, _ := protocol.Encode(protocol.Message{Type: protocol.TypeAuth, ClientID: p.Client.ID, Token: p.Client.Token})
	if err := conn.WriteMessage(protocol.OpcodeText, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_, body, err := conn.ReadMessage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	msg, err := protocol.Decode(body)
	if err != nil || msg.Type != protocol.TypeAuthOK {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) profileIndex(id string) int {
	for i := range a.store.Profiles {
		if a.store.Profiles[i].ID == id {
			return i
		}
	}
	return -1
}

func (a *API) profileByID(id string) (profile.Profile, bool) {
	idx := a.profileIndex(id)
	if idx < 0 {
		return profile.Profile{}, false
	}
	return a.store.Profiles[idx], true
}

func (a *API) save() error {
	if a.path == "" {
		return nil
	}
	return a.store.Save(a.path)
}

func (a *API) tunnelStatuses() map[string]profile.TunnelStatus {
	active, err := a.store.ActiveProfile()
	if err != nil {
		return map[string]profile.TunnelStatus{}
	}
	statuses := map[string]profile.TunnelStatus{}
	state := a.manager.Status().State
	for _, tunnel := range active.Tunnels {
		if !tunnel.Enabled {
			statuses[tunnel.ID] = profile.TunnelStatusDisabled
			continue
		}
		switch state {
		case runtime.StateConnected:
			statuses[tunnel.ID] = profile.TunnelStatusRegistered
		case runtime.StateError:
			statuses[tunnel.ID] = profile.TunnelStatusFailed
		default:
			statuses[tunnel.ID] = profile.TunnelStatusPending
		}
	}
	return statuses
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func loopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ip := net.ParseIP(host)
		if ip != nil && !ip.IsLoopback() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
