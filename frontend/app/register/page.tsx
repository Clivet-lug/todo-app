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
        <div className="root">
            <div className="card">
                <div className="brand">
                    <div className="brand-mark">T</div>
                    <span className="brand-name">Taskflow</span>
                </div>

                <h1 className="heading">Create your account</h1>
                <p className="subtext">Start managing tasks with your team</p>

                {error && <div className="error">{error}</div>}

                <form onSubmit={handleSubmit} className="form">
                    <div className="field">
                        <label className="label">Full name</label>
                        <input type="text" value={name} onChange={e => setName(e.target.value)}
                            className="input" placeholder="Alice Smith" required autoFocus />
                    </div>

                    <div className="field">
                        <label className="label">Email address</label>
                        <input type="email" value={email} onChange={e => setEmail(e.target.value)}
                            className="input" placeholder="you@example.com" required />
                    </div>

                    <div className="field">
                        <label className="label">Password</label>
                        <input type="password" value={password} onChange={e => setPassword(e.target.value)}
                            className="input" placeholder="At least 6 characters" required minLength={6} />
                    </div>

                    <div className="field">
                        <label className="label">Role</label>
                        <div className="role-group">
                            <button type="button"
                                className={`role-opt ${role === "member" ? "active" : ""}`}
                                onClick={() => setRole("member")}>
                                <span className="role-icon">👤</span>
                                <span className="role-text">
                                    <strong>Member</strong>
                                    <small>Works on assigned tasks</small>
                                </span>
                            </button>
                            <button type="button"
                                className={`role-opt ${role === "admin" ? "active" : ""}`}
                                onClick={() => setRole("admin")}>
                                <span className="role-icon">⚙️</span>
                                <span className="role-text">
                                    <strong>Admin</strong>
                                    <small>Creates & assigns tasks</small>
                                </span>
                            </button>
                        </div>
                    </div>

                    <button type="submit" className="btn" disabled={loading}>
                        {loading ? "Creating account…" : "Create account"}
                    </button>
                </form>

                <p className="footer">
                    Already have an account?{" "}
                    <Link href="/login" className="link">Sign in</Link>
                </p>
            </div>

            <style jsx>{`
        .root {
          min-height: 100vh;
          background: #f8f9fa;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 1.5rem;
        }

        .card {
          width: 100%;
          max-width: 420px;
          background: #ffffff;
          border: 1px solid #e9ecef;
          border-radius: 12px;
          padding: 2.5rem;
          box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.04);
        }

        .brand {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          margin-bottom: 2rem;
        }

        .brand-mark {
          width: 28px;
          height: 28px;
          background: #c0392b;
          color: #fff;
          border-radius: 6px;
          font-size: 0.8rem;
          font-weight: 700;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .brand-name {
          font-size: 0.95rem;
          font-weight: 600;
          color: #0f0f0f;
          letter-spacing: -0.01em;
        }

        .heading {
          font-size: 1.5rem;
          font-weight: 700;
          color: #0f0f0f;
          letter-spacing: -0.03em;
          margin-bottom: 0.35rem;
          line-height: 1.2;
        }

        .subtext {
          font-size: 0.875rem;
          color: #868e96;
          margin-bottom: 1.75rem;
        }

        .error {
          background: #fff5f5;
          border: 1px solid #ffc9c9;
          color: #c92a2a;
          font-size: 0.825rem;
          padding: 0.7rem 0.9rem;
          border-radius: 8px;
          margin-bottom: 1.25rem;
        }

        .form { display: flex; flex-direction: column; gap: 1rem; }
        .field { display: flex; flex-direction: column; gap: 0.4rem; }

        .label {
          font-size: 0.8rem;
          font-weight: 500;
          color: #495057;
        }

        .input {
          background: #f8f9fa;
          border: 1.5px solid #e9ecef;
          border-radius: 8px;
          padding: 0.65rem 0.85rem;
          font-size: 0.9rem;
          color: #0f0f0f;
          outline: none;
          transition: border-color 0.15s, box-shadow 0.15s;
          font-family: var(--font-sans);
          width: 100%;
        }

        .input::placeholder { color: #ced4da; }
        .input:focus {
          border-color: #c0392b;
          background: #fff;
          box-shadow: 0 0 0 3px rgba(192,57,43,0.08);
        }

        .role-group {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 0.5rem;
        }

        .role-opt {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          padding: 0.75rem;
          background: #f8f9fa;
          border: 1.5px solid #e9ecef;
          border-radius: 8px;
          cursor: pointer;
          text-align: left;
          transition: border-color 0.15s, background 0.15s;
          font-family: var(--font-sans);
        }

        .role-opt.active {
          border-color: #c0392b;
          background: #fff5f5;
        }

        .role-icon { font-size: 1rem; flex-shrink: 0; }

        .role-text {
          display: flex;
          flex-direction: column;
          gap: 1px;
        }

        .role-text strong {
          font-size: 0.8rem;
          font-weight: 600;
          color: #0f0f0f;
          display: block;
        }

        .role-text small {
          font-size: 0.68rem;
          color: #868e96;
          display: block;
          line-height: 1.3;
        }

        .btn {
          margin-top: 0.25rem;
          background: #0f0f0f;
          color: #fff;
          border: none;
          border-radius: 8px;
          padding: 0.75rem;
          font-size: 0.875rem;
          font-weight: 600;
          cursor: pointer;
          font-family: var(--font-sans);
          transition: background 0.15s, opacity 0.15s;
          letter-spacing: -0.01em;
        }

        .btn:hover:not(:disabled) { background: #c0392b; }
        .btn:disabled { opacity: 0.5; cursor: not-allowed; }

        .footer {
          margin-top: 1.5rem;
          font-size: 0.825rem;
          color: #868e96;
          text-align: center;
        }

        .link { color: #c0392b; text-decoration: none; font-weight: 500; }
        .link:hover { text-decoration: underline; }
      `}</style>
        </div>
    );
}