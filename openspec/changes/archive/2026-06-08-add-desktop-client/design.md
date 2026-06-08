## Context

The project currently provides a Go tunnel server and a Go CLI tunnel client. The CLI client loads a YAML file, connects to the tunnel server, registers tunnels, and forwards HTTP/WebSocket traffic. That is sufficient for server-side and script-driven usage, but desktop users still need to edit configuration files and inspect terminal logs.

This change adds a desktop client experience. The desktop client should be a local control plane for the existing tunnel client core: users configure profiles, server connection settings, credentials, and tunnels through GUI forms; a Go runtime manages connection lifecycle and tunnel registration.

```
Tauri Desktop UI
  |
  | Local HTTP API + SSE events
  v
Go Local Daemon / Runtime
  |
  | builds ClientConfig from active profile
  v
Existing Go Tunnel Client Core
  |
  | WebSocket tunnel
  v
Tunnel Server
```

## Goals / Non-Goals

**Goals:**

- Provide a visual desktop client for configuring and running tunnels without manual YAML editing.
- Use Tauri for the desktop UI and a Go local daemon/API for runtime control.
- Reuse the existing Go tunnel client core rather than duplicating forwarding logic.
- Support GUI-managed profiles with server URL, client ID, token, and tunnel definitions.
- Support multiple profiles and an active profile.
- Provide per-tunnel enabled toggles in the GUI.
- Apply enabled/disabled tunnel changes in the MVP by restarting the client runtime and re-registering only enabled tunnels.
- Show connection state, tunnel runtime status, recent events, recent request summaries, and recent errors.
- Provide tray actions for open window, start, stop, restart, and quit.
- Preserve compatibility with the existing CLI YAML via import/export.

**Non-Goals:**

- Managing the remote tunnel server from the desktop client.
- Changing the tunnel server protocol to support dynamic single-tunnel registration or unregistration in the MVP.
- Public endpoint access control.
- Long-term request log storage or analytics.
- Request/response body inspection.
- System credential storage for tokens in the MVP.
- Automatic updates or package signing in the MVP.

## Decisions

### Use Tauri UI with a Go local daemon/API

The desktop application will use Tauri for the window, tray, and frontend UI. A Go local daemon will expose localhost-only APIs for profile management, runtime control, status, and event streaming.

Rationale: Tauri is a good fit for a lightweight desktop shell, while Go should remain the owner of tunnel runtime behavior. A local API boundary avoids making the frontend parse CLI stdout or directly manage low-level networking.

Alternative considered: pure Go GUI with Wails or Fyne. This would reduce the number of technologies but would make frontend UI iteration less flexible. Alternative considered: Tauri directly spawning the existing CLI and parsing logs. This is simpler initially but becomes brittle for status, events, and configuration editing.

### Refactor client core into a controllable runtime

The existing client loop will be wrapped by a runtime layer that supports start, stop, restart, status, and events.

Rationale: CLI behavior is currently blocking and signal-driven. Desktop behavior needs explicit lifecycle control, observable state transitions, and restart support when profile or tunnel settings change.

The runtime state model will include at least:

```
Stopped -> Connecting -> Connected
Connected -> Reconnecting -> Connected
Any running state -> Stopped
Any running state -> Error
```

### Use GUI-managed profiles as the primary configuration model

The desktop client will store profile data in a GUI-owned profile store. Users will create and edit profiles through forms, not by editing YAML files.

Rationale: the desktop client should reduce configuration mistakes and make server/client/tunnel setup discoverable. The existing CLI YAML remains useful for automation and compatibility, but it is not the main desktop interaction.

MVP token storage will remain in the local profile store as plaintext. The UI must treat token fields as secrets by masking them by default. System credential storage can be added later without changing the profile-level user model.

### Keep CLI YAML compatibility through import/export

The desktop client will import existing CLI YAML into a profile and export a profile to CLI YAML.

Rationale: this preserves the current workflow and lets users move between CLI and desktop usage without hand-recreating tunnels.

