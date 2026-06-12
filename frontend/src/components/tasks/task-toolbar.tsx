"use client";

import { Search, Plus, ArrowUpDown, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import type { SortBy, SortDir, TaskStatus } from "@/lib/types";

export interface ToolbarState {
  search: string;
  status: TaskStatus | "";
  sortBy: SortBy;
  sortDir: SortDir;
}

interface Props {
  state: ToolbarState;
  onChange: (patch: Partial<ToolbarState>) => void;
  onNewTask: () => void;
}

export function TaskToolbar({ state, onChange, onNewTask }: Props) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
      {/* Search */}
      <div className="relative flex-1">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
        <Input
          value={state.search}
          onChange={(e) => onChange({ search: e.target.value })}
          placeholder="Search tasks by title…"
          className="pl-9 pr-8"
          aria-label="Search tasks"
        />
        {state.search && (
          <button
            onClick={() => onChange({ search: "" })}
            className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted hover:text-foreground"
            aria-label="Clear search"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        )}
      </div>

      <div className="flex items-center gap-2">
        {/* Status filter */}
        <Select
          value={state.status}
          onChange={(e) =>
            onChange({ status: e.target.value as TaskStatus | "" })
          }
          aria-label="Filter by status"
          className="w-[130px]"
        >
          <option value="">All status</option>
          <option value="todo">To do</option>
          <option value="in_progress">In progress</option>
          <option value="done">Done</option>
        </Select>

        {/* Sort field */}
        <Select
          value={state.sortBy}
          onChange={(e) => onChange({ sortBy: e.target.value as SortBy })}
          aria-label="Sort by"
          className="w-[140px]"
        >
          <option value="created_at">Created date</option>
          <option value="due_date">Due date</option>
          <option value="priority">Priority</option>
        </Select>

        {/* Sort direction */}
        <Button
          variant="secondary"
          size="icon"
          aria-label={`Sort ${state.sortDir === "asc" ? "ascending" : "descending"}`}
          title={state.sortDir === "asc" ? "Ascending" : "Descending"}
          onClick={() =>
            onChange({ sortDir: state.sortDir === "asc" ? "desc" : "asc" })
          }
        >
          <ArrowUpDown
            className={`h-4 w-4 transition-transform ${
              state.sortDir === "asc" ? "" : "rotate-180"
            }`}
          />
        </Button>

        <Button onClick={onNewTask} className="shrink-0">
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">New task</span>
        </Button>
      </div>
    </div>
  );
}
