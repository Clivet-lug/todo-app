"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { todoAPI, userAPI } from "../lib/api";
import { getUser, clearAuth } from "../lib/auth";
import { Todo, User, WorkflowStatus, COLUMNS } from "../types";

function PriorityBadge({ priority }: { priority: string }) {
  const styles: Record<string, { bg: string; color: string; label: string }> = {
    high:   { bg: "#fff5f5", color: "#c92a2a", label: "High"   },
    medium: { bg: "#fff9db", color: "#e67700", label: "Medium" },
    low:    { bg: "#ebfbee", color: "#2f9e44", label: "Low"    },
  };
  const s = styles[priority] ?? { bg: "#f1f3f5", color: "#868e96", label: priority };
  return (
    <span style={{
      background: s.bg, color: s.color,
      fontSize: "0.65rem", fontWeight: 600,
      padding: "2px 6px", borderRadius: 4,
      letterSpacing: "0.02em", flexShrink: 0,
    }}>{s.label}</span>
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
    <div className="card">
      <div className="card-header">
        <PriorityBadge priority={todo.priority} />
        {isAdmin && (
          <button className="card-del" onClick={() => onDelete(todo.id)} title="Delete">×</button>
        )}
      </div>

      <p className="card-title">{todo.title}</p>
      {todo.description && <p className="card-desc">{todo.description}</p>}

      {todo.assignee && (
        <div className="card-assignee">
          <div className="avatar">{todo.assignee.name.charAt(0).toUpperCase()}</div>
          <span className="assignee-name">{todo.assignee.name}</span>
        </div>
      )}

      {isAdmin && members.length > 0 && (
        <select className="card-select" value={todo.assigned_to ?? ""}
          onChange={e => onAssign(todo.id, Number(e.target.value))}>
          <option value="" disabled>{todo.assigned_to ? "Reassign…" : "Assign to member…"}</option>
          {members.map(m => <option key={m.id} value={m.id}>{m.name}</option>)}
        </select>
      )}

      <div className="card-nav">
        <button className="nav-btn" disabled={currentIdx === 0}
          onClick={() => onStatusChange(todo.id, statusOrder[currentIdx - 1])}>←</button>
        <span className="nav-label">{todo.status.replace("_", " ")}</span>
        <button className="nav-btn" disabled={currentIdx === statusOrder.length - 1}
          onClick={() => onStatusChange(todo.id, statusOrder[currentIdx + 1])}>→</button>
      </div>

      <style jsx>{`
        .card {
          background: #fff;
          border: 1px solid #e9ecef;
          border-radius: 8px;
          padding: 0.85rem;
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
          transition: box-shadow 0.15s, border-color 0.15s;
        }
        .card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.07); border-color: #dee2e6; }

        .card-header { display: flex; align-items: center; justify-content: space-between; }

        .card-del {
          background: none; border: none; color: #ced4da;
          font-size: 1rem; cursor: pointer; line-height: 1;
          transition: color 0.15s; padding: 0;
        }
        .card-del:hover { color: #c92a2a; }

        .card-title { font-size: 0.875rem; font-weight: 500; color: #212529; line-height: 1.45; }
        .card-desc  { font-size: 0.775rem; color: #868e96; line-height: 1.5; }

        .card-assignee { display: flex; align-items: center; gap: 0.4rem; }
        .avatar {
          width: 20px; height: 20px; border-radius: 50%;
          background: #212529; color: #fff;
          font-size: 0.6rem; font-weight: 700;
          display: flex; align-items: center; justify-content: center;
          flex-shrink: 0;
        }
        .assignee-name { font-size: 0.75rem; color: #495057; font-weight: 500; }

        .card-select {
          width: 100%;
          background: #f8f9fa;
          border: 1px solid #e9ecef;
          border-radius: 6px;
          padding: 0.35rem 0.5rem;
          font-size: 0.775rem;
          color: #495057;
          cursor: pointer;
          font-family: var(--font-sans);
          outline: none;
          transition: border-color 0.15s;
        }
        .card-select:focus { border-color: #c0392b; }

        .card-nav {
          display: flex; align-items: center;
          justify-content: space-between;
          padding-top: 0.25rem;
          border-top: 1px solid #f1f3f5;
          margin-top: 0.1rem;
        }
        .nav-btn {
          background: none; border: 1px solid #e9ecef;
          border-radius: 5px; color: #adb5bd;
          width: 26px; height: 22px; cursor: pointer;
          font-size: 0.8rem; display: flex;
          align-items: center; justify-content: center;
          transition: all 0.15s; font-family: var(--font-sans);
        }
        .nav-btn:hover:not(:disabled) { color: #c0392b; border-color: #c0392b; background: #fff5f5; }
        .nav-btn:disabled { opacity: 0.25; cursor: not-allowed; }
        .nav-label { font-size: 0.65rem; color: #adb5bd; text-transform: uppercase; letter-spacing: 0.05em; font-weight: 500; }
      `}</style>
    </div>
  );
}

