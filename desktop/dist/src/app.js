import {
  applyTunnelToggle,
  blankProfile,
  daemonBase,
  profileToPayload,
  renderStatusLabel,
} from "./state.js";

const state = {
  profiles: [],
  activeProfileId: "",
  runtime: { state: "stopped" },
  tunnels: {},
  events: [],
  editing: null,
};

const app = document.querySelector("#app");

async function api(path, options = {}) {
  const response = await fetch(`${daemonBase()}${path}`, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  if (response.status === 204) return null;
  return response.json();
}

async function refresh() {
  const status = await api("/api/status");
  state.profiles = status.profiles || [];
  state.activeProfileId = status.activeProfileId || "";
  state.runtime = status.runtime || { state: "stopped" };
  state.tunnels = status.tunnels || {};
  render();
}

function activeProfile() {
  return state.profiles.find((profile) => profile.id === state.activeProfileId) || state.profiles[0];
}

function render() {
  const profile = state.editing || activeProfile() || blankProfile();
  app.innerHTML = `
    <section class="shell">
      <header class="topbar">
        <div>
          <h1>MM Socket</h1>
          <p>${profile.name || "New profile"} · ${profile.client.server || "No server configured"}</p>
        </div>
        <strong class="status ${state.runtime.state}">${renderStatusLabel(state.runtime.state)}</strong>
      </header>

      <section class="toolbar">
        <button id="start">Start</button>
        <button id="stop">Stop</button>
        <button id="restart">Restart</button>
        <button id="new-profile">New Profile</button>
        <button id="save-profile">Save Profile</button>
        <button id="import-profile">Import YAML</button>
        <button id="export-profile">Export YAML</button>
        <input id="import-file" type="file" accept=".yaml,.yml" hidden />
      </section>

      <section class="grid">
        <form id="profile-form" class="panel">
          <h2>Profile</h2>
          <label>Name <input name="name" value="${escapeAttr(profile.name)}" /></label>
          <label>Server URL <input name="server" value="${escapeAttr(profile.client.server)}" placeholder="wss://callback.example.com/_tunnel/connect" /></label>
          <label>Client ID <input name="clientId" value="${escapeAttr(profile.client.id)}" /></label>
          <label>Token <input name="token" type="password" value="${escapeAttr(profile.client.token)}" /></label>
          <label>Reconnect <input name="reconnectInterval" value="${escapeAttr(profile.client.reconnectInterval || "5s")}" /></label>
          <div class="actions">
            <button type="button" id="test-server">Test Server</button>
            <button type="button" id="test-auth">Test Auth</button>
          </div>
        </form>

        <section class="panel">
          <div class="section-title">
            <h2>Tunnels</h2>
            <button id="add-tunnel">Add Tunnel</button>
          </div>
          <table>
            <thead><tr><th>Name</th><th>Public Path</th><th>Target</th><th>Enabled</th><th>Status</th><th></th></tr></thead>
            <tbody>
              ${(profile.tunnels || []).map((tunnel, index) => tunnelRow(tunnel, index)).join("")}
            </tbody>
          </table>
        </section>
      </section>

      <section class="panel">
        <h2>Recent Activity</h2>
        <ol class="events">
          ${state.events.slice(-8).reverse().map((event) => `<li>${escapeHtml(event.message || event.type || "event")}</li>`).join("")}
        </ol>
      </section>
    </section>
  `;
  bind(profile);
}

function tunnelRow(tunnel, index) {
  const status = state.tunnels[tunnel.id] || (tunnel.enabled ? "pending" : "disabled");
  return `
    <tr>
      <td><input data-tunnel="${index}" data-field="name" value="${escapeAttr(tunnel.name)}" /></td>
      <td><input data-tunnel="${index}" data-field="publicPath" value="${escapeAttr(tunnel.publicPath)}" /></td>
      <td><input data-tunnel="${index}" data-field="target" value="${escapeAttr(tunnel.target)}" /></td>
      <td><input data-tunnel="${index}" data-field="enabled" type="checkbox" ${tunnel.enabled ? "checked" : ""} /></td>
      <td><span class="badge ${status}">${status}</span></td>
      <td><button data-delete="${index}">Delete</button></td>
    </tr>
  `;
}

function bind(profile) {
  document.querySelector("#start").onclick = () => api("/api/runtime/start", { method: "POST" }).then(refresh).catch(showError);
  document.querySelector("#stop").onclick = () => api("/api/runtime/stop", { method: "POST" }).then(refresh).catch(showError);
  document.querySelector("#restart").onclick = () => api("/api/runtime/restart", { method: "POST" }).then(refresh).catch(showError);
  document.querySelector("#new-profile").onclick = () => {
    state.editing = blankProfile();
    render();
  };
  document.querySelector("#save-profile").onclick = () => saveProfile(profile).catch(showError);
  document.querySelector("#export-profile").onclick = () => exportProfile(profile).catch(showError);
  document.querySelector("#import-profile").onclick = () => document.querySelector("#import-file").click();
  document.querySelector("#import-file").onchange = (event) => importProfile(event.target.files[0]).catch(showError);
  document.querySelector("#add-tunnel").onclick = () => {
    profile.tunnels.push({ id: `tunnel-${Date.now()}`, name: "New Tunnel", publicPath: "/api", target: "http://127.0.0.1:3000", stripPath: true, enabled: true });
    render();
  };
  document.querySelector("#test-server").onclick = () => testProfile(profile, "test-server");
  document.querySelector("#test-auth").onclick = () => testProfile(profile, "test-auth");
  document.querySelectorAll("[data-delete]").forEach((button) => {
    button.onclick = () => {
      profile.tunnels.splice(Number(button.dataset.delete), 1);
      render();
    };
  });
  document.querySelectorAll("[data-tunnel]").forEach((input) => {
    input.onchange = async () => {
      const tunnel = profile.tunnels[Number(input.dataset.tunnel)];
      const field = input.dataset.field;
      tunnel[field] = input.type === "checkbox" ? input.checked : input.value;
      if (field === "enabled" && activeProfile()?.id === profile.id) {
        await applyTunnelToggle({ api, profile: profileToPayload(readProfileForm(profile)), tunnelId: tunnel.id, enabled: tunnel.enabled, runtimeState: state.runtime.state });
        state.events.push({ type: "log", message: `${tunnel.name} ${tunnel.enabled ? "enabled" : "disabled"}` });
        state.editing = null;
        await refresh();
      }
    };
  });
}

async function saveProfile(profile) {
  const payload = profileToPayload(readProfileForm(profile));
  const existing = state.profiles.some((item) => item.id === payload.id);
  await api(existing ? `/api/profiles/${payload.id}` : "/api/profiles", {
    method: existing ? "PUT" : "POST",
    body: JSON.stringify(payload),
  });
  state.editing = null;
  await refresh();
}

function readProfileForm(profile) {
  const form = new FormData(document.querySelector("#profile-form"));
  return {
    ...profile,
    id: profile.id || slug(form.get("name") || "profile"),
    name: String(form.get("name") || ""),
    client: {
      id: String(form.get("clientId") || ""),
      token: String(form.get("token") || ""),
      server: String(form.get("server") || ""),
      reconnectInterval: String(form.get("reconnectInterval") || "5s"),
    },
  };
}

async function testProfile(profile, action) {
  try {
    await saveProfile(profile);
    await api(`/api/profiles/${profile.id}/${action}`, { method: "POST" });
    state.events.push({ type: "log", message: `${action} succeeded` });
  } catch (error) {
    showError(error);
  }
  render();
}

async function exportProfile(profile) {
  if (!profile.id) throw new Error("Save the profile before exporting");
  const response = await fetch(`${daemonBase()}/api/profiles/${profile.id}/export`);
  if (!response.ok) throw new Error(await response.text());
  const text = await response.text();
  const blob = new Blob([text], { type: "application/x-yaml" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${profile.id}.yaml`;
  a.click();
  URL.revokeObjectURL(url);
}

async function importProfile(file) {
  if (!file) return;
  const id = slug(file.name.replace(/\.(ya?ml)$/i, ""));
  const response = await fetch(`${daemonBase()}/api/profiles/import?id=${encodeURIComponent(id)}&name=${encodeURIComponent(file.name)}`, {
    method: "POST",
    body: await file.text(),
  });
  if (!response.ok) throw new Error(await response.text());
  await refresh();
}

function showError(error) {
  state.events.push({ type: "error", message: String(error.message || error) });
  render();
}

function connectEvents() {
  const events = new EventSource(`${daemonBase()}/api/events`);
  for (const name of ["state", "log", "request", "error"]) {
    events.addEventListener(name, (event) => {
      state.events.push(JSON.parse(event.data));
      refresh().catch(() => {});
    });
  }
}

function slug(value) {
  return String(value).trim().toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "profile";
}

function escapeHtml(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[char]);
}

function escapeAttr(value) {
  return escapeHtml(value || "");
}

refresh().then(connectEvents).catch((error) => {
  state.events.push({ type: "error", message: `Unable to reach daemon: ${error.message}` });
  render();
});
