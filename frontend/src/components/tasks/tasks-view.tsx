"use client";

import { useMemo, useState } from "react";
import { Users, User as UserIcon } from "lucide-react";
import { AppHeader } from "@/components/app-header";
import { CommandPalette } from "@/components/command-palette";
import { TaskToolbar, type ToolbarState } from "./task-toolbar";
import { TaskList } from "./task-list";
import { Pagination } from "./pagination";
import { TaskFormModal } from "./task-form-modal";
import { ConfirmDialog } from "./confirm-dialog";
import { ActivityModal } from "./activity-modal";
import { useAuth } from "@/lib/auth-context";
import { useDebounce } from "@/lib/use-debounce";
import { useTasks, useTaskMutations, useTaskStream } from "@/lib/use-tasks";
import type { Task, TaskQuery } from "@/lib/types";
import type { TaskInput } from "@/lib/api";

const PAGE_SIZE = 10;

export function TasksView() {
  const { user } = useAuth();

  const [toolbar, setToolbar] = useState<ToolbarState>({
    search: "",
    status: "",
    sortBy: "created_at",
    sortDir: "desc",
  });
  const [page, setPage] = useState(1);
  const [adminView, setAdminView] = useState(false);

  // Modal / dialog state.
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Task | null>(null);
  const [deleting, setDeleting] = useState<Task | null>(null);
  const [historyFor, setHistoryFor] = useState<Task | null>(null);

  const debouncedSearch = useDebounce(toolbar.search, 300);

  const query: TaskQuery = useMemo(
    () => ({
      status: toolbar.status,
      search: debouncedSearch,
      sortBy: toolbar.sortBy,
      sortDir: toolbar.sortDir,
      page,
      pageSize: PAGE_SIZE,
    }),
    [toolbar.status, debouncedSearch, toolbar.sortBy, toolbar.sortDir, page],
  );

  const isAdmin = user?.role === "admin";
  const useAdmin = isAdmin && adminView;

  const { data, isLoading, isError, refetch, queryKey } = useTasksWithKey(
    query,
    useAdmin,
  );
  const { create, update, remove } = useTaskMutations(queryKey);

  // Real-time: refresh when the server pushes changes.
  useTaskStream(Boolean(user));

  const patchToolbar = (patch: Partial<ToolbarState>) => {
    setToolbar((s) => ({ ...s, ...patch }));
    setPage(1); // any filter/search/sort change resets to the first page
  };

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };
  const openEdit = (task: Task) => {
    setEditing(task);
    setFormOpen(true);
  };

  const handleSubmit = async (input: TaskInput) => {
    if (editing) {
      await update.mutateAsync({ id: editing.id, input });
    } else {
      await create.mutateAsync(input);
    }
    setFormOpen(false);
  };

  const toggleComplete = (task: Task) => {
    update.mutate({
      id: task.id,
      input: { status: task.status === "done" ? "todo" : "done" },
    });
  };

  const hasFilters = Boolean(
    debouncedSearch || toolbar.status,
  );

  return (
    <div className="min-h-dvh">
      <AppHeader onOpenPalette={() => setPaletteOpen(true)} />

      <main className="mx-auto max-w-5xl px-4 py-6 sm:px-6 sm:py-8">
        <div className="mb-6 flex items-end justify-between gap-4">
          <div>
            <h1 className="text-lg font-semibold tracking-tight">
              {useAdmin ? "All tasks" : "Your tasks"}
            </h1>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {useAdmin
                ? "Every task across all users."
                : "Plan, track, and complete your work."}
            </p>
          </div>

          {isAdmin && (
            <button
              onClick={() => {
                setAdminView((v) => !v);
                setPage(1);
              }}
              className="inline-flex items-center gap-2 rounded-md border border-border bg-surface-2 px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:border-border-strong hover:text-foreground"
            >
              {useAdmin ? (
                <UserIcon className="h-3.5 w-3.5" />
              ) : (
                <Users className="h-3.5 w-3.5" />
              )}
              {useAdmin ? "View my tasks" : "View all tasks"}
            </button>
          )}
        </div>

        <div className="mb-5">
          <TaskToolbar
            state={toolbar}
            onChange={patchToolbar}
            onNewTask={openCreate}
          />
        </div>

        <TaskList
          tasks={data?.tasks ?? []}
          isLoading={isLoading}
          isError={isError}
          hasFilters={hasFilters}
          showOwner={useAdmin}
          onRetry={() => refetch()}
          onToggleComplete={toggleComplete}
          onEdit={openEdit}
          onDelete={(task) => setDeleting(task)}
          onHistory={(task) => setHistoryFor(task)}
          onNewTask={openCreate}
        />

        {data && data.total > 0 && (
          <div className="mt-5">
            <Pagination
              page={data.page}
              totalPages={data.totalPages}
              total={data.total}
              pageSize={data.pageSize}
              onPageChange={setPage}
            />
          </div>
        )}
      </main>

      {/* Overlays */}
      <CommandPalette
        open={paletteOpen}
        onOpenChange={setPaletteOpen}
        onNewTask={openCreate}
        onFilterStatus={(status) => patchToolbar({ status })}
      />
      <TaskFormModal
        open={formOpen}
        task={editing}
        submitting={create.isPending || update.isPending}
        onClose={() => setFormOpen(false)}
        onSubmit={handleSubmit}
      />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete this task?"
        body={`"${deleting?.title ?? ""}" will be permanently removed. This can't be undone.`}
        onConfirm={() => deleting && remove.mutate(deleting.id)}
        onClose={() => setDeleting(null)}
      />
      <ActivityModal task={historyFor} onClose={() => setHistoryFor(null)} />
    </div>
  );
}

/**
 * Small wrapper so the component has the exact queryKey the mutations need to
 * target for optimistic updates.
 */
function useTasksWithKey(query: TaskQuery, admin: boolean) {
  const queryKey = ["tasks", { admin, ...query }] as const;
  const result = useTasks(query, admin);
  return { ...result, queryKey };
}
