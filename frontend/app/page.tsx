"use client";

import { useState, useEffect } from "react";

// ============================================================
// TYPES
// Match exactly what our Go API returns
// ============================================================
type Todo = {
  id: number;
  title: string;
  description: string;
  completed: boolean;
  priority: "low" | "medium" | "high";
  created_at: string;
  updated_at: string;
};

// ============================================================
// API URL - points to our Go backend
// ============================================================
// const API_URL = "http://localhost:9090";
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9090";

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================
export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Form state
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState("medium");

  // ============================================================
  // FETCH ALL TODOS from Go API
  // ============================================================
  const fetchTodos = async () => {
    try {
      const res = await fetch(`${API_URL}/todos`);
      const data = await res.json();
      if (data.success) {
        setTodos(data.data);
      }
    } catch {
      setError("Cannot connect to API. Is the Go server running?");
    } finally {
      setLoading(false);
    }
  };

  // Fetch todos when page loads
  useEffect(() => {
    fetchTodos();
  }, []);

  // ============================================================
  // CREATE TODO - POST /todos
  // ============================================================
  const createTodo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    try {
      const res = await fetch(`${API_URL}/todos`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description, priority }),
      });
      const data = await res.json();
      if (data.success) {
        setTitle("");
        setDescription("");
        setPriority("medium");
        fetchTodos(); // refresh list
      }
    } catch {
      setError("Error creating todo");
    }
  };

  // ============================================================
  // TOGGLE COMPLETE - PUT /todos/:id
  // ============================================================
  const toggleComplete = async (todo: Todo) => {
    try {
      await fetch(`${API_URL}/todos/${todo.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: todo.title,
          description: todo.description,
          priority: todo.priority,
          completed: !todo.completed,
        }),
      });
      fetchTodos();
    } catch {
      setError("Error updating todo");
    }
  };

  // ============================================================
  // DELETE TODO - DELETE /todos/:id
  // ============================================================
  const deleteTodo = async (id: number) => {
    try {
      await fetch(`${API_URL}/todos/${id}`, { method: "DELETE" });
      fetchTodos();
    } catch {
      setError("Error deleting todo");
    }
  };

  // ============================================================
  // PRIORITY BADGE COLORS
  // ============================================================
  const priorityColor = (priority: string) => {
    switch (priority) {
      case "high": return "bg-red-100 text-red-700";
      case "medium": return "bg-yellow-100 text-yellow-700";
      case "low": return "bg-green-100 text-green-700";
      default: return "bg-gray-100 text-gray-700";
    }
  };

  // Stats
  const completedCount = todos.filter((t) => t.completed).length;
  const pendingCount = todos.filter((t) => !t.completed).length;

  // ============================================================
  // RENDER
  // ============================================================
  return (
    <main className="min-h-screen bg-gray-50 py-10 px-4">
      <div className="max-w-2xl mx-auto">

        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">📝 Todo App</h1>
          <p className="text-gray-500 text-sm mt-1">
            Built with Go + PostgreSQL + Redis + Next.js
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          <div className="bg-white rounded-xl p-4 text-center shadow-sm border border-gray-100">
            <div className="text-2xl font-bold text-indigo-600">{todos.length}</div>
            <div className="text-xs text-gray-500 mt-1">Total</div>
          </div>
          <div className="bg-white rounded-xl p-4 text-center shadow-sm border border-gray-100">
            <div className="text-2xl font-bold text-green-600">{completedCount}</div>
            <div className="text-xs text-gray-500 mt-1">Completed</div>
          </div>
          <div className="bg-white rounded-xl p-4 text-center shadow-sm border border-gray-100">
            <div className="text-2xl font-bold text-orange-500">{pendingCount}</div>
            <div className="text-xs text-gray-500 mt-1">Pending</div>
          </div>
        </div>

        {/* Add Todo Form */}
        <form
          onSubmit={createTodo}
          className="bg-white rounded-xl p-6 shadow-sm border border-gray-100 mb-6"
        >
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            Add New Task
          </h2>
          <div className="flex flex-col gap-3">
            <input
              type="text"
              placeholder="Task title..."
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-indigo-400 placeholder-gray-400 text-gray-800"
              required
            />
            <input
              type="text"
              placeholder="Description (optional)"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-indigo-400 placeholder-gray-400 text-gray-800"
            />
            <div className="flex gap-3">
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
                className="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-indigo-400 placeholder-gray-400 text-gray-800"
              >
                <option value="low">Low Priority</option>
                <option value="medium">Medium Priority</option>
                <option value="high">High Priority</option>
              </select>
              <button
                type="submit"
                className="flex-1 bg-indigo-600 text-white px-6 py-2 rounded-lg text-sm font-semibold hover:bg-indigo-700 transition"
              >
                Add Task
              </button>
            </div>
          </div>
        </form>

        {/* Todo List */}
        <div className="flex flex-col gap-3">
          {loading ? (
            <div className="text-center text-gray-400 py-10">Loading...</div>
          ) : todos.length === 0 ? (
            <div className="text-center text-gray-400 py-10">
              No todos yet. Add one above! 👆
            </div>
          ) : (
            todos.map((todo) => (
              <div
                key={todo.id}
                className={`bg-white rounded-xl p-4 shadow-sm border transition ${todo.completed
                    ? "border-green-100 bg-green-50"
                    : "border-gray-100"
                  }`}
              >
                <div className="flex items-start gap-3">
                  {/* Checkbox */}
                  <button
                    onClick={() => toggleComplete(todo)}
                    className={`mt-1 w-5 h-5 rounded-full border-2 flex-shrink-0 transition ${todo.completed
                        ? "bg-green-500 border-green-500"
                        : "border-gray-300 hover:border-indigo-400"
                      }`}
                  />

                  {/* Content */}
                  <div className="flex-1">
                    <p
                      className={`text-sm font-medium ${todo.completed
                          ? "line-through text-gray-400"
                          : "text-gray-800"
                        }`}
                    >
                      {todo.title}
                    </p>
                    {todo.description && (
                      <p className="text-xs text-gray-400 mt-1">
                        {todo.description}
                      </p>
                    )}
                    <span
                      className={`inline-block mt-2 text-xs px-2 py-0.5 rounded-full font-medium ${priorityColor(todo.priority)}`}
                    >
                      {todo.priority}
                    </span>
                  </div>

                  {/* Delete */}
                  <button
                    onClick={() => deleteTodo(todo.id)}
                    className="text-gray-300 hover:text-red-400 transition text-lg"
                  >
                    ✕
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </main>
  );
}