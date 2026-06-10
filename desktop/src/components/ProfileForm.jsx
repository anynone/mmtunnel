import { Eye, Save } from "lucide-react";
import { useState } from "react";
import { Button } from "./ui/button.jsx";
import { Input } from "./ui/input.jsx";
import { Label } from "./ui/label.jsx";

export function ProfileForm({ profile, onChange, onSave, onTestServer, onTestAuth }) {
  const [showToken, setShowToken] = useState(false);
  const client = profile.client || {};
  const updateClient = (field, value) => onChange({ ...profile, client: { ...client, [field]: value } });
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Profile</h2>
        <Button size="sm" onClick={onSave}>
          <Save className="h-4 w-4" />
          Save
        </Button>
      </div>
      <div className="grid gap-3">
        <Field label="Name">
          <Input value={profile.name || ""} onChange={(event) => onChange({ ...profile, name: event.target.value })} />
        </Field>
        <Field label="Server URL">
          <Input value={client.server || ""} onChange={(event) => updateClient("server", event.target.value)} />
        </Field>
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="Client ID">
            <Input value={client.id || ""} onChange={(event) => updateClient("id", event.target.value)} />
          </Field>
          <Field label="Reconnect">
            <Input value={client.reconnectInterval || "5s"} onChange={(event) => updateClient("reconnectInterval", event.target.value)} />
          </Field>
        </div>
        <Field label="Token">
          <div className="flex gap-2">
            <Input type={showToken ? "text" : "password"} value={client.token || ""} onChange={(event) => updateClient("token", event.target.value)} />
            <Button type="button" variant="outline" size="icon" onClick={() => setShowToken((value) => !value)} aria-label="Toggle token visibility">
              <Eye className="h-4 w-4" />
            </Button>
          </div>
        </Field>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={onTestServer} disabled={!profile.id}>
            Test Server
          </Button>
          <Button type="button" variant="outline" onClick={onTestAuth} disabled={!profile.id}>
            Test Auth
          </Button>
        </div>
      </div>
    </section>
  );
}

function Field({ label, children }) {
  return (
    <div className="grid gap-1.5">
      <Label>{label}</Label>
      {children}
    </div>
  );
}
