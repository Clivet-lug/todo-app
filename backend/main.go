package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver - the _ means import for side effects
)

// DATA MODELS
type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}
 
// DATABASE CONNECTION 
var db *sql.DB
 
// CONNECT TO DATABASE
func connectDB() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Build connection string from environment variables
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Open the connection
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Ping to verify connection is actually working
	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot reach database:", err)
	}

	fmt.Println("✅ Connected to PostgreSQL successfully")
}
 
// CREATE TABLE
func createTable() {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id          SERIAL PRIMARY KEY,
			title       VARCHAR(255) NOT NULL,
			description TEXT,
			completed   BOOLEAN DEFAULT FALSE,
			priority    VARCHAR(50) DEFAULT 'medium',
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error creating todos table:", err)
	}

	fmt.Println("✅ Todos table ready")
}
 
// MAIN 
func main() {
	// Connect to database first
	connectDB()
	defer db.Close() // close connection when server stops

	// Create table if it doesn't exist
	createTable()

	// Register routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/", todoHandler)

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
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// func findTodoIndex(id int) int {
// 	for i, todo := range todos {
// 		if todo.ID == id {
// 			return i
// 		}
// 	}
// 	return -1
// }

// ROUTE HANDLERS 
func todosHandler(w http.ResponseWriter, r *http.Request) {
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

func todoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		sendJSON(w, http.StatusOK, nil)
		return
	}
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

// HEALTH CHECK 
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Also check if database is still reachable
	err := db.Ping()
	if err != nil {
		sendJSON(w, http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Message: "Database unreachable",
		})
		return
	}
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo API is healthy and database is connected",
	})
}

// GET ALL TODOS
func getTodos(w http.ResponseWriter, r *http.Request) {
	// Query database for all todos, newest first
	rows, err := db.Query(`
		SELECT id, title, description, completed, priority, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
	`)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error fetching todos",
		})
		return
	}
	defer rows.Close() // always close rows when done

	// Loop through results and build our slice
	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.Priority,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			continue
		}
		todos = append(todos, todo)
	}

	// Return empty array not null if no todos
	if todos == nil {
		todos = []Todo{}
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todos retrieved successfully",
		Data:    todos,
		Count:   len(todos),
	})
}

// GET SINGLE TODO 
func getTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	var todo Todo

	err := db.QueryRow(`
		SELECT id, title, description, completed, priority, created_at, updated_at
		FROM todos WHERE id = $1
	`, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	// sql.ErrNoRows means the todo wasn't found
	if err == sql.ErrNoRows {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error fetching todo",
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo retrieved successfully",
		Data:    todo,
	})
}

// CREATE TODO 
func createTodo(w http.ResponseWriter, r *http.Request) {
	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	if input.Priority == "" {
		input.Priority = "medium"
	}

	var todo Todo

	// INSERT and return the created row immediately using RETURNING
	err := db.QueryRow(`
		INSERT INTO todos (title, description, priority)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, completed, priority, created_at, updated_at
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Description),
		input.Priority,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error creating todo",
		})
		return
	}

	sendJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    todo,
	})
}

// UPDATE TODO 
func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	var todo Todo

	err := db.QueryRow(`
		UPDATE todos
		SET title       = COALESCE(NULLIF($1, ''), title),
		    description = COALESCE(NULLIF($2, ''), description),
		    completed   = $3,
		    priority    = COALESCE(NULLIF($4, ''), priority),
		    updated_at  = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, title, description, completed, priority, created_at, updated_at
	`,
		input.Title,
		input.Description,
		input.Completed,
		input.Priority,
		id,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error updating todo",
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo updated successfully",
		Data:    todo,
	})
}
 
// DELETE TODO
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error deleting todo",
		})
		return
	}

	// Check if any row was actually deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}
