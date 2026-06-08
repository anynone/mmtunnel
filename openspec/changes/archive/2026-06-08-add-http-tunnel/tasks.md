## 1. Project Setup

- [x] 1.1 Initialize the Go module and repository structure for server, client, shared protocol, and configuration packages.
- [x] 1.2 Add configuration loading for server YAML and client YAML with validation errors.
- [x] 1.3 Add logging initialization with `trace`, `debug`, `info`, `warning`, and `error` levels.
- [x] 1.4 Add shared duration, size, path, and URL validation helpers.

## 2. Tunnel Protocol

- [x] 2.1 Define client-server WebSocket message types for authentication, tunnel registration, request start/body/end, response start/body/end, WebSocket frames, errors, and keepalive handling.
- [x] 2.2 Implement request ID generation and correlation for concurrent HTTP requests and WebSocket sessions.
- [x] 2.3 Implement chunked body transfer with bounded buffering and backpressure-aware reads/writes.
- [x] 2.4 Add protocol-level tests for message encoding, decoding, request correlation, and error handling.

## 3. Server Core

- [x] 3.1 Implement server startup from YAML configuration, including listen address, trusted clients, timeouts, body size limit, and log level.
- [x] 3.2 Implement client WebSocket connection endpoint under `/_tunnel/connect`.
- [x] 3.3 Implement `clientId + token` authentication against server configuration.
- [x] 3.4 Implement ping/pong keepalive, pong timeout handling, and client disconnect cleanup.
- [x] 3.5 Implement the in-memory client and tunnel registry.
- [x] 3.6 Implement reserved `/_tunnel/*` path validation for tunnel registrations.
- [x] 3.7 Implement path conflict detection and the required conflict error message.
- [x] 3.8 Implement route removal and in-flight request failure when a client disconnects.

## 4. Server Routing And Forwarding

- [x] 4.1 Implement public HTTP handler that excludes reserved system paths from business routing.
- [x] 4.2 Implement longest-prefix route lookup for registered tunnels.
- [x] 4.3 Implement HTTP request forwarding from public requests into the tunnel protocol.
- [x] 4.4 Implement streaming response forwarding from clients back to public callers.
- [x] 4.5 Implement gateway or not-found responses for unmatched paths, offline clients, target failures, and tunnel failures.
- [x] 4.6 Implement request timeout, target connect timeout propagation, tunnel idle timeout, and maximum request body size enforcement.
- [x] 4.7 Implement standard WebSocket Upgrade forwarding from public callers into tunnel sessions.

## 5. Client Core

- [x] 5.1 Implement client startup from YAML configuration, including `clientId`, token, server URL, reconnect interval, and tunnels.
- [x] 5.2 Validate tunnel definitions including name, public path, target URL, and `stripPath` defaulting to `true`.
- [x] 5.3 Implement outbound WebSocket connection to the server.
- [x] 5.4 Implement authentication and tunnel registration after connection.
- [x] 5.5 Implement reconnect loop after disconnect or server restart.
- [x] 5.6 Implement client-side keepalive handling compatible with server ping/pong.

## 6. Client Forwarding

- [x] 6.1 Implement HTTP target request construction from tunnel protocol messages.
- [x] 6.2 Implement per-tunnel path stripping and query string preservation.
- [x] 6.3 Implement target `Host` behavior and forwarded headers: `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Tunnel-Client-Id`.
- [x] 6.4 Implement streaming request body forwarding to target services.
- [x] 6.5 Implement streaming target responses back through the tunnel protocol.
- [x] 6.6 Implement standard WebSocket target connection and bidirectional frame relay.
- [x] 6.7 Implement target connection timeout and target failure reporting.

## 7. Status And Operations

- [x] 7.1 Implement read-only status endpoints under `/_tunnel/*` for online clients and registered tunnels.
- [x] 7.2 Ensure status endpoints do not mutate client or tunnel state.
- [x] 7.3 Add operational logs for server startup, client connect/disconnect, tunnel registration, registration failure, request forwarding errors, and keepalive failures.
- [x] 7.4 Add example server and client YAML configuration files.
- [x] 7.5 Add deployment notes for Nginx HTTPS/WSS termination and optional blocking or allowlisting of `/_tunnel/*`.

## 8. Verification

- [x] 8.1 Add unit tests for configuration validation, reserved paths, conflict detection, and longest-prefix matching.
- [x] 8.2 Add server-client integration tests for successful authentication and tunnel registration.
- [x] 8.3 Add integration tests for HTTP GET, POST body forwarding, response status forwarding, and path stripping.
- [x] 8.4 Add integration tests for WebSocket Upgrade forwarding and bidirectional frame relay.
- [x] 8.5 Add integration tests for keepalive timeout, disconnect cleanup, and client reconnect.
- [x] 8.6 Add tests for status endpoint read-only behavior.
- [x] 8.7 Run OpenSpec validation and the Go test suite before marking the change ready to archive.
