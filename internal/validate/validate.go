package validate

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const SystemPathPrefix = "/_tunnel"

func Duration(value string, fallback time.Duration) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", value, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("duration %q must be positive", value)
	}
	return parsed, nil
}

func Size(value string, fallback int64) (int64, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fallback, nil
	}
	multiplier := int64(1)
	units := []struct {
		suffix     string
		multiplier int64
	}{
		{"gb", 1024 * 1024 * 1024},
		{"mb", 1024 * 1024},
		{"kb", 1024},
		{"b", 1},
	}
	for _, unit := range units {
		if strings.HasSuffix(value, unit.suffix) {
			multiplier = unit.multiplier
			value = strings.TrimSpace(strings.TrimSuffix(value, unit.suffix))
			break
		}
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", value)
	}
	if number <= 0 {
		return 0, fmt.Errorf("size must be positive")
	}
	return number * multiplier, nil
}

func PublicPath(path string) error {
	if path == "" {
		return fmt.Errorf("public path is required")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("public path %q must start with /", path)
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return fmt.Errorf("public path %q must not end with /", path)
	}
	if IsReservedPath(path) {
		return fmt.Errorf("public path %q uses reserved system prefix %s", path, SystemPathPrefix)
	}
	return nil
}

func IsReservedPath(path string) bool {
	return path == SystemPathPrefix || strings.HasPrefix(path, SystemPathPrefix+"/")
}

func TargetURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %q: %w", value, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("target URL %q must use http or https", value)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("target URL %q must include host", value)
	}
	return parsed, nil
}

func ServerURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q: %w", value, err)
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return nil, fmt.Errorf("server URL %q must use ws or wss", value)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("server URL %q must include host", value)
	}
	return parsed, nil
}

func PathsConflict(a, b string) bool {
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
