## Context

The project currently uses `mmsocket` as the Go module path, desktop package prefix, local configuration namespace, service install path, and OpenSpec/release text. The user requested the project be renamed to `mmtunnel`.

There are existing uncommitted desktop release-packaging changes. The rename must work with those changes rather than reverting them.

## Goals / Non-Goals

**Goals:**

- Make `mmtunnel` the canonical source-level project name.
- Make `MM Tunnel` the canonical user-facing desktop product name.
- Update package/module identifiers, local storage keys, default config path, service examples, docs, and active OpenSpec text.
- Preserve existing tunnel command names and runtime behavior.

**Non-Goals:**

- Renaming command binaries such as `tunnel-client`, `tunnel-server`, or `tunnel-daemon`.
- Migrating existing user profile files from the old local config directory.
- Renaming the repository directory on disk from `/home/anynone/project-internal/mmsocket`.
- Changing network protocol, API path, ports, YAML schema, or release packaging behavior.

## Decisions

### Rename source identifiers and imports

`go.mod` will change to `module mmtunnel`, and all internal Go imports will change from `mmsocket/...` to `mmtunnel/...`.

Rationale: keeping the old module name would leak the old identity into every Go package and test.

### Rename desktop package metadata and display name

Desktop package metadata will use `mmtunnel-desktop`, and Tauri/UI display text will use `MM Tunnel`.

Rationale: release artifacts, installers, and the visible app should all present the new project name.

### Rename local namespace without migration

The default desktop profile store will use the `mmtunnel` config namespace. The frontend localStorage key will move to `mmtunnel.daemon`.

Rationale: new installs should no longer create `mmsocket` state. Automatic migration is explicitly out of scope for this rename.

### Keep operational command names stable

The CLI/server/daemon binary names remain `tunnel-client`, `tunnel-server`, and `tunnel-daemon`.

Rationale: those names describe roles rather than the product brand, and renaming them would increase release and service migration scope.

## Risks / Trade-offs

- Existing local profiles under `mmsocket` will not be discovered automatically -> Document the new default path and leave migration for a future compatibility change if needed.
- OpenSpec archive may retain historical references to `mmsocket` -> Update active changes and current docs; archived historical changes can remain as history unless they affect current specs.
- Generated outputs can contain stale names -> Validate source files and regenerate frontend `dist` through the existing build.
