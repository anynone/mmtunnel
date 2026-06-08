package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"mmsocket/internal/logging"
	"mmsocket/internal/validate"
)

type ServerConfig struct {
	Listen               string
	LogLevel             string
	SystemPathPrefix     string
	RequestTimeout       time.Duration
	TargetConnectTimeout time.Duration
	TunnelIdleTimeout    time.Duration
	PingInterval         time.Duration
	PongTimeout          time.Duration
	MaxRequestBodySize   int64
	Clients              []TrustedClient
}

type TrustedClient struct {
	ID    string
	Token string
}

type ClientConfig struct {
	ID                string
	Token             string
	Server            string
	ReconnectInterval time.Duration
	Tunnels           []TunnelConfig
}

type TunnelConfig struct {
	Name       string
	PublicPath string
	Target     string
	StripPath  bool
}

func DefaultServer() ServerConfig {
	return ServerConfig{
		Listen:               ":8080",
		LogLevel:             string(logging.Info),
		SystemPathPrefix:     validate.SystemPathPrefix,
		RequestTimeout:       60 * time.Second,
		TargetConnectTimeout: 10 * time.Second,
		TunnelIdleTimeout:    60 * time.Second,
		PingInterval:         30 * time.Second,
		PongTimeout:          10 * time.Second,
		MaxRequestBodySize:   100 * 1024 * 1024,
	}
}

func DefaultClient() ClientConfig {
	return ClientConfig{ReconnectInterval: 5 * time.Second}
}

func LoadServer(path string) (ServerConfig, error) {
	cfg := DefaultServer()
	doc, err := parseSimpleYAML(path)
	if err != nil {
		return cfg, err
	}
	if v := doc.scalar("server.listen"); v != "" {
		cfg.Listen = v
	}
	if v := doc.scalar("server.logLevel"); v != "" {
		cfg.LogLevel = v
	}
	if v := doc.scalar("server.systemPathPrefix"); v != "" {
		cfg.SystemPathPrefix = v
	}
	if v := doc.scalar("server.requestTimeout"); v != "" {
		cfg.RequestTimeout, err = validate.Duration(v, cfg.RequestTimeout)
		if err != nil {
			return cfg, err
		}
	}
	if v := doc.scalar("server.targetConnectTimeout"); v != "" {
		cfg.TargetConnectTimeout, err = validate.Duration(v, cfg.TargetConnectTimeout)
		if err != nil {
			return cfg, err
		}
	}
	if v := doc.scalar("server.tunnelIdleTimeout"); v != "" {
		cfg.TunnelIdleTimeout, err = validate.Duration(v, cfg.TunnelIdleTimeout)
		if err != nil {
			return cfg, err
		}
	}
	if v := doc.scalar("server.pingInterval"); v != "" {
		cfg.PingInterval, err = validate.Duration(v, cfg.PingInterval)
		if err != nil {
			return cfg, err
		}
	}
	if v := doc.scalar("server.pongTimeout"); v != "" {
		cfg.PongTimeout, err = validate.Duration(v, cfg.PongTimeout)
		if err != nil {
			return cfg, err
		}
	}
	if v := doc.scalar("server.maxRequestBodySize"); v != "" {
		cfg.MaxRequestBodySize, err = validate.Size(v, cfg.MaxRequestBodySize)
		if err != nil {
			return cfg, err
		}
	}
	cfg.Clients = make([]TrustedClient, 0, len(doc.lists["clients"]))
	for _, item := range doc.lists["clients"] {
		client := TrustedClient{ID: item["id"], Token: item["token"]}
		if client.ID == "" || client.Token == "" {
			return cfg, fmt.Errorf("clients entries require id and token")
		}
		cfg.Clients = append(cfg.Clients, client)
	}
	if _, err := logging.ParseLevel(cfg.LogLevel); err != nil {
		return cfg, err
	}
	if cfg.SystemPathPrefix != validate.SystemPathPrefix {
		return cfg, fmt.Errorf("only systemPathPrefix %s is supported", validate.SystemPathPrefix)
	}
	return cfg, nil
}

