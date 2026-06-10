import { Activity, Network, Settings } from "lucide-react";
import { Button } from "./ui/button.jsx";
import { Badge } from "./ui/badge.jsx";
import { cn } from "@/lib/utils.js";

export function AppShell({ activeView, onViewChange, runtimeState, title, subtitle, children }) {
  const tabs = [
    { id: "overview", label: "Overview", icon: Network },
    { id: "activity", label: "Activity", icon: Activity },
    { id: "settings", label: "Settings", icon: Settings },
  ];
  return (
    <main className="min-h-screen bg-background text-foreground">
      <div className="mx-auto flex max-w-[1220px] gap-4 p-4">
        <aside className="hidden w-52 shrink-0 rounded-lg border bg-card p-3 text-card-foreground md:block">
          <div className="px-2 pb-4">
            <div className="text-lg font-semibold">MM Tunnel</div>
            <div className="mt-1 text-xs text-muted-foreground">Desktop control plane</div>
          </div>
          <nav className="grid gap-1">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <Button
                  key={tab.id}
                  variant={activeView === tab.id ? "secondary" : "ghost"}
                  className="justify-start"
                  onClick={() => onViewChange(tab.id)}
                >
                  <Icon className="h-4 w-4" />
                  {tab.label}
                </Button>
              );
            })}
          </nav>
        </aside>
        <section className="min-w-0 flex-1">
          <header className="mb-3 rounded-lg border bg-card p-4 text-card-foreground">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold">{title}</h1>
                <p className="mt-1 text-sm text-muted-foreground">{subtitle}</p>
              </div>
              <Badge className={cn("status-badge", `status-${runtimeState || "unknown"}`)}>{runtimeState || "unknown"}</Badge>
            </div>
            <nav className="mt-4 grid grid-cols-3 gap-2 md:hidden">
              {tabs.map((tab) => (
                <Button key={tab.id} variant={activeView === tab.id ? "secondary" : "outline"} size="sm" onClick={() => onViewChange(tab.id)}>
                  {tab.label}
                </Button>
              ))}
            </nav>
          </header>
          {children}
        </section>
      </div>
    </main>
  );
}
