"use client";

import { motion } from "framer-motion";
import { Check, Pencil, Trash2, Calendar, AlertTriangle, History } from "lucide-react";
import { format, isPast, isToday } from "date-fns";
import { StatusBadge, PriorityBadge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { Task } from "@/lib/types";

interface Props {
  task: Task;
  showOwner?: boolean;
  onToggleComplete: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onHistory: (task: Task) => void;
}

export function TaskCard({
  task,
  showOwner,
  onToggleComplete,
  onEdit,
  onDelete,
  onHistory,
}: Props) {
  const done = task.status === "done";
  const due = task.dueDate ? new Date(task.dueDate) : null;
  const overdue = due && !done && isPast(due) && !isToday(due);

  return (
    <motion.li
      layout
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      transition={{ duration: 0.18, ease: [0.16, 1, 0.3, 1] }}
      className="group relative flex items-start gap-3 rounded-lg border border-border bg-surface p-4 transition-colors hover:border-border-strong"
    >
      {/* Complete toggle */}
      <button
        onClick={() => onToggleComplete(task)}
        aria-label={done ? "Mark as not done" : "Mark as done"}
        className={cn(
          "mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border transition-all",
          done
            ? "border-success bg-success text-white"
            : "border-border-strong hover:border-accent",
        )}
      >
        {done && <Check className="h-3 w-3" strokeWidth={3} />}
      </button>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <h3
            className={cn(
              "text-sm font-medium leading-snug",
              done && "text-muted line-through",
            )}
          >
            {task.title}
          </h3>

          {/* Actions — appear on hover, always visible on touch */}
          <div className="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100 max-sm:opacity-100">
            <button
              onClick={() => onHistory(task)}
              aria-label="View activity"
              className="rounded-md p-1.5 text-muted hover:bg-surface-2 hover:text-foreground"
            >
              <History className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => onEdit(task)}
              aria-label="Edit task"
              className="rounded-md p-1.5 text-muted hover:bg-surface-2 hover:text-foreground"
            >
              <Pencil className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => onDelete(task)}
              aria-label="Delete task"
              className="rounded-md p-1.5 text-muted hover:bg-danger/10 hover:text-danger"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {task.description && (
          <p
            className={cn(
              "mt-1 line-clamp-2 text-xs leading-relaxed text-muted-foreground",
              done && "line-through opacity-60",
            )}
          >
            {task.description}
          </p>
        )}

        <div className="mt-2.5 flex flex-wrap items-center gap-2">
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />
          {due && (
            <span
              className={cn(
                "inline-flex items-center gap-1 text-[11px] font-medium",
                overdue ? "text-danger" : "text-muted-foreground",
              )}
            >
              {overdue ? (
                <AlertTriangle className="h-3 w-3" />
              ) : (
                <Calendar className="h-3 w-3" />
              )}
              {format(due, "MMM d, yyyy")}
            </span>
          )}
          {showOwner && (
            <span className="font-mono text-[10px] text-muted">
              {task.userId.slice(0, 8)}
            </span>
          )}
        </div>
      </div>
    </motion.li>
  );
}
