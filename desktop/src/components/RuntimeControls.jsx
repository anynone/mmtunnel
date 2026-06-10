import { Play, RotateCw, Square } from "lucide-react";
import { Button } from "./ui/button.jsx";

export function RuntimeControls({ runtimeState, onStart, onStop, onRestart }) {
  const running = ["connecting", "connected", "reconnecting"].includes(runtimeState);
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Runtime</h2>
        <span className="text-sm text-muted-foreground">{runtimeState || "unknown"}</span>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button onClick={onStart} disabled={running}>
          <Play className="h-4 w-4" />
          Start
        </Button>
        <Button variant="outline" onClick={onStop} disabled={!running}>
          <Square className="h-4 w-4" />
          Stop
        </Button>
        <Button variant="secondary" onClick={onRestart} disabled={!running}>
          <RotateCw className="h-4 w-4" />
          Restart
        </Button>
      </div>
    </section>
  );
}
