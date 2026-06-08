package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"mmtunnel/internal/protocol"
)

type targetWebSocket struct {
	conn *protocol.WSConn
}

func (c *Client) startTargetWebSocket(ctx context.Context, serverConn *protocol.WSConn, start protocol.Message, sessions map[string]*targetWebSocket, mu *sync.Mutex) {
	target, err := buildWebSocketTargetURL(start.Target, start.Path, start.Query)
	if err != nil {
		_ = write(serverConn, protocol.Message{Type: protocol.TypeError, RequestID: start.RequestID, Error: err.Error()})
		return
	}
	targetConn, _, err := protocol.Dial(target, 10_000_000_000)
	if err != nil {
		_ = write(serverConn, protocol.Message{Type: protocol.TypeError, RequestID: start.RequestID, Error: err.Error()})
		return
	}
	mu.Lock()
	sessions[start.RequestID] = &targetWebSocket{conn: targetConn}
	mu.Unlock()
	if err := write(serverConn, protocol.Message{Type: protocol.TypeWebSocketAccept, RequestID: start.RequestID}); err != nil {
		_ = targetConn.Close()
		return
	}
	go func() {
		defer func() {
			mu.Lock()
			delete(sessions, start.RequestID)
			mu.Unlock()
			_ = targetConn.Close()
			_ = write(serverConn, protocol.Message{Type: protocol.TypeWebSocketClose, RequestID: start.RequestID})
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			opcode, body, err := targetConn.ReadMessage()
			if err != nil {
				return
			}
			if err := write(serverConn, protocol.Message{Type: protocol.TypeWebSocketFrame, RequestID: start.RequestID, Opcode: opcode, Body: body}); err != nil {
				return
			}
		}
	}()
}

func buildWebSocketTargetURL(rawTarget, path, query string) (string, error) {
	target, err := url.Parse(rawTarget)
	if err != nil {
		return "", err
	}
	switch target.Scheme {
	case "http":
		target.Scheme = "ws"
	case "https":
		target.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported websocket target scheme %s", target.Scheme)
	}
	base := strings.TrimRight(target.Path, "/")
	target.Path = base + path
	target.RawQuery = query
	return target.String(), nil
}