func LoadClient(path string) (ClientConfig, error) {
	cfg := DefaultClient()
	doc, err := parseSimpleYAML(path)
	if err != nil {
		return cfg, err
	}
	cfg.ID = doc.scalar("client.id")
	cfg.Token = doc.scalar("client.token")
	cfg.Server = doc.scalar("client.server")
	if v := doc.scalar("client.reconnectInterval"); v != "" {
		cfg.ReconnectInterval, err = validate.Duration(v, cfg.ReconnectInterval)
		if err != nil {
			return cfg, err
		}
	}
	if cfg.ID == "" || cfg.Token == "" || cfg.Server == "" {
		return cfg, fmt.Errorf("client id, token, and server are required")
	}
	if _, err := validate.ServerURL(cfg.Server); err != nil {
		return cfg, err
	}
	for _, item := range doc.lists["tunnels"] {
		tunnel := TunnelConfig{
			Name:       item["name"],
			PublicPath: item["publicPath"],
			Target:     item["target"],
			StripPath:  true,
		}
		if raw, ok := item["stripPath"]; ok {
			tunnel.StripPath = strings.EqualFold(raw, "true")
		}
		if err := ValidateTunnel(tunnel); err != nil {
			return cfg, err
		}
		cfg.Tunnels = append(cfg.Tunnels, tunnel)
	}
	if len(cfg.Tunnels) == 0 {
		return cfg, fmt.Errorf("at least one tunnel is required")
	}
	for i := range cfg.Tunnels {
		for j := i + 1; j < len(cfg.Tunnels); j++ {
			if validate.PathsConflict(cfg.Tunnels[i].PublicPath, cfg.Tunnels[j].PublicPath) {
				return cfg, fmt.Errorf("tunnel path conflict between %s and %s", cfg.Tunnels[i].PublicPath, cfg.Tunnels[j].PublicPath)
			}
		}
	}
	return cfg, nil
}

func ValidateTunnel(tunnel TunnelConfig) error {
	if tunnel.Name == "" {
		return fmt.Errorf("tunnel name is required")
	}
	if err := validate.PublicPath(tunnel.PublicPath); err != nil {
		return err
	}
	if _, err := validate.TargetURL(tunnel.Target); err != nil {
		return err
	}
	return nil
}

type yamlDoc struct {
	scalars map[string]string
	lists   map[string][]map[string]string
}

func (d yamlDoc) scalar(key string) string {
	return d.scalars[key]
}

func parseSimpleYAML(path string) (yamlDoc, error) {
	file, err := os.Open(path)
	if err != nil {
		return yamlDoc{}, err
	}
	defer file.Close()

	doc := yamlDoc{scalars: map[string]string{}, lists: map[string][]map[string]string{}}
	var section string
	var currentList string
	var currentItem map[string]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(strings.Split(raw, "#")[0])
		if line == "" {
			continue
		}
		if !strings.Contains(line, ":") && !strings.HasPrefix(line, "- ") {
			return doc, fmt.Errorf("invalid yaml line: %s", raw)
		}
		if !strings.HasPrefix(raw, " ") && strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			currentList = ""
			currentItem = nil
			continue
		}
		if !strings.HasPrefix(raw, " ") {
			key, value := splitKV(line)
			if value == "" {
				currentList = key
				doc.lists[currentList] = nil
				currentItem = nil
			} else {
				doc.scalars[key] = value
			}
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(raw, " "), "- ") {
			if section != "" && section != "clients" && section != "tunnels" {
				currentList = section
			} else if currentList == "" {
				currentList = section
			}
			currentItem = map[string]string{}
			doc.lists[currentList] = append(doc.lists[currentList], currentItem)
			itemLine := strings.TrimPrefix(strings.TrimSpace(raw), "- ")
			if itemLine != "" {
				key, value := splitKV(itemLine)
				currentItem[key] = value
			}
			continue
		}
		key, value := splitKV(line)
		if currentItem != nil {
			currentItem[key] = value
			continue
		}
		if section != "" {
			doc.scalars[section+"."+key] = value
		} else {
			doc.scalars[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return doc, err
	}
	return doc, nil
}

func splitKV(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	key := strings.TrimSpace(parts[0])
	value := ""
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
	}
	return key, value
}
