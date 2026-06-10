## 1. Frontend Tooling

- [x] 1.1 Add React, React DOM, Vite, Tailwind CSS, PostCSS, and shadcn/ui support dependencies to the desktop package.
- [x] 1.2 Add Vite, Tailwind, PostCSS, and path alias configuration for the desktop frontend.
- [x] 1.3 Replace the copy-only desktop build with a Vite build that writes to `desktop/dist`.
- [x] 1.4 Update `desktop/index.html` to mount the React app and apply the initial theme safely.

## 2. shadcn/ui Foundation

- [x] 2.1 Add shared utilities for class name merging and UI component variants.
- [x] 2.2 Add local shadcn/ui-style components needed by the desktop UI: button, input, label, switch, select, dialog, table, badge, alert, tabs or navigation controls, and separator.
- [x] 2.3 Define shadcn/ui-compatible CSS variables for light and dark themes in the desktop stylesheet.
- [x] 2.4 Add status color treatments for connected, stopped, reconnecting, disabled, pending, and failed states.

## 3. React App Architecture

- [x] 3.1 Create the React entrypoint and top-level app component.
- [x] 3.2 Add a daemon API module for status, profile CRUD, active profile selection, runtime actions, connection tests, events, import, and export.
- [x] 3.3 Preserve or move pure state helpers with tests for profile payload conversion, status labels, tunnel toggle behavior, and slug generation.
- [x] 3.4 Add app-level state loading, refresh, error handling, and server-sent event subscription.
- [x] 3.5 Add a compact desktop app shell with primary areas for overview/profile, activity, and settings.

## 4. Complete Desktop Workflows

- [x] 4.1 Implement first-run setup when no profiles exist, including profile fields and initial tunnel definitions.
- [x] 4.2 Implement profile selection and active profile switching.
- [x] 4.3 Implement profile creation, editing, saving, and inactive profile deletion.
- [x] 4.4 Implement masked token editing with clear save behavior.
- [x] 4.5 Implement tunnel add, edit, delete, strip path, and enabled controls.
- [x] 4.6 Implement enabled toggle behavior that saves the profile and restarts the runtime when required by MVP semantics.
- [x] 4.7 Implement runtime start, stop, restart controls and overall runtime status display.
- [x] 4.8 Implement per-tunnel status badges that distinguish enabled state from runtime status.
- [x] 4.9 Implement server reachability and authentication test actions with visible success and failure feedback.
- [x] 4.10 Implement recent activity display for logs, request summaries, and errors without showing request or response bodies.
- [x] 4.11 Preserve CLI YAML import and export flows in the React UI.

## 5. Settings and Theme

- [x] 5.1 Add a settings view that fits the desktop control-plane layout.
- [x] 5.2 Add a theme mode control in settings with `system`, `light`, and `dark` modes.
- [x] 5.3 Persist theme preference in localStorage and apply `.dark` to the document root when appropriate.
- [x] 5.4 React to system color-scheme changes when the selected mode is `system`.
- [x] 5.5 Keep or replace the existing daemon base URL override in settings, depending on the daemon API needs discovered during implementation.

## 6. Documentation and Verification

- [x] 6.1 Update desktop documentation to describe the React/shadcn/ui frontend, theme settings, and build commands.
- [x] 6.2 Update frontend tests for preserved state helpers and theme helper behavior.
- [x] 6.3 Add workflow tests for setup, profile switching, inactive profile deletion, tunnel editing, enabled toggle, runtime controls, and settings where practical.
- [x] 6.4 Run desktop tests and build.
- [x] 6.5 Run Go tests to confirm the daemon/runtime contract remains unchanged.
- [x] 6.6 Run OpenSpec validation if the CLI is available in the implementation environment.
