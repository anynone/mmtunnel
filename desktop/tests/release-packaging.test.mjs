import { readFile } from "node:fs/promises";
import { test } from "node:test";
import assert from "node:assert/strict";

const tauriConfigPath = new URL("../src-tauri/tauri.conf.json", import.meta.url);
const workflowPath = new URL("../../.github/workflows/desktop-release.yml", import.meta.url);
const packageJsonPath = new URL("../package.json", import.meta.url);
const packageLockPath = new URL("../package-lock.json", import.meta.url);
const indexHtmlPath = new URL("../index.html", import.meta.url);
const appJsPath = new URL("../src/app.js", import.meta.url);
const stateJsPath = new URL("../src/state.js", import.meta.url);

test("tauri release bundle formats and sidecar path are explicit", async () => {
  const config = JSON.parse(await readFile(tauriConfigPath, "utf8"));

  assert.deepEqual(config.bundle.targets, ["dmg", "appimage", "deb", "nsis", "msi"]);
  assert.deepEqual(config.bundle.externalBin, ["binaries/tunnel-daemon"]);
});

test("desktop package and app identity use MM Tunnel naming", async () => {
  const config = JSON.parse(await readFile(tauriConfigPath, "utf8"));
  const packageJson = JSON.parse(await readFile(packageJsonPath, "utf8"));
  const packageLock = JSON.parse(await readFile(packageLockPath, "utf8"));
  const indexHtml = await readFile(indexHtmlPath, "utf8");
  const appJs = await readFile(appJsPath, "utf8");
  const stateJs = await readFile(stateJsPath, "utf8");

  assert.equal(packageJson.name, "mmtunnel-desktop");
  assert.equal(packageLock.name, "mmtunnel-desktop");
  assert.equal(packageLock.packages[""].name, "mmtunnel-desktop");
  assert.equal(config.productName, "MM Tunnel");
  assert.equal(config.identifier, "cn.anynone.mmtunnel");
  assert.match(indexHtml, /<title>MM Tunnel<\/title>/);
  assert.match(appJs, /<h1>MM Tunnel<\/h1>/);
  assert.match(stateJs, /mmtunnel\.daemon/);
  assert.doesNotMatch(`${indexHtml}\n${appJs}\n${stateJs}`, /MM Socket|mmsocket\.daemon/);
});

test("desktop release workflow only runs for published GitHub releases", async () => {
  const workflow = await readFile(workflowPath, "utf8");

  assert.match(workflow, /^on:\n  release:\n    types: \[published\]$/m);
  assert.doesNotMatch(workflow, /^\s*(push|pull_request):/m);
});

test("desktop release workflow covers required target matrix", async () => {
  const workflow = await readFile(workflowPath, "utf8");
  const targets = [
    "x86_64-apple-darwin",
    "aarch64-apple-darwin",
    "x86_64-unknown-linux-gnu",
    "aarch64-unknown-linux-gnu",
    "x86_64-pc-windows-msvc",
    "aarch64-pc-windows-msvc",
  ];

  for (const target of targets) {
    assert.match(workflow, new RegExp(`target: ${target}`));
  }
});

test("desktop release workflow builds sidecars and uploads bundles to the release", async () => {
  const workflow = await readFile(workflowPath, "utf8");

  assert.match(workflow, /go build -buildvcs=false/);
  assert.match(workflow, /desktop\/src-tauri\/binaries\/tunnel-daemon-\$\{\{ matrix\.target \}\}/);
  assert.match(workflow, /uses: tauri-apps\/tauri-action@/);
  assert.match(workflow, /projectPath: desktop/);
  assert.match(workflow, /releaseId: \$\{\{ github\.event\.release\.id \}\}/);
});
