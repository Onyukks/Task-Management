import type {
  Task,
  TaskList,
  TaskQuery,
  User,
  TaskStatus,
  TaskPriority,
} from "./types";

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

/** Shape of the backend's consistent error envelope. */
export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    fields?: Record<string, string>;
  };
}

export class ApiError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(status: number, body: ApiErrorBody) {
    super(body.error?.message ?? "Request failed");
    this.status = status;
    this.code = body.error?.code ?? "unknown";
    this.fields = body.error?.fields;
  }
}

/**
 * request is the single fetch wrapper used everywhere. It always sends the
 * auth cookie (credentials: "include") and normalises errors into ApiError.
 */
async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (!res.ok) {
    throw new ApiError(res.status, data as ApiErrorBody);
  }
  return data as T;
}

// ---- Auth ----

export const authApi = {
  signup: (input: { email: string; name: string; password: string }) =>
    request<{ user: User }>("/auth/signup", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  login: (input: { email: string; password: string }) =>
    request<{ user: User }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  logout: () => request<{ status: string }>("/auth/logout", { method: "POST" }),

  me: () => request<{ user: User }>("/auth/me"),
};

// ---- Tasks ----

export interface TaskInput {
  title: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  dueDate?: string | null;
}

function toQueryString(q: TaskQuery): string {
  const params = new URLSearchParams();
  if (q.status) params.set("status", q.status);
  if (q.search) params.set("search", q.search);
  if (q.sortBy) params.set("sortBy", q.sortBy);
  if (q.sortDir) params.set("sortDir", q.sortDir);
  if (q.page) params.set("page", String(q.page));
  if (q.pageSize) params.set("pageSize", String(q.pageSize));
  const s = params.toString();
  return s ? `?${s}` : "";
}

export const tasksApi = {
  list: (q: TaskQuery = {}, admin = false) =>
    request<TaskList>(`${admin ? "/admin/tasks" : "/tasks"}${toQueryString(q)}`),

  get: (id: string) => request<Task>(`/tasks/${id}`),

  create: (input: TaskInput) =>
    request<Task>("/tasks", { method: "POST", body: JSON.stringify(input) }),

  update: (id: string, input: Partial<TaskInput>) =>
    request<Task>(`/tasks/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    }),

  remove: (id: string) =>
    request<void>(`/tasks/${id}`, { method: "DELETE" }),
};
