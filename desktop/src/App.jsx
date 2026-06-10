import { Download, Upload } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { daemonApi } from "./lib/api.js";
import { applyTheme, getStoredTheme, setStoredTheme } from "./lib/theme.js";
import { blankProfile, blankTunnel, daemonBase, profileToPayload, slug } from "./state.js";
import { ActivityList } from "./components/ActivityList.jsx";
import { AppShell } from "./components/AppShell.jsx";
import { ProfileForm } from "./components/ProfileForm.jsx";
import { ProfileSelector } from "./components/ProfileSelector.jsx";
import { RuntimeControls } from "./components/RuntimeControls.jsx";
import { SettingsView } from "./components/SettingsView.jsx";
import { SetupFlow } from "./components/SetupFlow.jsx";
import { TunnelDialog } from "./components/TunnelDialog.jsx";
import { TunnelTable } from "./components/TunnelTable.jsx";
import { Alert } from "./components/ui/alert.jsx";
import { Button } from "./components/ui/button.jsx";

const runningStates = ["connecting", "connected", "reconnecting"];

export function App() {
  const [profiles, setProfiles] = useState([]);
  const [activeProfileId, setActiveProfileId] = useState("");
  const [runtime, setRuntime] = useState({ state: "stopped" });
  const [tunnels, setTunnels] = useState({});
  const [events, setEvents] = useState([]);
  const [editing, setEditing] = useState(blankProfile());
  const [view, setView] = useState("overview");
  const [notice, setNotice] = useState(null);
  const [dialogIndex, setDialogIndex] = useState(undefined);
  const [themeMode, setThemeMode] = useState(() => getStoredTheme());
  const [daemonUrl, setDaemonUrl] = useState(() => daemonBase());

  const activeProfile = useMemo(
    () => profiles.find((profile) => profile.id === activeProfileId) || profiles[0] || null,
    [profiles, activeProfileId],
  );

  useEffect(() => {
    const media = window.matchMedia?.("(prefers-color-scheme: dark)");
    const apply = () => applyTheme(themeMode, Boolean(media?.matches));
    apply();
    media?.addEventListener?.("change", apply);
    return () => media?.removeEventListener?.("change", apply);
  }, [themeMode]);

  useEffect(() => {
    refresh().catch(showError);
  }, []);

  useEffect(() => {
    const source = daemonApi.openEvents();
    for (const name of ["state", "log", "request", "error"]) {
      source.addEventListener(name, (event) => {
        const payload = JSON.parse(event.data);
        setEvents((items) => [...items.slice(-49), payload]);
        refresh().catch(() => {});
      });
    }
    source.onerror = () => setEvents((items) => [...items.slice(-49), { type: "error", message: "Event stream disconnected" }]);
    return () => source.close();
  }, []);

  async function refresh() {
    const status = await daemonApi.status();
    setProfiles(status.profiles || []);
    setActiveProfileId(status.activeProfileId || "");
    setRuntime(status.runtime || { state: "stopped" });
    setTunnels(status.tunnels || {});
    const selected = (status.profiles || []).find((profile) => profile.id === (status.activeProfileId || "")) || status.profiles?.[0] || blankProfile();
    setEditing(structuredCloneSafe(selected));
  }

  function showError(error) {
    const message = String(error.message || error);
    setNotice({ type: "error", message });
    setEvents((items) => [...items.slice(-49), { type: "error", message }]);
  }

  async function runAction(action, success) {
    try {
      await action();
      if (success) setNotice({ type: "success", message: success });
      await refresh();
    } catch (error) {
      showError(error);
    }
  }

  async function saveProfile(profile = editing) {
    const next = profileToPayload({ ...profile, id: profile.id || slug(profile.name), name: profile.name || "New Profile" });
    if (next.tunnels.length === 0) {
      next.tunnels = [];
    }
    const exists = profiles.some((item) => item.id === next.id);
    await (exists ? daemonApi.updateProfile(next) : daemonApi.createProfile(next));
    setEditing(structuredCloneSafe(next));
  }

  function updateTunnel(index, tunnel) {
    setEditing((current) => {
      const next = structuredCloneSafe(current);
      next.tunnels = next.tunnels || [];
      if (index === null || index === undefined) next.tunnels.push(tunnel);
      else next.tunnels[index] = tunnel;
      return next;
    });
  }

  async function toggleTunnel(index, enabled) {
    const next = structuredCloneSafe(editing);
    next.tunnels[index] = { ...next.tunnels[index], enabled };
    setEditing(next);
    if (!next.id) return;
    await runAction(async () => {
      await daemonApi.updateProfile(profileToPayload(next));
      if (next.id === activeProfileId && runningStates.includes(runtime.state)) {
        await daemonApi.restartRuntime();
      }
    }, `${next.tunnels[index].name} ${enabled ? "enabled" : "disabled"}`);
  }

  async function importProfile(file) {
    if (!file) return;
    const id = slug(file.name.replace(/\.(ya?ml)$/i, ""));
    const text = await file.text();
    await runAction(() => daemonApi.importProfile(id, file.name, text), "Profile imported");
  }

  async function exportProfile() {
    if (!editing.id) return;
    try {
      const text = await daemonApi.exportProfile(editing.id);
      const url = URL.createObjectURL(new Blob([text], { type: "application/x-yaml" }));
      const link = document.createElement("a");
      link.href = url;
      link.download = `${editing.id}.yaml`;
      link.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      showError(error);
    }
  }

  const title = activeProfile ? activeProfile.name || activeProfile.id : "Setup";
  const subtitle = activeProfile?.client?.server || "No profile configured";
  const setupMode = profiles.length === 0;

  return (
    <AppShell activeView={view} onViewChange={setView} runtimeState={runtime.state} title={title} subtitle={subtitle}>
      <div className="grid gap-3">
        {notice && <Alert variant={notice.type === "error" ? "destructive" : "default"}>{notice.message}</Alert>}

        {setupMode ? (
          <SetupFlow
            profile={editing}
            onProfileChange={setEditing}
            onSave={() => runAction(() => saveProfile(), "Profile created")}
            onAddTunnel={() => setDialogIndex(null)}
            onEditTunnel={setDialogIndex}
            onDeleteTunnel={(index) => setEditing(removeTunnel(editing, index))}
            onToggleTunnel={(index, checked) => setEditing(setTunnelEnabled(editing, index, checked))}
          />
        ) : view === "settings" ? (
          <SettingsView
            themeMode={themeMode}
            onThemeModeChange={(mode) => {
              const resolved = setStoredTheme(mode);
              setThemeMode(resolved);
            }}
            daemonUrl={daemonUrl}
            onDaemonUrlChange={setDaemonUrl}
          />
        ) : view === "activity" ? (
          <ActivityList events={events} />
        ) : (
          <>
            <div className="grid gap-3 lg:grid-cols-[320px_1fr]">
              <div className="grid content-start gap-3">
                <ProfileSelector
                  profiles={profiles}
                  activeProfileId={activeProfileId}
                  onSelect={(id) => runAction(() => daemonApi.setActiveProfile(id), "Active profile switched")}
                  onCreate={() => setEditing(blankProfile())}
                  onDelete={(id) => runAction(() => daemonApi.deleteProfile(id), "Profile deleted")}
                />
                <RuntimeControls
                  runtimeState={runtime.state}
                  onStart={() => runAction(daemonApi.startRuntime)}
                  onStop={() => runAction(daemonApi.stopRuntime)}
                  onRestart={() => runAction(daemonApi.restartRuntime)}
                />
                <ImportExport onImport={importProfile} onExport={exportProfile} disabled={!editing.id} />
              </div>
              <ProfileForm
                profile={editing}
                onChange={setEditing}
                onSave={() => runAction(() => saveProfile(), "Profile saved")}
                onTestServer={() => runAction(() => daemonApi.testServer(editing.id), "Server reachability succeeded")}
                onTestAuth={() => runAction(() => daemonApi.testAuth(editing.id), "Authentication succeeded")}
              />
            </div>
            <TunnelTable
              tunnels={editing.tunnels || []}
              statuses={editing.id === activeProfileId ? tunnels : {}}
              onAdd={() => setDialogIndex(null)}
              onEdit={setDialogIndex}
              onDelete={(index) => setEditing(removeTunnel(editing, index))}
              onToggle={toggleTunnel}
            />
            <ActivityList events={events.slice(-6)} />
          </>
        )}
      </div>
      <TunnelDialog
        open={dialogIndex !== undefined}
        tunnel={dialogIndex === null ? blankTunnel() : editing.tunnels?.[dialogIndex]}
        onClose={() => setDialogIndex(undefined)}
        onSave={(tunnel) => {
          updateTunnel(dialogIndex, tunnel);
          setDialogIndex(undefined);
        }}
      />
    </AppShell>
  );
}

function ImportExport({ onImport, onExport, disabled }) {
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <h2 className="mb-3 text-base font-semibold">YAML</h2>
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => document.querySelector("#import-profile-file")?.click()}>
          <Upload className="h-4 w-4" />
          Import
        </Button>
        <Button variant="outline" disabled={disabled} onClick={onExport}>
          <Download className="h-4 w-4" />
          Export
        </Button>
        <input id="import-profile-file" type="file" accept=".yaml,.yml" hidden onChange={(event) => onImport(event.target.files?.[0])} />
      </div>
    </section>
  );
}

function structuredCloneSafe(value) {
  return JSON.parse(JSON.stringify(value || blankProfile()));
}

function removeTunnel(profile, index) {
  const next = structuredCloneSafe(profile);
  next.tunnels = (next.tunnels || []).filter((_, itemIndex) => itemIndex !== index);
  return next;
}

function setTunnelEnabled(profile, index, enabled) {
  const next = structuredCloneSafe(profile);
  next.tunnels[index] = { ...next.tunnels[index], enabled };
  return next;
}
