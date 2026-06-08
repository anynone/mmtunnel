# http-tunnel Specification

## Purpose
TBD - created by archiving change add-http-tunnel. Update Purpose after archive.
## Requirements
### Requirement: Server deployment behind reverse proxy
The tunnel server SHALL accept plain HTTP requests and standard WebSocket Upgrade requests from an upstream reverse proxy and SHALL NOT require direct HTTPS certificate handling.

#### Scenario: Reverse proxy forwards HTTPS request as HTTP
- **WHEN** a reverse proxy terminates HTTPS and forwards an HTTP request to the tunnel server
- **THEN** the tunnel server routes and forwards the request according to registered tunnels

#### Scenario: Reverse proxy forwards WebSocket Upgrade
- **WHEN** a reverse proxy forwards a standard WebSocket Upgrade request to the tunnel server
- **THEN** the tunnel server handles the upgrade request as a tunnelable WebSocket session

### Requirement: Reserved system path prefix
The tunnel server SHALL reserve `/_tunnel/*` for system endpoints and SHALL reject any business tunnel whose public path uses the reserved prefix.

#### Scenario: Client registers reserved path
- **WHEN** an authenticated client attempts to register a tunnel with public path `/_tunnel/app`
- **THEN** the server rejects the registration

#### Scenario: Client registers non-reserved path
- **WHEN** an authenticated client registers a tunnel with public path `/api`
- **THEN** the server allows the path if it does not conflict with existing routes

### Requirement: Client authentication
The tunnel server SHALL authenticate tunnel clients using `clientId + token` from server configuration before accepting tunnel registrations.

#### Scenario: Client authenticates successfully
- **WHEN** a client connects with a configured `clientId` and matching token
- **THEN** the server accepts the tunnel connection and allows tunnel registration

#### Scenario: Client token is invalid
- **WHEN** a client connects with a configured `clientId` and an invalid token
- **THEN** the server rejects the tunnel connection

#### Scenario: Client id is unknown
- **WHEN** a client connects with an unknown `clientId`
- **THEN** the server rejects the tunnel connection

### Requirement: Client-sourced tunnel registration
The tunnel client SHALL load tunnel definitions from local YAML configuration and SHALL send those definitions to the server after authentication.

#### Scenario: Client registers multiple tunnels
- **WHEN** an authenticated client sends multiple valid tunnel definitions
- **THEN** the server registers each non-conflicting tunnel for that client

#### Scenario: Client tunnel target uses supported address
- **WHEN** a tunnel target points to localhost, loopback, a private network address, or an external HTTP/HTTPS service
- **THEN** the client accepts the target configuration

### Requirement: Path conflict rejection
The tunnel server SHALL reject tunnel registration when a submitted public path conflicts with an existing registered public path and SHALL report `路径冲突，已经存在当前路径 {existingPath}`.

#### Scenario: Exact path conflict
- **WHEN** a client registers `/api` and another tunnel registration attempts to register `/api`
- **THEN** the server rejects the new registration with `路径冲突，已经存在当前路径 /api`

#### Scenario: Overlapping path conflict
- **WHEN** a client registers `/api` and another tunnel registration attempts to register `/api/v1`
- **THEN** the server rejects the new registration with `路径冲突，已经存在当前路径 /api`

### Requirement: Longest-prefix routing
The tunnel server SHALL route public requests using longest-prefix path matching among registered tunnels.

#### Scenario: Request matches registered path
- **WHEN** `/api` is registered and a public request arrives for `/api/users`
- **THEN** the server routes the request to the `/api` tunnel

#### Scenario: Request has no matching tunnel
- **WHEN** a public request arrives for a path with no matching registered tunnel
- **THEN** the server returns a gateway or not-found response without forwarding the request to any client

### Requirement: Tunnel lifecycle
The tunnel server SHALL dynamically activate routes when a client registers tunnels and SHALL remove those routes when the client disconnects.

#### Scenario: Client comes online
- **WHEN** a client authenticates and registers valid tunnels
- **THEN** the server makes those tunnel routes available for public request routing

#### Scenario: Client disconnects
- **WHEN** a connected client disconnects or fails keepalive
- **THEN** the server removes that client's registered routes and fails in-flight requests for that client

### Requirement: HTTP request forwarding
The tunnel system SHALL forward HTTP methods, paths, query strings, headers, request bodies, response statuses, response headers, and response bodies between public requests and target services.

#### Scenario: GET request forwards to target
- **WHEN** a public `GET /api/users?id=1` request matches a tunnel target
- **THEN** the client sends a corresponding request to the target service and returns the target response to the public caller

#### Scenario: POST request body forwards to target
- **WHEN** a public `POST` request with a body matches a tunnel target
- **THEN** the client forwards the request body to the target service and returns the target response

