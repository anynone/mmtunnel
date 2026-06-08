## MODIFIED Requirements

### Requirement: Client-sourced tunnel registration
The tunnel client SHALL support tunnel definitions supplied from CLI YAML configuration or from desktop-client generated runtime configuration, and SHALL send the active tunnel definitions to the server after authentication.

#### Scenario: Client registers multiple tunnels from CLI YAML
- **WHEN** an authenticated CLI client loads multiple valid tunnel definitions from YAML
- **THEN** the server registers each non-conflicting tunnel for that client

#### Scenario: Client registers enabled tunnels from desktop profile
- **WHEN** a desktop runtime generates client configuration from a GUI profile with enabled and disabled tunnels
- **THEN** only enabled tunnel definitions are sent to the server for registration

#### Scenario: Client tunnel target uses supported address
- **WHEN** a tunnel target points to localhost, loopback, a private network address, or an external HTTP/HTTPS service
- **THEN** the client accepts the target configuration
