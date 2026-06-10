import { X } from "lucide-react";
import { Button } from "./button.jsx";
import { cn } from "@/lib/utils.js";

export function Dialog({ open, title, children, onClose, className }) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-background/70 p-4 backdrop-blur-sm">
      <section className={cn("w-full max-w-2xl rounded-lg border bg-card p-4 text-card-foreground shadow-lg", className)}>
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-base font-semibold">{title}</h2>
          <Button type="button" variant="ghost" size="icon" onClick={onClose} aria-label="Close dialog">
            <X className="h-4 w-4" />
          </Button>
        </div>
        {children}
      </section>
    </div>
  );
}
