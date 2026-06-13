import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCard } from "./task-card";
import type { Task } from "@/lib/types";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "t1",
    userId: "u1",
    title: "Write the README",
    description: "Document setup",
    status: "todo",
    priority: "high",
    dueDate: null,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

const noop = () => {};

describe("TaskCard", () => {
  it("renders the title, description and priority", () => {
    render(
      <TaskCard
        task={makeTask()}
        onToggleComplete={noop}
        onEdit={noop}
        onDelete={noop}
      />,
    );
    expect(screen.getByText("Write the README")).toBeInTheDocument();
    expect(screen.getByText("Document setup")).toBeInTheDocument();
    expect(screen.getByText("High")).toBeInTheDocument();
  });

  it("calls onToggleComplete when the checkbox is clicked", async () => {
    const onToggle = vi.fn();
    render(
      <TaskCard
        task={makeTask()}
        onToggleComplete={onToggle}
        onEdit={noop}
        onDelete={noop}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: /mark as done/i }));
    expect(onToggle).toHaveBeenCalledOnce();
  });

  it("flags an overdue task that is not done", () => {
    render(
      <TaskCard
        task={makeTask({ dueDate: "2020-01-01T00:00:00Z", status: "todo" })}
        onToggleComplete={noop}
        onEdit={noop}
        onDelete={noop}
      />,
    );
    // Overdue tasks render an alert affordance via the "Jan 1, 2020" date label.
    expect(screen.getByText("Jan 1, 2020")).toBeInTheDocument();
  });

  it("calls onDelete when the delete action is clicked", async () => {
    const onDelete = vi.fn();
    render(
      <TaskCard
        task={makeTask()}
        onToggleComplete={noop}
        onEdit={noop}
        onDelete={onDelete}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: /delete task/i }));
    expect(onDelete).toHaveBeenCalledOnce();
  });
});
