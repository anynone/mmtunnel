## 1. Regression Tests

- [x] 1.1 Add desktop release/package identity tests for `MM Tunnel`, `mmtunnel-desktop`, `cn.anynone.mmtunnel`, and `mmtunnel.daemon`.
- [x] 1.2 Add Go tests for the default desktop profile path using the `mmtunnel` namespace.

## 2. Source Rename

- [x] 2.1 Rename the Go module from `mmsocket` to `mmtunnel`.
- [x] 2.2 Update Go imports from `mmsocket/...` to `mmtunnel/...`.
- [x] 2.3 Rename desktop package metadata from `mmsocket-desktop` to `mmtunnel-desktop`.
- [x] 2.4 Rename Tauri product metadata and bundle identifier to `MM Tunnel` and `cn.anynone.mmtunnel`.
- [x] 2.5 Rename frontend UI text and localStorage namespace to `MM Tunnel` and `mmtunnel.daemon`.
- [x] 2.6 Rename the daemon default profile directory from `mmsocket` to `mmtunnel`.

## 3. Documentation And OpenSpec Text

- [x] 3.1 Update current documentation and service examples from old project naming to `mmtunnel` / `MM Tunnel`.
- [x] 3.2 Update active OpenSpec change text to use the new project identity where relevant.
- [x] 3.3 Confirm command binary names remain `tunnel-client`, `tunnel-server`, and `tunnel-daemon`.

## 4. Verification

- [x] 4.1 Validate OpenSpec change artifacts.
- [x] 4.2 Run desktop tests and build.
- [x] 4.3 Run Go tests.
- [x] 4.4 Search source files for stale `mmsocket` / `MM Socket` references outside archived history and intentional rename-change context.
