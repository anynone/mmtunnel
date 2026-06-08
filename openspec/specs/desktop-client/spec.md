# desktop-client Specification

## Purpose
TBD - created by archiving change add-desktop-client. Update Purpose after archive.
## Requirements
### Requirement: GUI server and client configuration
The desktop client SHALL allow users to configure server URL, client ID, token, and reconnect behavior through GUI forms without requiring manual YAML editing for normal operation.

#### Scenario: User creates first profile through setup wizard
- **WHEN** a user starts the desktop client with no existing profiles
- **THEN** the application presents a setup flow for entering profile name, server URL, client ID, token, and tunnel definitions

#### Scenario: User edits server settings through UI
- **WHEN** a user changes the server URL or client identity in the profile editor
- **THEN** the application persists the updated values through the desktop profile store

### Requirement: Profile management
The desktop client SHALL support multiple local profiles and one active profile.

#### Scenario: User switches active profile
- **WHEN** a user selects a different profile as active
- **THEN** the runtime uses that profile for subsequent start or restart operations

#### Scenario: User deletes inactive profile
- **WHEN** a user deletes a profile that is not active
- **THEN** the profile is removed from the local profile store

### Requirement: Tunnel form management
The desktop client SHALL allow users to create, edit, and delete tunnel definitions through GUI forms.

#### Scenario: User adds tunnel
- **WHEN** a user enters name, public path, target URL, and strip path settings and saves the form
- **THEN** the tunnel appears in the active profile tunnel list

#### Scenario: User edits tunnel
- **WHEN** a user changes an existing tunnel target URL and saves the form
- **THEN** the updated target URL is persisted in the active profile

#### Scenario: User deletes tunnel
- **WHEN** a user deletes a tunnel from a profile
- **THEN** the tunnel is removed from the profile and is not included in future runtime registration

### Requirement: Tunnel enabled toggle
The desktop client SHALL provide an enabled toggle for each tunnel and SHALL exclude disabled tunnels from runtime registration.

#### Scenario: Disabled tunnel is not registered
- **WHEN** a tunnel has `enabled: false` and the runtime starts
- **THEN** that tunnel is not included in the generated runtime tunnel registration

#### Scenario: Enabled tunnel is registered
- **WHEN** a tunnel has `enabled: true` and the runtime starts
- **THEN** that tunnel is included in the generated runtime tunnel registration

### Requirement: Tunnel toggle applies by reconnect in MVP
The desktop client MVP SHALL apply tunnel enabled state changes by restarting or reconnecting the client runtime and re-registering only enabled tunnels.

#### Scenario: User disables running tunnel
- **WHEN** the runtime is connected and a user toggles an enabled tunnel off
- **THEN** the application saves the profile, restarts the runtime, and reconnects without registering that tunnel

#### Scenario: User enables disabled tunnel
- **WHEN** the runtime is connected and a user toggles a disabled tunnel on
- **THEN** the application saves the profile, restarts the runtime, and reconnects with that tunnel included in registration

#### Scenario: Runtime shows reconnecting during toggle apply
- **WHEN** a tunnel enabled state change triggers runtime restart
- **THEN** the UI shows a reconnecting state until the runtime reaches connected, stopped, or error state

### Requirement: Runtime lifecycle controls
The desktop client SHALL provide start, stop, and restart controls for the local tunnel runtime.

#### Scenario: User starts runtime
- **WHEN** a user clicks Start for a valid active profile
- **THEN** the Go runtime connects to the configured tunnel server and registers enabled tunnels

#### Scenario: User stops runtime
- **WHEN** a user clicks Stop while the runtime is running
- **THEN** the runtime disconnects from the tunnel server and stops forwarding traffic

#### Scenario: User restarts runtime
- **WHEN** a user clicks Restart while the runtime is running
- **THEN** the runtime stops the active connection and starts again using the active profile

### Requirement: Runtime status display
The desktop client SHALL show runtime state and per-tunnel runtime status.

#### Scenario: Runtime connected
- **WHEN** the runtime successfully connects and registers enabled tunnels
- **THEN** the UI displays overall state `Connected`

#### Scenario: Tunnel registered
- **WHEN** an enabled tunnel is included in successful runtime registration
- **THEN** the UI displays that tunnel status as `Registered`

#### Scenario: Tunnel disabled
- **WHEN** a tunnel has `enabled: false`
- **THEN** the UI displays that tunnel status as `Disabled`

#### Scenario: Tunnel registration failed
- **WHEN** runtime registration fails for a tunnel or profile
- **THEN** the UI displays a failed status and an error message visible to the user

### Requirement: Connection testing
The desktop client SHALL allow users to test server reachability and client authentication from the GUI before starting the normal runtime.

#### Scenario: Server reachability test succeeds
- **WHEN** the server URL accepts a WebSocket upgrade
- **THEN** the GUI reports server reachability success

#### Scenario: Authentication test fails
- **WHEN** the server rejects the configured client ID and token
- **THEN** the GUI reports an authentication failure without registering business tunnels

### Requirement: Local daemon API
The desktop client SHALL expose a loopback-only local API for profile management, runtime lifecycle control, status, logs, and event streaming.

#### Scenario: UI requests status
- **WHEN** the Tauri UI calls the local status API
- **THEN** the Go daemon returns current runtime state, active profile, and tunnel statuses

#### Scenario: UI receives events
- **WHEN** runtime state or request activity changes
- **THEN** the Go daemon emits an event consumable by the Tauri UI

### Requirement: Event and activity display
The desktop client SHALL display recent connection events, request summaries, and errors without showing request or response bodies.

#### Scenario: Request completes
- **WHEN** a forwarded HTTP request completes
- **THEN** the UI can show method, path, status, duration, and tunnel name

#### Scenario: Request fails
- **WHEN** a forwarded request fails
- **THEN** the UI can show method, path, tunnel name, and error summary

### Requirement: Tray workflow
The desktop client SHALL provide a tray workflow with status visibility and quick runtime actions.

#### Scenario: User opens tray menu
- **WHEN** the user opens the tray menu
- **THEN** the menu shows current connection status and actions for opening the window, starting, stopping, restarting, and quitting

#### Scenario: User closes main window
- **WHEN** the user closes the main window without quitting the application
- **THEN** the application continues running from the tray if the runtime is active

### Requirement: CLI YAML import and export
The desktop client SHALL support importing existing CLI YAML into a GUI profile and exporting a GUI profile as CLI-compatible YAML.

#### Scenario: User imports CLI config
- **WHEN** a user imports an existing CLI client YAML file
- **THEN** the application creates a profile with matching server, client identity, token, and tunnel definitions

#### Scenario: User exports profile
- **WHEN** a user exports a profile as CLI YAML
- **THEN** the exported file can be used by the existing CLI tunnel client

