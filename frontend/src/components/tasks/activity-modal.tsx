"use client";

import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import {
  Plus,
  Pencil,
  AlignLeft,
  CircleDot,
  Flag,
  Calendar,
  Loader2,
  History,
} from "lucide-react";
import { Modal } from "@/components/ui/modal";
import { tasksApi } from "@/lib/api";
import type { Task } from "@/lib/types";

const iconFor: Record<string, React.ReactNode> = {
  created: <Plus className="h-3.5 w-3.5" />,
  title: <Pencil className="h-3.5 w-3.5" />,
  description: <AlignLeft className="h-3.5 w-3.5" />,
  status: <CircleDot className="h-3.5 w-3.5" />,
  priority: <Flag className="h-3.5 w-3.5" />,
  due_date: <Calendar className="h-3.5 w-3.5" />,
};

export function ActivityModal({
  task,
  onClose,
}: {
  task: Task | null;
  onClose: () => void;
}) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["activity", task?.id],
    queryFn: () => tasksApi.activity(task!.id),
    enabled: Boolean(task),
  });

  const activity = data?.activity ?? [];

  return (
    <Modal
      open={Boolean(task)}
      onClose={onClose}
      title="Activity"
      description={task?.title}
    >
      {isLoading && (
        <div className="flex justify-center py-8">
          <Loader2 className="h-5 w-5 animate-spin text-muted" />
        </div>
      )}

      {isError && (
        <p className="py-6 text-center text-sm text-danger">
          Couldn&apos;t load activity.
        </p>
      )}

      {!isLoading && !isError && activity.length === 0 && (
        <p className="py-6 text-center text-sm text-muted-foreground">
          No activity yet.
        </p>
      )}

      {activity.length > 0 && (
        <ol className="relative space-y-4 before:absolute before:left-[11px] before:top-1 before:h-[calc(100%-0.5rem)] before:w-px before:bg-border">
          {activity.map((a) => (
            <li key={a.id} className="relative flex gap-3">
              <span className="z-10 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border bg-surface-2 text-muted">
                {iconFor[a.action] ?? <History className="h-3.5 w-3.5" />}
              </span>
              <div className="min-w-0 flex-1 pt-0.5">
                <p className="text-sm leading-snug">{a.detail}</p>
                <p className="mt-0.5 text-[11px] text-muted">
                  {a.actorName} ·{" "}
                  {formatDistanceToNow(new Date(a.createdAt), {
                    addSuffix: true,
                  })}
                </p>
              </div>
            </li>
          ))}
        </ol>
      )}
    </Modal>
  );
}
