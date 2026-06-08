package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mmtunnel/internal/config"
	"mmtunnel/internal/validate"
)

type TunnelStatus string

const (
	TunnelStatusDisabled   TunnelStatus = "disabled"
	TunnelStatusPending    TunnelStatus = "pending"
	TunnelStatusRegistered TunnelStatus = "registered"
	TunnelStatusFailed     TunnelStatus = "failed"
)

type Store struct {
	ActiveProfileID string    `json:"activeProfileId"`
	Profiles        []Profile `json:"profiles"`
}

type Profile struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Client   ClientIdentity `json:"client"`
	Tunnels  []Tunnel       `json:"tunnels"`
	Metadata Metadata       `json:"metadata,omitempty"`
}

type Metadata struct {
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type ClientIdentity struct {
	ID                string `json:"id"`
	Token             string `json:"token"`
	Server            string `json:"server"`
	ReconnectInterval string `json:"reconnectInterval,omitempty"`
}

type Tunnel struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	PublicPath string       `json:"publicPath"`
	Target     string       `json:"target"`
	StripPath  bool         `json:"stripPath"`
	Enabled    bool         `json:"enabled"`
	Status     TunnelStatus `json:"status,omitempty"`
	LastError  string       `json:"lastError,omitempty"`
}

func Load(path string) (Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Store{}, nil
		}
		return Store{}, err
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return Store{}, err
	}
	return store, store.Validate()
}

func (s Store) Save(path string) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (s Store) Validate() error {
	seenProfiles := map[string]bool{}
	for _, profile := range s.Profiles {
		if profile.ID == "" {
			return fmt.Errorf("profile id is required")
		}
		if seenProfiles[profile.ID] {
			return fmt.Errorf("duplicate profile id %s", profile.ID)
		}
		seenProfiles[profile.ID] = true
		if profile.Client.ID == "" || profile.Client.Token == "" || profile.Client.Server == "" {
			return fmt.Errorf("profile %s client id, token, and server are required", profile.ID)
		}
		if _, err := validate.ServerURL(profile.Client.Server); err != nil {
			return err
		}
		enabledPaths := []string{}
		for _, tunnel := range profile.Tunnels {
			if tunnel.ID == "" || tunnel.Name == "" {
				return fmt.Errorf("profile %s tunnel id and name are required", profile.ID)
			}
			if err := validate.PublicPath(tunnel.PublicPath); err != nil {
				return err
			}
			if _, err := validate.TargetURL(tunnel.Target); err != nil {
				return err
			}
			if tunnel.Enabled {
				for _, existing := range enabledPaths {
					if validate.PathsConflict(existing, tunnel.PublicPath) {
						return fmt.Errorf("enabled tunnel path conflict between %s and %s", existing, tunnel.PublicPath)
					}
				}
				enabledPaths = append(enabledPaths, tunnel.PublicPath)
			}
		}
	}
	if s.ActiveProfileID != "" && !seenProfiles[s.ActiveProfileID] {
		return fmt.Errorf("active profile %s does not exist", s.ActiveProfileID)
	}
	return nil
}

func (s Store) ActiveProfile() (Profile, error) {
	for _, profile := range s.Profiles {
		if profile.ID == s.ActiveProfileID {
			return profile, nil
		}
	}
	return Profile{}, fmt.Errorf("active profile %s does not exist", s.ActiveProfileID)
}

func (s Store) ActiveClientConfig() (config.ClientConfig, map[string]TunnelStatus, error) {
	profile, err := s.ActiveProfile()
	if err != nil {
		return config.ClientConfig{}, nil, err
	}
	return profile.ClientConfig()
}

func (s Store) ActiveClientRuntimeConfig() (config.ClientConfig, error) {
	cfg, _, err := s.ActiveClientConfig()
	return cfg, err
}

func (s Store) ActiveProfileIDValue() string {
	return s.ActiveProfileID
}

func (p Profile) ClientConfig() (config.ClientConfig, map[string]TunnelStatus, error) {
	cfg := config.DefaultClient()
	cfg.ID = p.Client.ID
	cfg.Token = p.Client.Token
	cfg.Server = p.Client.Server
	if p.Client.ReconnectInterval != "" {
		parsed, err := validate.Duration(p.Client.ReconnectInterval, cfg.ReconnectInterval)
		if err != nil {
			return cfg, nil, err
		}
		cfg.ReconnectInterval = parsed
	}
	statuses := map[string]TunnelStatus{}
	for _, tunnel := range p.Tunnels {
		if !tunnel.Enabled {
			statuses[tunnel.ID] = TunnelStatusDisabled
			continue
		}
		cfg.Tunnels = append(cfg.Tunnels, config.TunnelConfig{
			Name:       tunnel.Name,
			PublicPath: tunnel.PublicPath,
			Target:     tunnel.Target,
			StripPath:  tunnel.StripPath,
		})
		statuses[tunnel.ID] = TunnelStatusPending
	}
	if len(cfg.Tunnels) == 0 {
		return cfg, statuses, nil
	}
	return cfg, statuses, nil
}

func ImportCLIYAML(id, name, path string) (Profile, error) {
	cfg, err := config.LoadClient(path)
	if err != nil {
		return Profile{}, err
	}
	profile := Profile{
		ID:   id,
		Name: name,
		Client: ClientIdentity{
			ID:                cfg.ID,
			Token:             cfg.Token,
			Server:            cfg.Server,
			ReconnectInterval: cfg.ReconnectInterval.String(),
		},
	}
	for _, tunnel := range cfg.Tunnels {
		profile.Tunnels = append(profile.Tunnels, Tunnel{
			ID:         safeID(tunnel.Name),
			Name:       tunnel.Name,
			PublicPath: tunnel.PublicPath,
			Target:     tunnel.Target,
			StripPath:  tunnel.StripPath,
			Enabled:    true,
		})
	}
	return profile, nil
}

func (s Store) ExportCLIYAML(profileID, path string) error {
	var selected *Profile
	for i := range s.Profiles {
		if s.Profiles[i].ID == profileID {
			selected = &s.Profiles[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("profile %s does not exist", profileID)
	}
	cfg, _, err := selected.ClientConfig()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("client:\n")
	b.WriteString(fmt.Sprintf("  id: %s\n", cfg.ID))
	b.WriteString(fmt.Sprintf("  token: %q\n", cfg.Token))
	b.WriteString(fmt.Sprintf("  server: %q\n", cfg.Server))
	b.WriteString(fmt.Sprintf("  reconnectInterval: %s\n\n", cfg.ReconnectInterval))
	b.WriteString("tunnels:\n")
	for _, tunnel := range cfg.Tunnels {
		b.WriteString(fmt.Sprintf("  - name: %s\n", tunnel.Name))
		b.WriteString(fmt.Sprintf("    publicPath: %s\n", tunnel.PublicPath))
		b.WriteString(fmt.Sprintf("    target: %s\n", tunnel.Target))
		b.WriteString(fmt.Sprintf("    stripPath: %t\n", tunnel.StripPath))
	}
	return os.WriteFile(path, []byte(b.String()), 0600)
}

func safeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return "tunnel"
	}
	return value
}
