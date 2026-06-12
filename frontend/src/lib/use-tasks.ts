"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { useEffect } from "react";
import { toast } from "sonner";
import { tasksApi, type TaskInput, API_URL } from "./api";
import type { Task, TaskList, TaskQuery } from "./types";

const tasksKey = (q: TaskQuery, admin: boolean) =>
  ["tasks", { admin, ...q }] as const;

/** useTasks fetches a filtered/sorted/paginated page of tasks. */
export function useTasks(query: TaskQuery, admin = false) {
  return useQuery({
    queryKey: tasksKey(query, admin),
    queryFn: () => tasksApi.list(query, admin),
    placeholderData: keepPreviousData, // keep the old page visible while refetching
  });
}

/**
 * useTaskMutations exposes create/update/delete with optimistic cache updates
 * and automatic rollback on failure. We optimistically mutate every cached
 * tasks page, then invalidate on settle so server ordering wins.
 */
export function useTaskMutations(activeKey: readonly unknown[]) {
  const qc = useQueryClient();

  const invalidate = () =>
    qc.invalidateQueries({ queryKey: ["tasks"] });

  const create = useMutation({
    mutationFn: (input: TaskInput) => tasksApi.create(input),
    onSuccess: () => toast.success("Task created"),
    onError: (e: Error) => toast.error(e.message || "Could not create task"),
    onSettled: invalidate,
  });

  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<TaskInput> }) =>
      tasksApi.update(id, input),
    // Optimistically patch the task in the currently visible list.
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: activeKey });
      const previous = qc.getQueryData<TaskList>(activeKey);
      if (previous) {
        qc.setQueryData<TaskList>(activeKey, {
          ...previous,
          tasks: previous.tasks.map((t) =>
            t.id === id ? { ...t, ...input } as Task : t,
          ),
        });
      }
      return { previous };
    },
    onError: (e: Error, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(activeKey, ctx.previous);
      toast.error(e.message || "Could not update task");
    },
    onSettled: invalidate,
  });

  const remove = useMutation({
    mutationFn: (id: string) => tasksApi.remove(id),
    // Optimistically drop the task from the visible list.
    onMutate: async (id: string) => {
      await qc.cancelQueries({ queryKey: activeKey });
      const previous = qc.getQueryData<TaskList>(activeKey);
      if (previous) {
        qc.setQueryData<TaskList>(activeKey, {
          ...previous,
          tasks: previous.tasks.filter((t) => t.id !== id),
          total: Math.max(0, previous.total - 1),
        });
      }
      return { previous };
    },
    onError: (e: Error, _id, ctx) => {
      if (ctx?.previous) qc.setQueryData(activeKey, ctx.previous);
      toast.error(e.message || "Could not delete task");
    },
    onSuccess: () => toast.success("Task deleted"),
    onSettled: invalidate,
  });

  return { create, update, remove };
}

/**
 * useTaskStream subscribes to the backend SSE stream and refreshes the task
 * list whenever a change is pushed (e.g. from another tab/device).
 */
export function useTaskStream(enabled: boolean) {
  const qc = useQueryClient();
  useEffect(() => {
    if (!enabled) return;
    const es = new EventSource(`${API_URL}/tasks/stream`, {
      withCredentials: true,
    });
    const refresh = () => qc.invalidateQueries({ queryKey: ["tasks"] });
    es.addEventListener("task.created", refresh);
    es.addEventListener("task.updated", refresh);
    es.addEventListener("task.deleted", refresh);
    return () => es.close();
  }, [enabled, qc]);
}
