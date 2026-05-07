package main

import (
	"context"
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
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// ============================================================
// DATA MODELS
// ============================================================
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

// ============================================================
// GLOBAL CONNECTIONS
// ============================================================
var db *sql.DB
var rdb *redis.Client          // Redis client
var ctx = context.Background() // Redis needs a context

// ============================================================
// CONNECT TO POSTGRESQL
// ============================================================
func connectDB() {
    godotenv.Load()

    // Railway provides DATABASE_URL automatically
    // Fall back to individual variables for local dev
    connStr := os.Getenv("DATABASE_URL")
    
    if connStr == "" {
        // Local dev - build from individual variables
        connStr = fmt.Sprintf(
            "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
            os.Getenv("DB_HOST"),
            os.Getenv("DB_PORT"),
            os.Getenv("DB_USER"),
            os.Getenv("DB_PASSWORD"),
            os.Getenv("DB_NAME"),
        )
    }

    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error connecting to database:", err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal("Cannot reach database:", err)
    }

    fmt.Println("✅ Connected to PostgreSQL successfully")
}

// ============================================================
// CONNECT TO REDIS
// ============================================================
func connectRedis() {
    // Use REDIS_URL if available (Railway)
    redisURL := os.Getenv("REDIS_URL")
    
    var opts *redis.Options
    var err error
    
    if redisURL != "" {
        opts, err = redis.ParseURL(redisURL)
        if err != nil {
            log.Fatal("Cannot parse Redis URL:", err)
        }
    } else {
        // Local dev
        opts = &redis.Options{
            Addr: fmt.Sprintf("%s:%s",
                os.Getenv("REDIS_HOST"),
                os.Getenv("REDIS_PORT"),
            ),
        }
    }

    rdb = redis.NewClient(opts)

    _, err = rdb.Ping(ctx).Result()
    if err != nil {
        log.Fatal("Cannot reach Redis:", err)
    }

    fmt.Println("✅ Connected to Redis successfully")
}

// ============================================================
// CREATE TABLE
// ============================================================
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

// ============================================================
// MAIN
// ============================================================
// func main() {
// 	connectDB()
// 	defer db.Close()

// 	connectRedis()
// 	defer rdb.Close()

// 	createTable()

// 	http.HandleFunc("/health", healthCheck)
// 	http.HandleFunc("/todos", todosHandler)
// 	http.HandleFunc("/todos/", todoHandler)

// 	fmt.Println("🚀 Todo API running on http://localhost:9090")
// 	fmt.Println("📋 Endpoints:")
// 	fmt.Println("   GET    /health")
// 	fmt.Println("   GET    /todos")
// 	fmt.Println("   POST   /todos")
// 	fmt.Println("   GET    /todos/{id}")
// 	fmt.Println("   PUT    /todos/{id}")
// 	fmt.Println("   DELETE /todos/{id}")

// 	http.ListenAndServe(":9090", nil)
// }

func main() {
    connectDB()
    defer db.Close()

    connectRedis()
    defer rdb.Close()

    createTable()

    http.HandleFunc("/health", healthCheck)
    http.HandleFunc("/todos", todosHandler)
    http.HandleFunc("/todos/", todoHandler)

    // Read port from environment - Railway injects PORT automatically
    port := os.Getenv("PORT")
    if port == "" {
        port = "9090" // fallback for local dev
    }

    fmt.Println("🚀 Todo API running on http://localhost:" + port)
    http.ListenAndServe(":"+port, nil)
}

// ============================================================
// HELPERS
// ============================================================
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "https://todo-app-sable-zeta-21.vercel.app")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// ============================================================
// ROUTE HANDLERS
// ============================================================
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

// ============================================================
// HEALTH CHECK
// Checks both PostgreSQL and Redis
// ============================================================
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check PostgreSQL
	if err := db.Ping(); err != nil {
		sendJSON(w, http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Message: "PostgreSQL unreachable",
		})
		return
	}

	// Check Redis
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		sendJSON(w, http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Message: "Redis unreachable",
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "API healthy - PostgreSQL and Redis connected",
	})
}

// ============================================================
// GET ALL TODOS
// Flow: Check Redis cache first → if miss, query PostgreSQL
//
//	and store result in Redis for next time
//
// ============================================================
func getTodos(w http.ResponseWriter, r *http.Request) {
	cacheKey := "todos:all"

	// Step 1 - Try to get from Redis cache first
	cached, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache HIT - return cached data directly
		// This is much faster than hitting the database
		fmt.Println("📦 Cache HIT - returning from Redis")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Cache", "HIT") // header so you can see in Postman
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cached))
		return
	}

	// Step 2 - Cache MISS - query PostgreSQL
	fmt.Println("🔍 Cache MISS - querying PostgreSQL")
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
	defer rows.Close()

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

	if todos == nil {
		todos = []Todo{}
	}

	response := APIResponse{
		Success: true,
		Message: "Todos retrieved successfully",
		Data:    todos,
		Count:   len(todos),
	}

	// Step 3 - Store in Redis for 30 seconds
	// Next request within 30s will be served from cache
	responseJSON, _ := json.Marshal(response)
	rdb.Set(ctx, cacheKey, responseJSON, 30*time.Second)
	fmt.Println("💾 Stored in Redis cache for 30 seconds")

	sendJSON(w, http.StatusOK, response)
}

// ============================================================
// INVALIDATE CACHE
// Called after create, update, delete so cache stays fresh
// ============================================================
func invalidateCache() {
	rdb.Del(ctx, "todos:all")
	fmt.Println("🗑️  Cache invalidated")
}

// ============================================================
// GET SINGLE TODO
// ============================================================
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

// ============================================================
// CREATE TODO
// ============================================================
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

	// Clear cache so next GET /todos shows fresh data
	invalidateCache()

	sendJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    todo,
	})
}

// ============================================================
// UPDATE TODO
// ============================================================
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

	// Clear cache
	invalidateCache()

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo updated successfully",
		Data:    todo,
	})
}

// ============================================================
// DELETE TODO
// ============================================================
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Error deleting todo",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	// Clear cache
	invalidateCache()

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}
