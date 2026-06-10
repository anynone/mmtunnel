import { Check, Plus, Trash2 } from "lucide-react";
import { Button } from "./ui/button.jsx";

export function ProfileSelector({ profiles, activeProfileId, onSelect, onCreate, onDelete }) {
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Profiles</h2>
        <Button size="sm" variant="outline" onClick={onCreate}>
          <Plus className="h-4 w-4" />
          New
        </Button>
      </div>
      <div className="grid gap-1">
        {profiles.map((profile) => {
          const active = profile.id === activeProfileId;
          return (
            <div key={profile.id} className="flex items-center gap-1 rounded-md border bg-background p-1">
              <Button variant={active ? "secondary" : "ghost"} className="min-w-0 flex-1 justify-start" onClick={() => !active && onSelect(profile.id)}>
                {active && <Check className="h-4 w-4" />}
                <span className="truncate">{profile.name || profile.id}</span>
              </Button>
              <Button variant="ghost" size="icon" disabled={active} onClick={() => onDelete(profile.id)} aria-label={`Delete ${profile.name || profile.id}`}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          );
        })}
      </div>
      <p className="mt-2 text-xs text-muted-foreground">Active profiles are protected from deletion. Switch first to remove another profile.</p>
    </section>
  );
}
