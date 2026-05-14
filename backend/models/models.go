package models

import "time"

// ─── USER ────────────────────────────────────────────────────────────────────

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never sent in JSON responses
	Role      string    `json:"role"` // "admin" | "member"
	CreatedAt time.Time `json:"created_at"`
}

// ─── TODO ────────────────────────────────────────────────────────────────────

// WorkflowStatus represents the four pipeline stages
type WorkflowStatus string

const (
	StatusTodo       WorkflowStatus = "todo"
	StatusInProgress WorkflowStatus = "in_progress"
	StatusReview     WorkflowStatus = "review"
	StatusDone       WorkflowStatus = "done"
)

// ValidStatuses for input validation
var ValidStatuses = map[WorkflowStatus]bool{
	StatusTodo:       true,
	StatusInProgress: true,
	StatusReview:     true,
	StatusDone:       true,
}

type Todo struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Completed   bool           `json:"completed"`   // kept for backward compat; mirrors status==done
	Priority    string         `json:"priority"`    // "low" | "medium" | "high"
	Status      WorkflowStatus `json:"status"`      // workflow stage
	AssignedTo  *int           `json:"assigned_to"` // nullable user ID
	AssignedBy  *int           `json:"assigned_by"` // nullable user ID (admin who assigned)
	Assignee    *UserSummary   `json:"assignee,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// UserSummary is a lightweight user view embedded in Todo responses
type UserSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ─── AUTH ─────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional; defaults to "member"
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ─── API RESPONSE ────────────────────────────────────────────────────────────

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}