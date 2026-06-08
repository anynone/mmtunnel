## Why

The current tunnel client is usable from the command line, but day-to-day desktop use still requires editing YAML files and watching terminal logs. A visual client should let developers configure server connection details, client identity, tokens, tunnels, and runtime state from a desktop UI while reusing the existing Go tunnel core.

## What Changes

- Add a desktop client application based on Tauri for the UI and a Go local daemon/API for tunnel runtime control.
- Add a GUI-first profile model for configuring server URL, client ID, token, and tunnels without requiring manual YAML editing.
- Add multi-profile support so users can switch between company, local, and test tunnel configurations.
- Add per-tunnel enabled toggles in the UI.
- For MVP, apply tunnel enabled changes by restarting/reconnecting the client runtime and registering only enabled tunnels.
- Add runtime controls for start, stop, restart, and status.
- Add read-only runtime event visibility for connection state, tunnel registration state, recent request summaries, and recent errors.
- Add local profile persistence managed by the desktop client.
- Add optional import/export compatibility with the existing CLI YAML format.
- Add a tray-oriented desktop workflow with status visibility and quick runtime actions.

## Capabilities

### New Capabilities

- `desktop-client`: Defines the visual desktop client, including GUI configuration, profile management, tunnel enablement, local daemon/API, runtime controls, event display, and tray behavior.

### Modified Capabilities

- `http-tunnel`: Clarifies that CLI clients can continue loading YAML directly while desktop clients may generate runtime tunnel configuration from GUI-managed profiles.

## Impact

- Introduces a Tauri desktop application and associated frontend code.
- Introduces a Go local daemon/API that wraps the existing client runtime.
- Requires refactoring the existing client logic into a controllable runtime with observable state and events.
- Adds a profile store distinct from the current CLI YAML config while preserving import/export compatibility.
- Adds local-only APIs for profile CRUD, runtime lifecycle control, connection testing, status, and event streaming.
- Does not require tunnel server protocol changes for the MVP tunnel enable/disable behavior.
