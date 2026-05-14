"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { todoAPI, userAPI } from "../lib/api";
import { getUser, clearAuth } from "../lib/auth";
import { Todo, User, WorkflowStatus, COLUMNS } from "../types";

function PriorityDot({ priority }: { priority: string }) {
  const colors: Record<string, string> = {
    high: "#dc2626",
    medium: "#d97706",
    low: "#16a34a",
  };
  return (
    <span style={{
      display: "inline-block", width: 7, height: 7,
      borderRadius: "50%", background: colors[priority] ?? "#666", flexShrink: 0,
    }} title={priority} />
  );
}

function TodoCard({ todo, isAdmin, members, onStatusChange, onAssign, onDelete }: {
  todo: Todo; isAdmin: boolean; members: User[];
  onStatusChange: (id: number, status: WorkflowStatus) => void;
  onAssign: (id: number, userId: number) => void;
  onDelete: (id: number) => void;
}) {
  const statusOrder: WorkflowStatus[] = ["todo", "in_progress", "review", "done"];
  const currentIdx = statusOrder.indexOf(todo.status);

  return (
    <div className="todo-card">
      <div className="card-top">
        <PriorityDot priority={todo.priority} />
        <span className="card-title">{todo.title}</span>
        {isAdmin && (
          <button className="card-delete" onClick={() => onDelete(todo.id)} title="Delete">×</button>
        )}
      </div>
      {todo.description && <p className="card-desc">{todo.description}</p>}
      {todo.assignee && (
        <div className="card-assignee">
          <span className="assignee-avatar">{todo.assignee.name.charAt(0).toUpperCase()}</span>
          <span className="assignee-name">{todo.assignee.name}</span>
        </div>
      )}
      {isAdmin && members.length > 0 && (
        <select className="card-select" value={todo.assigned_to ?? ""}
          onChange={e => onAssign(todo.id, Number(e.target.value))}>
          <option value="" disabled>{todo.assigned_to ? "Reassign…" : "Assign to…"}</option>
          {members.map(m => <option key={m.id} value={m.id}>{m.name}</option>)}
        </select>
      )}
      <div className="card-nav">
        <button className="nav-btn" disabled={currentIdx === 0}
          onClick={() => onStatusChange(todo.id, statusOrder[currentIdx - 1])} title="Move back">←</button>
        <span className="card-status-label">{todo.status.replace("_", " ")}</span>
        <button className="nav-btn" disabled={currentIdx === statusOrder.length - 1}
          onClick={() => onStatusChange(todo.id, statusOrder[currentIdx + 1])} title="Move forward">→</button>
      </div>
    </div>
  );
}

