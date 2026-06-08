export function daemonBase() {
  return localStorage.getItem("mmsocket.daemon") || "http://127.0.0.1:19081";
}

export function blankProfile() {
  return {
    id: "",
    name: "",
    client: { id: "", token: "", server: "wss://callback.example.com/_tunnel/connect", reconnectInterval: "5s" },
    tunnels: [],
  };
}

export function profileToPayload(profile) {
  return {
    ...profile,
    tunnels: (profile.tunnels || []).map((tunnel) => ({
      stripPath: true,
      enabled: true,
      ...tunnel,
    })),
  };
}

export async function applyTunnelToggle({ api, profile, tunnelId, enabled, runtimeState }) {
  const tunnel = profile.tunnels.find((item) => item.id === tunnelId);
  if (!tunnel) throw new Error(`Tunnel ${tunnelId} not found`);
  tunnel.enabled = enabled;
  await api(`/api/profiles/${profile.id}`, { method: "PUT", body: JSON.stringify(profile) });
  if (["connecting", "connected", "reconnecting"].includes(runtimeState)) {
    await api("/api/runtime/restart", { method: "POST" });
  }
}

export function renderStatusLabel(state) {
  const labels = {
    stopped: "Stopped",
    connecting: "Connecting",
    connected: "Connected",
    reconnecting: "Reconnecting",
    stopping: "Stopping",
    error: "Error",
  };
  return labels[state] || "Unknown";
}
