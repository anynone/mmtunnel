import { cn } from "@/lib/utils.js";

export function Badge({ className, variant = "default", ...props }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium",
        variant === "default" && "border-transparent bg-secondary text-secondary-foreground",
        variant === "outline" && "border-border text-foreground",
        variant === "destructive" && "border-transparent bg-destructive text-destructive-foreground",
        className,
      )}
      {...props}
    />
  );
}
