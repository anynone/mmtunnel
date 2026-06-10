## Why

The current desktop UI is a minimal vanilla JavaScript control surface that renders HTML strings and uses a small hand-written stylesheet. It proves the local daemon workflow, but it does not yet provide the complete desktop-client experience described by the specification: first-run setup, profile switching and deletion, full tunnel form management, clear test feedback, a settings surface, or user-selectable light and dark themes.

Rewriting the desktop frontend with React, Tailwind CSS, and shadcn/ui gives the application a maintainable component model, a consistent design system, and a standard theme-token foundation while keeping the existing Tauri shell and Go daemon boundary intact.

## What Changes

- Migrate the desktop frontend from vanilla JavaScript and hand-built HTML strings to React running through Vite.
- Add Tailwind CSS and shadcn/ui as the desktop UI component foundation.
- Rebuild the desktop client experience with shadcn/ui-style components for runtime controls, profile editing, tunnel management, activity display, dialogs, forms, switches, badges, and settings.
- Complete the desktop-client UI capabilities already described by the active spec, including first-run setup, profile switching, inactive profile deletion, tunnel add/edit/delete, strip path editing, enabled toggles, connection test feedback, and recent activity visibility.
- Add a settings view that includes theme mode selection.
- Support `light`, `dark`, and `system` theme modes, persist the user's preference locally, and apply the theme through shadcn/ui-compatible CSS variables.
- Preserve the existing local daemon API, Tauri shell, profile/runtime behavior, and CLI YAML import/export semantics.

## Capabilities

### Modified Capabilities

- `desktop-client`: Clarifies that the desktop UI uses React and shadcn/ui, completes the existing specified GUI workflows, and adds user-selectable light/dark/system theme behavior through a settings view.

## Impact

- Replaces `desktop/src/app.js` string-rendering UI with a React component tree.
- Changes the desktop build path from static file copying to a Vite production build.
- Adds frontend dependencies for React, Vite, Tailwind CSS, shadcn/ui support libraries, and icons.
- Adds shadcn/ui-compatible theme tokens to the desktop stylesheet.
- Requires frontend tests to move from DOM-free state-only coverage toward React component and workflow coverage where practical.
- Does not change the Go daemon API contract, tunnel runtime behavior, Tauri sidecar architecture, or tunnel server protocol.