function CreateModal({ onClose, onCreate }: {
  onClose: () => void;
  onCreate: (title: string, desc: string, priority: string) => Promise<void>;
}) {
  const [title, setTitle] = useState("");
  const [desc, setDesc] = useState("");
  const [priority, setPriority] = useState("medium");
  const [loading, setLoading] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setLoading(true);
    await onCreate(title.trim(), desc.trim(), priority);
    setLoading(false);
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <span className="modal-title">New task</span>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>
        <form onSubmit={submit} className="modal-form">
          <input autoFocus className="modal-input" placeholder="Task title"
            value={title} onChange={e => setTitle(e.target.value)} required />
          <textarea className="modal-input modal-textarea" placeholder="Description (optional)"
            value={desc} onChange={e => setDesc(e.target.value)} rows={3} />
          <div className="modal-row">
            <select className="modal-input modal-select" value={priority}
              onChange={e => setPriority(e.target.value)}>
              <option value="low">Low priority</option>
              <option value="medium">Medium priority</option>
              <option value="high">High priority</option>
            </select>
            <button type="submit" className="modal-btn" disabled={loading}>
              {loading ? "Adding…" : "Add task"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const router = useRouter();

  // ── FIX: read user only on client to avoid hydration mismatch ──
  const [user, setUser] = useState<User | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const u = getUser();
    setUser(u);
    setMounted(true);
    if (!u) router.replace("/login");
  }, [router]);

  const isAdmin = user?.role === "admin";

  const [todos,      setTodos]      = useState<Todo[]>([]);
  const [members,    setMembers]    = useState<User[]>([]);
  const [loading,    setLoading]    = useState(true);
  const [error,      setError]      = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const fetchTodos = useCallback(async () => {
    setError("");
    try {
      const res = await todoAPI.getAll();
      if (res.success) setTodos((res.data as Todo[]) ?? []);
      else setError(res.message);
    } catch {
      setError("Cannot reach API");
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchMembers = useCallback(async () => {
    if (!isAdmin) return;
    try {
      const res = await userAPI.getMembers();
      if (res.success) setMembers((res.data as User[]) ?? []);
    } catch { /* non-critical */ }
  }, [isAdmin]);

  useEffect(() => {
    if (mounted && user) {
      fetchTodos();
      fetchMembers();
    }
  }, [mounted, user, fetchTodos, fetchMembers]);

  const handleCreate = async (title: string, desc: string, priority: string) => {
    const res = await todoAPI.create(title, desc, priority);
    if (res.success) fetchTodos();
    else setError(res.message);
  };

  const handleStatusChange = async (id: number, status: WorkflowStatus) => {
    const res = await todoAPI.updateStatus(id, status);
    if (res.success) setTodos(prev => prev.map(t => t.id === id ? { ...t, ...res.data as Todo } : t));
  };

  const handleAssign = async (todoId: number, userId: number) => {
    const res = await todoAPI.assign(todoId, userId);
    if (res.success) fetchTodos();
    else setError(res.message);
  };

  const handleDelete = async (id: number) => {
    const res = await todoAPI.delete(id);
    if (res.success) setTodos(prev => prev.filter(t => t.id !== id));
    else setError(res.message);
  };

  const handleLogout = () => { clearAuth(); router.replace("/login"); };

  const todosInColumn = (status: WorkflowStatus) => todos.filter(t => t.status === status);

  // Don't render until mounted (avoids hydration mismatch from localStorage)
  if (!mounted) return null;
  if (!user) return null;

  return (
    <div className="dash-root">
      <header className="topbar">
        <div className="topbar-left">
          <span className="topbar-brand">TASKFLOW</span>
          <span className="topbar-divider" />
          <span className="topbar-role">{user.role}</span>
        </div>
        <div className="topbar-right">
          <span className="topbar-user">{user.name}</span>
          {isAdmin && (
            <button className="topbar-create" onClick={() => setShowCreate(true)}>+ New task</button>
          )}
          <button className="topbar-logout" onClick={handleLogout}>Sign out</button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          {error}
          <button onClick={() => setError("")}>×</button>
        </div>
      )}

      <div className="stats-row">
        <div className="stat"><span className="stat-num">{todos.length}</span><span className="stat-lbl">Total</span></div>
        {COLUMNS.map(col => (
          <div className="stat" key={col.status}>
            <span className="stat-num">{todosInColumn(col.status).length}</span>
            <span className="stat-lbl">{col.label}</span>
          </div>
        ))}
      </div>

      {loading ? (
        <div className="loading">Loading tasks…</div>
      ) : (
        <div className="board">
          {COLUMNS.map(col => (
            <div className="column" key={col.status}>
              <div className="col-header">
                <span className={`col-dot col-dot--${col.status}`} />
                <span className="col-label">{col.label}</span>
                <span className="col-count">{todosInColumn(col.status).length}</span>
              </div>
              <div className="col-body">
                {todosInColumn(col.status).length === 0
                  ? <div className="col-empty">No tasks</div>
                  : todosInColumn(col.status).map(todo => (
                    <TodoCard key={todo.id} todo={todo} isAdmin={isAdmin} members={members}
                      onStatusChange={handleStatusChange} onAssign={handleAssign} onDelete={handleDelete} />
                  ))
                }
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && <CreateModal onClose={() => setShowCreate(false)} onCreate={handleCreate} />}

      <style jsx>{`
        .dash-root { min-height: 100vh; background: #f8f8f8; display: flex; flex-direction: column; font-family: 'Georgia', serif; }

        /* Topbar */
        .topbar { display: flex; align-items: center; justify-content: space-between; padding: 0 1.5rem; height: 54px; background: #111; border-bottom: 3px solid #c0392b; flex-shrink: 0; }
        .topbar-left, .topbar-right { display: flex; align-items: center; gap: 0.75rem; }
        .topbar-brand { font-family: 'Courier New', monospace; font-size: 0.75rem; font-weight: 700; letter-spacing: 0.2em; color: #fff; }
        .topbar-divider { width: 1px; height: 16px; background: #333; }
        .topbar-role { font-size: 0.65rem; letter-spacing: 0.12em; text-transform: uppercase; color: #c0392b; background: #1a1a1a; border: 1px solid #333; padding: 2px 8px; border-radius: 2px; font-family: 'Courier New', monospace; }
        .topbar-user { font-size: 0.78rem; color: #999; font-family: 'Courier New', monospace; }
        .topbar-create { background: #c0392b; color: #fff; border: none; border-radius: 2px; padding: 0.35rem 0.9rem; font-size: 0.72rem; font-weight: 700; cursor: pointer; font-family: 'Courier New', monospace; transition: background 0.15s; }
        .topbar-create:hover { background: #a93226; }
        .topbar-logout { background: transparent; border: 1px solid #333; color: #888; border-radius: 2px; padding: 0.35rem 0.9rem; font-size: 0.7rem; cursor: pointer; font-family: 'Courier New', monospace; transition: all 0.15s; }
        .topbar-logout:hover { border-color: #888; color: #fff; }

        /* Error */
        .error-banner { background: #fef2f2; border-bottom: 1px solid #fca5a5; color: #dc2626; font-size: 0.8rem; padding: 0.6rem 1.5rem; display: flex; justify-content: space-between; align-items: center; font-family: 'Courier New', monospace; }
        .error-banner button { background: none; border: none; color: #dc2626; font-size: 1rem; cursor: pointer; }

        /* Stats */
        .stats-row { display: flex; gap: 1px; background: #ddd; border-bottom: 1px solid #ddd; flex-shrink: 0; }
        .stat { flex: 1; display: flex; flex-direction: column; align-items: center; padding: 0.65rem; background: #fff; gap: 2px; }
        .stat-num { font-size: 1.1rem; font-weight: 700; color: #111; font-family: 'Courier New', monospace; }
        .stat-lbl { font-size: 0.6rem; letter-spacing: 0.1em; text-transform: uppercase; color: #999; font-family: 'Courier New', monospace; }

        /* Board */
        .board { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1px; background: #e0e0e0; flex: 1; }
        .column { background: #f4f4f4; display: flex; flex-direction: column; min-height: calc(100vh - 140px); }
        .col-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.8rem 1rem; border-bottom: 1px solid #e0e0e0; background: #fff; position: sticky; top: 0; z-index: 1; }
        .col-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
        .col-dot--todo { background: #9ca3af; }
        .col-dot--in_progress { background: #d97706; }
        .col-dot--review { background: #7c3aed; }
        .col-dot--done { background: #16a34a; }
        .col-label { font-size: 0.68rem; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; color: #444; flex: 1; font-family: 'Courier New', monospace; }
        .col-count { font-size: 0.65rem; color: #999; background: #f0f0f0; border: 1px solid #e0e0e0; padding: 1px 6px; border-radius: 10px; font-family: 'Courier New', monospace; }
        .col-body { padding: 0.75rem; display: flex; flex-direction: column; gap: 0.6rem; flex: 1; overflow-y: auto; }
        .col-empty { font-size: 0.7rem; color: #ccc; text-align: center; padding: 2rem 0; font-family: 'Courier New', monospace; letter-spacing: 0.08em; }

        /* Cards */
        :global(.todo-card) { background: #fff; border: 1px solid #e5e5e5; border-radius: 3px; padding: 0.7rem 0.75rem; display: flex; flex-direction: column; gap: 0.45rem; transition: border-color 0.15s, box-shadow 0.15s; }
        :global(.todo-card:hover) { border-color: #bbb; box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
        :global(.card-top) { display: flex; align-items: flex-start; gap: 0.4rem; }
        :global(.card-title) { flex: 1; font-size: 0.82rem; color: #111; line-height: 1.4; word-break: break-word; font-family: 'Georgia', serif; }
        :global(.card-delete) { background: none; border: none; color: #ddd; font-size: 1rem; cursor: pointer; padding: 0; line-height: 1; flex-shrink: 0; transition: color 0.15s; }
        :global(.card-delete:hover) { color: #c0392b; }
        :global(.card-desc) { font-size: 0.72rem; color: #999; line-height: 1.5; margin: 0; font-family: 'Courier New', monospace; }
        :global(.card-assignee) { display: flex; align-items: center; gap: 0.35rem; }
        :global(.assignee-avatar) { width: 18px; height: 18px; border-radius: 50%; background: #111; color: #fff; font-size: 0.6rem; font-weight: 700; display: flex; align-items: center; justify-content: center; font-family: 'Courier New', monospace; }
        :global(.assignee-name) { font-size: 0.68rem; color: #888; font-family: 'Courier New', monospace; }
        :global(.card-select) { width: 100%; background: #f8f8f8; border: 1px solid #e0e0e0; border-radius: 2px; padding: 0.3rem 0.5rem; font-size: 0.68rem; color: #666; cursor: pointer; font-family: 'Courier New', monospace; outline: none; }
        :global(.card-select:focus) { border-color: #c0392b; }
        :global(.card-nav) { display: flex; align-items: center; justify-content: space-between; margin-top: 0.1rem; }
        :global(.nav-btn) { background: none; border: 1px solid #e0e0e0; border-radius: 2px; color: #bbb; font-size: 0.75rem; width: 24px; height: 20px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; font-family: 'Courier New', monospace; }
        :global(.nav-btn:hover:not(:disabled)) { color: #c0392b; border-color: #c0392b; }
        :global(.nav-btn:disabled) { opacity: 0.2; cursor: not-allowed; }
        :global(.card-status-label) { font-size: 0.6rem; color: #bbb; letter-spacing: 0.08em; text-transform: uppercase; font-family: 'Courier New', monospace; }

        /* Loading */
        .loading { flex: 1; display: flex; align-items: center; justify-content: center; font-size: 0.8rem; color: #bbb; letter-spacing: 0.1em; font-family: 'Courier New', monospace; }

        /* Modal */
        :global(.modal-overlay) { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 50; padding: 1rem; }
        :global(.modal-box) { background: #fff; border: 1px solid #e0e0e0; border-top: 3px solid #c0392b; border-radius: 3px; width: 100%; max-width: 420px; padding: 1.5rem; }
        :global(.modal-header) { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.25rem; }
        :global(.modal-title) { font-size: 0.75rem; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; color: #333; font-family: 'Courier New', monospace; }
        :global(.modal-close) { background: none; border: none; color: #bbb; font-size: 1.2rem; cursor: pointer; transition: color 0.15s; font-family: 'Courier New', monospace; }
        :global(.modal-close:hover) { color: #111; }
        :global(.modal-form) { display: flex; flex-direction: column; gap: 0.75rem; }
        :global(.modal-input) { width: 100%; background: #f8f8f8; border: 1px solid #e0e0e0; border-radius: 2px; padding: 0.6rem 0.75rem; font-size: 0.82rem; color: #111; outline: none; font-family: 'Georgia', serif; transition: border-color 0.15s; box-sizing: border-box; }
        :global(.modal-input::placeholder) { color: #ccc; font-family: 'Courier New', monospace; }
        :global(.modal-input:focus) { border-color: #c0392b; }
        :global(.modal-textarea) { resize: vertical; font-family: 'Courier New', monospace; font-size: 0.78rem; }
        :global(.modal-row) { display: flex; gap: 0.6rem; }
        :global(.modal-select) { flex: 1; font-family: 'Courier New', monospace; font-size: 0.75rem; }
        :global(.modal-btn) { background: #111; color: #fff; border: none; border-radius: 2px; padding: 0 1rem; font-size: 0.75rem; font-weight: 700; cursor: pointer; font-family: 'Courier New', monospace; white-space: nowrap; transition: background 0.15s; }
        :global(.modal-btn:hover:not(:disabled)) { background: #c0392b; }
        :global(.modal-btn:disabled) { opacity: 0.5; cursor: not-allowed; }

        /* Responsive */
        @media (max-width: 900px) { .board { grid-template-columns: repeat(2, 1fr); } }
        @media (max-width: 560px) { .board { grid-template-columns: 1fr; } }
      `}</style>
    </div>
  );
}