# Desktop Release Packaging Design

## Goal

Generate desktop installers only when a GitHub Release is published. The release job must attach installers for Windows, Linux, and macOS, limited to x64 and aarch64 architectures.

## Outputs

The workflow will produce these installer families:

- macOS x64 and aarch64: `dmg`
- Linux x64 and aarch64: `AppImage` and `deb`
- Windows x64 and aarch64: NSIS `exe` installer and WiX `msi` installer

## Architecture

Use a single GitHub Actions workflow triggered by:

```yaml
on:
  release:
    types: [published]
```

The workflow will use a matrix with one row per operating system and architecture target. Each matrix row will:

1. Check out the repository.
2. Install Node.js and Rust.
3. Install platform-specific dependencies.
4. Build the Go `tunnel-daemon` sidecar for the matrix target.
5. Place the sidecar under `desktop/src-tauri/binaries/tunnel-daemon-<target-triple>` with `.exe` for Windows.
6. Run `tauri-apps/tauri-action` from `projectPath: desktop`.
7. Upload generated Tauri bundles to the existing GitHub Release.

The Tauri bundle configuration will use explicit targets instead of `"all"`:

```json
["dmg", "appimage", "deb", "nsis", "msi"]
```

Unsupported targets for the current platform are ignored by Tauri bundling; the action matrix determines which OS-specific outputs are produced.

## Sidecar Handling

Tauri v2 expects external binaries to exist with a target triple suffix. The current `externalBin` path will move from `../../bin/tunnel-daemon` to `binaries/tunnel-daemon`, relative to `desktop/src-tauri/tauri.conf.json`.

The workflow will compile the sidecar from `./cmd/tunnel-daemon` using Go cross-compilation:

- `x86_64-apple-darwin`: `GOOS=darwin GOARCH=amd64`
- `aarch64-apple-darwin`: `GOOS=darwin GOARCH=arm64`
- `x86_64-unknown-linux-gnu`: `GOOS=linux GOARCH=amd64`
- `aarch64-unknown-linux-gnu`: `GOOS=linux GOARCH=arm64`
- `x86_64-pc-windows-msvc`: `GOOS=windows GOARCH=amd64`
- `aarch64-pc-windows-msvc`: `GOOS=windows GOARCH=arm64`

The workflow will use a writable `GOCACHE` under the runner temp directory so CI does not depend on user-local cache paths.

## Release Behavior

The workflow will not run on push, pull request, or tag creation by itself. Publishing a GitHub Release is the only trigger.

Each matrix row uploads its own artifacts to the same release using `GITHUB_TOKEN` with `contents: write`.

## Testing

Local verification will include:

- JSON syntax validation for `desktop/src-tauri/tauri.conf.json`
- YAML parse validation for the GitHub Actions workflow
- `npm test` under `desktop`
- A local frontend/Tauri configuration sanity check where possible without requiring cross-platform packaging tools

Full installer generation requires GitHub-hosted runners for all target platforms.

## Non-Goals

- Code signing and notarization are not configured in this change.
- Auto-update metadata is not configured in this change.
- Building on push, pull request, or tag-only events is intentionally excluded.
