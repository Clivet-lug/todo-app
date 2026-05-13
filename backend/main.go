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

	"github.com/Clivet-lug/todo-app/backend/databases"
	"github.com/Clivet-lug/todo-app/backend/handlers"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/repositories"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)


var rdb *redis.Client          // Redis client
var ctx = context.Background() // Redis needs a context

// CONNECT TO REDIS
func connectRedis() {
	redisURL := os.Getenv("REDIS_URL")

	var opts *redis.Options
	var err error

	if redisURL != "" {
		opts, err = redis.ParseURL(redisURL)
		if err != nil {
			fmt.Println("Cannot parse Redis URL, running without cache")
			return
		}
	} else {
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
		fmt.Println("Redis unavailable, running without cache")
		rdb = nil
		return
	}

	fmt.Println("✅ Connected to Redis successfully")
}


func main() {
	db:= databases.ConnectDB()
	defer db.Close()

	connectRedis()
	if rdb != nil {
		defer rdb.Close()
	}

	connector := repositories.NewConnector(db)
	if err := connector.CreateTable(); err != nil {
		log.Fatal("Error creating todos table:", err)
	}

	http.HandleFunc("/health", handlers.HealthCheck)
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

// HELPERS
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

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
		sendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{
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
		sendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid todo ID",
		})
		return
	}
	connectDB := databases.ConnectDB()
	defer connectDB.Close()

	
	switch r.Method {
	case http.MethodGet:
		todo, err := repositories.NewConnector(connectDB).GetTodoByID(id)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Error fetching todos",
			})
			return
		}
		sendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Todos retrieved successfully",
			Data:    todo,
		})
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		deleteTodo(w, r, id)
	default:
		sendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
	}
}


// GET ALL TODOS
func getTodos(w http.ResponseWriter, r *http.Request) {
	cacheKey := "todos:all"

	// Only use cache if Redis is available
	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			fmt.Println("📦 Cache HIT - returning from Redis")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cached))
			return
		}
	}

	fmt.Println("📂 Cache MISS - querying database")

	connectDB := databases.ConnectDB()
	defer connectDB.Close()

	todos, err := repositories.NewConnector(connectDB).GetTodos()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error fetching todos",
		})
		return
	}

	if todos == nil {
		todos = []*models.Todo{}
	}

	response := models.APIResponse{
		Success: true,
		Message: "Todos retrieved successfully",
		Data:    todos,
		Count:   len(todos),
	}

	// Only cache if Redis available
	if rdb != nil {
		responseJSON, _ := json.Marshal(response)
		rdb.Set(ctx, cacheKey, responseJSON, 30*time.Second)
	}

	sendJSON(w, http.StatusOK, response)
}

// INVALIDATE CACHE
// Called after create, update, delete so cache stays fresh
func invalidateCache() {
	if rdb != nil {
		rdb.Del(ctx, "todos:all")
		fmt.Println("🗑️ Cache invalidated")
	}
}

// CREATE TODO
func createTodo(w http.ResponseWriter, r *http.Request) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	if input.Priority == "" {
		input.Priority = "medium"
	}

	var todo models.Todo

	connectDB := databases.ConnectDB()
	defer connectDB.Close()

	err := connectDB.QueryRow(`
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
		sendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error creating todo",
		})
		return
	}

	// Clear cache so next GET /todos shows fresh data
	invalidateCache()

	sendJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    todo,
	})
}

// UPDATE TODO
func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	var todo models.Todo
	connectDB := databases.ConnectDB()
	defer connectDB.Close()

	err := connectDB.QueryRow(`
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
		sendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error updating todo",
		})
		return
	}

	// Clear cache
	invalidateCache()

	sendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo updated successfully",
		Data:    todo,
	})
}

// DELETE TODO
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {

	connectDB := databases.ConnectDB()
	defer connectDB.Close()

	result, err := connectDB.Exec(`DELETE FROM todos WHERE id = $1`, id)
	
	// result, err := repositories.NewConnector(connectDB).GetTodos()

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error deleting todo",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Todo with ID %d not found", id),
		})
		return
	}

	// Clear cache
	invalidateCache()

	sendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}
