package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"mmtunnel/internal/protocol"
)

type Session struct {
	ClientID    string
	ConnectedAt time.Time
	ReadTimeout time.Duration
	conn        *protocol.WSConn
	log         *slog.Logger
	writeMu     sync.Mutex
	pendingHTTP map[string]chan protocol.Message
	pendingWS   map[string]chan protocol.Message
	pendingMu   sync.Mutex
	closed      chan struct{}
	closeOnce   sync.Once
}

func NewSession(clientID string, conn *protocol.WSConn, log *slog.Logger) *Session {
	return &Session{
		ClientID:    clientID,
		ConnectedAt: time.Now(),
		conn:        conn,
		log:         log,
		pendingHTTP: map[string]chan protocol.Message{},
		pendingWS:   map[string]chan protocol.Message{},
		closed:      make(chan struct{}),
	}
}

func (s *Session) StartReader(onClose func()) {
	go func() {
		defer onClose()
		defer s.Close()
		s.extendReadDeadline()
		for {
			opcode, data, err := s.conn.ReadMessage()
			if err != nil {
				s.log.Debug("client tunnel read ended", "clientId", s.ClientID, "error", err)
				return
			}
			s.extendReadDeadline()
			if opcode == protocol.OpcodePong {
				continue
			}
			if opcode != protocol.OpcodeText {
				continue
			}
			msg, err := protocol.Decode(data)
			if err != nil {
				s.log.Warn("invalid tunnel message", "clientId", s.ClientID, "error", err)
				continue
			}
			switch msg.Type {
			case protocol.TypeResponseStart, protocol.TypeResponseBody, protocol.TypeResponseEnd, protocol.TypeError:
				s.deliver(s.pendingHTTP, msg)
			case protocol.TypeWebSocketAccept, protocol.TypeWebSocketFrame, protocol.TypeWebSocketClose:
				s.deliver(s.pendingWS, msg)
			case protocol.TypePong:
			default:
				s.log.Debug("ignored client message", "clientId", s.ClientID, "type", msg.Type)
			}
		}
	}()
}

func (s *Session) extendReadDeadline() {
	if s.ReadTimeout > 0 {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	}
}

func (s *Session) Send(msg protocol.Message) error {
	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn.WriteMessage(protocol.OpcodeText, data)
}

func (s *Session) Ping() error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn.WriteMessage(protocol.OpcodePing, []byte("ping"))
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		_ = s.conn.Close()
		s.pendingMu.Lock()
		for id, ch := range s.pendingHTTP {
			delete(s.pendingHTTP, id)
			close(ch)
		}
		for id, ch := range s.pendingWS {
			delete(s.pendingWS, id)
			close(ch)
		}
		s.pendingMu.Unlock()
	})
}

func (s *Session) ForwardHTTP(ctx context.Context, route Route, r *http.Request, maxBody int64, w http.ResponseWriter) error {
	requestID := protocol.NewRequestID()
	ch := s.addPending(s.pendingHTTP, requestID)
	defer s.removePending(s.pendingHTTP, requestID)

	path := forwardPath(route, r.URL.Path)
	start := protocol.Message{
		Type:      protocol.TypeRequestStart,
		RequestID: requestID,
		Method:    r.Method,
		Path:      path,
		Query:     r.URL.RawQuery,
		Target:    route.Target,
		Headers:   protocol.CopyHeaders(r.Header),
	}
	addForwardedHeaders(start.Headers, r, route.ClientID)
	if err := s.Send(start); err != nil {
		return err
	}
	body := http.MaxBytesReader(nil, r.Body, maxBody)
	buf := make([]byte, 32*1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if sendErr := s.Send(protocol.Message{Type: protocol.TypeRequestBody, RequestID: requestID, Body: append([]byte(nil), buf[:n]...)}); sendErr != nil {
				return sendErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if err := s.Send(protocol.Message{Type: protocol.TypeRequestEnd, RequestID: requestID}); err != nil {
		return err
	}

	started := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return fmt.Errorf("client disconnected")
			}
			switch msg.Type {
			case protocol.TypeResponseStart:
				for name, values := range msg.Headers {
					for _, value := range values {
						w.Header().Add(name, value)
					}
				}
				status := msg.Status
				if status == 0 {
					status = http.StatusOK
				}
				w.WriteHeader(status)
				started = true
			case protocol.TypeResponseBody:
				if !started {
					w.WriteHeader(http.StatusOK)
					started = true
				}
				if _, err := w.Write(msg.Body); err != nil {
					return err
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case protocol.TypeResponseEnd:
				if !started {
					w.WriteHeader(http.StatusOK)
				}
				return nil
			case protocol.TypeError:
				return errors.New(msg.Error)
			}
		}
	}
}

