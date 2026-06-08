package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"mmtunnel/internal/config"
	"mmtunnel/internal/protocol"
	"mmtunnel/internal/validate"
)

type Server struct {
	cfg      config.ServerConfig
	log      *slog.Logger
	registry *Registry
	clients  map[string]string
}

func New(cfg config.ServerConfig, log *slog.Logger) *Server {
	clients := map[string]string{}
	for _, client := range cfg.Clients {
		clients[client.ID] = client.Token
	}
	return &Server{cfg: cfg, log: log, registry: NewRegistry(), clients: clients}
}

func (s *Server) Run(ctx context.Context) error {
	httpServer := &http.Server{Addr: s.cfg.Listen, Handler: s}
	errs := make(chan error, 1)
	go func() {
		s.log.Info("tunnel server starting", "listen", s.cfg.Listen)
		errs <- httpServer.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errs:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == validate.SystemPathPrefix+"/connect" {
		s.handleConnect(w, r)
		return
	}
	if validate.IsReservedPath(r.URL.Path) {
		s.handleStatus(w, r)
		return
	}
	if protocol.IsWebSocketRequest(r) {
		s.handlePublicWebSocket(w, r)
		return
	}
	s.handlePublicHTTP(w, r)
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := protocol.Upgrade(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	opcode, data, err := conn.ReadMessage()
	if err != nil || opcode != protocol.OpcodeText {
		_ = conn.Close()
		return
	}
	auth, err := protocol.Decode(data)
	if err != nil || auth.Type != protocol.TypeAuth {
		_ = conn.Close()
		return
	}
	if !s.authenticate(auth.ClientID, auth.Token) {
		_ = writeProtocol(conn, protocol.Message{Type: protocol.TypeError, Error: "invalid client credentials"})
		_ = conn.Close()
		s.log.Warn("client authentication failed", "clientId", auth.ClientID)
		return
	}
	session := NewSession(auth.ClientID, conn, s.log.With("clientId", auth.ClientID))
	session.ReadTimeout = s.cfg.PingInterval + s.cfg.PongTimeout
	if err := session.Send(protocol.Message{Type: protocol.TypeAuthOK, ClientID: auth.ClientID}); err != nil {
		session.Close()
		return
	}
	opcode, data, err = conn.ReadMessage()
	if err != nil || opcode != protocol.OpcodeText {
		session.Close()
		return
	}
	reg, err := protocol.Decode(data)
	if err != nil || reg.Type != protocol.TypeRegister {
		_ = session.Send(protocol.Message{Type: protocol.TypeError, Error: "expected register message"})
		session.Close()
		return
	}
	if err := s.registry.RegisterClient(session, reg.Tunnels); err != nil {
		_ = session.Send(protocol.Message{Type: protocol.TypeError, Error: err.Error()})
		session.Close()
		s.log.Warn("tunnel registration failed", "clientId", auth.ClientID, "error", err)
		return
	}
	if err := session.Send(protocol.Message{Type: protocol.TypeRegisterOK}); err != nil {
		s.registry.UnregisterClient(auth.ClientID)
		session.Close()
		return
	}
	s.log.Info("client connected", "clientId", auth.ClientID, "tunnels", len(reg.Tunnels))
	session.StartReader(func() {
		s.registry.UnregisterClient(auth.ClientID)
		s.log.Info("client disconnected", "clientId", auth.ClientID)
	})
	s.startKeepalive(session)
}

func (s *Server) authenticate(clientID, token string) bool {
	expected, ok := s.clients[clientID]
	return ok && expected == token
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "status endpoints are read-only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"clients": s.registry.Status(),
	})
}

func (s *Server) handlePublicHTTP(w http.ResponseWriter, r *http.Request) {
	route, ok := s.registry.Match(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	tracked := &trackingResponseWriter{ResponseWriter: w}
	if err := route.session.ForwardHTTP(ctx, route, r, s.cfg.MaxRequestBodySize, tracked); err != nil {
		s.log.Warn("http forwarding failed", "path", r.URL.Path, "clientId", route.ClientID, "error", err)
		if !tracked.wrote {
			http.Error(w, "bad gateway", http.StatusBadGateway)
		}
		return
	}
}

func (s *Server) handlePublicWebSocket(w http.ResponseWriter, r *http.Request) {
	route, ok := s.registry.Match(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	public, err := protocol.Upgrade(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	if err := route.session.ForwardWebSocket(ctx, route, public, r); err != nil && !strings.Contains(err.Error(), "closed") {
		s.log.Warn("websocket forwarding failed", "path", r.URL.Path, "clientId", route.ClientID, "error", err)
	}
}

func (s *Server) startKeepalive(session *Session) {
	go func() {
		ticker := time.NewTicker(s.cfg.PingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-session.closed:
				return
			case <-ticker.C:
				if err := session.Ping(); err != nil {
					s.log.Warn("keepalive failed", "clientId", session.ClientID, "error", err)
					session.Close()
					return
				}
			}
		}
	}()
}

func writeProtocol(conn *protocol.WSConn, msg protocol.Message) error {
	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(protocol.OpcodeText, data)
}

func ConflictError(path string) error {
	return fmt.Errorf("路径冲突，已经存在当前路径 %s", path)
}

type trackingResponseWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *trackingResponseWriter) WriteHeader(statusCode int) {
	w.wrote = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *trackingResponseWriter) Write(data []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(data)
}

func (w *trackingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
