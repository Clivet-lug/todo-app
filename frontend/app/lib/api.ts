import { getToken } from "./auth";
import { APIResponse, Todo, User, WorkflowStatus } from "../types";

const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9090";

// Attaches Bearer token to every request automatically
function headers(extra?: Record<string, string>): Record<string, string> {
  const token = getToken();
  return {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...extra,
  };
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: headers(options.headers as Record<string, string>),
  });
  return res.json();
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export const authAPI = {
  register: (name: string, email: string, password: string, role = "member") =>
    request("/auth/register", {
      method: "POST",
      body: JSON.stringify({ name, email, password, role }),
    }),

  login: (email: string, password: string) =>
    request("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
};

// ── Todos ─────────────────────────────────────────────────────────────────────

export const todoAPI = {
  getAll: () =>
    request<Todo[]>("/todos"),

  create: (title: string, description: string, priority: string) =>
    request<Todo>("/todos", {
      method: "POST",
      body: JSON.stringify({ title, description, priority }),
    }),

  update: (id: number, data: Partial<Todo>) =>
    request<Todo>(`/todos/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    request(`/todos/${id}`, { method: "DELETE" }),

  assign: (todoId: number, assignedTo: number) =>
    request<Todo>(`/todos/${todoId}/assign`, {
      method: "PUT",
      body: JSON.stringify({ assigned_to: assignedTo }),
    }),

  updateStatus: (todoId: number, status: WorkflowStatus) =>
    request<Todo>(`/todos/${todoId}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),
};

// ── Users ─────────────────────────────────────────────────────────────────────

export const userAPI = {
  getMembers: () =>
    request<User[]>("/users/members"),
};