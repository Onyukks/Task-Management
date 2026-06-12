"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";
import { Input, Textarea, Label } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import type { Task } from "@/lib/types";
import type { TaskInput } from "@/lib/api";

const schema = z.object({
  title: z.string().min(1, "Title is required").max(200, "Too long (max 200)"),
  description: z.string().max(2000, "Too long (max 2000)").optional(),
  status: z.enum(["todo", "in_progress", "done"]),
  priority: z.enum(["low", "medium", "high"]),
  dueDate: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  task: Task | null; // null => create mode
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: TaskInput) => Promise<void> | void;
}

/** Convert an ISO timestamp to the yyyy-MM-dd a date input expects. */
function toDateInput(iso: string | null | undefined): string {
  if (!iso) return "";
  return new Date(iso).toISOString().slice(0, 10);
}

export function TaskFormModal({ open, task, submitting, onClose, onSubmit }: Props) {
  const isEdit = Boolean(task);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: "",
      description: "",
      status: "todo",
      priority: "medium",
      dueDate: "",
    },
  });

  // Reset the form whenever we open it for a different task (or for create).
  useEffect(() => {
    if (!open) return;
    reset({
      title: task?.title ?? "",
      description: task?.description ?? "",
      status: task?.status ?? "todo",
      priority: task?.priority ?? "medium",
      dueDate: toDateInput(task?.dueDate),
    });
  }, [open, task, reset]);

  const submit = handleSubmit(async (values) => {
    const input: TaskInput = {
      title: values.title.trim(),
      description: values.description ?? "",
      status: values.status,
      priority: values.priority,
      // Empty date => null (clears it on edit); otherwise send an ISO timestamp.
      dueDate: values.dueDate ? new Date(values.dueDate).toISOString() : null,
    };
    await onSubmit(input);
  });

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit task" : "New task"}
      description={
        isEdit ? "Update the details below." : "Add a task to your list."
      }
    >
      <form onSubmit={submit} className="space-y-4">
        <div className="space-y-1.5">
          <Label>Title</Label>
          <Input
            {...register("title")}
            placeholder="e.g. Ship the landing page"
            autoFocus
          />
          {errors.title && (
            <p className="text-xs text-danger">{errors.title.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label>Description</Label>
          <Textarea
            {...register("description")}
            placeholder="Add any details… (optional)"
          />
          {errors.description && (
            <p className="text-xs text-danger">{errors.description.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label>Status</Label>
            <Select {...register("status")}>
              <option value="todo">To do</option>
              <option value="in_progress">In progress</option>
              <option value="done">Done</option>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>Priority</Label>
            <Select {...register("priority")}>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
            </Select>
          </div>
        </div>

        <div className="space-y-1.5">
          <Label>Due date</Label>
          <Input type="date" {...register("dueDate")} className="w-full" />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={submitting}>
            {isEdit ? "Save changes" : "Create task"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
