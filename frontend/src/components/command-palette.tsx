"use client";

import { useEffect } from "react";
import { Command } from "cmdk";
import { AnimatePresence, motion } from "framer-motion";
import {
  Plus,
  Sun,
  Moon,
  LogOut,
  ListFilter,
  CheckCircle2,
  Circle,
  Clock,
} from "lucide-react";
import { useTheme } from "next-themes";
import { useAuth } from "@/lib/auth-context";
import type { TaskStatus } from "@/lib/types";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNewTask: () => void;
  onFilterStatus: (status: TaskStatus | "") => void;
}

export function CommandPalette({
  open,
  onOpenChange,
  onNewTask,
  onFilterStatus,
}: Props) {
  const { setTheme, resolvedTheme } = useTheme();
  const { logout } = useAuth();

  // Global ⌘K / Ctrl+K toggles the palette.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        onOpenChange(!open);
      }
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [open, onOpenChange]);

  const run = (fn: () => void) => {
    fn();
    onOpenChange(false);
  };

  return (
    <AnimatePresence>
      {open && (
        <div className="fixed inset-0 z-[60] flex items-start justify-center p-4 pt-[15vh]">
          <motion.div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.12 }}
            onClick={() => onOpenChange(false)}
          />
          <motion.div
            initial={{ opacity: 0, scale: 0.98, y: -6 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.98, y: -6 }}
            transition={{ duration: 0.16, ease: [0.16, 1, 0.3, 1] }}
            className="relative z-10 w-full max-w-lg overflow-hidden rounded-xl border border-border bg-surface shadow-2xl shadow-black/50"
          >
            <Command label="Command palette" className="outline-none">
              <Command.Input
                autoFocus
                placeholder="Type a command or search…"
                className="w-full border-b border-border bg-transparent px-4 py-3.5 text-sm outline-none placeholder:text-muted"
              />
              <Command.List className="max-h-80 overflow-y-auto p-2">
                <Command.Empty className="py-6 text-center text-sm text-muted">
                  No results found.
                </Command.Empty>

                <Group heading="Actions">
                  <Item onSelect={() => run(onNewTask)} icon={<Plus />}>
                    Create new task
                  </Item>
                </Group>

                <Group heading="Filter by status">
                  <Item
                    onSelect={() => run(() => onFilterStatus(""))}
                    icon={<ListFilter />}
                  >
                    All tasks
                  </Item>
                  <Item
                    onSelect={() => run(() => onFilterStatus("todo"))}
                    icon={<Circle />}
                  >
                    To do
                  </Item>
                  <Item
                    onSelect={() => run(() => onFilterStatus("in_progress"))}
                    icon={<Clock />}
                  >
                    In progress
                  </Item>
                  <Item
                    onSelect={() => run(() => onFilterStatus("done"))}
                    icon={<CheckCircle2 />}
                  >
                    Done
                  </Item>
                </Group>

                <Group heading="Preferences">
                  <Item
                    onSelect={() =>
                      run(() =>
                        setTheme(resolvedTheme === "dark" ? "light" : "dark"),
                      )
                    }
                    icon={resolvedTheme === "dark" ? <Sun /> : <Moon />}
                  >
                    Toggle theme
                  </Item>
                  <Item onSelect={() => run(() => logout())} icon={<LogOut />}>
                    Sign out
                  </Item>
                </Group>
              </Command.List>
            </Command>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}

function Group({
  heading,
  children,
}: {
  heading: string;
  children: React.ReactNode;
}) {
  return (
    <Command.Group
      heading={heading}
      className="px-1 py-1 text-[11px] font-medium uppercase tracking-wide text-muted [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5"
    >
      {children}
    </Command.Group>
  );
}

function Item({
  onSelect,
  icon,
  children,
}: {
  onSelect: () => void;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <Command.Item
      onSelect={onSelect}
      className="flex cursor-pointer items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-foreground transition-colors data-[selected=true]:bg-[var(--accent-soft)] data-[selected=true]:text-accent [&_svg]:h-4 [&_svg]:w-4 [&_svg]:text-muted data-[selected=true]:[&_svg]:text-accent"
    >
      {icon}
      {children}
    </Command.Item>
  );
}
