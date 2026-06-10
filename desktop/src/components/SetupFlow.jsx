import { Button } from "./ui/button.jsx";
import { ProfileForm } from "./ProfileForm.jsx";
import { TunnelTable } from "./TunnelTable.jsx";

export function SetupFlow({ profile, onProfileChange, onSave, onAddTunnel, onEditTunnel, onDeleteTunnel, onToggleTunnel }) {
  return (
    <div className="grid gap-4">
      <section className="rounded-lg border bg-card p-4 text-card-foreground">
        <h2 className="text-base font-semibold">Create your first profile</h2>
        <p className="mt-1 text-sm text-muted-foreground">Configure the local client identity and at least one tunnel when you are ready.</p>
      </section>
      <ProfileForm profile={profile} onChange={onProfileChange} onSave={onSave} onTestServer={() => {}} onTestAuth={() => {}} />
      <TunnelTable tunnels={profile.tunnels || []} statuses={{}} onAdd={onAddTunnel} onEdit={onEditTunnel} onDelete={onDeleteTunnel} onToggle={onToggleTunnel} />
      <div className="flex justify-end">
        <Button onClick={onSave}>Create Profile</Button>
      </div>
    </div>
  );
}
