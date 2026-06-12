"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import { useAuth } from "@/lib/auth-context";
import { TasksView } from "@/components/tasks/tasks-view";

export default function TasksPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  // Client-side route guard: bounce unauthenticated users to login.
  useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [user, loading, router]);

  if (loading || !user) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted" />
      </div>
    );
  }

  return <TasksView />;
}
