package protocol

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const acceptSalt = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type WSConn struct {
	netConn net.Conn
	reader  *bufio.Reader
	mu      sync.Mutex
	masked  bool
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WSConn, error) {
	if !IsWebSocketRequest(r) {
		return nil, fmt.Errorf("request is not websocket upgrade")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, fmt.Errorf("missing Sec-WebSocket-Key")
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("response writer does not support hijacking")
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}
	accept := websocketAccept(key)
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := conn.Write([]byte(response)); err != nil {
		conn.Close()
		return nil, err
	}
	return &WSConn{netConn: conn, reader: rw.Reader, masked: false}, nil
}

func Dial(rawURL string, timeout time.Duration) (*WSConn, *http.Response, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, err
	}
	portHost := parsed.Host
	if !strings.Contains(parsed.Host, ":") {
		if parsed.Scheme == "wss" {
			portHost += ":443"
		} else {
			portHost += ":80"
		}
	}
	var conn net.Conn
	if parsed.Scheme == "wss" {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", portHost, &tls.Config{ServerName: parsed.Hostname()})
	} else {
		conn, err = net.DialTimeout("tcp", portHost, timeout)
	}
	if err != nil {
		return nil, nil, err
	}
	key := randomKey()
	path := parsed.RequestURI()
	if path == "" {
		path = "/"
	}
	req := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + parsed.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, nil, err
	}
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: http.MethodGet})
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, resp, fmt.Errorf("websocket upgrade failed: %s", resp.Status)
	}
	if resp.Header.Get("Sec-WebSocket-Accept") != websocketAccept(key) {
		conn.Close()
		return nil, resp, fmt.Errorf("invalid websocket accept header")
	}
	return &WSConn{netConn: conn, reader: reader, masked: true}, resp, nil
}

func IsWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") && strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func (c *WSConn) ReadMessage() (int, []byte, error) {
	for {
		opcode, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch opcode {
		case OpcodePing:
			_ = c.WriteMessage(OpcodePong, payload)
		case OpcodePong:
			return opcode, payload, nil
		default:
			return opcode, payload, nil
		}
	}
}

func (c *WSConn) WriteMessage(opcode int, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeFrame(opcode, payload)
}

func (c *WSConn) Close() error {
	_ = c.WriteMessage(OpcodeClose, nil)
	return c.netConn.Close()
}

func (c *WSConn) SetReadDeadline(t time.Time) error {
	return c.netConn.SetReadDeadline(t)
}

func (c *WSConn) SetWriteDeadline(t time.Time) error {
	return c.netConn.SetWriteDeadline(t)
}

func (c *WSConn) readFrame() (int, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.reader, header); err != nil {
		return 0, nil, err
	}
	opcode := int(header[0] & 0x0f)
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7f)
	if length == 126 {
		var ext [2]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	} else if length == 127 {
		var ext [8]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext[:])
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(c.reader, mask[:]); err != nil {
			return 0, nil, err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, payload, nil
}

func (c *WSConn) writeFrame(opcode int, payload []byte) error {
	header := []byte{0x80 | byte(opcode)}
	length := len(payload)
	maskBit := byte(0)
	if c.masked {
		maskBit = 0x80
	}
	if length < 126 {
		header = append(header, maskBit|byte(length))
	} else if length <= 65535 {
		header = append(header, maskBit|126)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	} else {
		header = append(header, maskBit|127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}
	if c.masked {
		var mask [4]byte
		if _, err := rand.Read(mask[:]); err != nil {
			return err
		}
		header = append(header, mask[:]...)
		maskedPayload := append([]byte(nil), payload...)
		for i := range maskedPayload {
			maskedPayload[i] ^= mask[i%4]
		}
		payload = maskedPayload
	}
	if _, err := c.netConn.Write(header); err != nil {
		return err
	}
	_, err := c.netConn.Write(payload)
	return err
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + acceptSalt))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func randomKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "AAAAAAAAAAAAAAAAAAAAAA=="
	}
	return base64.StdEncoding.EncodeToString(b[:])
}
