## MODIFIED Requirements

### Requirement: GUI server and client configuration
The desktop client SHALL provide a React and shadcn/ui-based GUI that allows users to configure server URL, client ID, token, and reconnect behavior through forms without requiring manual YAML editing for normal operation.

#### Scenario: User creates first profile through setup wizard
- **WHEN** a user starts the desktop client with no existing profiles
- **THEN** the application presents a setup flow for entering profile name, server URL, client ID, token, and tunnel definitions

#### Scenario: User edits server settings through UI
- **WHEN** a user changes the server URL or client identity in the profile editor
- **THEN** the application persists the updated values through the desktop profile store

#### Scenario: User edits token through UI
- **WHEN** a user edits the client token
- **THEN** the token input is masked by default and the saved profile contains the updated token value

### Requirement: Profile management
The desktop client SHALL support multiple local profiles, one active profile, profile switching, profile creation, profile editing, and inactive profile deletion through the GUI.

#### Scenario: User switches active profile
- **WHEN** a user selects a different profile as active
- **THEN** the runtime uses that profile for subsequent start or restart operations

#### Scenario: User deletes inactive profile
- **WHEN** a user deletes a profile that is not active
- **THEN** the profile is removed from the local profile store

#### Scenario: User attempts to delete active profile
- **WHEN** a user views the active profile
- **THEN** the UI prevents accidental active profile deletion or requires the user to switch active profiles first

### Requirement: Tunnel form management
The desktop client SHALL allow users to create, edit, and delete tunnel definitions through GUI forms, including name, public path, target URL, strip path, and enabled state.

#### Scenario: User adds tunnel
- **WHEN** a user enters name, public path, target URL, strip path settings, and enabled state and saves the form
- **THEN** the tunnel appears in the active profile tunnel list

#### Scenario: User edits tunnel
- **WHEN** a user changes an existing tunnel target URL or strip path setting and saves the form
- **THEN** the updated tunnel definition is persisted in the active profile

#### Scenario: User deletes tunnel
- **WHEN** a user deletes a tunnel from a profile
- **THEN** the tunnel is removed from the profile and is not included in future runtime registration

### Requirement: Connection testing
The desktop client SHALL allow users to test server reachability and client authentication from the GUI before starting the normal runtime and SHALL show visible success or failure feedback for each test.

#### Scenario: Server reachability test succeeds
- **WHEN** the server URL accepts a WebSocket upgrade
- **THEN** the GUI reports server reachability success

#### Scenario: Authentication test fails
- **WHEN** the server rejects the configured client ID and token
- **THEN** the GUI reports an authentication failure without registering business tunnels

#### Scenario: Connection test fails
- **WHEN** a reachability or authentication test returns an error
- **THEN** the GUI displays an error summary visible to the user

### Requirement: Event and activity display
The desktop client SHALL display recent connection events, request summaries, and errors in the GUI without showing request or response bodies.

#### Scenario: Request completes
- **WHEN** a forwarded HTTP request completes
- **THEN** the UI can show method, path, status, duration, and tunnel name

#### Scenario: Request fails
- **WHEN** a forwarded HTTP request fails
- **THEN** the UI can show method, path, tunnel name, and error summary

#### Scenario: Runtime event occurs
- **WHEN** the daemon emits a runtime state, log, request, or error event
- **THEN** the UI adds the event to the recent activity display

## ADDED Requirements

### Requirement: React shadcn/ui desktop frontend
The desktop client SHALL use a React frontend with shadcn/ui-style components and Tailwind CSS theme tokens while preserving the existing Tauri and local daemon architecture.

#### Scenario: Desktop frontend builds for Tauri
- **WHEN** the desktop frontend build runs
- **THEN** it produces the static assets consumed by the existing Tauri configuration

#### Scenario: User opens the desktop app
- **WHEN** the desktop app loads
- **THEN** the main UI is rendered through React components using the local daemon API for data and actions

### Requirement: Desktop appearance settings
The desktop client SHALL provide a settings view with a theme mode control.

#### Scenario: User selects light theme
- **WHEN** a user selects `light` theme mode in settings
- **THEN** the UI applies the light theme and persists the preference locally

#### Scenario: User selects dark theme
- **WHEN** a user selects `dark` theme mode in settings
- **THEN** the UI applies the dark theme and persists the preference locally

#### Scenario: User selects system theme
- **WHEN** a user selects `system` theme mode in settings
- **THEN** the UI follows the operating system color scheme and persists the preference locally

#### Scenario: User restarts the desktop app
- **WHEN** a user reopens the desktop app after selecting a theme mode
- **THEN** the saved theme preference is applied without requiring the user to select it again
