// Matches models.User from Go backend
export type User = {
  id: number;
  name: string;
  email: string;
  role: "admin" | "member";
  created_at: string;
};

// Workflow stages — matches Go WorkflowStatus constants
export type WorkflowStatus = "todo" | "in_progress" | "review" | "done";

// Matches models.Todo from Go backend
export type Todo = {
  id: number;
  title: string;
  description: string;
  completed: boolean;
  priority: "low" | "medium" | "high";
  status: WorkflowStatus;
  assigned_to: number | null;
  assigned_by: number | null;
  assignee?: {
    id: number;
    name: string;
    email: string;
  } | null;
  created_at: string;
  updated_at: string;
};

// Auth API response shape
export type AuthResponse = {
  token: string;
  user: User;
};

export type APIResponse<T = unknown> = {
  success: boolean;
  message: string;
  data?: T;
  count?: number;
};

// Kanban column config
export const COLUMNS: { status: WorkflowStatus; label: string; color: string; dot: string }[] = [
  { status: "todo",        label: "To Do",       color: "border-slate-300",  dot: "bg-slate-400"  },
  { status: "in_progress", label: "In Progress", color: "border-amber-400",  dot: "bg-amber-400"  },
  { status: "review",      label: "Review",      color: "border-violet-400", dot: "bg-violet-400" },
  { status: "done",        label: "Done",        color: "border-emerald-400",dot: "bg-emerald-400"},
];