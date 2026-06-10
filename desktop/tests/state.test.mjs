import test from "node:test";
import assert from "node:assert/strict";
import { applyTunnelToggle, profileToPayload, renderStatusLabel, slug } from "../src/state.js";
import { resolveThemeMode, themeClassForMode } from "../src/lib/theme.js";

test("profileToPayload defaults tunnel enabled and stripPath", () => {
  const payload = profileToPayload({ tunnels: [{ id: "api", name: "API" }] });
  assert.equal(payload.tunnels[0].enabled, true);
  assert.equal(payload.tunnels[0].stripPath, true);
});

test("applyTunnelToggle saves profile and restarts running runtime", async () => {
  const calls = [];
  await applyTunnelToggle({
    runtimeState: "connected",
    tunnelId: "api",
    enabled: false,
    profile: { id: "company", tunnels: [{ id: "api", enabled: true }] },
    api: async (path, options) => calls.push([path, options.method]),
  });
  assert.deepEqual(calls, [
    ["/api/profiles/company", "PUT"],
    ["/api/runtime/restart", "POST"],
  ]);
});

test("renderStatusLabel maps connected state", () => {
  assert.equal(renderStatusLabel("connected"), "Connected");
});

test("slug normalizes profile and tunnel names", () => {
  assert.equal(slug("Company API Tunnel"), "company-api-tunnel");
  assert.equal(slug("  "), "profile");
});

test("themeClassForMode applies explicit and system dark modes", () => {
  assert.equal(themeClassForMode("light", true), "");
  assert.equal(themeClassForMode("dark", false), "dark");
  assert.equal(themeClassForMode("system", true), "dark");
  assert.equal(themeClassForMode("system", false), "");
});

test("resolveThemeMode rejects invalid stored preferences", () => {
  assert.equal(resolveThemeMode("dark"), "dark");
  assert.equal(resolveThemeMode("sepia"), "system");
  assert.equal(resolveThemeMode(null), "system");
});
