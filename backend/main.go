package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Clivet-lug/todo-app/backend/databases"
	"github.com/Clivet-lug/todo-app/backend/handlers"
	"github.com/Clivet-lug/todo-app/backend/middleware"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/repositories"
	"github.com/Clivet-lug/todo-app/backend/utils"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

// Put it in middleware
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

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
			Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		}
	}
	rdb = redis.NewClient(opts)
	if _, err = rdb.Ping(ctx).Result(); err != nil {
		fmt.Println("Redis unavailable, running without cache")
		rdb = nil
		return
	}
	fmt.Println("✅ Connected to Redis successfully")
}

func invalidateCache() {
	if rdb != nil {
		rdb.Del(ctx, "todos:all")
	}
}

func sendCachedOrFetch(w http.ResponseWriter, cacheKey string, fetch func() (interface{}, error)) {
	if rdb != nil {
		if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cached))
			return
		}
	}
	data, err := fetch()
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Error fetching data"})
		return
	}
	response := models.APIResponse{Success: true, Message: "Retrieved successfully", Data: data}
	if rdb != nil {
		if b, err := json.Marshal(response); err == nil {
			rdb.Set(ctx, cacheKey, b, 30*time.Second)
		}
	}
	utils.SendJSON(w, http.StatusOK, response)
}

func extractID(path, prefix string) (int, string, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(trimmed, "/", 2)
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", false
	}
	remainder := ""
	if len(parts) == 2 {
		remainder = parts[1]
	}
	return id, remainder, true
}

func main() {
	db := databases.ConnectDB()
	defer db.Close()

	connectRedis()
	if rdb != nil {
		defer rdb.Close()
	}

	userRepo := repositories.NewUserRepository(db)
	if err := userRepo.CreateUsersTable(); err != nil {
		panic("Error creating users table: " + err.Error())
	}
	connector := repositories.NewConnector(db)
	if err := connector.CreateTable(); err != nil {
		panic("Error creating todos table: " + err.Error())
	}

	http.HandleFunc("/auth/register", withCORS(authMethodGate(http.MethodPost, handlers.Register)))
	http.HandleFunc("/auth/login", withCORS(authMethodGate(http.MethodPost, handlers.Login)))
	http.HandleFunc("/health", withCORS(handlers.HealthCheck))
	http.HandleFunc("/todos", withCORS(middleware.RequireAuth(todosHandler)))
	http.HandleFunc("/todos/", withCORS(middleware.RequireAuth(todoHandler)))
	http.HandleFunc("/users/members", withCORS(middleware.RequireAdmin(handlers.ListMembers)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	fmt.Println("🚀 Todo API running on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}

func authMethodGate(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
			return
		}
		h(w, r)
	}
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		claims := middleware.GetClaims(r)
		cacheKey := fmt.Sprintf("todos:user:%d", claims.UserID)
		if claims.Role == "admin" {
			cacheKey = "todos:all"
		}
		sendCachedOrFetch(w, cacheKey, func() (interface{}, error) {
			db := databases.ConnectDB()
			defer db.Close()
			return repositories.NewConnector(db).GetTodos(claims.UserID, claims.Role == "admin")
		})
	case http.MethodPost:
		claims := middleware.GetClaims(r)
		if claims.Role != "admin" {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can create todos"})
			return
		}
		handlers.CreateTodo(w, r)
		invalidateCache()
	default:
		utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	id, remainder, ok := extractID(r.URL.Path, "/todos/")
	if !ok {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid todo ID"})
		return
	}

	switch remainder {
	case "assign":
		if r.Method != http.MethodPut {
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
			return
		}
		middleware.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
			handlers.AssignTodo(w, r, id)
			invalidateCache()
		})(w, r)
	case "status":
		if r.Method != http.MethodPut {
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
			return
		}
		handlers.UpdateTodoStatus(w, r, id)
		invalidateCache()
	case "":
		switch r.Method {
		case http.MethodGet:
			handlers.GetTodoByID(w, r, id)
		case http.MethodPut:
			claims := middleware.GetClaims(r)
			if claims.Role != "admin" {
				utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can edit todo details"})
				return
			}
			handlers.UpdateTodo(w, r, id)
			invalidateCache()
		case http.MethodDelete:
			claims := middleware.GetClaims(r)
			if claims.Role != "admin" {
				utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can delete todos"})
				return
			}
			handlers.DeleteTodo(w, r, id)
			invalidateCache()
		default:
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
		}
	default:
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Message: "Route not found"})
	}
}