#### Scenario: Target returns error status
- **WHEN** the target service returns an HTTP error status
- **THEN** the tunnel system returns that status and response body to the public caller

### Requirement: Path stripping
Each tunnel SHALL support a `stripPath` setting that controls whether the public path prefix is removed before forwarding, and `stripPath` SHALL default to `true`.

#### Scenario: Strip path enabled
- **WHEN** a tunnel has `publicPath: /api`, `target: http://127.0.0.1:8080`, and `stripPath: true`, and a request arrives for `/api/users?id=1`
- **THEN** the client forwards the request to `http://127.0.0.1:8080/users?id=1`

#### Scenario: Strip path disabled
- **WHEN** a tunnel has `publicPath: /api`, `target: http://127.0.0.1:8080`, and `stripPath: false`, and a request arrives for `/api/users?id=1`
- **THEN** the client forwards the request to `http://127.0.0.1:8080/api/users?id=1`

#### Scenario: Strip path omitted
- **WHEN** a tunnel omits `stripPath` and a request arrives under the tunnel public path
- **THEN** the client treats `stripPath` as `true`

### Requirement: Forwarded header handling
The tunnel system SHALL NOT preserve the public request's original `Host` header when forwarding to the target and SHALL add `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Tunnel-Client-Id`.

#### Scenario: Host header rewritten for target
- **WHEN** a public request for `gateway.example.com/api` is forwarded to `http://127.0.0.1:8080`
- **THEN** the forwarded request uses the target host rather than `gateway.example.com`

#### Scenario: Forwarding metadata headers added
- **WHEN** a public request is forwarded through a tunnel
- **THEN** the forwarded request includes `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Tunnel-Client-Id`

### Requirement: Standard WebSocket forwarding
The tunnel system SHALL support forwarding standard WebSocket Upgrade sessions between public callers and target services.

#### Scenario: Public WebSocket connects to target
- **WHEN** a public request performs a standard WebSocket Upgrade on a registered tunnel path
- **THEN** the tunnel system establishes a corresponding WebSocket connection to the target service

#### Scenario: WebSocket frames are relayed
- **WHEN** a public WebSocket session and target WebSocket session are established
- **THEN** the tunnel system relays WebSocket frames in both directions until either side closes

### Requirement: Tunnel keepalive
The client-server tunnel connection SHALL use WebSocket ping/pong keepalive and SHALL clean up disconnected clients.

#### Scenario: Pong received
- **WHEN** the server sends a ping and receives a pong within the configured timeout
- **THEN** the server keeps the client connection active

#### Scenario: Pong timeout
- **WHEN** the server sends a ping and does not receive a pong within the configured timeout
- **THEN** the server closes the client connection and removes that client's registered routes

### Requirement: Client reconnect
The tunnel client SHALL attempt to reconnect after a tunnel connection is lost.

#### Scenario: Server restarts
- **WHEN** the tunnel server closes the client connection during a restart
- **THEN** the client attempts to reconnect using its configured reconnect interval

#### Scenario: Client reconnects and re-registers tunnels
- **WHEN** a disconnected client reconnects successfully
- **THEN** it re-authenticates and re-registers its configured tunnels

### Requirement: Status endpoints
The tunnel server SHALL provide read-only status endpoints under `/_tunnel/*` that expose online clients and registered tunnel state.

#### Scenario: Status endpoint returns online client
- **WHEN** a client is connected and has registered tunnels
- **THEN** a status endpoint under `/_tunnel/*` can report that client and its registered tunnel paths

#### Scenario: Status endpoint is read-only
- **WHEN** a caller accesses status endpoints under `/_tunnel/*`
- **THEN** the endpoints do not create, update, or delete clients or tunnels

### Requirement: Configurable logging
The tunnel server SHALL support configured log levels `trace`, `debug`, `info`, `warning`, and `error`.

#### Scenario: Server starts with configured log level
- **WHEN** the server configuration sets `logLevel: debug`
- **THEN** the server initializes logging at debug level

#### Scenario: Invalid log level
- **WHEN** the server configuration contains an unsupported log level
- **THEN** the server reports a configuration error

### Requirement: Timeout and size controls
The tunnel system SHALL provide configurable defaults for request timeout, target connect timeout, tunnel idle timeout, ping interval, pong timeout, and maximum request body size.

#### Scenario: Target connection timeout
- **WHEN** a target service cannot be connected within the configured target connect timeout
- **THEN** the tunnel system returns a gateway error to the public caller

#### Scenario: Request exceeds maximum body size
- **WHEN** a public HTTP request exceeds the configured maximum request body size
- **THEN** the tunnel system rejects the request instead of forwarding an unbounded body

