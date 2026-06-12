import { cn } from "@/lib/utils";
import type { TaskPriority, TaskStatus } from "@/lib/types";

const statusStyles: Record<TaskStatus, string> = {
  todo: "bg-surface-2 text-muted-foreground border-border",
  in_progress: "bg-[var(--accent-soft)] text-accent border-accent/30",
  done: "bg-success/10 text-success border-success/30",
};

const statusLabels: Record<TaskStatus, string> = {
  todo: "To do",
  in_progress: "In progress",
  done: "Done",
};

const priorityStyles: Record<TaskPriority, string> = {
  low: "bg-surface-2 text-muted-foreground border-border",
  medium: "bg-warning/10 text-warning border-warning/30",
  high: "bg-danger/10 text-danger border-danger/30",
};

function Pill({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium leading-none",
        className,
      )}
    >
      {children}
    </span>
  );
}

export function StatusBadge({ status }: { status: TaskStatus }) {
  return (
    <Pill className={statusStyles[status]}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {statusLabels[status]}
    </Pill>
  );
}

export function PriorityBadge({ priority }: { priority: TaskPriority }) {
  return (
    <Pill className={priorityStyles[priority]}>
      {priority.charAt(0).toUpperCase() + priority.slice(1)}
    </Pill>
  );
}
