## Context

The desktop client currently uses a single vanilla JavaScript module to fetch daemon status, render HTML with template strings, and bind event handlers after each render. Styling is handled by one CSS file. This keeps the MVP small, but it makes the next layer of desktop behavior expensive to maintain: setup flow, profile management, settings, theme state, dialogs, validation feedback, and richer table/form interactions all compete inside one render function.

The product decision for this change is to migrate the desktop frontend to React and use shadcn/ui components. The rewrite should also complete the UI-facing requirements already present in the `desktop-client` spec, rather than only changing visual styling.

## Goals / Non-Goals

**Goals:**

- Migrate the desktop frontend to React and Vite.
- Use Tailwind CSS and shadcn/ui as the UI component foundation.
- Preserve the existing Go daemon API and Tauri shell boundary.
- Provide a complete desktop UI for setup, profile management, runtime controls, tunnel management, activity, import/export, and settings.
- Add a settings view with `light`, `dark`, and `system` theme selection.
- Persist theme preference locally and apply it before or during app startup to avoid an obvious wrong-theme flash.
- Keep the interface compact and desktop-tool oriented, not marketing-like.
- Keep state conversion and daemon API code testable outside large UI components.

**Non-Goals:**

- Changing the local daemon API shape unless a missing endpoint is already required by the existing desktop-client spec.
- Changing tunnel runtime semantics or the restart-based enabled-toggle behavior.
- Adding OS credential storage for tokens.
- Adding long-term analytics, request body inspection, or remote tunnel server management.
- Redesigning tray behavior beyond what is already specified.

## Decisions

### Use React as the desktop UI runtime

The desktop frontend will be rewritten around a React component tree. React is the expected foundation for shadcn/ui and provides a better fit for the required stateful workflows than repeated full-page string rendering.

Representative structure:

```text
desktop/
  src/
    main.jsx
    App.jsx
    styles.css
    lib/
      api.js
      profiles.js
      state.js
      theme.js
      utils.js
    components/
      ui/
      AppShell.jsx
      RuntimeControls.jsx
      ProfileSelector.jsx
      ProfileForm.jsx
      TunnelTable.jsx
      TunnelDialog.jsx
      ActivityList.jsx
      SettingsView.jsx
      ThemeModeControl.jsx
      SetupFlow.jsx
```

The existing pure functions in `state.js` should be preserved or moved with tests rather than buried inside React components.

### Use shadcn/ui components and tokens

The UI should use shadcn/ui-style local components under `desktop/src/components/ui`. Components should be checked into the repo, not consumed from a runtime package. The implementation should prefer standard shadcn/ui patterns for:

- `Button` for commands.
- `Input` and `Label` for form fields.
- `Textarea` only if a multiline field becomes necessary.
- `Switch` for tunnel enabled state and binary settings.
- `Select` or radio-style controls for profile and theme choices.
- `Tabs` or a compact sidebar for primary views if settings needs a separate surface.
- `Dialog` or `Sheet` for add/edit tunnel flows.
- `Badge` for runtime and per-tunnel status.
- `Table` for tunnel rows.
- `Alert` or inline messages for test and error feedback.

The visual direction should stay utilitarian and information-dense: restrained surfaces, predictable alignment, compact controls, 8px or smaller card radii, and no decorative hero treatment.

### Keep daemon interaction in a small API layer

React components should not build fetch URLs or know daemon endpoint details directly. A small API module should own:

```text
GET    /api/status
POST   /api/profiles
PUT    /api/profiles/{id}
DELETE /api/profiles/{id}
POST   /api/profiles/{id}/test-server
POST   /api/profiles/{id}/test-auth
POST   /api/runtime/start
POST   /api/runtime/stop
POST   /api/runtime/restart
GET    /api/events
GET    /api/profiles/{id}/export
POST   /api/profiles/import
```

If the current daemon has an active-profile endpoint, the UI should use it for profile switching. If it does not, implementation should first verify the daemon surface before introducing any frontend-only workaround.

### Complete the specified GUI workflows

The rewrite should make the existing `desktop-client` spec visible in the UI:

- If no profiles exist, show a setup flow instead of an empty editor.
- Provide profile selection and active profile switching.
- Allow deleting inactive profiles, while protecting the active profile from accidental deletion.
- Provide profile creation and editing for server URL, client ID, token, and reconnect interval.
- Keep token inputs masked by default.
- Provide tunnel add, edit, delete, strip path, and enabled controls.
- Save enabled changes and trigger runtime restart when required by the existing MVP behavior.
- Show overall runtime state and per-tunnel status.
- Show clear feedback for server and auth tests.
- Show recent logs, request summaries, and errors without request or response bodies.
- Preserve import and export for CLI YAML compatibility.

### Put theme controls in Settings

The app will include a settings view. The first settings capabilities should be:

- Theme mode: `system`, `light`, or `dark`.
- Daemon base URL, if the current `mmtunnel.daemon` localStorage override remains useful for local development.

Theme preference should be stored in `localStorage`, for example:

```text
mmtunnel.theme = system | light | dark
```

Application of theme should follow shadcn/ui convention:

```text
light/system-light -> document.documentElement has no dark class
dark/system-dark   -> document.documentElement has class "dark"
```

The CSS should define shadcn/ui-compatible variables for both themes and avoid a single-hue palette. Status colors should remain semantically distinct for connected, stopped, reconnecting, failed, and disabled states.

### Use Vite as the frontend build

The desktop build should run Vite and produce `desktop/dist` for Tauri. The current copy-only build script should be replaced or adapted so `npm run build` produces the same output location expected by `src-tauri/tauri.conf.json`.

Tests should remain runnable with `npm test`. The first implementation can keep Node test coverage for pure state and theme helpers, and add component tests only where the repo's dependency and runtime choices make them practical.

## Risks / Trade-offs

- Dependency growth: shadcn/ui brings React, Tailwind, Radix-style primitives, class merging utilities, and icons. This is acceptable because the current UI requirements now exceed what the vanilla renderer can comfortably support.
- Migration size: rewriting the UI and completing missing workflows in one change is larger than a visual-only pass. The upside is avoiding a temporary React shell that still lacks required desktop-client behavior.
- Theme flash: applying theme only after React mounts can briefly show the wrong theme. Add a small startup theme application path or keep the mount fast enough that this is not visible.
- API mismatch: profile switching or deletion may reveal missing daemon endpoints. Verify the daemon contract before assuming frontend-only state can represent active profile changes.
- Test churn: string-rendered UI tests will not map directly to React. Keep pure state tests stable and add workflow coverage around the most failure-prone user actions.

## Migration Plan

1. Add React, Vite, Tailwind CSS, shadcn/ui support libraries, and build configuration.
2. Establish shadcn/ui-compatible CSS variables for light and dark themes.
3. Build the React app shell, daemon API layer, status/event subscription, and theme provider.
4. Recreate existing visible UI behavior with shadcn/ui components.
5. Add the missing specified workflows: setup flow, profile switching, inactive profile deletion, tunnel dialogs, strip path editing, and test feedback.
6. Add the settings view with theme mode selection and daemon URL override if retained.
7. Replace the copy-only desktop build with a Vite build that writes to `dist`.
8. Update tests and docs to reflect the React/shadcn/ui desktop frontend.

Rollback is straightforward at the product level: the Go daemon, CLI, server protocol, profile store, and Tauri shell remain unchanged. The old vanilla frontend can be restored if the React migration is not accepted.
