# Desktop Client

The desktop client is a local control plane for the existing Go tunnel client.

## Architecture

```text
Tauri desktop UI
  |
  | http://127.0.0.1:19081
  | server-sent events
  v
Go tunnel daemon
  |
  | builds config.ClientConfig from active GUI profile
  v
Existing Go tunnel client core
  |
  v
Tunnel server
```

The desktop UI does not require users to edit YAML for normal operation. Users configure the server URL, client ID, token, reconnect interval, and tunnels through forms. The daemon persists those settings in a local profile store.

## Profile Store

By default the daemon stores profiles at:

```text
$XDG_CONFIG_HOME/mmtunnel/profiles.json
```

or the platform-specific directory returned by Go's `os.UserConfigDir()`.

Override the location:

```bash
./bin/tunnel-daemon -profiles /path/to/profiles.json
```

MVP token storage is plaintext in the profile store. The UI masks token fields, but this is not equivalent to OS credential storage.

## Running Locally

Build the daemon:

```bash
GOCACHE=/tmp/go-build go build -buildvcs=false -o bin/tunnel-daemon ./cmd/tunnel-daemon
```

Start the daemon:

```bash
./bin/tunnel-daemon -listen 127.0.0.1:19081
```

Build the frontend:

```bash
cd desktop
npm run build
```

Run frontend tests:

```bash
cd desktop
npm test
```

## Tauri

The Tauri scaffold lives in `desktop/src-tauri`. It is configured to use the built frontend in `desktop/dist` and to bundle `desktop/src-tauri/binaries/tunnel-daemon-<target-triple>` as an external sidecar binary.

Typical commands after installing the Tauri CLI and dependencies:

```bash
cd desktop
npm run build
npm run tauri:dev
npm run tauri:build
```

## Tunnel Enabled Toggle

Each tunnel has a configured `enabled` value and a runtime status.

```text
enabled=true  -> included in runtime registration
enabled=false -> kept in profile, not registered
```

For MVP, changing a tunnel enabled state while the runtime is running restarts the runtime. During restart, all tunnels may be briefly unavailable and in-flight HTTP/WebSocket sessions may be interrupted.

## CLI YAML Compatibility

The desktop client can import an existing CLI YAML file into a profile and export a profile as CLI-compatible YAML. Disabled desktop tunnels are omitted from exported CLI YAML.
