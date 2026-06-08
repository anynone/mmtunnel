## 1. Tauri Packaging Configuration

- [x] 1.1 Update `desktop/src-tauri/tauri.conf.json` to use explicit bundle targets for `dmg`, `appimage`, `deb`, `nsis`, and `msi`.
- [x] 1.2 Move the Tauri `externalBin` logical path to `binaries/tunnel-daemon`.
- [x] 1.3 Add or update desktop package scripts for release builds if needed.

## 2. Release Workflow

- [x] 2.1 Add `.github/workflows/desktop-release.yml` with `release.published` as the only trigger.
- [x] 2.2 Define the GitHub Actions matrix for macOS, Linux, and Windows x64/aarch64 targets.
- [x] 2.3 Install Node.js, Rust targets, Go, and platform dependencies in the workflow.
- [x] 2.4 Build `cmd/tunnel-daemon` for each matrix target using writable CI cache paths.
- [x] 2.5 Place the sidecar under `desktop/src-tauri/binaries/tunnel-daemon-<target-triple>` with `.exe` for Windows.
- [x] 2.6 Run `tauri-apps/tauri-action` with `projectPath: desktop`, target args, and release asset upload to the triggering release.

## 3. Verification

- [x] 3.1 Validate OpenSpec change artifacts.
- [x] 3.2 Validate JSON syntax for `desktop/src-tauri/tauri.conf.json`.
- [x] 3.3 Validate GitHub Actions workflow YAML syntax.
- [x] 3.4 Run desktop frontend tests with `npm test` under `desktop`.
- [x] 3.5 Review workflow paths and release asset behavior against the required installer matrix.
