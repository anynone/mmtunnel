import { useEffect, useState } from "react";
import { Button } from "./ui/button.jsx";
import { Dialog } from "./ui/dialog.jsx";
import { Input } from "./ui/input.jsx";
import { Label } from "./ui/label.jsx";
import { Switch } from "./ui/switch.jsx";
import { blankTunnel, slug } from "../state.js";

export function TunnelDialog({ open, tunnel, onClose, onSave }) {
  const [draft, setDraft] = useState(blankTunnel());
  useEffect(() => {
    setDraft(tunnel ? { ...tunnel } : blankTunnel());
  }, [tunnel, open]);
  const update = (field, value) => setDraft((current) => ({ ...current, [field]: value }));
  const save = () => {
    const name = draft.name || "New Tunnel";
    onSave({ ...draft, id: draft.id || slug(name, "tunnel"), name });
  };
  return (
    <Dialog open={open} title={tunnel?.id ? "Edit tunnel" : "Add tunnel"} onClose={onClose}>
      <div className="grid gap-3">
        <Field label="Name">
          <Input value={draft.name || ""} onChange={(event) => update("name", event.target.value)} />
        </Field>
        <Field label="Public Path">
          <Input value={draft.publicPath || ""} onChange={(event) => update("publicPath", event.target.value)} />
        </Field>
        <Field label="Target URL">
          <Input value={draft.target || ""} onChange={(event) => update("target", event.target.value)} />
        </Field>
        <div className="flex flex-wrap gap-6">
          <Toggle label="Strip path" checked={Boolean(draft.stripPath)} onChange={(value) => update("stripPath", value)} />
          <Toggle label="Enabled" checked={Boolean(draft.enabled)} onChange={(value) => update("enabled", value)} />
        </div>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="button" onClick={save}>
            Save Tunnel
          </Button>
        </div>
      </div>
    </Dialog>
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

function Toggle({ label, checked, onChange }) {
  return (
    <label className="flex items-center gap-2 text-sm">
      <Switch checked={checked} onCheckedChange={onChange} />
      {label}
    </label>
  );
}
