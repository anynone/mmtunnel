import { cn } from "@/lib/utils.js";

export function Alert({ className, variant = "default", ...props }) {
  return (
    <div
      className={cn(
        "rounded-md border border-border bg-card px-3 py-2 text-sm text-card-foreground",
        variant === "destructive" && "border-destructive/40 bg-destructive/10 text-destructive",
        className,
      )}
      {...props}
    />
  );
}
