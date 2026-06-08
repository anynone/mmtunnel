## Context

This change introduces a new HTTP tunnel system for enterprise-style development and integration use cases. The repository currently contains OpenSpec planning artifacts only, so this design establishes the initial architecture rather than modifying an existing implementation.

The system consists of a public tunnel server and one or more private tunnel clients. The server is deployed behind Nginx or another reverse proxy, which terminates HTTPS/WSS and forwards HTTP plus WebSocket Upgrade traffic to the tunnel server. Clients initiate outbound WebSocket connections to the server, authenticate with `clientId + token`, register path-based tunnels from local YAML configuration, and forward requests to configured targets.

```
Public user
  |
  | HTTPS / WSS
  v
Nginx or external reverse proxy
  |
  | HTTP / WebSocket Upgrade
  v
Tunnel Server (Go)
  |
  | WebSocket tunnel
  v
Tunnel Client (Go CLI)
  |
  | HTTP / WebSocket
  v
Target service
```

## Goals / Non-Goals

**Goals:**

- Provide HTTP request forwarding through client-initiated tunnels.
- Provide standard WebSocket Upgrade forwarding through the same tunnel system.
- Support multiple authenticated clients and multiple tunnels per client.
- Use path-prefix routing with longest-prefix matching.
- Dynamically register and unregister routes as clients connect and disconnect.
- Keep tunnel configuration local to the client for the first release.
- Preserve enough forwarding metadata for target services to understand the original request source and scheme.
- Provide read-only server status under reserved system paths.
- Keep server deployment compatible with reverse proxies that handle HTTPS certificates.

**Non-Goals:**

- Direct TLS certificate parsing or certificate automation in the tunnel server.
- TCP or UDP tunneling.
- Public endpoint authentication.
- Load balancing multiple clients on the same public path.
- Server-managed tunnel configuration or a management dashboard.
- Billing, tenant management, or audit workflows.
- Client GUI in the first release.

## Decisions

### Deploy the server behind a reverse proxy

The tunnel server will accept HTTP and WebSocket Upgrade traffic from a reverse proxy instead of terminating HTTPS itself.

Rationale: this keeps certificate handling, TLS policy, and public ingress controls in existing infrastructure such as Nginx. It also allows deployments to block or restrict `/_tunnel/*` status paths at the proxy layer.

Alternative considered: direct HTTPS termination in the tunnel server. This was rejected for the first release because certificate automation and TLS operations are not required for the target development scenarios.

### Use Go for server and client core

The server and initial CLI client will be implemented in Go.

Rationale: Go is a strong fit for network services, concurrent request handling, cross-platform CLI distribution, and a future local tray application that can reuse the same core logic.

Alternative considered: JavaScript or Rust. JavaScript would simplify some GUI paths but is less attractive for a single-binary network tool. Rust offers strong safety but increases implementation complexity for the first release.

### Use WebSocket as the client-server tunnel transport

Clients will establish a long-lived WebSocket connection to the server. The connection will carry control messages, HTTP request/response frames, and WebSocket forwarding frames identified by request IDs.

Rationale: WebSocket is broadly supported, works through common reverse proxies, and matches the requirement to support ping/pong keepalive. It is simpler than HTTP/2 stream multiplexing or QUIC for the first implementation.

Alternative considered: HTTP/2 streams or QUIC. These may provide cleaner multiplexing or better performance, but they introduce more deployment and implementation complexity.

### Authenticate clients with `clientId + token`

The server configuration file will contain trusted client IDs and tokens. A client must authenticate before registering tunnels.

Rationale: the public endpoints intentionally have no tunnel-level access control, so route creation must be limited to trusted clients. Server-side configuration is enough for the first release and avoids introducing a database.

Alternative considered: database-backed client identity. This is deferred until management UI or dynamic provisioning is needed.

### Keep tunnel definitions in the client YAML

Each client configures its own tunnels locally and sends the tunnel list to the server after authentication.

Rationale: this supports developer-owned local workflows and avoids a server-side control plane in the first release.

Alternative considered: server-owned tunnel configuration. This is better for centralized enterprise administration, but requires additional management APIs and state storage.

### Reserve `/_tunnel/*` for system endpoints

The server will reserve `/_tunnel/*` for connection and status endpoints. Business tunnels must not register paths with this prefix.

Rationale: system endpoints need a stable namespace that cannot be shadowed by user routes.

### Route by longest path prefix and reject conflicts

The server will match public requests using longest-prefix matching. Registration fails if a new tunnel's public path conflicts with an existing path.

Rationale: longest-prefix matching is predictable when paths overlap. Rejecting conflicts avoids ambiguous routing and avoids load-balancing behavior that is out of scope.

Conflict detection should treat exact duplicate paths as conflicts. It should also reject overlapping prefixes that would make ownership ambiguous, such as `/api` and `/api/v1`, unless a later design explicitly introduces hierarchical ownership rules.

### Support per-tunnel path stripping

Each tunnel has `stripPath`, defaulting to `true`. When enabled, the public path prefix is removed before forwarding to the target.

Rationale: most local services are not aware that they are mounted under a public gateway prefix. A per-tunnel switch preserves compatibility with services that do expect the prefix.

### Do not preserve the original Host header

The forwarded request will use the target service host as `Host`. The server/client forwarding layer will add `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Tunnel-Client-Id`.

Rationale: target services usually expect their own host, while forwarded headers retain useful source metadata.

### Provide read-only status under `/_tunnel/*`

The server will expose read-only status for online clients, registered tunnels, and basic connection state under reserved system paths.

Rationale: status visibility is necessary for operations and debugging. Access control is handled by deployment configuration, especially Nginx path blocking or allowlisting.

### Use configured timeouts and keepalive defaults

The implementation will provide default values for request timeout, target connect timeout, tunnel idle timeout, ping interval, and pong timeout. These values will be configurable.

Rationale: request forwarding and WebSocket tunnels need bounded failure behavior. Defaults make first-run behavior usable while allowing operators to tune production deployments.

## Risks / Trade-offs

- Public routes have no built-in access control -> Deployments must treat exposed paths as public and use upstream proxy controls or target-service auth when needed.
- Client-owned target URLs can forward to external services -> This is intentional but relies on token trust and operator control over who can register clients.
- WebSocket multiplexing can suffer head-of-line or flow-control issues -> Implement request IDs, chunked body frames, backpressure, and configurable concurrency limits.
- Route conflict rules can be stricter than necessary -> Start strict to prevent surprising routing; revisit once hierarchical route ownership is needed.
- Status endpoints under `/_tunnel/*` depend on proxy-level protection -> Document deployment requirements and make status read-only.
- Client disconnects can leave in-flight requests incomplete -> Server must unregister routes and fail in-flight requests promptly when a tunnel closes.
- Large uploads or downloads can stress memory -> Forward bodies as streams/chunks and enforce configurable request size and timeout limits.

## Migration Plan

This is a new system, so no data migration is required.

1. Implement server and client binaries with local YAML configuration.
2. Deploy the server behind Nginx or an equivalent reverse proxy.
3. Configure Nginx to terminate HTTPS/WSS and forward HTTP/WebSocket Upgrade traffic to the tunnel server.
4. Configure Nginx to block or restrict `/_tunnel/*` status paths where appropriate.
5. Start trusted clients with `clientId + token` and local tunnel definitions.

Rollback consists of stopping the tunnel server and removing or disabling the reverse-proxy routes that forward traffic to it.

## Open Questions

- Exact default values for maximum request body size and maximum concurrent requests per client should be finalized during implementation.
- The future tray GUI architecture should be designed separately once the CLI tunnel core is stable.
