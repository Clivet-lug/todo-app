package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DATA MODELS 
// Todo represents a single todo item
type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority"` // low, medium, high
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Standard wrapper for all responses
// Every endpoint returns this same structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}
 
// IN-MEMORY STORAGE 
var todos = []Todo{
	{
		ID:          1,
		Title:       "Learn Go",
		Description: "Study Go fundamentals and build APIs",
		Completed:   true,
		Priority:    "high",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	{
		ID:          2,
		Title:       "Learn PostgreSQL",
		Description: "Connect Go to a real database",
		Completed:   false,
		Priority:    "high",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	{
		ID:          3,
		Title:       "Learn Docker",
		Description: "Containerize the todo app",
		Completed:   false,
		Priority:    "medium",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

var nextID = 4

// MAIN - Register routes and start server 
func main() {
	// Health check
	http.HandleFunc("/health", healthCheck)

	// Todo routes - we check the HTTP method inside each handler
	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/", todoHandler) // handles /todos/{id}

	fmt.Println("🚀 Todo API running on http://localhost:9090")
	fmt.Println("📋 Endpoints:")
	fmt.Println("   GET    /health")
	fmt.Println("   GET    /todos")
	fmt.Println("   POST   /todos")
	fmt.Println("   GET    /todos/{id}")
	fmt.Println("   PUT    /todos/{id}")
	fmt.Println("   DELETE /todos/{id}")

	http.ListenAndServe(":9090", nil)
}

 
// HELPERS
// sendJSON writes a JSON response with the correct headers
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// Allow requests from our Next.js frontend
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// findTodoIndex finds a todo by ID and returns its index, -1 if not found
func findTodoIndex(id int) int {
	for i, todo := range todos {
		if todo.ID == id {
			return i
		}
	}
	return -1
}

// HEALTH CHECK
// GET /health 
func healthCheck(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo API is healthy",
	})
}
 
// TODOS HANDLER - routes GET /todos and POST /todos 
func todosHandler(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS request (needed for Next.js)
	if r.Method == http.MethodOptions {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTodos(w, r)
	case http.MethodPost:
		createTodo(w, r)
	default:
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

// SINGLE TODO HANDLER - routes GET/PUT/DELETE /todos/{id}
func todoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	// Extract ID from URL e.g. /todos/3 → "3"
	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid todo ID",
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTodoByID(w, r, id)
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		deleteTodo(w, r, id)
	default:
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

// GET ALL TODOS
// GET /todos 
func getTodos(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todos retrieved successfully",
		Data:    todos,
		Count:   len(todos),
	})
}

// GET SINGLE TODO
func getTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	index := findTodoIndex(id)

	if index == -1 {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo retrieved successfully",
		Data:    todos[index],
	})
}
 
// CREATE TODO 
func createTodo(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a Todo struct
	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(input.Title) == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	// Set default priority if not provided
	if input.Priority == "" {
		input.Priority = "medium"
	}

	// Build the new todo
	newTodo := Todo{
		ID:          nextID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Completed:   false,
		Priority:    input.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	todos = append(todos, newTodo)
	nextID++

	sendJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    newTodo,
	})
}

// UPDATE TODO
func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	index := findTodoIndex(id)

	if index == -1 {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Update only provided fields, keep existing values
	existing := todos[index]

	if strings.TrimSpace(input.Title) != "" {
		existing.Title = strings.TrimSpace(input.Title)
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.Priority != "" {
		existing.Priority = input.Priority
	}

	// Completed can be explicitly set to true or false
	existing.Completed = input.Completed
	existing.UpdatedAt = time.Now()

	todos[index] = existing

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo updated successfully",
		Data:    existing,
	})
}

// DELETE TODO
// DELETE /todos/{id} 
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	index := findTodoIndex(id)

	if index == -1 {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	todos = append(todos[:index], todos[index+1:]...)

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}
