import { Save } from "lucide-react";
import { daemonBase } from "../state.js";
import { Button } from "./ui/button.jsx";
import { Input } from "./ui/input.jsx";
import { Label } from "./ui/label.jsx";
import { ThemeModeControl } from "./ThemeModeControl.jsx";

export function SettingsView({ themeMode, onThemeModeChange, daemonUrl, onDaemonUrlChange }) {
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <h2 className="text-base font-semibold">Settings</h2>
      <div className="mt-4 grid max-w-xl gap-4">
        <div className="grid gap-1.5">
          <Label>Theme</Label>
          <ThemeModeControl value={themeMode} onChange={onThemeModeChange} />
        </div>
        <div className="grid gap-1.5">
          <Label>Daemon URL</Label>
          <div className="flex gap-2">
            <Input value={daemonUrl} onChange={(event) => onDaemonUrlChange(event.target.value)} placeholder={daemonBase()} />
            <Button
              variant="outline"
              size="icon"
              aria-label="Save daemon URL"
              onClick={() => {
                localStorage.setItem("mmtunnel.daemon", daemonUrl || "http://127.0.0.1:19081");
              }}
            >
              <Save className="h-4 w-4" />
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">Used for local development or custom daemon ports.</p>
        </div>
      </div>
    </section>
  );
}
