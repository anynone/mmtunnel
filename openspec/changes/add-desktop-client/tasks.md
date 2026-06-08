## 1. Runtime Refactor

- [x] 1.1 Extract the existing CLI client loop into a reusable runtime package with start, stop, restart, and status operations.
- [x] 1.2 Add runtime states for stopped, connecting, connected, reconnecting, stopping, and error.
- [x] 1.3 Preserve existing CLI behavior by making `cmd/tunnel-client` use the runtime package.
- [x] 1.4 Add runtime event emission for connection state changes, registration results, request summaries, and errors.
- [x] 1.5 Add lightweight in-memory event and log buffers with bounded retention.

## 2. Profile Store

- [x] 2.1 Define desktop profile models for profile metadata, server URL, client identity, token, reconnect settings, and tunnels.
- [x] 2.2 Add per-tunnel `enabled` and runtime status fields to the desktop profile/runtime model.
- [x] 2.3 Implement local profile persistence with active profile support.
- [x] 2.4 Validate profile server URL, client identity, token, public paths, target URLs, and enabled tunnel conflicts.
- [x] 2.5 Implement conversion from active desktop profile to existing `config.ClientConfig`, filtering out disabled tunnels.
- [x] 2.6 Implement import from existing CLI YAML into a desktop profile.
- [x] 2.7 Implement export from a desktop profile to CLI-compatible YAML.

## 3. Local Daemon API

- [x] 3.1 Add a Go desktop daemon command that binds only to loopback.
- [x] 3.2 Implement profile CRUD APIs.
- [x] 3.3 Implement active profile selection API.
- [x] 3.4 Implement runtime start, stop, and restart APIs.
- [x] 3.5 Implement runtime status API with active profile and per-tunnel statuses.
- [x] 3.6 Implement server reachability test API without registering business tunnels.
- [x] 3.7 Implement authentication test API without registering business tunnels.
- [x] 3.8 Implement events stream API for runtime status, logs, and request summaries.
- [x] 3.9 Implement logs API for recent bounded event history.

## 4. Tauri Desktop Shell

- [x] 4.1 Create the Tauri desktop app structure and configure it to start or connect to the Go daemon.
- [x] 4.2 Implement first-run setup flow for profile name, server URL, client ID, token, and initial tunnels.
- [x] 4.3 Implement profile list, create, edit, delete, and active profile selection screens.
- [x] 4.4 Implement server/client settings editor with masked token input.
- [x] 4.5 Implement tunnel list screen with add, edit, delete, strip path, and enabled toggle controls.
- [x] 4.6 Implement tunnel enabled toggle behavior that saves the profile and triggers runtime restart when running.
- [x] 4.7 Implement runtime controls for start, stop, and restart.
- [x] 4.8 Implement runtime status display and per-tunnel status badges.
- [x] 4.9 Implement recent activity and error display from daemon events.
- [x] 4.10 Implement import and export flows for CLI YAML compatibility.

## 5. Tray Workflow

- [x] 5.1 Add tray menu with current connection status.
- [x] 5.2 Add tray actions for open window, start, stop, restart, and quit.
- [x] 5.3 Keep the runtime active when the main window closes but the tray app remains running.
- [x] 5.4 Ensure Quit stops the runtime and exits the daemon/UI cleanly.

## 6. Documentation

- [x] 6.1 Document the desktop architecture and local daemon API.
- [x] 6.2 Document profile storage location and plaintext token caveat for MVP.
- [x] 6.3 Document CLI YAML import/export behavior.
- [x] 6.4 Document tunnel enabled toggle behavior and the MVP reconnect trade-off.
- [x] 6.5 Document desktop build and packaging commands for the supported MVP platforms.

## 7. Verification

- [x] 7.1 Add unit tests for profile validation, enabled tunnel filtering, and CLI YAML import/export.
- [x] 7.2 Add runtime tests for start, stop, restart, state transitions, and event emission.
- [x] 7.3 Add daemon API tests for profile CRUD, runtime status, start/stop/restart, test-server, and test-auth endpoints.
- [x] 7.4 Add tests proving disabled tunnels are not included in runtime registration.
- [x] 7.5 Add tests proving toggling a tunnel while running triggers runtime restart in MVP.
- [x] 7.6 Add frontend tests for setup, profile editor, tunnel editor, enabled toggle, and runtime controls where practical.
- [x] 7.7 Run OpenSpec validation, Go tests, and frontend build/test commands before marking the change ready to archive.
