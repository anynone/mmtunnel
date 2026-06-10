import { cn } from "@/lib/utils.js";

export function Table({ className, ...props }) {
  return <table className={cn("w-full caption-bottom text-sm", className)} {...props} />;
}

export function Th({ className, ...props }) {
  return <th className={cn("h-9 border-b px-2 text-left align-middle font-medium text-muted-foreground", className)} {...props} />;
}

export function Td({ className, ...props }) {
  return <td className={cn("border-b px-2 py-2 align-middle", className)} {...props} />;
}
