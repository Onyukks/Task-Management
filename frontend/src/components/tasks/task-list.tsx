"use client";

import { AnimatePresence, motion } from "framer-motion";
import { Inbox, AlertCircle, SearchX } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TaskCard } from "./task-card";
import type { Task } from "@/lib/types";

interface Props {
  tasks: Task[];
  isLoading: boolean;
  isError: boolean;
  hasFilters: boolean;
  showOwner?: boolean;
  onRetry: () => void;
  onToggleComplete: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onHistory: (task: Task) => void;
  onNewTask: () => void;
}

export function TaskList(props: Props) {
  const { tasks, isLoading, isError, hasFilters } = props;

  // ---- Loading: skeleton rows ----
  if (isLoading) {
    return (
      <ul className="space-y-2.5">
        {Array.from({ length: 4 }).map((_, i) => (
          <li
            key={i}
            className="h-[92px] animate-pulse rounded-lg border border-border bg-surface"
            style={{ animationDelay: `${i * 80}ms` }}
          />
        ))}
      </ul>
    );
  }

  // ---- Error ----
  if (isError) {
    return (
      <State
        icon={<AlertCircle className="h-6 w-6 text-danger" />}
        title="Couldn't load your tasks"
        body="Something went wrong reaching the server. Check your connection and try again."
      >
        <Button variant="secondary" onClick={props.onRetry}>
          Try again
        </Button>
      </State>
    );
  }

  // ---- Empty ----
  if (tasks.length === 0) {
    return hasFilters ? (
      <State
        icon={<SearchX className="h-6 w-6 text-muted" />}
        title="No matching tasks"
        body="No tasks match your current search and filters. Try adjusting them."
      />
    ) : (
      <State
        icon={<Inbox className="h-6 w-6 text-muted" />}
        title="No tasks yet"
        body="Create your first task to get started. Tip: press ⌘K anytime."
      >
        <Button onClick={props.onNewTask}>Create a task</Button>
      </State>
    );
  }

  // ---- Data ----
  return (
    <motion.ul layout className="space-y-2.5">
      <AnimatePresence mode="popLayout" initial={false}>
        {tasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            showOwner={props.showOwner}
            onToggleComplete={props.onToggleComplete}
            onEdit={props.onEdit}
            onDelete={props.onDelete}
            onHistory={props.onHistory}
          />
        ))}
      </AnimatePresence>
    </motion.ul>
  );
}

function State({
  icon,
  title,
  body,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  body: string;
  children?: React.ReactNode;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface/40 px-6 py-16 text-center"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-surface-2">
        {icon}
      </div>
      <h3 className="text-sm font-semibold">{title}</h3>
      <p className="mt-1 max-w-xs text-xs text-muted-foreground">{body}</p>
      {children && <div className="mt-5">{children}</div>}
    </motion.div>
  );
}
