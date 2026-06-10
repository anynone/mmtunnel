import { Alert } from "./ui/alert.jsx";

export function ActivityList({ events }) {
  const rows = (events || []).slice(-20).reverse();
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <h2 className="text-base font-semibold">Recent Activity</h2>
      <div className="mt-3 grid gap-2">
        {rows.length === 0 && <p className="text-sm text-muted-foreground">No recent activity.</p>}
        {rows.map((event, index) => (
          <Alert key={`${event.type || "event"}-${index}`} variant={event.type === "error" ? "destructive" : "default"}>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-medium uppercase text-muted-foreground">{event.type || "event"}</span>
              <span>{formatEvent(event)}</span>
            </div>
          </Alert>
        ))}
      </div>
    </section>
  );
}

function formatEvent(event) {
  if (event.message) return event.message;
  const parts = [event.method, event.path, event.status, event.duration, event.tunnelName || event.tunnel];
  const text = parts.filter(Boolean).join(" ");
  return text || "Runtime event";
}
