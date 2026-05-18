package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Clivet-lug/todo-app/backend/databases"
	"github.com/Clivet-lug/todo-app/backend/middleware"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/repositories"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

// GET ALL TODOS
// Admins see everything. Members see only their assigned todos.
// GET /todos
func GetTodos(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	isAdmin := claims.Role == "admin"

	db := databases.ConnectDB()
	defer db.Close()

	todos, err := repositories.NewConnector(db).GetTodos(claims.UserID, isAdmin)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error fetching todos",
		})
		return
	}

	if todos == nil {
		todos = []*models.Todo{}
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todos retrieved successfully",
		Data:    todos,
		Count:   len(todos),
	})
}

// GET SINGLE TODO
// GET /todos/{id}
func GetTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)

	db := databases.ConnectDB()
	defer db.Close()

	todo, err := repositories.NewConnector(db).GetTodoByID(id)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	// Members can only view todos assigned to them
	if claims.Role != "admin" {
		if todo.AssignedTo == nil || *todo.AssignedTo != claims.UserID {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Access denied",
			})
			return
		}
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo retrieved successfully",
		Data:    todo,
	})
}

// CREATE TODO
// Admin only. POST /todos
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	if input.Priority == "" {
		input.Priority = "medium"
	}

	// Default status for new todos
	status := models.StatusTodo

	db := databases.ConnectDB()
	defer db.Close()

	var todo models.Todo
	err := db.QueryRow(`
		INSERT INTO todos (title, description, priority, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Description),
		input.Priority,
		status,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.Status,
		&todo.AssignedTo,
		&todo.AssignedBy,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error creating todo",
		})
		return
	}

	utils.SendJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    todo,
	})
}

// UPDATE TODO
// Admin only (title, description, priority edits). PUT /todos/{id}
func UpdateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	db := databases.ConnectDB()
	defer db.Close()

	var todo models.Todo
	err := db.QueryRow(`
		UPDATE todos
		SET title       = COALESCE(NULLIF($1, ''), title),
		    description = COALESCE(NULLIF($2, ''), description),
		    priority    = COALESCE(NULLIF($3, ''), priority),
		    updated_at  = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at
	`,
		input.Title,
		input.Description,
		input.Priority,
		id,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.Status,
		&todo.AssignedTo,
		&todo.AssignedBy,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo updated successfully",
		Data:    todo,
	})
}

// DELETE TODO
// Admin only. DELETE /todos/{id}
func DeleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	db := databases.ConnectDB()
	defer db.Close()

	result, err := db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error deleting todo",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}

// ASSIGN TODO
// Admin only. PUT /todos/{id}/assign
func AssignTodo(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)

	var body struct {
		AssignedTo int `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AssignedTo == 0 {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "assigned_to (user ID) is required",
		})
		return
	}

	db := databases.ConnectDB()
	defer db.Close()

	// Verify the target user exists and is a member
	userRepo := repositories.NewUserRepository(db)
	targetUser, err := userRepo.FindByID(body.AssignedTo)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Target user not found",
		})
		return
	}
	if targetUser.Role != "member" {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Todos can only be assigned to members",
		})
		return
	}

	todo, err := repositories.NewConnector(db).AssignTodo(id, body.AssignedTo, claims.UserID)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Todo assigned to %s", targetUser.Name),
		Data:    todo,
	})
}

// UPDATE STATUS
// Members update their own assigned todos; admins can update any.
// PUT /todos/{id}/status
// Body: { "status": "in_progress" }
func UpdateTodoStatus(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)

	var body struct {
		Status models.WorkflowStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if !models.ValidStatuses[body.Status] {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: `status must be one of: "todo", "in_progress", "review", "done"`,
		})
		return
	}

	db := databases.ConnectDB()
	defer db.Close()

	connector := repositories.NewConnector(db)

	// Members can only move their own todos
	if claims.Role != "admin" {
		todo, err := connector.GetTodoByID(id)
		if err != nil {
			utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Todo with ID %d not found", id),
			})
			return
		}
		if todo.AssignedTo == nil || *todo.AssignedTo != claims.UserID {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "You can only update status of todos assigned to you",
			})
			return
		}
	}

	todo, err := connector.UpdateStatus(id, body.Status)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Status updated successfully",
		Data:    todo,
	})
}

// ─── LIST MEMBERS ─────────────────────────────────────────────────────────────
// Admin only. GET /users/members
// Returns all member-role users (for the assignment dropdown).
func ListMembers(w http.ResponseWriter, r *http.Request) {
	db := databases.ConnectDB()
	defer db.Close()

	members, err := repositories.NewUserRepository(db).ListMembers()
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error fetching members",
		})
		return
	}

	if members == nil {
		members = []*models.User{}
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Members retrieved successfully",
		Data:    members,
		Count:   len(members),
	})
}
