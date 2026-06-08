package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const (
	TypeAuth            = "auth"
	TypeAuthOK          = "auth_ok"
	TypeRegister        = "register"
	TypeRegisterOK      = "register_ok"
	TypeRequestStart    = "request_start"
	TypeRequestBody     = "request_body"
	TypeRequestEnd      = "request_end"
	TypeResponseStart   = "response_start"
	TypeResponseBody    = "response_body"
	TypeResponseEnd     = "response_end"
	TypeWebSocketStart  = "websocket_start"
	TypeWebSocketAccept = "websocket_accept"
	TypeWebSocketFrame  = "websocket_frame"
	TypeWebSocketClose  = "websocket_close"
	TypeError           = "error"
	TypePing            = "ping"
	TypePong            = "pong"
)

const (
	OpcodeText   = 1
	OpcodeBinary = 2
	OpcodeClose  = 8
	OpcodePing   = 9
	OpcodePong   = 10
)

type Tunnel struct {
	Name       string `json:"name"`
	PublicPath string `json:"publicPath"`
	Target     string `json:"target"`
	StripPath  bool   `json:"stripPath"`
}

type Message struct {
	Type      string              `json:"type"`
	RequestID string              `json:"requestId,omitempty"`
	ClientID  string              `json:"clientId,omitempty"`
	Token     string              `json:"token,omitempty"`
	Tunnels   []Tunnel            `json:"tunnels,omitempty"`
	Method    string              `json:"method,omitempty"`
	Path      string              `json:"path,omitempty"`
	Query     string              `json:"query,omitempty"`
	Target    string              `json:"target,omitempty"`
	Headers   map[string][]string `json:"headers,omitempty"`
	Body      []byte              `json:"body,omitempty"`
	Status    int                 `json:"status,omitempty"`
	Opcode    int                 `json:"opcode,omitempty"`
	Error     string              `json:"error,omitempty"`
}

func Encode(msg Message) ([]byte, error) {
	if msg.Type == "" {
		return nil, fmt.Errorf("message type is required")
	}
	return json.Marshal(msg)
}

func Decode(data []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	if msg.Type == "" {
		return msg, fmt.Errorf("message type is required")
	}
	return msg, nil
}

func NewRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprintf("%p", &b)))
	}
	return hex.EncodeToString(b[:])
}

func CopyHeaders(in http.Header) map[string][]string {
	out := make(map[string][]string, len(in))
	for k, values := range in {
		out[k] = append([]string(nil), values...)
	}
	return out
}

type Pending[T any] struct {
	mu    sync.Mutex
	items map[string]chan T
}

func NewPending[T any]() *Pending[T] {
	return &Pending[T]{items: map[string]chan T{}}
}

func (p *Pending[T]) Add(id string) <-chan T {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch := make(chan T, 1)
	p.items[id] = ch
	return ch
}

func (p *Pending[T]) Complete(id string, value T) bool {
	p.mu.Lock()
	ch, ok := p.items[id]
	if ok {
		delete(p.items, id)
	}
	p.mu.Unlock()
	if ok {
		ch <- value
		close(ch)
	}
	return ok
}

func (p *Pending[T]) Delete(id string) {
	p.mu.Lock()
	ch, ok := p.items[id]
	if ok {
		delete(p.items, id)
		close(ch)
	}
	p.mu.Unlock()
}
