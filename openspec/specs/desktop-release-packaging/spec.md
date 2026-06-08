# desktop-release-packaging Specification

## Purpose
TBD - created by archiving change add-desktop-release-packaging. Update Purpose after archive.
## Requirements
### Requirement: Release-only desktop packaging workflow
The project SHALL provide a GitHub Actions workflow that builds desktop installers only when a GitHub Release is published.

#### Scenario: Published release triggers packaging
- **WHEN** a GitHub Release is published
- **THEN** the desktop packaging workflow runs

#### Scenario: Non-release events do not trigger packaging
- **WHEN** code is pushed, a pull request is opened, or a tag is created without publishing a GitHub Release
- **THEN** the desktop packaging workflow does not run

### Requirement: Desktop installer target matrix
The desktop packaging workflow SHALL build x64 and aarch64 installers for macOS, Linux, and Windows.

#### Scenario: macOS release artifacts
- **WHEN** the desktop packaging workflow runs
- **THEN** it builds `dmg` installers for `x86_64-apple-darwin` and `aarch64-apple-darwin`

#### Scenario: Linux release artifacts
- **WHEN** the desktop packaging workflow runs
- **THEN** it builds `AppImage` and `deb` installers for `x86_64-unknown-linux-gnu` and `aarch64-unknown-linux-gnu`

#### Scenario: Windows release artifacts
- **WHEN** the desktop packaging workflow runs
- **THEN** it builds NSIS `exe` and WiX `msi` installers for `x86_64-pc-windows-msvc` and `aarch64-pc-windows-msvc`

### Requirement: Explicit Tauri bundle formats
The desktop Tauri configuration SHALL declare the required bundle formats explicitly.

#### Scenario: Tauri bundle targets are configured
- **WHEN** Tauri builds the desktop app for release
- **THEN** the configured bundle targets include `dmg`, `appimage`, `deb`, `nsis`, and `msi`

### Requirement: Target-specific daemon sidecar
The desktop release build SHALL bundle a `tunnel-daemon` sidecar binary compiled for the same target as the Tauri app.

#### Scenario: Sidecar is built for matrix target
- **WHEN** a release matrix row builds a target platform and architecture
- **THEN** it compiles `cmd/tunnel-daemon` for the matching operating system and architecture

#### Scenario: Sidecar uses Tauri v2 target suffix
- **WHEN** the sidecar is placed for Tauri bundling
- **THEN** its file name includes the Rust target triple suffix expected by Tauri v2

#### Scenario: Windows sidecar has executable extension
- **WHEN** the sidecar is built for a Windows target
- **THEN** its file name ends with `.exe`

### Requirement: Release asset upload
The desktop packaging workflow SHALL upload generated installer artifacts to the GitHub Release that triggered the workflow.

#### Scenario: Artifacts attach to triggering release
- **WHEN** a release matrix row finishes generating installers
- **THEN** the generated installers are uploaded as assets on the triggering GitHub Release

