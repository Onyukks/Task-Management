"use client";

import { CheckCircle2, LogOut, Command } from "lucide-react";
import { useAuth } from "@/lib/auth-context";
import { Button } from "./ui/button";
import { ThemeToggle } from "./theme-toggle";

export function AppHeader({ onOpenPalette }: { onOpenPalette: () => void }) {
  const { user, logout } = useAuth();

  return (
    <header className="sticky top-0 z-30 border-b border-border bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
        <div className="flex items-center gap-2.5">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent text-accent-foreground">
            <CheckCircle2 className="h-4 w-4" />
          </div>
          <span className="text-sm font-semibold tracking-tight">Tasks</span>
          {user?.role === "admin" && (
            <span className="rounded-full border border-accent/30 bg-[var(--accent-soft)] px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-accent">
              Admin
            </span>
          )}
        </div>

        <div className="flex items-center gap-1.5">
          <button
            onClick={onOpenPalette}
            className="hidden items-center gap-2 rounded-md border border-border bg-surface-2 px-2.5 py-1.5 text-xs text-muted-foreground transition-colors hover:border-border-strong hover:text-foreground sm:flex"
          >
            <Command className="h-3 w-3" />
            <span>Quick actions</span>
            <kbd className="rounded bg-background px-1.5 py-0.5 font-mono text-[10px]">
              ⌘K
            </kbd>
          </button>

          <ThemeToggle />

          <div className="ml-1 hidden text-right sm:block">
            <p className="text-xs font-medium leading-tight">{user?.name}</p>
            <p className="text-[11px] leading-tight text-muted">{user?.email}</p>
          </div>

          <Button
            variant="ghost"
            size="icon"
            aria-label="Sign out"
            onClick={() => logout()}
          >
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </header>
  );
}
