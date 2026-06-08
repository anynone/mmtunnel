## Why

Teams need a controlled way to expose HTTP services running on local machines or private networks through a public gateway for development, testing, and internal integration scenarios. A dedicated tunnel service lets trusted clients register path-based routes dynamically without requiring inbound connectivity to the private network.

## What Changes

- Add a Go-based tunnel server that runs behind Nginx or another reverse proxy and accepts HTTP requests plus standard WebSocket Upgrade requests.
- Add a Go-based tunnel client that connects outbound to the server over WebSocket, authenticates with `clientId + token`, and registers locally configured tunnels.
- Support path-prefix routing with longest-prefix matching.
- Reserve `/_tunnel/*` for system endpoints and reject business tunnels that attempt to use this prefix.
- Allow each client to register multiple tunnels from a local YAML configuration.
- Reject duplicate public paths with an explicit path conflict error.
- Forward HTTP requests and standard WebSocket Upgrade sessions to client-side targets.
- Allow tunnel targets to point to localhost, loopback addresses, private network addresses, or external HTTP/HTTPS services.
- Add configurable path prefix stripping per tunnel, defaulting to enabled.
- Add server-side read-only status endpoints under `/_tunnel/*`.
- Add configurable server log levels: `trace`, `debug`, `info`, `warning`, and `error`.

## Capabilities

### New Capabilities

- `http-tunnel`: Defines the HTTP and WebSocket tunnel system, including client authentication, dynamic route registration, path routing, forwarding behavior, keepalive, status visibility, and configuration requirements.

### Modified Capabilities

- None.

## Impact

- Introduces new server and client applications, both implemented primarily in Go.
- Introduces YAML configuration for server clients/tokens and client tunnel definitions.
- Introduces a WebSocket-based client-server tunnel protocol.
- Requires deployment behind a reverse proxy for HTTPS/WSS termination in production.
- Establishes `/_tunnel/*` as reserved system API space.
- Future client GUI work can build on the Go client core using a local tray application approach such as Tauri plus Go.