### Apply tunnel enabled toggles by runtime restart in MVP

Each tunnel will have a desired `enabled` setting. Disabled tunnels are kept in the profile but are not included in the generated runtime `ClientConfig`.

For MVP, changing any tunnel enabled state will save the profile and restart the runtime if it is running. The reconnected client will register only enabled tunnels.

Rationale: this avoids changing the server protocol while still giving users a practical GUI switch for whether a public route can reach the local target.

Trade-off: all tunnels are briefly unavailable during restart, and in-flight HTTP/WebSocket sessions can be interrupted. The UI must show a reconnecting state and make this behavior explicit.

### Distinguish desired enabled state from runtime status

The tunnel table will show both the user's desired enabled setting and the actual runtime status.

Examples:

- `Enabled: on`, `Status: registered`
- `Enabled: off`, `Status: disabled`
- `Enabled: on`, `Status: pending`
- `Enabled: on`, `Status: failed`

Rationale: a tunnel can be enabled in configuration but fail registration due to path conflict or connection failure. The UI should not imply that the route is active until runtime confirms it.

### Expose local API and event stream

The Go daemon will expose localhost-only endpoints for profile CRUD, connection testing, runtime control, status, logs, and events. Runtime events will be streamed to the UI, preferably with server-sent events.

Rationale: status updates, request summaries, and log messages are event-driven. SSE is simple, browser-native, and adequate for one local UI consumer.

Representative API surface:

```
GET    /api/status
GET    /api/profiles
POST   /api/profiles
GET    /api/profiles/{id}
PUT    /api/profiles/{id}
DELETE /api/profiles/{id}
POST   /api/profiles/{id}/test-server
POST   /api/profiles/{id}/test-auth
POST   /api/runtime/start
POST   /api/runtime/stop
POST   /api/runtime/restart
GET    /api/events
GET    /api/logs
```

### Provide connection tests from the GUI

The GUI will support testing server reachability and client authentication before starting normal runtime.

Rationale: the most common setup failures are wrong `ws`/`wss`, reverse proxy redirects, server reachability, and invalid token. A form-level test gives direct feedback without requiring users to inspect terminal output.

MVP connection tests should not register real business tunnels, to avoid creating path conflicts or disrupting an active profile.

## Risks / Trade-offs

- Plaintext token storage -> Mask token fields in the UI, keep file permissions restrictive where possible, and defer system credential storage to a follow-up change.
- Tauri plus Go daemon adds build complexity -> Keep the daemon API small and isolate desktop-specific code from the tunnel forwarding core.
- Restart-based tunnel toggles interrupt all tunnels -> Show reconnecting status and document that MVP toggles can interrupt in-flight requests.
- Local API exposure can become a security surface -> Bind to loopback only, use a random local port when practical, and avoid listening on external interfaces.
- Profile store and CLI YAML can drift -> Treat GUI profile store as canonical for desktop usage and provide explicit import/export actions.
- Request event capture can become too detailed -> MVP stores only lightweight metadata such as method, path, status, duration, tunnel name, and error, never request or response bodies.

## Migration Plan

This is an additive desktop client feature.

1. Refactor existing client runtime to support lifecycle control and event emission while preserving CLI behavior.
2. Add profile store and import/export compatibility for existing CLI YAML.
3. Add Go local daemon/API for profile management, runtime control, status, and event streaming.
4. Add Tauri desktop UI and tray shell.
5. Add packaging/build documentation for supported desktop platforms.

Rollback consists of continuing to use the existing CLI client. The tunnel server protocol and existing CLI YAML workflow remain compatible.

## Open Questions

- Which desktop platforms should be packaged first: Windows, Linux, macOS, or a smaller subset?
- Whether token storage should move to OS credential storage in the first production release or remain a follow-up after MVP.
- Whether the Go daemon should run as a Tauri sidecar process only, or also support standalone execution for debugging and automation.
