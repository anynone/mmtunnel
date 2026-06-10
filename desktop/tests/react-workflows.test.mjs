import { readFile } from "node:fs/promises";
import { test } from "node:test";
import assert from "node:assert/strict";

const appPath = new URL("../src/App.jsx", import.meta.url);
const apiPath = new URL("../src/lib/api.js", import.meta.url);
const settingsPath = new URL("../src/components/SettingsView.jsx", import.meta.url);
const tunnelDialogPath = new URL("../src/components/TunnelDialog.jsx", import.meta.url);

test("React UI wires profile switching, deletion, import, export, and runtime actions", async () => {
  const app = await readFile(appPath, "utf8");
  const api = await readFile(apiPath, "utf8");

  assert.match(api, /\/api\/profiles\/.*\/active/);
  assert.match(api, /deleteProfile/);
  assert.match(api, /importProfile/);
  assert.match(api, /exportProfile/);
  assert.match(app, /daemonApi\.setActiveProfile/);
  assert.match(app, /daemonApi\.deleteProfile/);
  assert.match(app, /daemonApi\.startRuntime/);
  assert.match(app, /daemonApi\.stopRuntime/);
  assert.match(app, /daemonApi\.restartRuntime/);
});

test("React UI includes settings theme control and tunnel strip path editing", async () => {
  const settings = await readFile(settingsPath, "utf8");
  const tunnelDialog = await readFile(tunnelDialogPath, "utf8");

  assert.match(settings, /ThemeModeControl/);
  assert.match(settings, /mmtunnel\.daemon/);
  assert.match(tunnelDialog, /Strip path/);
  assert.match(tunnelDialog, /Enabled/);
});
