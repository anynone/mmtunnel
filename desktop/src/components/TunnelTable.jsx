import { Edit, Plus, Trash2 } from "lucide-react";
import { Badge } from "./ui/badge.jsx";
import { Button } from "./ui/button.jsx";
import { Switch } from "./ui/switch.jsx";
import { Table, Td, Th } from "./ui/table.jsx";
import { cn } from "@/lib/utils.js";

export function TunnelTable({ tunnels, statuses, onAdd, onEdit, onDelete, onToggle }) {
  return (
    <section className="rounded-lg border bg-card p-4 text-card-foreground">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Tunnels</h2>
        <Button size="sm" variant="outline" onClick={onAdd}>
          <Plus className="h-4 w-4" />
          Add Tunnel
        </Button>
      </div>
      <div className="overflow-x-auto">
        <Table>
          <thead>
            <tr>
              <Th>Name</Th>
              <Th>Public Path</Th>
              <Th>Target</Th>
              <Th>Strip</Th>
              <Th>Enabled</Th>
              <Th>Status</Th>
              <Th className="w-24"></Th>
            </tr>
          </thead>
          <tbody>
            {(tunnels || []).map((tunnel, index) => {
              const status = statuses?.[tunnel.id] || (tunnel.enabled ? "pending" : "disabled");
              return (
                <tr key={tunnel.id || index}>
                  <Td className="font-medium">{tunnel.name || "Untitled"}</Td>
                  <Td className="font-mono text-xs">{tunnel.publicPath}</Td>
                  <Td className="max-w-[260px] truncate font-mono text-xs">{tunnel.target}</Td>
                  <Td>{tunnel.stripPath ? "yes" : "no"}</Td>
                  <Td>
                    <Switch checked={Boolean(tunnel.enabled)} onCheckedChange={(checked) => onToggle(index, checked)} />
                  </Td>
                  <Td>
                    <Badge className={cn("status-badge", `status-${status}`)}>{status}</Badge>
                  </Td>
                  <Td>
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" onClick={() => onEdit(index)} aria-label={`Edit ${tunnel.name}`}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => onDelete(index)} aria-label={`Delete ${tunnel.name}`}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </Td>
                </tr>
              );
            })}
          </tbody>
        </Table>
      </div>
      {(!tunnels || tunnels.length === 0) && <p className="mt-3 text-sm text-muted-foreground">No tunnels configured.</p>}
    </section>
  );
}
