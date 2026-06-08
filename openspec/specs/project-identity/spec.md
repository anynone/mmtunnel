# project-identity Specification

## Purpose
TBD - created by archiving change rename-project-to-mmtunnel. Update Purpose after archive.
## Requirements
### Requirement: Canonical project identity
The project SHALL use `mmtunnel` as its canonical source-level project name and `MM Tunnel` as its canonical user-facing product name.

#### Scenario: Source identifiers use canonical project name
- **WHEN** source files, package metadata, or build configuration refer to the project name
- **THEN** they use `mmtunnel` or an `mmtunnel`-derived identifier

#### Scenario: User-facing app text uses canonical product name
- **WHEN** the desktop app, installer metadata, or documentation displays the product name to users
- **THEN** it uses `MM Tunnel`

### Requirement: Desktop application identifiers
The desktop application SHALL use `mmtunnel-desktop` for package metadata and `cn.anynone.mmtunnel` for its Tauri bundle identifier.

#### Scenario: Desktop package metadata is read
- **WHEN** package metadata for the desktop app is inspected
- **THEN** the package name is `mmtunnel-desktop`

#### Scenario: Tauri bundle identifier is read
- **WHEN** the Tauri configuration is inspected
- **THEN** the bundle identifier is `cn.anynone.mmtunnel`

### Requirement: Local configuration namespace
The desktop daemon and frontend SHALL use `mmtunnel` as the default local configuration namespace.

#### Scenario: Default profile path is resolved
- **WHEN** the desktop daemon resolves the default profile store path
- **THEN** the path uses a `mmtunnel` directory name

#### Scenario: Frontend daemon URL override is stored
- **WHEN** the desktop frontend reads or writes the daemon URL override
- **THEN** it uses the `mmtunnel.daemon` localStorage key

### Requirement: Stable tunnel command names
The project SHALL keep the existing tunnel command binary names unchanged by this rename.

#### Scenario: Command names are referenced
- **WHEN** build scripts, release workflows, or service examples refer to tunnel binaries
- **THEN** they continue using `tunnel-client`, `tunnel-server`, or `tunnel-daemon`