func (s *Session) ForwardWebSocket(ctx context.Context, route Route, public *protocol.WSConn, r *http.Request) error {
	requestID := protocol.NewRequestID()
	ch := s.addPending(s.pendingWS, requestID)
	defer s.removePending(s.pendingWS, requestID)

	headers := protocol.CopyHeaders(r.Header)
	addForwardedHeaders(headers, r, route.ClientID)
	if err := s.Send(protocol.Message{
		Type:      protocol.TypeWebSocketStart,
		RequestID: requestID,
		Method:    r.Method,
		Path:      forwardPath(route, r.URL.Path),
		Query:     r.URL.RawQuery,
		Target:    route.Target,
		Headers:   headers,
	}); err != nil {
		return err
	}
	select {
	case msg, ok := <-ch:
		if !ok {
			return fmt.Errorf("client disconnected")
		}
		if msg.Type == protocol.TypeError {
			return errors.New(msg.Error)
		}
		if msg.Type != protocol.TypeWebSocketAccept {
			return fmt.Errorf("unexpected websocket response %s", msg.Type)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	errs := make(chan error, 2)
	go func() {
		for {
			opcode, body, err := public.ReadMessage()
			if err != nil {
				errs <- err
				return
			}
			if err := s.Send(protocol.Message{Type: protocol.TypeWebSocketFrame, RequestID: requestID, Opcode: opcode, Body: body}); err != nil {
				errs <- err
				return
			}
		}
	}()
	go func() {
		for msg := range ch {
			switch msg.Type {
			case protocol.TypeWebSocketFrame:
				if err := public.WriteMessage(msg.Opcode, msg.Body); err != nil {
					errs <- err
					return
				}
			case protocol.TypeWebSocketClose:
				errs <- nil
				return
			case protocol.TypeError:
				errs <- errors.New(msg.Error)
				return
			}
		}
		errs <- fmt.Errorf("client disconnected")
	}()
	err := <-errs
	_ = s.Send(protocol.Message{Type: protocol.TypeWebSocketClose, RequestID: requestID})
	_ = public.Close()
	return err
}

func (s *Session) addPending(m map[string]chan protocol.Message, requestID string) chan protocol.Message {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	ch := make(chan protocol.Message, 32)
	m[requestID] = ch
	return ch
}

func (s *Session) removePending(m map[string]chan protocol.Message, requestID string) {
	s.pendingMu.Lock()
	ch, ok := m[requestID]
	if ok {
		delete(m, requestID)
		close(ch)
	}
	s.pendingMu.Unlock()
}

func (s *Session) deliver(m map[string]chan protocol.Message, msg protocol.Message) {
	s.pendingMu.Lock()
	ch := m[msg.RequestID]
	s.pendingMu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	case <-s.closed:
	}
}

func forwardPath(route Route, requestPath string) string {
	if !route.StripPath {
		return requestPath
	}
	trimmed := strings.TrimPrefix(requestPath, route.PublicPath)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func addForwardedHeaders(headers map[string][]string, r *http.Request, clientID string) {
	if headers == nil {
		return
	}
	remote := r.RemoteAddr
	if idx := strings.LastIndex(remote, ":"); idx > -1 {
		remote = remote[:idx]
	}
	headers["X-Forwarded-For"] = append(headers["X-Forwarded-For"], remote)
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	headers["X-Forwarded-Proto"] = []string{proto}
	headers["X-Tunnel-Client-Id"] = []string{clientID}
	delete(headers, "Host")
}
