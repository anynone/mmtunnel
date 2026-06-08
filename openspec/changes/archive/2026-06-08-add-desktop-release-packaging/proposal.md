## Why

The desktop client needs reproducible release installers for Windows, Linux, and macOS. The current Tauri setup can build locally, but it does not define a release-only CI pipeline or a complete x64/aarch64 installer matrix.

## What Changes

- Add a GitHub Actions workflow that runs only when a GitHub Release is published.
- Build and upload desktop installer artifacts for macOS, Linux, and Windows.
- Support x64 and aarch64 architectures for each desktop operating system.
- Configure Tauri bundle targets explicitly for `dmg`, `appimage`, `deb`, `nsis`, and `msi`.
- Build the Go `tunnel-daemon` sidecar for each release target and place it using Tauri v2 sidecar naming.

## Capabilities

### New Capabilities

- `desktop-release-packaging`: Defines the release-triggered desktop packaging matrix, required installer formats, and bundled sidecar behavior.

### Modified Capabilities

- None.

## Impact

- Adds a GitHub Actions workflow under `.github/workflows/`.
- Updates `desktop/src-tauri/tauri.conf.json` packaging configuration.
- Adds or updates desktop package scripts as needed for release builds.
- Uses GitHub-hosted macOS, Linux, and Windows runners.
- Requires GitHub `contents: write` permission for uploading release assets.
