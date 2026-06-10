import { daemonBase } from "../state.js";

async function request(path, options = {}) {
  const response = await fetch(`${daemonBase()}${path}`, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  if (response.status === 204) return null;
  const type = response.headers.get("Content-Type") || "";
  return type.includes("application/json") ? response.json() : response.text();
}

export const daemonApi = {
  status: () => request("/api/status"),
  logs: () => request("/api/logs"),
  createProfile: (profile) => request("/api/profiles", { method: "POST", body: JSON.stringify(profile) }),
  updateProfile: (profile) => request(`/api/profiles/${encodeURIComponent(profile.id)}`, { method: "PUT", body: JSON.stringify(profile) }),
  deleteProfile: (id) => request(`/api/profiles/${encodeURIComponent(id)}`, { method: "DELETE" }),
  setActiveProfile: (id) => request(`/api/profiles/${encodeURIComponent(id)}/active`, { method: "POST" }),
  testServer: (id) => request(`/api/profiles/${encodeURIComponent(id)}/test-server`, { method: "POST" }),
  testAuth: (id) => request(`/api/profiles/${encodeURIComponent(id)}/test-auth`, { method: "POST" }),
  startRuntime: () => request("/api/runtime/start", { method: "POST" }),
  stopRuntime: () => request("/api/runtime/stop", { method: "POST" }),
  restartRuntime: () => request("/api/runtime/restart", { method: "POST" }),
  importProfile: (id, name, text) =>
    request(`/api/profiles/import?id=${encodeURIComponent(id)}&name=${encodeURIComponent(name)}`, {
      method: "POST",
      headers: { "Content-Type": "application/x-yaml" },
      body: text,
    }),
  exportProfile: async (id) => {
    const response = await fetch(`${daemonBase()}/api/profiles/${encodeURIComponent(id)}/export`);
    if (!response.ok) throw new Error(await response.text());
    return response.text();
  },
  openEvents: () => new EventSource(`${daemonBase()}/api/events`),
};