function CreateModal({ onClose, onCreate }: {
  onClose: () => void;
  onCreate: (title: string, desc: string, priority: string) => Promise<void>;
}) {
  const [title, setTitle]       = useState("");
  const [desc, setDesc]         = useState("");
  const [priority, setPriority] = useState("medium");
  const [loading, setLoading]   = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setLoading(true);
    await onCreate(title.trim(), desc.trim(), priority);
    setLoading(false);
    onClose();
  };

  return (
    <div className="overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-head">
          <h2 className="modal-title">New task</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>
        <form onSubmit={submit} className="modal-form">
          <div className="mfield">
            <label className="mlabel">Title</label>
            <input autoFocus className="minput" placeholder="What needs to be done?"
              value={title} onChange={e => setTitle(e.target.value)} required />
          </div>
          <div className="mfield">
            <label className="mlabel">Description <span style={{color:"#adb5bd",fontWeight:400}}>(optional)</span></label>
            <textarea className="minput mtextarea" placeholder="Add more details…"
              value={desc} onChange={e => setDesc(e.target.value)} rows={3} />
          </div>
          <div className="mfield">
            <label className="mlabel">Priority</label>
            <div className="priority-group">
              {(["low","medium","high"] as const).map(p => (
                <button key={p} type="button"
                  className={`priority-opt priority-opt--${p} ${priority === p ? "active" : ""}`}
                  onClick={() => setPriority(p)}>
                  {p.charAt(0).toUpperCase() + p.slice(1)}
                </button>
              ))}
            </div>
          </div>
          <div className="modal-actions">
            <button type="button" className="btn-cancel" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn-submit" disabled={loading}>
              {loading ? "Adding…" : "Add task"}
            </button>
          </div>
        </form>
      </div>

      <style jsx>{`
        .overlay {
          position: fixed; inset: 0;
          background: rgba(0,0,0,0.3);
          display: flex; align-items: center; justify-content: center;
          z-index: 50; padding: 1rem;
          backdrop-filter: blur(2px);
        }
        .modal {
          background: #fff;
          border: 1px solid #e9ecef;
          border-radius: 12px;
          width: 100%; max-width: 440px;
          padding: 1.5rem;
          box-shadow: 0 8px 32px rgba(0,0,0,0.12);
        }
        .modal-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.25rem; }
        .modal-title { font-size: 1rem; font-weight: 600; color: #0f0f0f; letter-spacing: -0.02em; }
        .modal-close { background: none; border: none; color: #adb5bd; font-size: 1.25rem; cursor: pointer; transition: color 0.15s; }
        .modal-close:hover { color: #212529; }

        .modal-form { display: flex; flex-direction: column; gap: 1rem; }
        .mfield { display: flex; flex-direction: column; gap: 0.35rem; }
        .mlabel { font-size: 0.8rem; font-weight: 500; color: #495057; }
        .minput {
          background: #f8f9fa; border: 1.5px solid #e9ecef; border-radius: 8px;
          padding: 0.65rem 0.85rem; font-size: 0.875rem; color: #0f0f0f;
          outline: none; transition: border-color 0.15s, box-shadow 0.15s;
          font-family: var(--font-sans); width: 100%;
        }
        .minput::placeholder { color: #ced4da; }
        .minput:focus { border-color: #c0392b; background: #fff; box-shadow: 0 0 0 3px rgba(192,57,43,0.08); }
        .mtextarea { resize: vertical; }

        .priority-group { display: flex; gap: 0.5rem; }
        .priority-opt {
          flex: 1; padding: 0.5rem; border-radius: 7px; border: 1.5px solid #e9ecef;
          background: #f8f9fa; font-size: 0.775rem; font-weight: 500;
          cursor: pointer; font-family: var(--font-sans); transition: all 0.15s; color: #495057;
        }
        .priority-opt--low.active    { background: #ebfbee; border-color: #2f9e44; color: #2f9e44; }
        .priority-opt--medium.active { background: #fff9db; border-color: #e67700; color: #e67700; }
        .priority-opt--high.active   { background: #fff5f5; border-color: #c92a2a; color: #c92a2a; }

        .modal-actions { display: flex; gap: 0.5rem; justify-content: flex-end; padding-top: 0.25rem; }
        .btn-cancel {
          background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 7px;
          padding: 0.6rem 1rem; font-size: 0.825rem; font-weight: 500;
          cursor: pointer; font-family: var(--font-sans); color: #495057; transition: all 0.15s;
        }
        .btn-cancel:hover { background: #e9ecef; }
        .btn-submit {
          background: #0f0f0f; color: #fff; border: none; border-radius: 7px;
          padding: 0.6rem 1.25rem; font-size: 0.825rem; font-weight: 600;
          cursor: pointer; font-family: var(--font-sans); transition: background 0.15s;
        }
        .btn-submit:hover:not(:disabled) { background: #c0392b; }
        .btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }
      `}</style>
    </div>
  );
}

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser]           = useState<User | null>(null);
  const [mounted, setMounted]     = useState(false);
  const [todos, setTodos]         = useState<Todo[]>([]);
  const [members, setMembers]     = useState<User[]>([]);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState("");
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    const u = getUser();
    setUser(u);
    setMounted(true);
    if (!u) router.replace("/login");
  }, [router]);

  const isAdmin = user?.role === "admin";

  const fetchTodos = useCallback(async () => {
    setError("");
    try {
      const res = await todoAPI.getAll();
      if (res.success) setTodos((res.data as Todo[]) ?? []);
      else setError(res.message);
    } catch { setError("Cannot reach API"); }
    finally { setLoading(false); }
  }, []);

  const fetchMembers = useCallback(async () => {
    if (!isAdmin) return;
    try {
      const res = await userAPI.getMembers();
      if (res.success) setMembers((res.data as User[]) ?? []);
    } catch { /* non-critical */ }
  }, [isAdmin]);

  useEffect(() => {
    if (mounted && user) { fetchTodos(); fetchMembers(); }
  }, [mounted, user, fetchTodos, fetchMembers]);

  const handleCreate = async (title: string, desc: string, priority: string) => {
    const res = await todoAPI.create(title, desc, priority);
    if (res.success) fetchTodos(); else setError(res.message);
  };

  const handleStatusChange = async (id: number, status: WorkflowStatus) => {
    const res = await todoAPI.updateStatus(id, status);
    if (res.success) setTodos(prev => prev.map(t => t.id === id ? { ...t, ...res.data as Todo } : t));
  };

  const handleAssign = async (todoId: number, userId: number) => {
    const res = await todoAPI.assign(todoId, userId);
    if (res.success) fetchTodos(); else setError(res.message);
  };

  const handleDelete = async (id: number) => {
    const res = await todoAPI.delete(id);
    if (res.success) setTodos(prev => prev.filter(t => t.id !== id));
    else setError(res.message);
  };

  const todosInColumn = (status: WorkflowStatus) => todos.filter(t => t.status === status);

  if (!mounted || !user) return null;

  return (
    <div className="root">
      {/* Topbar */}
      <header className="topbar">
        <div className="topbar-left">
          <div className="logo-mark">T</div>
          <span className="logo-name">Taskflow</span>
          <div className="role-chip">{user.role}</div>
        </div>
        <div className="topbar-right">
          <span className="user-name">{user.name}</span>
          {isAdmin && (
            <button className="btn-new" onClick={() => setShowCreate(true)}>
              <span>+</span> New task
            </button>
          )}
          <button className="btn-signout" onClick={() => { clearAuth(); router.replace("/login"); }}>
            Sign out
          </button>
        </div>
      </header>

      {error && (
        <div className="error-bar">
          <span>{error}</span>
          <button onClick={() => setError("")}>×</button>
        </div>
      )}

      {/* Stats */}
      <div className="stats">
        <div className="stat">
          <span className="stat-n">{todos.length}</span>
          <span className="stat-l">Total</span>
        </div>
        {COLUMNS.map(col => (
          <div className="stat" key={col.status}>
            <span className="stat-n">{todosInColumn(col.status).length}</span>
            <span className="stat-l">{col.label}</span>
          </div>
        ))}
      </div>

      {/* Board */}
      {loading ? (
        <div className="loading">Loading tasks…</div>
      ) : (
        <div className="board">
          {COLUMNS.map(col => (
            <div className="col" key={col.status}>
              <div className="col-head">
                <span className={`col-dot dot-${col.status}`} />
                <span className="col-label">{col.label}</span>
                <span className="col-badge">{todosInColumn(col.status).length}</span>
              </div>
              <div className="col-body">
                {todosInColumn(col.status).length === 0
                  ? <p className="col-empty">No tasks here</p>
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
        .root { min-height: 100vh; background: #f1f3f5; display: flex; flex-direction: column; }

        /* Topbar */
        .topbar {
          display: flex; align-items: center; justify-content: space-between;
          padding: 0 1.5rem; height: 56px;
          background: #fff; border-bottom: 1px solid #e9ecef;
          flex-shrink: 0;
        }
        .topbar-left, .topbar-right { display: flex; align-items: center; gap: 0.75rem; }

        .logo-mark {
          width: 28px; height: 28px; background: #c0392b; color: #fff;
          border-radius: 6px; font-size: 0.8rem; font-weight: 700;
          display: flex; align-items: center; justify-content: center;
        }
        .logo-name { font-size: 0.95rem; font-weight: 600; color: #0f0f0f; letter-spacing: -0.02em; }

        .role-chip {
          font-size: 0.65rem; font-weight: 600; text-transform: uppercase;
          letter-spacing: 0.06em; color: #c92a2a;
          background: #fff5f5; border: 1px solid #ffc9c9;
          padding: 2px 8px; border-radius: 20px;
        }

        .user-name { font-size: 0.825rem; color: #868e96; font-weight: 400; }

        .btn-new {
          display: flex; align-items: center; gap: 0.3rem;
          background: #0f0f0f; color: #fff; border: none;
          border-radius: 7px; padding: 0.45rem 0.9rem;
          font-size: 0.825rem; font-weight: 600; cursor: pointer;
          font-family: var(--font-sans); transition: background 0.15s;
          letter-spacing: -0.01em;
        }
        .btn-new:hover { background: #c0392b; }

        .btn-signout {
          background: transparent; border: 1px solid #e9ecef; color: #868e96;
          border-radius: 7px; padding: 0.45rem 0.9rem;
          font-size: 0.8rem; cursor: pointer; font-family: var(--font-sans);
          transition: all 0.15s;
        }
        .btn-signout:hover { border-color: #adb5bd; color: #212529; }

        /* Error */
        .error-bar {
          background: #fff5f5; border-bottom: 1px solid #ffc9c9;
          color: #c92a2a; font-size: 0.825rem;
          padding: 0.6rem 1.5rem;
          display: flex; justify-content: space-between; align-items: center;
        }
        .error-bar button { background: none; border: none; color: #c92a2a; font-size: 1rem; cursor: pointer; }

        /* Stats */
        .stats {
          display: flex; gap: 1px; background: #e9ecef;
          border-bottom: 1px solid #e9ecef; flex-shrink: 0;
        }
        .stat {
          flex: 1; display: flex; flex-direction: column; align-items: center;
          padding: 0.75rem; background: #fff; gap: 2px;
        }
        .stat-n { font-size: 1.15rem; font-weight: 700; color: #0f0f0f; letter-spacing: -0.03em; }
        .stat-l { font-size: 0.65rem; color: #adb5bd; font-weight: 500; text-transform: uppercase; letter-spacing: 0.05em; }

        /* Board */
        .board {
          display: grid; grid-template-columns: repeat(4, 1fr);
          gap: 1px; background: #e9ecef; flex: 1;
        }
        .col { background: #f8f9fa; display: flex; flex-direction: column; min-height: calc(100vh - 148px); }

        .col-head {
          display: flex; align-items: center; gap: 0.5rem;
          padding: 0.85rem 1rem; border-bottom: 1px solid #e9ecef;
          background: #fff; position: sticky; top: 0; z-index: 1;
        }
        .col-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
        .dot-todo        { background: #adb5bd; }
        .dot-in_progress { background: #f59f00; }
        .dot-review      { background: #7950f2; }
        .dot-done        { background: #2f9e44; }

        .col-label { font-size: 0.75rem; font-weight: 600; color: #212529; flex: 1; letter-spacing: -0.01em; }
        .col-badge {
          font-size: 0.65rem; color: #868e96; background: #f1f3f5;
          border: 1px solid #e9ecef; padding: 1px 7px; border-radius: 20px; font-weight: 500;
        }

        .col-body { padding: 0.75rem; display: flex; flex-direction: column; gap: 0.5rem; flex: 1; overflow-y: auto; }
        .col-empty { font-size: 0.775rem; color: #ced4da; text-align: center; padding: 2rem 0; }

        /* Loading */
        .loading { flex: 1; display: flex; align-items: center; justify-content: center; font-size: 0.875rem; color: #adb5bd; }

        /* Responsive */
        @media (max-width: 900px) { .board { grid-template-columns: repeat(2, 1fr); } }
        @media (max-width: 560px) { .board { grid-template-columns: 1fr; } .stats { flex-wrap: wrap; } }
      `}</style>
    </div>
  );
}