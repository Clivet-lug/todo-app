"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { authAPI } from "../lib/api";
import { saveAuth } from "../lib/auth";
import { AuthResponse } from "../types";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await authAPI.login(email, password);
      if (res.success && res.data) {
        const { token, user } = res.data as AuthResponse;
        saveAuth(token, user);
        router.replace("/dashboard");
      } else {
        setError(res.message || "Login failed");
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

        <h1 className="heading">Welcome back</h1>
        <p className="subtext">Sign in to your account to continue</p>

        {error && <div className="error">{error}</div>}

        <form onSubmit={handleSubmit} className="form">
          <div className="field">
            <label className="label">Email address</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              className="input"
              placeholder="you@example.com"
              required
              autoFocus
            />
          </div>
          <div className="field">
            <label className="label">Password</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="input"
              placeholder="••••••••"
              required
            />
          </div>
          <button type="submit" className="btn" disabled={loading}>
            {loading ? "Signing in…" : "Sign in"}
          </button>
        </form>

        <p className="footer">
          Don&apos;t have an account?{" "}
          <Link href="/register" className="link">Create one</Link>
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
          max-width: 400px;
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
          font-family: var(--font-sans);
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
          font-weight: 400;
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

        .form {
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }

        .field {
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
        }

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