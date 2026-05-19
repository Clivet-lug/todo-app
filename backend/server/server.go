package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Clivet-lug/todo-app/backend/handlers"
	"github.com/Clivet-lug/todo-app/backend/middleware"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
	"github.com/redis/go-redis/v9"
)

// Server owns the HTTP mux and all handlers.
// main.go builds it once and calls Start().
type Server struct {
	mux         *http.ServeMux
	port        string
	rdb         *redis.Client
	authHandler *handlers.AuthHandler
	todoHandler *handlers.TodoHandler
}

func New(
	port string,
	rdb *redis.Client,
	authHandler *handlers.AuthHandler,
	todoHandler *handlers.TodoHandler,
) *Server {
	s := &Server{
		mux:         http.NewServeMux(),
		port:        port,
		rdb:         rdb,
		authHandler: authHandler,
		todoHandler: todoHandler,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Start() error {
	fmt.Println("🚀 Todo API running on http://localhost:" + s.port)
	return http.ListenAndServe(":"+s.port, s.mux)
}

// ROUTE REGISTRATION

func (s *Server) registerRoutes() {
	// Public
	s.mux.HandleFunc("/health",         middleware.WithCORS(handlers.HealthCheck))
	s.mux.HandleFunc("/auth/register",  middleware.WithCORS(s.methodGate(http.MethodPost, s.authHandler.Register)))
	s.mux.HandleFunc("/auth/login",     middleware.WithCORS(s.methodGate(http.MethodPost, s.authHandler.Login)))

	// Protected
	s.mux.HandleFunc("/todos",          middleware.WithCORS(middleware.RequireAuth(s.todosCollection)))
	s.mux.HandleFunc("/todos/",         middleware.WithCORS(middleware.RequireAuth(s.todosItem)))
	s.mux.HandleFunc("/users/members",  middleware.WithCORS(middleware.RequireAdmin(s.todoHandler.ListMembers)))
}

// CACHE HELPERS   

func (s *Server) cacheGet(key string) (string, bool) {
	if s.rdb == nil {
		return "", false
	}
	val, err := s.rdb.Get(context.Background(), key).Result()
	return val, err == nil
}

func (s *Server) cacheSet(key string, data interface{}) {
	if s.rdb == nil {
		return
	}
	b, err := json.Marshal(data)
	if err == nil {
		s.rdb.Set(context.Background(), key, b, 30*time.Second)
	}
}

func (s *Server) cacheInvalidate() {
	if s.rdb != nil {
		s.rdb.Del(context.Background(), "todos:all")
		fmt.Println("🗑️  Cache invalidated")
	}
}

func (s *Server) serveCached(w http.ResponseWriter, cacheKey string, fresh func() (interface{}, error)) {
	if cached, ok := s.cacheGet(cacheKey); ok {
		fmt.Println("Cache HIT")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cached))
		return
	}
	fmt.Println("Cache MISS")
	data, err := fresh()
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Error fetching data"})
		return
	}
	resp := models.APIResponse{Success: true, Message: "Retrieved successfully", Data: data}
	s.cacheSet(cacheKey, resp)
	utils.SendJSON(w, http.StatusOK, resp)
}

// COLLECTION HANDLER  /todos 

func (s *Server) todosCollection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	isAdmin := claims.Role == "admin"

	switch r.Method {
	case http.MethodGet:
		cacheKey := fmt.Sprintf("todos:user:%d", claims.UserID)
		if isAdmin {
			cacheKey = "todos:all"
		}
		s.serveCached(w, cacheKey, func() (interface{}, error) {
			return s.todoHandler.GetTodosData(claims.UserID, isAdmin)
		})

	case http.MethodPost:
		if !isAdmin {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can create todos"})
			return
		}
		s.todoHandler.CreateTodo(w, r)
		s.cacheInvalidate()

	default:
		utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
	}
}

// ITEM HANDLER  /todos/{id}[/sub-route] 

func (s *Server) todosItem(w http.ResponseWriter, r *http.Request) {
	id, remainder, ok := extractID(r.URL.Path, "/todos/")
	if !ok {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid todo ID"})
		return
	}

	switch remainder {
	case "assign":
		middleware.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
			s.todoHandler.AssignTodo(w, r, id)
			s.cacheInvalidate()
		})(w, r)

	case "status":
		if r.Method != http.MethodPut {
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
			return
		}
		s.todoHandler.UpdateStatus(w, r, id)
		s.cacheInvalidate()

	case "":
		claims := middleware.GetClaims(r)
		switch r.Method {
		case http.MethodGet:
			s.todoHandler.GetTodoByID(w, r, id)
		case http.MethodPut:
			if claims.Role != "admin" {
				utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can edit todos"})
				return
			}
			s.todoHandler.UpdateTodo(w, r, id)
			s.cacheInvalidate()
		case http.MethodDelete:
			if claims.Role != "admin" {
				utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Only admins can delete todos"})
				return
			}
			s.todoHandler.DeleteTodo(w, r, id)
			s.cacheInvalidate()
		default:
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
		}

	default:
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Message: "Route not found"})
	}
}

// HELPERS

func (s *Server) methodGate(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			utils.SendJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
			return
		}
		h(w, r)
	}
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