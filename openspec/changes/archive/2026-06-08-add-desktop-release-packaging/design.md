## Context

The repository contains a Tauri desktop app under `desktop/` and a Go `tunnel-daemon` sidecar under `cmd/tunnel-daemon`. The current desktop package config enables bundling but does not define an explicit release-only GitHub Actions pipeline or a complete installer matrix.

Tauri v2 external binaries are target-specific sidecars. The configured `externalBin` value identifies the logical sidecar path, while the actual file on disk must include the Rust target triple suffix for the build target.

## Goals / Non-Goals

**Goals:**

- Build desktop installers only when a GitHub Release is published.
- Produce macOS `dmg`, Linux `AppImage` and `deb`, Windows NSIS `exe`, and Windows WiX `msi`.
- Produce x64 and aarch64 artifacts for macOS, Linux, and Windows.
- Build and bundle the Go `tunnel-daemon` sidecar for each matrix target.
- Keep release packaging configuration in repo-local CI and Tauri config.

**Non-Goals:**

- Code signing, notarization, or installer trust-chain setup.
- Auto-update metadata.
- Pull request, push, or tag-only build workflows.
- Adding 32-bit, ARMv7, RPM, Snap, or App Store packaging targets.

## Decisions

### Trigger builds from published GitHub Releases

The workflow will use `on.release.types: [published]`.

Rationale: the requested behavior is to build only when generating a release. This avoids consuming runners on normal pushes, pull requests, or tag creation.

Alternative considered: triggering on version tags. That is common for release pipelines, but it would build before a GitHub Release exists and does not match the requested "生成release时进行构建" behavior as directly.

### Use one build matrix row per OS and architecture

The workflow matrix will include:

- `macos-15-intel` with `x86_64-apple-darwin`
- `macos-latest` with `aarch64-apple-darwin`
- `ubuntu-24.04` with `x86_64-unknown-linux-gnu`
- `ubuntu-24.04-arm` with `aarch64-unknown-linux-gnu`
- `windows-latest` with `x86_64-pc-windows-msvc`
- `windows-11-arm` with `aarch64-pc-windows-msvc`

Rationale: native OS and architecture runners are the least surprising way to produce platform installers where GitHub provides them. Architecture is still declared through Rust targets and Tauri build arguments so artifact names and sidecar names remain explicit.

### Use Tauri action with `projectPath: desktop`

The workflow will use `tauri-apps/tauri-action` to run the release build and upload artifacts to the published release.

Rationale: the action is the standard Tauri release integration and supports subdirectory projects through `projectPath`.

### Configure explicit bundle targets

`desktop/src-tauri/tauri.conf.json` will replace `"targets": "all"` with:

```json
["dmg", "appimage", "deb", "nsis", "msi"]
```

Rationale: explicit targets document the required installer families and prevent unintended bundle formats from becoming release artifacts.

### Move sidecar path under `desktop/src-tauri/binaries`

The Tauri `externalBin` value will become `binaries/tunnel-daemon`. Each workflow matrix row will compile `cmd/tunnel-daemon` and write the actual file as:

```text
desktop/src-tauri/binaries/tunnel-daemon-<target-triple>
desktop/src-tauri/binaries/tunnel-daemon-<target-triple>.exe
```

Rationale: this follows Tauri v2 sidecar naming and avoids relying on root-level `bin/` files, which are host-specific build outputs.

### Build the Go sidecar in CI before Tauri packaging

Each matrix row will map Rust target triples to Go `GOOS` and `GOARCH`, set `GOCACHE` to the runner temp directory, and run `go build -buildvcs=false ./cmd/tunnel-daemon`.

Rationale: the installer must contain the matching daemon binary. Building the sidecar in the same matrix row keeps the bundled daemon aligned with the Tauri app target.

## Risks / Trade-offs

- Linux aarch64 packaging may require additional cross-compilation dependencies on GitHub runners -> Start with Tauri's documented Ubuntu dependency set and keep the matrix explicit so failures identify the affected target.
- macOS and Windows artifacts are unsigned -> Document this as out of scope; unsigned installers may show OS trust warnings.
- Windows ARM64 NSIS installer process may still use an x86 installer wrapper -> Accept this Tauri limitation as long as the app binary itself targets ARM64.
- Publishing the release starts all matrix jobs after the release is visible -> Failed jobs can leave a release with partial artifacts; rerunning failed jobs should upload missing assets to the same release.

## Migration Plan

1. Add the release workflow under `.github/workflows/desktop-release.yml`.
2. Update Tauri bundle targets and sidecar path.
3. Add desktop package scripts only if they reduce duplication or make local release commands clearer.
4. Validate config syntax and run local desktop tests.
5. Verify full installer generation through the next published GitHub Release.
