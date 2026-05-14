"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { authAPI } from "../lib/api";
import { saveAuth } from "../lib/auth";
import { AuthResponse } from "../types";

export default function RegisterPage() {
    const router = useRouter();
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [role, setRole] = useState<"member" | "admin">("member");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);

        try {
            const res = await authAPI.register(name, email, password, role);
            if (res.success && res.data) {
                const { token, user } = res.data as AuthResponse;
                saveAuth(token, user);
                router.replace("/dashboard");
            } else {
                setError(res.message || "Registration failed");
            }
        } catch {
            setError("Cannot reach server. Is the API running?");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-root">
            <div className="auth-grid-bg" />

            <div className="auth-card">
                <div className="auth-brand">
                    <span className="auth-brand-icon">▣</span>
                    <span className="auth-brand-name">TASKFLOW</span>
                </div>

                <h1 className="auth-heading">Create account</h1>
                <p className="auth-sub">Get started with Taskflow</p>

                {error && <div className="auth-error">{error}</div>}

                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="auth-field">
                        <label className="auth-label">Full name</label>
                        <input
                            type="text"
                            value={name}
                            onChange={e => setName(e.target.value)}
                            className="auth-input"
                            placeholder="Alice Smith"
                            required
                            autoFocus
                        />
                    </div>

                    <div className="auth-field">
                        <label className="auth-label">Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={e => setEmail(e.target.value)}
                            className="auth-input"
                            placeholder="you@example.com"
                            required
                        />
                    </div>

                    <div className="auth-field">
                        <label className="auth-label">Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                            className="auth-input"
                            placeholder="At least 6 characters"
                            required
                            minLength={6}
                        />
                    </div>

                    <div className="auth-field">
                        <label className="auth-label">Role</label>
                        <div className="role-toggle">
                            <button
                                type="button"
                                className={`role-btn ${role === "member" ? "active" : ""}`}
                                onClick={() => setRole("member")}
                            >
                                Member
                            </button>
                            <button
                                type="button"
                                className={`role-btn ${role === "admin" ? "active" : ""}`}
                                onClick={() => setRole("admin")}
                            >
                                Admin
                            </button>
                        </div>
                        <p className="role-hint">
                            {role === "admin"
                                ? "Admins create tasks, assign members, manage everything"
                                : "Members work on tasks assigned to them"}
                        </p>
                    </div>

                    <button type="submit" className="auth-btn" disabled={loading}>
                        {loading ? "Creating account…" : "Create account →"}
                    </button>
                </form>

                <p className="auth-footer">
                    Already have an account?{" "}
                    <Link href="/login" className="auth-link">
                        Sign in
                    </Link>
                </p>
            </div>

            <style jsx>{`
        .auth-root {
          min-height: 100vh;
          background: #0c0c0f;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 2rem;
          position: relative;
          overflow: hidden;
        }

        .auth-grid-bg {
          position: absolute;
          inset: 0;
          background-image:
            linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px);
          background-size: 40px 40px;
        }

        .auth-card {
          position: relative;
          width: 100%;
          max-width: 420px;
          background: #141418;
          border: 1px solid #2a2a35;
          border-radius: 4px;
          padding: 2.5rem;
        }

        .auth-brand {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          margin-bottom: 2rem;
        }

        .auth-brand-icon { font-size: 1.2rem; color: #7c6af7; }

        .auth-brand-name {
          font-family: 'Courier New', monospace;
          font-size: 0.75rem;
          font-weight: 700;
          letter-spacing: 0.2em;
          color: #7c6af7;
        }

        .auth-heading {
          font-family: 'Georgia', serif;
          font-size: 1.75rem;
          font-weight: 400;
          color: #f0f0f5;
          margin: 0 0 0.25rem;
          letter-spacing: -0.02em;
        }

        .auth-sub {
          font-size: 0.8rem;
          color: #555568;
          margin: 0 0 1.75rem;
          font-family: 'Courier New', monospace;
        }

        .auth-error {
          background: rgba(239, 68, 68, 0.1);
          border: 1px solid rgba(239, 68, 68, 0.3);
          color: #f87171;
          font-size: 0.8rem;
          padding: 0.75rem 1rem;
          border-radius: 2px;
          margin-bottom: 1.25rem;
          font-family: 'Courier New', monospace;
        }

        .auth-form { display: flex; flex-direction: column; gap: 1rem; }

        .auth-field { display: flex; flex-direction: column; gap: 0.4rem; }

        .auth-label {
          font-size: 0.7rem;
          font-weight: 600;
          letter-spacing: 0.12em;
          color: #555568;
          text-transform: uppercase;
          font-family: 'Courier New', monospace;
        }

        .auth-input {
          background: #0c0c0f;
          border: 1px solid #2a2a35;
          border-radius: 2px;
          padding: 0.65rem 0.85rem;
          font-size: 0.875rem;
          color: #f0f0f5;
          outline: none;
          transition: border-color 0.15s;
          font-family: 'Courier New', monospace;
        }

        .auth-input::placeholder { color: #33333f; }
        .auth-input:focus { border-color: #7c6af7; }

        .role-toggle {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 0;
          border: 1px solid #2a2a35;
          border-radius: 2px;
          overflow: hidden;
        }

        .role-btn {
          background: #0c0c0f;
          border: none;
          padding: 0.6rem;
          font-size: 0.78rem;
          font-weight: 600;
          letter-spacing: 0.05em;
          color: #555568;
          cursor: pointer;
          font-family: 'Courier New', monospace;
          transition: background 0.15s, color 0.15s;
        }

        .role-btn.active {
          background: #7c6af7;
          color: #fff;
        }

        .role-hint {
          font-size: 0.7rem;
          color: #3a3a50;
          font-family: 'Courier New', monospace;
          margin: 0;
          line-height: 1.5;
        }

        .auth-btn {
          margin-top: 0.5rem;
          background: #7c6af7;
          color: #fff;
          border: none;
          border-radius: 2px;
          padding: 0.75rem;
          font-size: 0.8rem;
          font-weight: 700;
          letter-spacing: 0.08em;
          cursor: pointer;
          font-family: 'Courier New', monospace;
          transition: background 0.15s, opacity 0.15s;
        }

        .auth-btn:hover:not(:disabled) { background: #6a58e5; }
        .auth-btn:disabled { opacity: 0.5; cursor: not-allowed; }

        .auth-footer {
          margin-top: 1.5rem;
          font-size: 0.78rem;
          color: #555568;
          text-align: center;
          font-family: 'Courier New', monospace;
        }

        .auth-link { color: #7c6af7; text-decoration: none; }
        .auth-link:hover { text-decoration: underline; }
      `}</style>
        </div>
    );
}