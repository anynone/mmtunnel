package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"mmtunnel/internal/config"
	"mmtunnel/internal/protocol"
)

type Client struct {
	cfg  config.ClientConfig
	log  *slog.Logger
	http *http.Client
	hook Hook
}

type Hook func(Event)

type Event struct {
	Type     string
	Method   string
	Path     string
	Status   int
	Duration time.Duration
	Error    string
}

type Option func(*Client)

func WithHook(hook Hook) Option {
	return func(c *Client) {
		c.hook = hook
	}
}

type incomingRequest struct {
	start protocol.Message
	body  bytes.Buffer
}

func New(cfg config.ClientConfig, log *slog.Logger, options ...Option) *Client {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	c := &Client{
		cfg:  cfg,
		log:  log,
		http: &http.Client{Timeout: 0},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *Client) Run(ctx context.Context) error {
	for {
		if err := c.runOnce(ctx); err != nil {
			c.log.Warn("client connection ended", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.cfg.ReconnectInterval):
		}
	}
}

func (c *Client) RunOnce(ctx context.Context) error {
	return c.runOnce(ctx)
}

func (c *Client) runOnce(ctx context.Context) error {
	conn, _, err := protocol.Dial(c.cfg.Server, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := write(conn, protocol.Message{Type: protocol.TypeAuth, ClientID: c.cfg.ID, Token: c.cfg.Token}); err != nil {
		return err
	}
	if err := expect(conn, protocol.TypeAuthOK); err != nil {
		return err
	}
	tunnels := make([]protocol.Tunnel, 0, len(c.cfg.Tunnels))
	for _, tunnel := range c.cfg.Tunnels {
		tunnels = append(tunnels, protocol.Tunnel{
			Name:       tunnel.Name,
			PublicPath: tunnel.PublicPath,
			Target:     tunnel.Target,
			StripPath:  tunnel.StripPath,
		})
	}
	if err := write(conn, protocol.Message{Type: protocol.TypeRegister, Tunnels: tunnels}); err != nil {
		return err
	}
	if err := expect(conn, protocol.TypeRegisterOK); err != nil {
		return err
	}
	c.log.Info("client connected and registered", "tunnels", len(tunnels))
	c.emit(Event{Type: "connected"})

	done := make(chan error, 1)
	go c.readLoop(ctx, conn, done)
	go c.keepalive(ctx, conn)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (c *Client) readLoop(ctx context.Context, conn *protocol.WSConn, done chan<- error) {
	requests := map[string]*incomingRequest{}
	wsSessions := map[string]*targetWebSocket{}
	var mu sync.Mutex

	for {
		opcode, data, err := conn.ReadMessage()
		if err != nil {
			done <- err
			return
		}
		if opcode == protocol.OpcodePing {
			_ = conn.WriteMessage(protocol.OpcodePong, data)
			continue
		}
		if opcode != protocol.OpcodeText {
			continue
		}
		msg, err := protocol.Decode(data)
		if err != nil {
			c.log.Warn("invalid server message", "error", err)
			continue
		}
		switch msg.Type {
		case protocol.TypeRequestStart:
			mu.Lock()
			requests[msg.RequestID] = &incomingRequest{start: msg}
			mu.Unlock()
		case protocol.TypeRequestBody:
			mu.Lock()
			req := requests[msg.RequestID]
			if req != nil {
				_, _ = req.body.Write(msg.Body)
			}
			mu.Unlock()
		case protocol.TypeRequestEnd:
			mu.Lock()
			req := requests[msg.RequestID]
			delete(requests, msg.RequestID)
			mu.Unlock()
			if req != nil {
				go c.forwardHTTP(ctx, conn, *req)
			}
		case protocol.TypeWebSocketStart:
			go c.startTargetWebSocket(ctx, conn, msg, wsSessions, &mu)
		case protocol.TypeWebSocketFrame:
			mu.Lock()
			ws := wsSessions[msg.RequestID]
			mu.Unlock()
			if ws != nil {
				_ = ws.conn.WriteMessage(msg.Opcode, msg.Body)
			}
		case protocol.TypeWebSocketClose:
			mu.Lock()
			ws := wsSessions[msg.RequestID]
			delete(wsSessions, msg.RequestID)
			mu.Unlock()
			if ws != nil {
				_ = ws.conn.Close()
			}
		}
	}
}

func (c *Client) forwardHTTP(ctx context.Context, conn *protocol.WSConn, req incomingRequest) {
	started := time.Now()
	target, err := buildTargetURL(req.start.Target, req.start.Path, req.start.Query)
	if err != nil {
		_ = write(conn, protocol.Message{Type: protocol.TypeError, RequestID: req.start.RequestID, Error: err.Error()})
		c.emit(Event{Type: "request_failed", Method: req.start.Method, Path: req.start.Path, Duration: time.Since(started), Error: err.Error()})
		return
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.start.Method, target, bytes.NewReader(req.body.Bytes()))
	if err != nil {
		_ = write(conn, protocol.Message{Type: protocol.TypeError, RequestID: req.start.RequestID, Error: err.Error()})
		c.emit(Event{Type: "request_failed", Method: req.start.Method, Path: req.start.Path, Duration: time.Since(started), Error: err.Error()})
		return
	}
	for name, values := range req.start.Headers {
		for _, value := range values {
			httpReq.Header.Add(name, value)
		}
	}
	httpReq.Host = ""
	resp, err := c.http.Do(httpReq)
	if err != nil {
		_ = write(conn, protocol.Message{Type: protocol.TypeError, RequestID: req.start.RequestID, Error: err.Error()})
		c.emit(Event{Type: "request_failed", Method: req.start.Method, Path: req.start.Path, Duration: time.Since(started), Error: err.Error()})
		return
	}
	defer resp.Body.Close()
	if err := write(conn, protocol.Message{Type: protocol.TypeResponseStart, RequestID: req.start.RequestID, Status: resp.StatusCode, Headers: protocol.CopyHeaders(resp.Header)}); err != nil {
		return
	}
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if sendErr := write(conn, protocol.Message{Type: protocol.TypeResponseBody, RequestID: req.start.RequestID, Body: append([]byte(nil), buf[:n]...)}); sendErr != nil {
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = write(conn, protocol.Message{Type: protocol.TypeError, RequestID: req.start.RequestID, Error: err.Error()})
			c.emit(Event{Type: "request_failed", Method: req.start.Method, Path: req.start.Path, Status: resp.StatusCode, Duration: time.Since(started), Error: err.Error()})
			return
		}
	}
	_ = write(conn, protocol.Message{Type: protocol.TypeResponseEnd, RequestID: req.start.RequestID})
	c.emit(Event{Type: "request_completed", Method: req.start.Method, Path: req.start.Path, Status: resp.StatusCode, Duration: time.Since(started)})
}

func (c *Client) emit(event Event) {
	if c.hook != nil {
		c.hook(event)
	}
}

func buildTargetURL(rawTarget, path, query string) (string, error) {
	target, err := url.Parse(rawTarget)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(target.Path, "/")
	target.Path = base + path
	target.RawQuery = query
	return target.String(), nil
}

func (c *Client) keepalive(ctx context.Context, conn *protocol.WSConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = write(conn, protocol.Message{Type: protocol.TypePing})
		}
	}
}

func expect(conn *protocol.WSConn, want string) error {
	opcode, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if opcode != protocol.OpcodeText {
		return fmt.Errorf("expected text message")
	}
	msg, err := protocol.Decode(data)
	if err != nil {
		return err
	}
	if msg.Type == protocol.TypeError {
		return errors.New(msg.Error)
	}
	if msg.Type != want {
		return fmt.Errorf("expected %s, got %s", want, msg.Type)
	}
	return nil
}

func write(conn *protocol.WSConn, msg protocol.Message) error {
	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(protocol.OpcodeText, data)
}
